package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/port"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/outbox"
)

type PaymentUseCase struct {
	repo     port.PaymentRepository
	provider contracts.PaymentProvider
	dbPool   *db.Pool
}

func NewPaymentUseCase(
	repo port.PaymentRepository,
	provider contracts.PaymentProvider,
	dbPool *db.Pool,
) *PaymentUseCase {
	return &PaymentUseCase{
		repo:     repo,
		provider: provider,
		dbPool:   dbPool,
	}
}

type PaymentEventPayload struct {
	PaymentID         uuid.UUID `json:"payment_id"`
	OrderID           uuid.UUID `json:"order_id"`
	Provider          string    `json:"provider"`
	ProviderPaymentID string    `json:"provider_payment_id"`
	AmountRUB         string    `json:"amount_rub"`
	Status            string    `json:"status"`
	Reason            string    `json:"reason,omitempty"`
}

// InitiatePayment запускает создание платежа у выбранного провайдера.
func (uc *PaymentUseCase) InitiatePayment(
	ctx context.Context,
	orderID uuid.UUID,
	userID uuid.UUID,
	returnURL string,
) (*domain.Payment, error) {
	// Проверяем, существует ли уже платёж для этого заказа
	existing, err := uc.repo.GetByOrderID(ctx, orderID)
	if err == nil && existing != nil {
		return existing, nil
	}

	// 1. Инициируем платёж в провайдере
	res, err := uc.provider.InitiatePayment(ctx, contracts.InitiatePaymentRequest{
		OrderID:     orderID,
		UserID:      userID,
		AmountRUB:   domain.FixedPaymentAmountRUB, // INV-01
		Description: fmt.Sprintf("Оплата заказа %s", orderID),
		ReturnURL:   returnURL,
	})
	if err != nil {
		return nil, fmt.Errorf("initiating payment in %s: %w", uc.provider.Name(), err)
	}

	// 2. Создаём доменную сущность
	payment, err := domain.NewPayment(
		uuid.New(),
		orderID,
		uc.provider.Name(),
		res.ProviderPaymentID,
		res.ConfirmationURL,
	)
	if err != nil {
		return nil, fmt.Errorf("creating payment aggregate: %w", err)
	}

	// 3. Сохраняем в транзакции с outbox-событием
	if uc.dbPool != nil {
		err = uc.dbPool.WithTx(ctx, func(txCtx context.Context) error {
			if err := uc.repo.Save(txCtx, payment); err != nil {
				return fmt.Errorf("saving payment: %w", err)
			}

			outboxEvent, err := outbox.NewEvent(
				"dopamine.events",
				orderID.String(),
				"payment.initiated.v1",
				PaymentEventPayload{
					PaymentID:         payment.ID(),
					OrderID:           payment.OrderID(),
					Provider:          payment.Provider(),
					ProviderPaymentID: payment.ProviderPaymentID(),
					AmountRUB:         payment.AmountRUB(),
					Status:            string(payment.Status()),
				},
			)
			if err != nil {
				return fmt.Errorf("creating outbox event: %w", err)
			}

			return outbox.SaveInTx(txCtx, uc.dbPool, "payment", outboxEvent)
		})
		if err != nil {
			return nil, fmt.Errorf("persisting payment transaction: %w", err)
		}
	} else {
		// Mock/test fallback
		if err := uc.repo.Save(ctx, payment); err != nil {
			return nil, err
		}
	}

	return payment, nil
}

// HandleWebhook обрабатывает входящий вебхук от платежной системы с гарантией идемпотентности.
func (uc *PaymentUseCase) HandleWebhook(ctx context.Context, rawBody []byte, headers map[string]string) error {
	res, err := uc.provider.VerifyWebhook(ctx, rawBody, headers)
	if err != nil {
		return fmt.Errorf("verifying webhook: %w", err)
	}

	if res.Status == contracts.PaymentStatusPending {
		return nil
	}

	// Идемпотентность через processed_events
	eventKey := fmt.Sprintf("webhook:%s:%s", res.ProviderPaymentID, res.Status)

	handler := func(ctx context.Context) error {
		payment, err := uc.repo.GetByOrderID(ctx, res.OrderID)
		if err != nil {
			// fallback by provider_payment_id
			payment, err = uc.repo.GetByProviderPaymentID(ctx, res.ProviderPaymentID)
			if err != nil {
				return fmt.Errorf("payment not found for order %s: %w", res.OrderID, err)
			}
		}

		var eventType string
		switch res.Status {
		case contracts.PaymentStatusSucceeded:
			if err := payment.MarkSucceeded(); err != nil {
				return fmt.Errorf("marking payment succeeded: %w", err)
			}
			eventType = "payment.succeeded.v1"
		case contracts.PaymentStatusFailed:
			if err := payment.MarkFailed(res.FailureReason); err != nil {
				return fmt.Errorf("marking payment failed: %w", err)
			}
			eventType = "payment.failed.v1"
		default:
			return nil
		}

		if uc.dbPool != nil {
			return uc.dbPool.WithTx(ctx, func(txCtx context.Context) error {
				if err := uc.repo.Update(txCtx, payment); err != nil {
					return fmt.Errorf("updating payment: %w", err)
				}

				outboxEvent, err := outbox.NewEvent(
					"dopamine.events",
					payment.OrderID().String(),
					eventType,
					PaymentEventPayload{
						PaymentID:         payment.ID(),
						OrderID:           payment.OrderID(),
						Provider:          payment.Provider(),
						ProviderPaymentID: payment.ProviderPaymentID(),
						AmountRUB:         payment.AmountRUB(),
						Status:            string(payment.Status()),
						Reason:            payment.FailureReason(),
					},
				)
				if err != nil {
					return fmt.Errorf("creating outbox event: %w", err)
				}

				return outbox.SaveInTx(txCtx, uc.dbPool, "payment", outboxEvent)
			})
		}

		return uc.repo.Update(ctx, payment)
	}

	if uc.dbPool != nil {
		return uc.processIdempotent(ctx, eventKey, "payment-webhook", handler)
	}

	return handler(ctx)
}

func (uc *PaymentUseCase) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}

func (uc *PaymentUseCase) processIdempotent(ctx context.Context, eventID, consumerGroup string, handler func(ctx context.Context) error) error {
	var processedAt time.Time
	checkQuery := `SELECT processed_at FROM payment.processed_events WHERE event_id = $1 AND consumer_group = $2`
	err := uc.dbPool.Raw().QueryRow(ctx, checkQuery, eventID, consumerGroup).Scan(&processedAt)
	if err == nil {
		// Уже обработано — пропускаем
		return nil
	}

	if err := handler(ctx); err != nil {
		return err
	}

	insertQuery := `INSERT INTO payment.processed_events (event_id, consumer_group, processed_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING`
	_, _ = uc.dbPool.Raw().Exec(ctx, insertQuery, eventID, consumerGroup)
	return nil
}

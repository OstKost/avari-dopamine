package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/contracts"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/port"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/outbox"
)

type OrderUseCase struct {
	repo          port.OrderRepository
	cartLookup    contracts.CartLookup
	productLookup contracts.ProductLookup
	pickupLookup  contracts.PickupPointLookup
	dbPool        *db.Pool
}

func NewOrderUseCase(
	repo port.OrderRepository,
	cartLookup contracts.CartLookup,
	productLookup contracts.ProductLookup,
	pickupLookup contracts.PickupPointLookup,
	dbPool *db.Pool,
) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		cartLookup:    cartLookup,
		productLookup: productLookup,
		pickupLookup:  pickupLookup,
		dbPool:        dbPool,
	}
}

type CreateOrderResult struct {
	Order *domain.Order
}

// CreateOrder оформляет заказ из текущей корзины пользователя (FR-ORDER-01).
func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID uuid.UUID) (*domain.Order, error) {
	// 1. Получаем снимок корзины
	cart, err := uc.cartLookup.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching cart: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, domain.ErrCartEmpty
	}

	if cart.PickupPointID == nil {
		return nil, domain.ErrPickupPointRequired
	}

	// 2. Проверяем INV-02: пользователь не может иметь два активных payment_pending заказа
	hasPending, err := uc.repo.HasActivePendingOrder(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("checking pending orders: %w", err)
	}
	if hasPending {
		return nil, domain.ErrDuplicatePendingOrder
	}

	// 3. Загружаем данные ПВЗ
	pickupSnapshot, err := uc.pickupLookup.GetByID(ctx, *cart.PickupPointID)
	if err != nil {
		return nil, fmt.Errorf("getting pickup point details: %w", err)
	}

	pickupInfo := domain.PickupPointInfo{
		ID:             pickupSnapshot.ID,
		Name:           pickupSnapshot.Name,
		Latitude:       pickupSnapshot.Latitude,
		Longitude:      pickupSnapshot.Longitude,
		DistanceMeters: pickupSnapshot.DistanceMeters,
	}

	// 4. Загружаем актуальные данные товаров для создания снапшота
	prodIDs := make([]uuid.UUID, len(cart.Items))
	for i, item := range cart.Items {
		prodIDs[i] = item.ProductID
	}

	products, err := uc.productLookup.GetByIDs(ctx, prodIDs)
	if err != nil {
		return nil, fmt.Errorf("getting product snapshots: %w", err)
	}

	orderItems := make([]domain.OrderItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		prod, ok := products[item.ProductID]
		if !ok {
			return nil, fmt.Errorf("product %s not found in catalog", item.ProductID)
		}

		orderItems = append(orderItems, domain.NewOrderItem(
			uuid.New(),
			item.ProductID,
			prod.Name,
			prod.CategoryName,
			prod.PriceRUB,
			prod.ImageSeed,
			item.Quantity,
		))
	}

	// 5. Создаём агрегат Order (с фиксированной суммой 10.00 RUB по INV-01)
	orderID := uuid.New()
	order, err := domain.NewOrder(orderID, userID, pickupInfo, orderItems, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	// Автоматический перевод в payment_pending при оформлении
	_ = order.TransitionTo(domain.StatusPaymentPending)

	// 6. Атомарное сохранение в БД + Outbox событие в одной транзакции (ADR-003)
	outboxEvent, err := outbox.NewEvent("order-events", order.ID().String(), "order.created.v1", map[string]interface{}{
		"order_id":         order.ID().String(),
		"user_id":          order.UserID().String(),
		"status":           order.Status().String(),
		"total_amount_rub": order.TotalAmountRUB().StringFixed(2),
		"created_at":       order.CreatedAt().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("creating outbox event: %w", err)
	}

	if uc.dbPool != nil {
		err = uc.dbPool.WithTx(ctx, func(txCtx context.Context) error {
			if err := uc.repo.Create(txCtx, order); err != nil {
				return err
			}
			if err := outbox.SaveInTx(txCtx, uc.dbPool, "order", outboxEvent); err != nil {
				return err
			}
			return nil
		})
	} else {
		err = uc.repo.Create(ctx, order)
	}
	if err != nil {
		return nil, fmt.Errorf("persisting order in transaction: %w", err)
	}

	// 7. Очищаем корзину после успешного коммита
	_ = uc.cartLookup.ClearCart(ctx, userID)

	return order, nil
}

// TransitionOrderStatus выполняет переход статуса заказа и публикует событие в outbox (ADR-003).
func (uc *OrderUseCase) TransitionOrderStatus(ctx context.Context, orderID uuid.UUID, targetStatus domain.Status, reason string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	fromStatus := order.Status()
	if err := order.TransitionTo(targetStatus); err != nil {
		return nil, err
	}

	outboxEvent, err := outbox.NewEvent("order-events", order.ID().String(), "order.status_changed.v1", map[string]interface{}{
		"order_id":    order.ID().String(),
		"user_id":     order.UserID().String(),
		"from_status": fromStatus.String(),
		"to_status":   targetStatus.String(),
		"reason":      reason,
		"updated_at":  order.UpdatedAt().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("creating outbox event: %w", err)
	}

	if uc.dbPool != nil {
		err = uc.dbPool.WithTx(ctx, func(txCtx context.Context) error {
			if err := uc.repo.UpdateStatus(txCtx, order.ID(), fromStatus, targetStatus, reason); err != nil {
				return err
			}
			if err := outbox.SaveInTx(txCtx, uc.dbPool, "order", outboxEvent); err != nil {
				return err
			}
			return nil
		})
	} else {
		err = uc.repo.UpdateStatus(ctx, order.ID(), fromStatus, targetStatus, reason)
	}
	if err != nil {
		return nil, fmt.Errorf("persisting status transition: %w", err)
	}

	return order, nil
}

func (uc *OrderUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUseCase) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Order, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return uc.repo.ListByUserID(ctx, userID, limit, offset)
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, userID, orderID uuid.UUID, reason string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID() != userID {
		return nil, domain.ErrOrderNotFound
	}
	if reason == "" {
		reason = "cancelled by user"
	}
	return uc.TransitionOrderStatus(ctx, orderID, domain.StatusCancelled, reason)
}

func (uc *OrderUseCase) GetStatusHistory(ctx context.Context, orderID uuid.UUID) ([]port.StatusHistoryRecord, error) {
	return uc.repo.GetStatusHistory(ctx, orderID)
}

// ProcessExternalEvent идемпотентно обрабатывает внешние события (payment, delivery) через таблицу processed_events (ADR-003).
func (uc *OrderUseCase) ProcessExternalEvent(ctx context.Context, eventID, consumerName string, fn func(ctx context.Context) error) error {
	processed, err := uc.repo.IsEventProcessed(ctx, eventID, consumerName)
	if err != nil {
		return fmt.Errorf("checking event idempotency: %w", err)
	}
	if processed {
		return nil // событие уже обработано
	}

	if err := fn(ctx); err != nil {
		return err
	}

	return uc.repo.MarkEventProcessed(ctx, eventID, consumerName)
}

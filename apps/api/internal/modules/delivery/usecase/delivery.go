package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/port"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/outbox"
	"github.com/ostkost/dopamine-market/api/internal/platform/random"
	"github.com/ostkost/dopamine-market/api/internal/platform/scheduler"
)

var defaultCourierNames = []string{
	"Алексей Смирнов",
	"Дмитрий Иванов",
	"Михаил Кузнецов",
	"Иван Попов",
	"Сергей Васильев",
	"Анна Петрова",
	"Екатерина Соколова",
	"Артём Морозов",
	"Максим Новиков",
	"Никита Федоров",
}

// Config задаёт параметры временных интервалов симуляции доставки (ADR-005, ADR-012).
type Config struct {
	AssemblingMinDuration      time.Duration
	AssemblingMaxDuration      time.Duration
	CourierAssignedMinDuration time.Duration
	CourierAssignedMaxDuration time.Duration
	InTransitMinDuration       time.Duration
	InTransitMaxDuration       time.Duration
	DelayedMinDuration         time.Duration
	DelayedMaxDuration         time.Duration
	DelayProbability           float64
}

// DefaultConfig возвращает стандартные интервалы согласно PRD / FR-DELIVERY-01.
func DefaultConfig() Config {
	return Config{
		AssemblingMinDuration:      10 * time.Second,
		AssemblingMaxDuration:      30 * time.Second,
		CourierAssignedMinDuration: 15 * time.Second,
		CourierAssignedMaxDuration: 45 * time.Second,
		InTransitMinDuration:       60 * time.Second,
		InTransitMaxDuration:       180 * time.Second,
		DelayedMinDuration:         30 * time.Second,
		DelayedMaxDuration:         60 * time.Second,
		DelayProbability:           0.10,
	}
}

// FastTestConfig возвращает ускоренные интервалы для unit/интеграционных тестов.
func FastTestConfig() Config {
	return Config{
		AssemblingMinDuration:      10 * time.Millisecond,
		AssemblingMaxDuration:      30 * time.Millisecond,
		CourierAssignedMinDuration: 15 * time.Millisecond,
		CourierAssignedMaxDuration: 45 * time.Millisecond,
		InTransitMinDuration:       60 * time.Millisecond,
		InTransitMaxDuration:       180 * time.Millisecond,
		DelayedMinDuration:         30 * time.Millisecond,
		DelayedMaxDuration:         60 * time.Millisecond,
		DelayProbability:           0.10,
	}
}

type DeliveryUseCase struct {
	repo port.DeliveryRepository
	pool *db.Pool
	rnd  random.Source
	cfg  Config
}

func NewDeliveryUseCase(
	repo port.DeliveryRepository,
	pool *db.Pool,
	rnd random.Source,
	cfg Config,
) *DeliveryUseCase {
	return &DeliveryUseCase{
		repo: repo,
		pool: pool,
		rnd:  rnd,
		cfg:  cfg,
	}
}

// DeliveryStatusPayload — полезная нагрузка для Kafka-события delivery.status_changed.v1.
type DeliveryStatusPayload struct {
	DeliveryID            uuid.UUID `json:"delivery_id"`
	OrderID               uuid.UUID `json:"order_id"`
	Status                string    `json:"status"`
	CourierName           string    `json:"courier_name"`
	CourierRating         float64   `json:"courier_rating"`
	EstimatedCompletionAt time.Time `json:"estimated_completion_at"`
	Reason                string    `json:"reason,omitempty"`
}

// StartDelivery создаёт доставку и планирует первый переход (FR-DELIVERY-01).
func (uc *DeliveryUseCase) StartDelivery(ctx context.Context, orderID uuid.UUID) (*domain.Delivery, error) {
	// Проверка на повторный запуск (идемпотентность)
	existing, err := uc.repo.FindByOrderID(ctx, orderID)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, domain.ErrDeliveryNotFound) {
		return nil, fmt.Errorf("checking existing delivery: %w", err)
	}

	courierName := uc.rnd.Pick(defaultCourierNames)
	rawRating := uc.rnd.Float64Range(4.2, 5.0)
	courierRating := math.Round(rawRating*100) / 100

	assemblingSec := uc.durationRange(uc.cfg.AssemblingMinDuration, uc.cfg.AssemblingMaxDuration)
	totalEstimate := assemblingSec +
		uc.durationRange(uc.cfg.CourierAssignedMinDuration, uc.cfg.CourierAssignedMaxDuration) +
		uc.durationRange(uc.cfg.InTransitMinDuration, uc.cfg.InTransitMaxDuration)

	del := domain.NewDelivery(
		uuid.New(),
		orderID,
		courierName,
		courierRating,
		time.Now().UTC(),
		totalEstimate,
	)

	err = uc.pool.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.repo.Save(txCtx, del); err != nil {
			return fmt.Errorf("saving delivery: %w", err)
		}

		fireAt := time.Now().UTC().Add(assemblingSec)
		if err := scheduler.ScheduleInTx(txCtx, uc.pool, "delivery", del.ID, string(domain.StatusCourierAssigned), fireAt, nil); err != nil {
			return fmt.Errorf("scheduling initial transition: %w", err)
		}

		eventPayload := DeliveryStatusPayload{
			DeliveryID:            del.ID,
			OrderID:               del.OrderID,
			Status:                string(del.Status),
			CourierName:           del.CourierName,
			CourierRating:         del.CourierRating,
			EstimatedCompletionAt: del.EstimatedCompletionAt,
		}

		evt, err := outbox.NewEvent("dopamine.events", del.OrderID.String(), "delivery.status_changed.v1", eventPayload)
		if err != nil {
			return fmt.Errorf("creating outbox event: %w", err)
		}

		return outbox.SaveInTx(txCtx, uc.pool, "delivery", evt)
	})

	if err != nil {
		return nil, err
	}

	return del, nil
}

// AdvanceDeliveryState выполняет переход состояния по таймеру планировщика (ADR-005).
func (uc *DeliveryUseCase) AdvanceDeliveryState(ctx context.Context, deliveryID uuid.UUID, targetStatus domain.Status) (*domain.Delivery, error) {
	var del *domain.Delivery

	err := uc.pool.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		del, err = uc.repo.FindByID(txCtx, deliveryID)
		if err != nil {
			return fmt.Errorf("finding delivery %s: %w", deliveryID, err)
		}

		if del.Status == targetStatus {
			return nil
		}

		if err := del.TransitionTo(targetStatus); err != nil {
			return fmt.Errorf("transitioning delivery state: %w", err)
		}

		if err := uc.repo.Update(txCtx, del); err != nil {
			return fmt.Errorf("updating delivery: %w", err)
		}

		// Планирование следующего шага
		switch del.Status {
		case domain.StatusCourierAssigned:
			delay := uc.durationRange(uc.cfg.CourierAssignedMinDuration, uc.cfg.CourierAssignedMaxDuration)
			fireAt := time.Now().UTC().Add(delay)
			if err := scheduler.ScheduleInTx(txCtx, uc.pool, "delivery", del.ID, string(domain.StatusInTransit), fireAt, nil); err != nil {
				return fmt.Errorf("scheduling in_transit transition: %w", err)
			}

		case domain.StatusInTransit:
			isDelayed := uc.rnd.Bool(uc.cfg.DelayProbability)
			if isDelayed {
				delay := uc.durationRange(uc.cfg.DelayedMinDuration, uc.cfg.DelayedMaxDuration)
				fireAt := time.Now().UTC().Add(delay)
				if err := scheduler.ScheduleInTx(txCtx, uc.pool, "delivery", del.ID, string(domain.StatusDeliveryDelayed), fireAt, nil); err != nil {
					return fmt.Errorf("scheduling delivery_delayed transition: %w", err)
				}
			} else {
				delay := uc.durationRange(uc.cfg.InTransitMinDuration, uc.cfg.InTransitMaxDuration)
				fireAt := time.Now().UTC().Add(delay)
				if err := scheduler.ScheduleInTx(txCtx, uc.pool, "delivery", del.ID, string(domain.StatusDelivered), fireAt, nil); err != nil {
					return fmt.Errorf("scheduling delivered transition: %w", err)
				}
			}

		case domain.StatusDeliveryDelayed:
			delay := uc.durationRange(uc.cfg.DelayedMinDuration, uc.cfg.DelayedMaxDuration)
			fireAt := time.Now().UTC().Add(delay)
			if err := scheduler.ScheduleInTx(txCtx, uc.pool, "delivery", del.ID, string(domain.StatusDelivered), fireAt, nil); err != nil {
				return fmt.Errorf("scheduling delivered after delayed transition: %w", err)
			}

		case domain.StatusDelivered:
			// Финальное событие завершения доставки
			completedEvt, err := outbox.NewEvent("dopamine.events", del.OrderID.String(), "delivery.completed.v1", DeliveryStatusPayload{
				DeliveryID:            del.ID,
				OrderID:               del.OrderID,
				Status:                string(del.Status),
				CourierName:           del.CourierName,
				CourierRating:         del.CourierRating,
				EstimatedCompletionAt: del.EstimatedCompletionAt,
			})
			if err != nil {
				return fmt.Errorf("creating delivery.completed outbox event: %w", err)
			}
			if err := outbox.SaveInTx(txCtx, uc.pool, "delivery", completedEvt); err != nil {
				return fmt.Errorf("saving delivery.completed outbox event: %w", err)
			}
		}

		// Публикуем событие изменения статуса для всех переходов
		statusEvt, err := outbox.NewEvent("dopamine.events", del.OrderID.String(), "delivery.status_changed.v1", DeliveryStatusPayload{
			DeliveryID:            del.ID,
			OrderID:               del.OrderID,
			Status:                string(del.Status),
			CourierName:           del.CourierName,
			CourierRating:         del.CourierRating,
			EstimatedCompletionAt: del.EstimatedCompletionAt,
		})
		if err != nil {
			return fmt.Errorf("creating delivery.status_changed outbox event: %w", err)
		}

		return outbox.SaveInTx(txCtx, uc.pool, "delivery", statusEvt)
	})

	if err != nil {
		return nil, err
	}

	return del, nil
}

// GetDeliveryByOrderID возвращает информацию о доставке по ID заказа.
func (uc *DeliveryUseCase) GetDeliveryByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Delivery, error) {
	return uc.repo.FindByOrderID(ctx, orderID)
}

// ProcessExternalEvent гарантирует идемпотентность обработки Kafka-событий (ADR-003).
func (uc *DeliveryUseCase) ProcessExternalEvent(ctx context.Context, eventID, consumerGroup string, handler func(ctx context.Context) error) error {
	conn := uc.pool.Conn(ctx)

	// Проверяем, обрабатывалось ли событие ранее
	var exists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM delivery.processed_events
			WHERE event_id = $1 AND consumer_group = $2
		)
	`
	if err := conn.QueryRow(ctx, checkQuery, eventID, consumerGroup).Scan(&exists); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("checking processed event: %w", err)
	}
	if exists {
		return nil
	}

	return uc.pool.WithTx(ctx, func(txCtx context.Context) error {
		if err := handler(txCtx); err != nil {
			return err
		}

		insertQuery := `
			INSERT INTO delivery.processed_events (event_id, consumer_group, processed_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (event_id, consumer_group) DO NOTHING
		`
		_, err := uc.pool.Conn(txCtx).Exec(txCtx, insertQuery, eventID, consumerGroup, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("marking event as processed: %w", err)
		}
		return nil
	})
}

func (uc *DeliveryUseCase) durationRange(minD, maxD time.Duration) time.Duration {
	if minD >= maxD {
		return minD
	}
	minSec := int(minD.Seconds())
	maxSec := int(maxD.Seconds())
	if minSec > 0 && maxSec > 0 {
		return time.Duration(uc.rnd.IntRange(minSec, maxSec)) * time.Second
	}
	minMs := int(minD.Milliseconds())
	maxMs := int(maxD.Milliseconds())
	return time.Duration(uc.rnd.IntRange(minMs, maxMs)) * time.Millisecond
}

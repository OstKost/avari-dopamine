package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
)

// Transition представляет запланированный переход состояния в БД (ADR-005).
type Transition struct {
	ID           uuid.UUID       `json:"id"`
	DeliveryID   uuid.UUID       `json:"delivery_id"`
	TargetStatus string          `json:"target_status"`
	FireAt       time.Time       `json:"fire_at"`
	Payload      json.RawMessage `json:"payload"`
	CreatedAt    time.Time       `json:"created_at"`
	ProcessedAt  *time.Time      `json:"processed_at,omitempty"`
}

// Handler — сигнатура функции-обработчика сработавшего перехода.
type Handler func(ctx context.Context, item Transition) error

// Scheduler — примитив таймер-стейт-машины, устойчивый к рестартам сервиса (ADR-005).
type Scheduler struct {
	pool     *db.Pool
	logger   *slog.Logger
	interval time.Duration
}

// New создаёт новый экземпляр планировщика.
func New(pool *db.Pool, logger *slog.Logger, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 1 * time.Second
	}
	return &Scheduler{
		pool:     pool,
		logger:   logger,
		interval: interval,
	}
}

// ScheduleInTx регистрирует запланированный переход в схеме указанного модуля в рамках активной транзакции.
func ScheduleInTx(ctx context.Context, pool *db.Pool, schema string, deliveryID uuid.UUID, targetStatus string, fireAt time.Time, payload any) error {
	conn := pool.Conn(ctx)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling scheduler transition payload: %w", err)
	}

	id := uuid.New()
	query := fmt.Sprintf(`
		INSERT INTO "%s".scheduled_transitions (id, delivery_id, target_status, fire_at, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, schema)

	_, err = conn.Exec(ctx, query, id, deliveryID, targetStatus, fireAt.UTC(), payloadBytes, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("inserting scheduled transition in schema %q: %w", schema, err)
	}
	return nil
}

// Run запускает бесконечный цикл опроса созревших переходов для указанной схемы БД.
func (s *Scheduler) Run(ctx context.Context, schema string, handler Handler) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.processPending(ctx, schema, handler); err != nil && !isCancellation(err) {
				s.logger.Error("scheduler processing error", slog.String("schema", schema), slog.String("error", err.Error()))
			}
		}
	}
}

func (s *Scheduler) processPending(ctx context.Context, schema string, handler Handler) error {
	for {
		var item Transition
		var found bool

		// Выбираем один созревший переход с блокировкой SKIP LOCKED в короткой транзакции
		err := s.pool.WithTx(ctx, func(txCtx context.Context) error {
			query := fmt.Sprintf(`
				SELECT id, delivery_id, target_status, fire_at, payload, created_at
				FROM "%s".scheduled_transitions
				WHERE fire_at <= NOW() AND processed_at IS NULL
				ORDER BY fire_at ASC
				LIMIT 1
				FOR UPDATE SKIP LOCKED
			`, schema)

			conn := s.pool.Conn(txCtx)
			row := conn.QueryRow(txCtx, query)
			if err := row.Scan(&item.ID, &item.DeliveryID, &item.TargetStatus, &item.FireAt, &item.Payload, &item.CreatedAt); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					found = false
					return nil
				}
				return fmt.Errorf("querying pending scheduled transition: %w", err)
			}
			found = true

			// Помечаем сразу как обработанный в рамках этой же транзакции, если обработчик выполнится успешно
			markQuery := fmt.Sprintf(`
				UPDATE "%s".scheduled_transitions
				SET processed_at = NOW()
				WHERE id = $1
			`, schema)
			_, err := conn.Exec(txCtx, markQuery, item.ID)
			return err
		})

		if err != nil {
			return err
		}

		if !found {
			return nil
		}

		// Вызываем бизнес-обработчик перехода
		if err := handler(ctx, item); err != nil {
			s.logger.Error("failed to process scheduled transition item",
				slog.String("schema", schema),
				slog.String("transition_id", item.ID.String()),
				slog.String("delivery_id", item.DeliveryID.String()),
				slog.String("target_status", item.TargetStatus),
				slog.String("error", err.Error()),
			)
			// При ошибке обработчика продолжаем цикл, чтобы не заблокировать остальные задачи
		}
	}
}

func isCancellation(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

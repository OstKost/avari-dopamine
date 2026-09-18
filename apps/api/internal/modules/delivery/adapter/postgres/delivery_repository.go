package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ostkost/dopamine-market/api/internal/modules/delivery/domain"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
)

type DeliveryRepository struct {
	pool *db.Pool
}

func NewDeliveryRepository(pool *db.Pool) *DeliveryRepository {
	return &DeliveryRepository{pool: pool}
}

func (r *DeliveryRepository) Save(ctx context.Context, d *domain.Delivery) error {
	conn := r.pool.Conn(ctx)

	query := `
		INSERT INTO delivery.deliveries (
			id, order_id, status, courier_name, courier_rating, started_at, estimated_completion_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := conn.Exec(ctx, query,
		d.ID,
		d.OrderID,
		string(d.Status),
		d.CourierName,
		d.CourierRating,
		d.StartedAt,
		d.EstimatedCompletionAt,
		d.CreatedAt,
		d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting delivery: %w", err)
	}
	return nil
}

func (r *DeliveryRepository) Update(ctx context.Context, d *domain.Delivery) error {
	conn := r.pool.Conn(ctx)

	query := `
		UPDATE delivery.deliveries
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := conn.Exec(ctx, query,
		string(d.Status),
		d.UpdatedAt,
		d.ID,
	)
	if err != nil {
		return fmt.Errorf("updating delivery: %w", err)
	}
	return nil
}

func (r *DeliveryRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT id, order_id, status, courier_name, courier_rating, started_at, estimated_completion_at, created_at, updated_at
		FROM delivery.deliveries
		WHERE id = $1
	`

	return r.scanDelivery(conn.QueryRow(ctx, query, id))
}

func (r *DeliveryRepository) FindByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Delivery, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT id, order_id, status, courier_name, courier_rating, started_at, estimated_completion_at, created_at, updated_at
		FROM delivery.deliveries
		WHERE order_id = $1
	`

	return r.scanDelivery(conn.QueryRow(ctx, query, orderID))
}

func (r *DeliveryRepository) scanDelivery(row pgx.Row) (*domain.Delivery, error) {
	var (
		id                    uuid.UUID
		orderID               uuid.UUID
		statusStr             string
		courierName           string
		courierRating         float64
		startedAt             time.Time
		estimatedCompletionAt time.Time
		createdAt             time.Time
		updatedAt             time.Time
	)

	err := row.Scan(
		&id,
		&orderID,
		&statusStr,
		&courierName,
		&courierRating,
		&startedAt,
		&estimatedCompletionAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("scanning delivery: %w", err)
	}

	return &domain.Delivery{
		ID:                    id,
		OrderID:               orderID,
		Status:                domain.Status(statusStr),
		CourierName:           courierName,
		CourierRating:         courierRating,
		StartedAt:             startedAt,
		EstimatedCompletionAt: estimatedCompletionAt,
		CreatedAt:             createdAt,
		UpdatedAt:             updatedAt,
	}, nil
}

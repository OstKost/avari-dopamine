package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/order/port"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/shopspring/decimal"
)

type OrderRepository struct {
	pool *db.Pool
}

func NewOrderRepository(pool *db.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	conn := r.pool.Conn(ctx)

	// 1. Insert order
	orderQuery := `
		INSERT INTO "order".orders (
			id, user_id, status, total_amount_rub,
			pickup_point_id, pickup_point_name, pickup_point_latitude, pickup_point_longitude, pickup_point_distance_meters,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := conn.Exec(ctx, orderQuery,
		o.ID(),
		o.UserID(),
		o.Status().String(),
		o.TotalAmountRUB(),
		o.PickupPoint().ID,
		o.PickupPoint().Name,
		o.PickupPoint().Latitude,
		o.PickupPoint().Longitude,
		o.PickupPoint().DistanceMeters,
		o.CreatedAt(),
		o.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("inserting order: %w", err)
	}

	// 2. Insert items
	itemQuery := `
		INSERT INTO "order".order_items (
			id, order_id, product_id, name, category_name, price_rub, image_seed, quantity, subtotal_rub, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	for _, it := range o.Items() {
		_, err := conn.Exec(ctx, itemQuery,
			it.ID(),
			o.ID(),
			it.ProductID(),
			it.Name(),
			it.CategoryName(),
			it.PriceRUB(),
			it.ImageSeed(),
			it.Quantity(),
			it.SubtotalRUB(),
			o.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("inserting order item %s: %w", it.ID(), err)
		}
	}

	// 3. Insert initial status history
	historyQuery := `
		INSERT INTO "order".order_status_history (
			id, order_id, from_status, to_status, reason, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = conn.Exec(ctx, historyQuery,
		uuid.New(),
		o.ID(),
		"",
		o.Status().String(),
		"order created",
		o.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("inserting status history: %w", err)
	}

	return nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID uuid.UUID, fromStatus, toStatus domain.Status, reason string) error {
	conn := r.pool.Conn(ctx)

	now := time.Now().UTC()

	updateQuery := `
		UPDATE "order".orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	tag, err := conn.Exec(ctx, updateQuery, toStatus.String(), now, orderID)
	if err != nil {
		return fmt.Errorf("updating order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}

	historyQuery := `
		INSERT INTO "order".order_status_history (
			id, order_id, from_status, to_status, reason, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = conn.Exec(ctx, historyQuery,
		uuid.New(),
		orderID,
		fromStatus.String(),
		toStatus.String(),
		reason,
		now,
	)
	if err != nil {
		return fmt.Errorf("inserting status history record: %w", err)
	}

	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT id, user_id, status, total_amount_rub,
		       pickup_point_id, pickup_point_name, pickup_point_latitude, pickup_point_longitude, pickup_point_distance_meters,
		       created_at, updated_at
		FROM "order".orders
		WHERE id = $1
	`

	var (
		orderID        uuid.UUID
		userID         uuid.UUID
		statusStr      string
		totalAmount    decimal.Decimal
		pointID        uuid.UUID
		pointName      string
		pointLat       float64
		pointLon       float64
		pointDist      float64
		createdAt      time.Time
		updatedAt      time.Time
	)

	err := conn.QueryRow(ctx, query, id).Scan(
		&orderID,
		&userID,
		&statusStr,
		&totalAmount,
		&pointID,
		&pointName,
		&pointLat,
		&pointLon,
		&pointDist,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("querying order: %w", err)
	}

	items, err := r.getOrderItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	pickupInfo := domain.PickupPointInfo{
		ID:             pointID,
		Name:           pointName,
		Latitude:       pointLat,
		Longitude:      pointLon,
		DistanceMeters: pointDist,
	}

	return domain.ReconstituteOrder(
		orderID,
		userID,
		domain.Status(statusStr),
		totalAmount,
		pickupInfo,
		items,
		createdAt,
		updatedAt,
	), nil
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Order, int, error) {
	conn := r.pool.Conn(ctx)

	countQuery := `SELECT COUNT(*) FROM "order".orders WHERE user_id = $1`
	var total int
	if err := conn.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting user orders: %w", err)
	}

	query := `
		SELECT id, user_id, status, total_amount_rub,
		       pickup_point_id, pickup_point_name, pickup_point_latitude, pickup_point_longitude, pickup_point_distance_meters,
		       created_at, updated_at
		FROM "order".orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := conn.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying user orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var (
			orderID     uuid.UUID
			uid         uuid.UUID
			statusStr   string
			totalAmount decimal.Decimal
			pointID     uuid.UUID
			pointName   string
			pointLat    float64
			pointLon    float64
			pointDist   float64
			createdAt   time.Time
			updatedAt   time.Time
		)

		if err := rows.Scan(
			&orderID,
			&uid,
			&statusStr,
			&totalAmount,
			&pointID,
			&pointName,
			&pointLat,
			&pointLon,
			&pointDist,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning order row: %w", err)
		}

		pickupInfo := domain.PickupPointInfo{
			ID:             pointID,
			Name:           pointName,
			Latitude:       pointLat,
			Longitude:      pointLon,
			DistanceMeters: pointDist,
		}

		orders = append(orders, domain.ReconstituteOrder(
			orderID,
			uid,
			domain.Status(statusStr),
			totalAmount,
			pickupInfo,
			nil, // lazy loaded items for list
			createdAt,
			updatedAt,
		))
	}

	return orders, total, nil
}

func (r *OrderRepository) HasActivePendingOrder(ctx context.Context, userID uuid.UUID) (bool, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT COUNT(*)
		FROM "order".orders
		WHERE user_id = $1 AND status = 'payment_pending'
	`
	var count int
	if err := conn.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return false, fmt.Errorf("checking active pending order: %w", err)
	}
	return count > 0, nil
}

func (r *OrderRepository) GetStatusHistory(ctx context.Context, orderID uuid.UUID) ([]port.StatusHistoryRecord, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT id, order_id, from_status, to_status, reason, created_at
		FROM "order".order_status_history
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	rows, err := conn.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("querying status history: %w", err)
	}
	defer rows.Close()

	var records []port.StatusHistoryRecord
	for rows.Next() {
		var (
			rec        port.StatusHistoryRecord
			fromStatus string
			toStatus   string
		)
		if err := rows.Scan(&rec.ID, &rec.OrderID, &fromStatus, &toStatus, &rec.Reason, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning history record: %w", err)
		}
		rec.FromStatus = domain.Status(fromStatus)
		rec.ToStatus = domain.Status(toStatus)
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *OrderRepository) IsEventProcessed(ctx context.Context, eventID, consumerName string) (bool, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT COUNT(*)
		FROM "order".processed_events
		WHERE event_id = $1 AND consumer_name = $2
	`
	var count int
	if err := conn.QueryRow(ctx, query, eventID, consumerName).Scan(&count); err != nil {
		return false, fmt.Errorf("checking processed event: %w", err)
	}
	return count > 0, nil
}

func (r *OrderRepository) MarkEventProcessed(ctx context.Context, eventID, consumerName string) error {
	conn := r.pool.Conn(ctx)

	query := `
		INSERT INTO "order".processed_events (event_id, consumer_name, processed_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (event_id) DO NOTHING
	`
	_, err := conn.Exec(ctx, query, eventID, consumerName)
	if err != nil {
		return fmt.Errorf("marking event as processed: %w", err)
	}
	return nil
}

func (r *OrderRepository) getOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT id, product_id, name, category_name, price_rub, image_seed, quantity, subtotal_rub
		FROM "order".order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	rows, err := conn.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("querying order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var (
			id           uuid.UUID
			prodID       uuid.UUID
			name         string
			categoryName string
			priceRUB     decimal.Decimal
			imageSeed    string
			quantity     int
			subtotalRUB  decimal.Decimal
		)
		if err := rows.Scan(&id, &prodID, &name, &categoryName, &priceRUB, &imageSeed, &quantity, &subtotalRUB); err != nil {
			return nil, fmt.Errorf("scanning order item: %w", err)
		}
		items = append(items, domain.NewOrderItem(id, prodID, name, categoryName, priceRUB, imageSeed, quantity))
	}
	return items, rows.Err()
}

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/domain"
)

type PickupRepository struct {
	pool *pgxpool.Pool
}

func NewPickupRepository(pool *pgxpool.Pool) *PickupRepository {
	return &PickupRepository{pool: pool}
}

func (r *PickupRepository) SavePickupPoints(ctx context.Context, points []*domain.PickupPoint) error {
	if len(points) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `
		INSERT INTO pickup.pickup_points (id, user_id, name, latitude, longitude, distance_meters, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, p := range points {
		_, err := tx.Exec(ctx, query,
			p.ID(),
			p.UserID(),
			p.Name(),
			p.Location().Latitude,
			p.Location().Longitude,
			p.DistanceMeters(),
			p.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("inserting pickup point: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing tx: %w", err)
	}

	return nil
}

func (r *PickupRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.PickupPoint, error) {
	query := `
		SELECT id, user_id, name, latitude, longitude, distance_meters, created_at
		FROM pickup.pickup_points
		WHERE user_id = $1
		ORDER BY distance_meters ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying pickup points: %w", err)
	}
	defer rows.Close()

	var points []*domain.PickupPoint
	for rows.Next() {
		var (
			id             uuid.UUID
			uid            uuid.UUID
			name           string
			lat            float64
			lon            float64
			distanceMeters float64
			createdAt      time.Time
		)
		if err := rows.Scan(&id, &uid, &name, &lat, &lon, &distanceMeters, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning pickup point: %w", err)
		}
		loc := domain.LatLng{Latitude: lat, Longitude: lon}
		points = append(points, domain.NewPickupPoint(id, uid, name, loc, distanceMeters, createdAt))
	}

	return points, rows.Err()
}

func (r *PickupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PickupPoint, error) {
	query := `
		SELECT id, user_id, name, latitude, longitude, distance_meters, created_at
		FROM pickup.pickup_points
		WHERE id = $1
	`
	var (
		uid            uuid.UUID
		name           string
		lat            float64
		lon            float64
		distanceMeters float64
		createdAt      time.Time
	)
	err := r.pool.QueryRow(ctx, query, id).Scan(&id, &uid, &name, &lat, &lon, &distanceMeters, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPickupPointNotFound
		}
		return nil, fmt.Errorf("getting pickup point: %w", err)
	}
	loc := domain.LatLng{Latitude: lat, Longitude: lon}
	return domain.NewPickupPoint(id, uid, name, loc, distanceMeters, createdAt), nil
}

func (r *PickupRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM pickup.pickup_points WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("deleting user pickup points: %w", err)
	}
	return nil
}

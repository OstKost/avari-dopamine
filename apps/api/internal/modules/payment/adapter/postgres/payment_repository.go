package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ostkost/dopamine-market/api/internal/modules/payment/domain"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
)

type PaymentRepository struct {
	pool *db.Pool
}

func NewPaymentRepository(pool *db.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Save(ctx context.Context, p *domain.Payment) error {
	conn := r.pool.Conn(ctx)

	query := `
		INSERT INTO payment.payments (
			id, order_id, provider, provider_payment_id, amount_rub, status, confirmation_url, failure_reason, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := conn.Exec(ctx, query,
		p.ID(),
		p.OrderID(),
		p.Provider(),
		p.ProviderPaymentID(),
		p.AmountRUB(),
		string(p.Status()),
		nullString(p.ConfirmationURL()),
		nullString(p.FailureReason()),
		p.CreatedAt(),
		p.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("inserting payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT id, order_id, provider, provider_payment_id, amount_rub, status, confirmation_url, failure_reason, created_at, updated_at
		FROM payment.payments
		WHERE id = $1
	`

	return r.scanPayment(conn.QueryRow(ctx, query, id))
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Payment, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT id, order_id, provider, provider_payment_id, amount_rub, status, confirmation_url, failure_reason, created_at, updated_at
		FROM payment.payments
		WHERE order_id = $1
	`

	return r.scanPayment(conn.QueryRow(ctx, query, orderID))
}

func (r *PaymentRepository) GetByProviderPaymentID(ctx context.Context, providerPaymentID string) (*domain.Payment, error) {
	conn := r.pool.Conn(ctx)

	query := `
		SELECT id, order_id, provider, provider_payment_id, amount_rub, status, confirmation_url, failure_reason, created_at, updated_at
		FROM payment.payments
		WHERE provider_payment_id = $1
	`

	return r.scanPayment(conn.QueryRow(ctx, query, providerPaymentID))
}

func (r *PaymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	conn := r.pool.Conn(ctx)

	query := `
		UPDATE payment.payments
		SET status = $1, failure_reason = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := conn.Exec(ctx, query,
		string(p.Status()),
		nullString(p.FailureReason()),
		p.UpdatedAt(),
		p.ID(),
	)
	if err != nil {
		return fmt.Errorf("updating payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) scanPayment(row pgx.Row) (*domain.Payment, error) {
	var (
		id                uuid.UUID
		orderID           uuid.UUID
		provider          string
		providerPaymentID string
		amountRUB         string
		statusStr         string
		confirmURL        sql.NullString
		failureReason     sql.NullString
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := row.Scan(
		&id,
		&orderID,
		&provider,
		&providerPaymentID,
		&amountRUB,
		&statusStr,
		&confirmURL,
		&failureReason,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("scanning payment: %w", err)
	}

	return domain.ReconstitutePayment(
		id,
		orderID,
		provider,
		providerPaymentID,
		amountRUB,
		domain.Status(statusStr),
		confirmURL.String,
		failureReason.String,
		createdAt,
		updatedAt,
	), nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

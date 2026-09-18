package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO identity.users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, user.ID(), user.Email().String(), user.PasswordHash(), user.CreatedAt())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM identity.users
		WHERE LOWER(email) = LOWER($1)
	`
	var (
		id           uuid.UUID
		rawEmail     string
		passwordHash string
		createdAt    time.Time
	)

	err := r.pool.QueryRow(ctx, query, email.String()).Scan(&id, &rawEmail, &passwordHash, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	parsedEmail, err := domain.NewEmail(rawEmail)
	if err != nil {
		return nil, fmt.Errorf("parsing db email: %w", err)
	}

	return domain.NewUser(id, parsedEmail, passwordHash, createdAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM identity.users
		WHERE id = $1
	`
	var (
		rawEmail     string
		passwordHash string
		createdAt    time.Time
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(&id, &rawEmail, &passwordHash, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}

	parsedEmail, err := domain.NewEmail(rawEmail)
	if err != nil {
		return nil, fmt.Errorf("parsing db email: %w", err)
	}

	return domain.NewUser(id, parsedEmail, passwordHash, createdAt)
}

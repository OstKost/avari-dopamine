// Package contracts содержит межмодульные интерфейсы и DTO (ADR-004).
// Прямой импорт internal/modules/X из internal/modules/Y запрещён.
package contracts

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserInfo — публичная информация о пользователе, доступная другим модулям.
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// IdentityLookup — интерфейс для получения информации о пользователе.
type IdentityLookup interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (UserInfo, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// User — доменный агрегат пользователя.
type User struct {
	id           uuid.UUID
	email        Email
	passwordHash string
	createdAt    time.Time
}

// NewUser создаёт новую сущность пользователя.
func NewUser(id uuid.UUID, email Email, passwordHash string, createdAt time.Time) (*User, error) {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if email.IsZero() {
		return nil, ErrInvalidEmail
	}
	if passwordHash == "" {
		return nil, ErrWeakPassword
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		createdAt:    createdAt,
	}, nil
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

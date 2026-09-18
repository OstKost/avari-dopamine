package domain

import (
	"net/mail"
	"strings"
)

// Email — value object адреса электронной почты.
// Гарантирует lowercase и базовую валидацию формата (FR-AUTH-01).
type Email struct {
	value string
}

// NewEmail валидирует и нормализует строку в Email.
func NewEmail(raw string) (Email, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Email{}, ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return Email{}, ErrInvalidEmail
	}

	// Нормализация к нижнему регистру (case-insensitive уникальность)
	lower := strings.ToLower(trimmed)
	return Email{value: lower}, nil
}

// String возвращает строковое представление email.
func (e Email) String() string {
	return e.value
}

// IsZero проверяет, пустой ли email.
func (e Email) IsZero() bool {
	return e.value == ""
}

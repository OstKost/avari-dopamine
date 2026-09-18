package domain

import "errors"

var (
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrWeakPassword        = errors.New("password must be at least 8 characters long")
	ErrEmailAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrRefreshTokenReused  = errors.New("refresh token reuse detected: session terminated")
)

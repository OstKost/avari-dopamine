package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/port"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	FamilyID     uuid.UUID
}

type AuthResponse struct {
	User   *domain.User
	Tokens TokenPair
}

type AuthUseCase struct {
	userRepo        port.UserRepository
	tokenStore      port.RefreshTokenStore
	hasher          port.PasswordHasher
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthUseCase(
	userRepo port.UserRepository,
	tokenStore port.RefreshTokenStore,
	hasher port.PasswordHasher,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:        userRepo,
		tokenStore:      tokenStore,
		hasher:          hasher,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, rawEmail, password string) (*AuthResponse, error) {
	email, err := domain.NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}

	if len(password) < 8 {
		return nil, domain.ErrWeakPassword
	}

	// Проверяем, существует ли уже пользователь
	existing, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("checking email existence: %w", err)
	}

	passwordHash, err := uc.hasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user, err := domain.NewUser(uuid.New(), email, passwordHash, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Генерируем токены для новой сессии
	tokens, err := uc.generateTokenPair(ctx, user.ID())
	if err != nil {
		return nil, fmt.Errorf("generating tokens: %w", err)
	}

	return &AuthResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, rawEmail, password string) (*AuthResponse, error) {
	email, err := domain.NewEmail(rawEmail)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("finding user: %w", err)
	}

	valid, err := uc.hasher.ComparePassword(user.PasswordHash(), password)
	if err != nil || !valid {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := uc.generateTokenPair(ctx, user.ID())
	if err != nil {
		return nil, fmt.Errorf("generating tokens: %w", err)
	}

	return &AuthResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (uc *AuthUseCase) Refresh(ctx context.Context, refreshTokenString string) (*TokenPair, error) {
	claims, err := uc.parseRefreshToken(refreshTokenString)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	familyID, err := uuid.Parse(claims.FamilyID)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}
	tokenID, err := uuid.Parse(claims.TokenID)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	newTokenID := uuid.New()
	if err := uc.tokenStore.ValidateAndRotate(ctx, familyID, tokenID, newTokenID, userID, uc.refreshTokenTTL); err != nil {
		return nil, err
	}

	accessToken, err := uc.generateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	newRefreshToken, err := uc.generateRefreshTokenString(userID, familyID, newTokenID)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		FamilyID:     familyID,
	}, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, refreshTokenString string) error {
	claims, err := uc.parseRefreshToken(refreshTokenString)
	if err != nil {
		return nil // если токен уже невалиден, считаем logout успешным
	}

	familyID, err := uuid.Parse(claims.FamilyID)
	if err != nil {
		return nil
	}

	return uc.tokenStore.RevokeFamily(ctx, familyID)
}

func (uc *AuthUseCase) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

// Helpers for JWT

type RefreshTokenClaims struct {
	FamilyID string `json:"fid"`
	TokenID  string `json:"tid"`
	jwt.RegisteredClaims
}

func (uc *AuthUseCase) generateTokenPair(ctx context.Context, userID uuid.UUID) (TokenPair, error) {
	familyID := uuid.New()
	tokenID := uuid.New()

	if err := uc.tokenStore.SaveToken(ctx, familyID, tokenID, userID, uc.refreshTokenTTL); err != nil {
		return TokenPair{}, err
	}

	access, err := uc.generateAccessToken(userID)
	if err != nil {
		return TokenPair{}, err
	}

	refresh, err := uc.generateRefreshTokenString(userID, familyID, tokenID)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		FamilyID:     familyID,
	}, nil
}

func (uc *AuthUseCase) generateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": now.Unix(),
		"exp": now.Add(uc.accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *AuthUseCase) generateRefreshTokenString(userID, familyID, tokenID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := RefreshTokenClaims{
		FamilyID: familyID.String(),
		TokenID:  tokenID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(uc.refreshTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}

func (uc *AuthUseCase) parseRefreshToken(tokenStr string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(uc.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, domain.ErrInvalidRefreshToken
	}

	return claims, nil
}

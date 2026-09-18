package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/domain"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/usecase"
)

type mockUserRepo struct {
	usersByEmail map[string]*domain.User
	usersByID    map[uuid.UUID]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByEmail: make(map[string]*domain.User),
		usersByID:    make(map[uuid.UUID]*domain.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.usersByEmail[user.Email().String()]; exists {
		return domain.ErrEmailAlreadyExists
	}
	m.usersByEmail[user.Email().String()] = user
	m.usersByID[user.ID()] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	u, ok := m.usersByEmail[email.String()]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

type mockTokenStore struct {
	families map[uuid.UUID]uuid.UUID // familyID -> activeTokenID
	revoked  map[uuid.UUID]bool
}

func newMockTokenStore() *mockTokenStore {
	return &mockTokenStore{
		families: make(map[uuid.UUID]uuid.UUID),
		revoked:  make(map[uuid.UUID]bool),
	}
}

func (m *mockTokenStore) SaveToken(ctx context.Context, familyID, tokenID, userID uuid.UUID, ttl time.Duration) error {
	m.families[familyID] = tokenID
	m.revoked[familyID] = false
	return nil
}

func (m *mockTokenStore) ValidateAndRotate(ctx context.Context, familyID, tokenID, newID, userID uuid.UUID, ttl time.Duration) error {
	if m.revoked[familyID] {
		return domain.ErrRefreshTokenReused
	}
	activeID, exists := m.families[familyID]
	if !exists {
		return domain.ErrInvalidRefreshToken
	}
	if activeID != tokenID {
		// Reuse detected!
		m.revoked[familyID] = true
		return domain.ErrRefreshTokenReused
	}
	m.families[familyID] = newID
	return nil
}

func (m *mockTokenStore) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	delete(m.families, familyID)
	m.revoked[familyID] = true
	return nil
}

type mockHasher struct{}

func (m *mockHasher) HashPassword(p string) (string, error) {
	return "hash_" + p, nil
}

func (m *mockHasher) ComparePassword(hash, password string) (bool, error) {
	return hash == "hash_"+password, nil
}

func TestAuthUseCase_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepo()
	store := newMockTokenStore()
	hasher := &mockHasher{}
	secret := "test-secret-at-least-32-bytes-long!!"

	uc := usecase.NewAuthUseCase(repo, store, hasher, secret, 15*time.Minute, 30*24*time.Hour)
	ctx := context.Background()

	// 1. Weak password
	_, err := uc.Register(ctx, "test@example.com", "short")
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}

	// 2. Successful register
	resp, err := uc.Register(ctx, "Test@Example.Com", "password123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if resp.User.Email().String() != "test@example.com" {
		t.Errorf("email not normalized, got %q", resp.User.Email().String())
	}
	if resp.Tokens.AccessToken == "" || resp.Tokens.RefreshToken == "" {
		t.Errorf("expected tokens to be non-empty")
	}

	// 3. Duplicate email
	_, err = uc.Register(ctx, "test@example.com", "password123")
	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}

	// 4. Successful login
	loginResp, err := uc.Login(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.User.ID() != resp.User.ID() {
		t.Errorf("user ID mismatch on login")
	}

	// 5. Invalid credentials
	_, err = uc.Login(ctx, "test@example.com", "wrongpassword")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthUseCase_RefreshRotationAndReuseDetection(t *testing.T) {
	repo := newMockUserRepo()
	store := newMockTokenStore()
	hasher := &mockHasher{}
	secret := "test-secret-at-least-32-bytes-long!!"

	uc := usecase.NewAuthUseCase(repo, store, hasher, secret, 15*time.Minute, 30*24*time.Hour)
	ctx := context.Background()

	resp, err := uc.Register(ctx, "user@example.com", "securepassword")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	firstRefreshToken := resp.Tokens.RefreshToken

	// 1. Valid Refresh
	newTokens, err := uc.Refresh(ctx, firstRefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if newTokens.RefreshToken == firstRefreshToken {
		t.Errorf("refresh token was not rotated!")
	}

	// 2. Reuse detection: using old firstRefreshToken again must fail with ErrRefreshTokenReused
	_, err = uc.Refresh(ctx, firstRefreshToken)
	if !errors.Is(err, domain.ErrRefreshTokenReused) {
		t.Fatalf("expected ErrRefreshTokenReused, got %v", err)
	}

	// 3. Subsequent attempts with newTokens must also fail because family is revoked!
	_, err = uc.Refresh(ctx, newTokens.RefreshToken)
	if !errors.Is(err, domain.ErrRefreshTokenReused) {
		t.Fatalf("expected family to be completely revoked after reuse, got %v", err)
	}
}

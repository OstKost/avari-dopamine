package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ostkost/dopamine-market/api/internal/modules/identity/domain"
)

func TestEmail_Validation(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{"valid lowercase", "user@example.com", "user@example.com", false},
		{"valid uppercase", "USER@EXAMPLE.COM", "user@example.com", false},
		{"valid mixed case and whitespace", "  User.Name+tag@Sub.Domain.Org  ", "user.name+tag@sub.domain.org", false},
		{"invalid empty", "", "", true},
		{"invalid spaces only", "   ", "", true},
		{"invalid missing @", "userexample.com", "", true},
		{"invalid missing domain", "user@", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewEmail(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewEmail(%q) err = %v, wantErr = %v", tt.raw, err, tt.wantErr)
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("NewEmail(%q) = %q, want %q", tt.raw, got.String(), tt.want)
			}
		})
	}
}

func TestUser_Creation(t *testing.T) {
	email, err := domain.NewEmail("test@example.com")
	if err != nil {
		t.Fatalf("failed to create email: %v", err)
	}

	userID := uuid.New()
	user, err := domain.NewUser(userID, email, "hashed_password", time.Now())
	if err != nil {
		t.Fatalf("NewUser failed: %v", err)
	}

	if user.ID() != userID {
		t.Errorf("user ID mismatch")
	}
	if user.Email().String() != "test@example.com" {
		t.Errorf("user Email mismatch")
	}
}

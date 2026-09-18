package config

import (
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/dopamine_market")
	t.Setenv("AUTH_JWT_SECRET", "test-secret-must-be-long-enough")
}

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Env != "local" {
		t.Errorf("expected default Env=local, got %q", cfg.Env)
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("expected default HTTP.Port=8080, got %d", cfg.HTTP.Port)
	}
	if cfg.Payment.Provider != "mock" {
		t.Errorf("expected default Payment.Provider=mock, got %q", cfg.Payment.Provider)
	}
	if cfg.Payment.MockSuccessRate != 0.95 {
		t.Errorf("expected default MockSuccessRate=0.95, got %v", cfg.Payment.MockSuccessRate)
	}
}

func TestLoad_MissingRequiredFails(t *testing.T) {
	// Явно не устанавливаем DATABASE_URL/AUTH_JWT_SECRET.
	t.Setenv("DATABASE_URL", "")
	t.Setenv("AUTH_JWT_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when required env vars are missing, got nil")
	}
}

func TestLoad_YooKassaWithoutSecretsFails(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PAYMENT_PROVIDER", "yookassa")
	// Секреты YooKassa намеренно не устанавливаем.

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when PAYMENT_PROVIDER=yookassa without shop credentials")
	}
}

func TestLoad_YooKassaWithSecretsSucceeds(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PAYMENT_PROVIDER", "yookassa")
	t.Setenv("YOOKASSA_SHOP_ID", "test-shop")
	t.Setenv("YOOKASSA_SECRET_KEY", "test-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Payment.Provider != "yookassa" {
		t.Errorf("expected Provider=yookassa, got %q", cfg.Payment.Provider)
	}
}

func TestLoad_InvalidPaymentProviderFails(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PAYMENT_PROVIDER", "stripe")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for unsupported PAYMENT_PROVIDER value")
	}
}

func TestLoad_KafkaBrokersSplitting(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("KAFKA_BROKERS", "broker1:9092,broker2:9092")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if len(cfg.Kafka.Brokers) != 2 {
		t.Errorf("expected 2 brokers, got %d: %v", len(cfg.Kafka.Brokers), cfg.Kafka.Brokers)
	}
}

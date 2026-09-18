// Package config загружает конфигурацию приложения из переменных окружения.
//
// Принцип: fail-fast. Если обязательная переменная отсутствует или не парсится,
// приложение должно упасть при старте с понятной ошибкой, а не в середине
// работы с nil-полем или пустой строкой там, где ожидалось значение.
//
// См. ADR-006 (PAYMENT_PROVIDER выбор адаптера) и AGENTS.md раздел 3.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config — корневая конфигурация процесса. cmd/server, cmd/worker, cmd/seed
// и cmd/migrator используют общий Load(), но не все поля обязательны для
// каждого бинарника — валидация обязательности выполняется на уровне
// конкретного cmd, а не здесь (Load сам по себе не решает, кто его вызывает).
type Config struct {
	Env     string // local | staging | production — влияет на формат логов (ADR-010)
	HTTP    HTTPConfig
	DB      DBConfig
	Redis   RedisConfig
	Kafka   KafkaConfig
	Logger  LoggerConfig
	Auth    AuthConfig
	Payment PaymentConfig
}

// HTTPConfig настраивает HTTP-сервер (cmd/server), см. internal/platform/httpserver.
type HTTPConfig struct {
	Port            int
	ShutdownTimeout time.Duration
}

// DBConfig настраивает пул PostgreSQL, см. internal/platform/db.
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig настраивает клиент Redis, см. internal/platform/redis.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// KafkaConfig настраивает продюсер/консьюмер Kafka, см. internal/platform/kafka.
type KafkaConfig struct {
	Brokers []string
	// ClientID идентифицирует продюсера/консьюмера в логах брокера — полезно при дебаге.
	ClientID string
}

// LoggerConfig настраивает slog+GELF логгер, см. internal/platform/logger и ADR-010.
type LoggerConfig struct {
	// Level: debug | info | warn | error
	Level string
	// GraylogHost/Port — GELF UDP endpoint (ADR-010). Пустой Host отключает GELF-транспорт
	// (например, при ENV=local без поднятого Graylog — тогда используется только stdout).
	GraylogHost string
	GraylogPort int
}

// AuthConfig настраивает JWT-аутентификацию, см. ADR-007 (реализация — EPIC-01).
type AuthConfig struct {
	// JWTSecret подписывает access-токены (ADR-007, HS256 на данном масштабе).
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// PaymentConfig согласно ADR-006: выбор адаптера через конфигурацию, домен не знает деталей.
type PaymentConfig struct {
	// Provider: "mock" | "yookassa"
	Provider string
	// MockSuccessRate — доля успешных платежей MockProvider, по умолчанию 0.95.
	MockSuccessRate   float64
	YooKassaShopID    string
	YooKassaSecretKey string
}

// Load читает конфигурацию из переменных окружения, применяя разумные
// значения по умолчанию там, где отсутствие переменной не критично
// (тайминги, порты), и требуя явного значения там, где дефолт был бы
// опасен (секреты, DSN).
func Load() (Config, error) {
	var errs []string

	cfg := Config{
		Env: getEnvDefault("ENV", "local"),
		HTTP: HTTPConfig{
			Port:            getEnvIntDefault("HTTP_PORT", 8080),
			ShutdownTimeout: getEnvDurationDefault("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		DB: DBConfig{
			DSN:             mustGetEnv("DATABASE_URL", &errs),
			MaxOpenConns:    getEnvIntDefault("DB_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    getEnvIntDefault("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDurationDefault("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     getEnvDefault("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvIntDefault("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Brokers:  strings.Split(getEnvDefault("KAFKA_BROKERS", "localhost:9092"), ","),
			ClientID: getEnvDefault("KAFKA_CLIENT_ID", "dopamine-market-api"),
		},
		Logger: LoggerConfig{
			Level:       getEnvDefault("LOG_LEVEL", "info"),
			GraylogHost: os.Getenv("GRAYLOG_HOST"),
			GraylogPort: getEnvIntDefault("GRAYLOG_PORT", 12201),
		},
		Auth: AuthConfig{
			JWTSecret:       mustGetEnv("AUTH_JWT_SECRET", &errs),
			AccessTokenTTL:  getEnvDurationDefault("AUTH_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvDurationDefault("AUTH_REFRESH_TOKEN_TTL", 30*24*time.Hour),
		},
		Payment: PaymentConfig{
			Provider:          getEnvDefault("PAYMENT_PROVIDER", "mock"),
			MockSuccessRate:   getEnvFloatDefault("MOCK_PAYMENT_SUCCESS_RATE", 0.95),
			YooKassaShopID:    os.Getenv("YOOKASSA_SHOP_ID"),
			YooKassaSecretKey: os.Getenv("YOOKASSA_SECRET_KEY"),
		},
	}

	// Кросс-полевая валидация: ADR-006 требует fail-fast, если выбран yookassa
	// без секретов — тихий fallback на mock запрещён (может привести к
	// незамеченной потере интеграции в проде).
	if cfg.Payment.Provider == "yookassa" {
		if cfg.Payment.YooKassaShopID == "" {
			errs = append(errs, "YOOKASSA_SHOP_ID is required when PAYMENT_PROVIDER=yookassa")
		}
		if cfg.Payment.YooKassaSecretKey == "" {
			errs = append(errs, "YOOKASSA_SECRET_KEY is required when PAYMENT_PROVIDER=yookassa")
		}
	} else if cfg.Payment.Provider != "mock" {
		errs = append(errs, fmt.Sprintf("PAYMENT_PROVIDER must be 'mock' or 'yookassa', got %q", cfg.Payment.Provider))
	}

	if len(errs) > 0 {
		return Config{}, fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return cfg, nil
}

func mustGetEnv(key string, errs *[]string) string {
	v := os.Getenv(key)
	if v == "" {
		*errs = append(*errs, fmt.Sprintf("%s is required", key))
	}
	return v
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvIntDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvFloatDefault(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func getEnvDurationDefault(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ostkost/dopamine-market/api/internal/platform/db"
	"github.com/ostkost/dopamine-market/api/internal/platform/kafka"
)

// Event представляет outbox-событие для гарантированной публикации в Kafka (ADR-003).
type Event struct {
	ID          uuid.UUID       `json:"id"`
	Topic       string          `json:"topic"`
	EventKey    string          `json:"event_key"`
	EventType   string          `json:"event_type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
}

// NewEvent создаёт новое outbox-событие с маршалингом полезной нагрузки.
func NewEvent(topic, eventKey, eventType string, payload interface{}) (Event, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("marshaling outbox payload: %w", err)
	}

	return Event{
		ID:        uuid.New(),
		Topic:     topic,
		EventKey:  eventKey,
		EventType: eventType,
		Payload:   bytes,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// SaveInTx сохраняет outbox-событие в таблицу указанной схемы в рамках активной транзакции (ADR-003 / ADR-011).
func SaveInTx(ctx context.Context, pool *db.Pool, schema string, event Event) error {
	conn := pool.Conn(ctx)

	query := fmt.Sprintf(`
		INSERT INTO "%s".outbox_events (id, topic, event_key, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, schema)

	_, err := conn.Exec(ctx, query, event.ID, event.Topic, event.EventKey, event.EventType, event.Payload, event.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting outbox event in schema %q: %w", schema, err)
	}
	return nil
}

// Relay отвечает за периодический опрос таблиц outbox_events и публикацию событий в Kafka.
type Relay struct {
	pool     *db.Pool
	producer *kafka.Producer
	schemas  []string
	logger   *slog.Logger
	interval time.Duration
}

func NewRelay(pool *db.Pool, producer *kafka.Producer, schemas []string, logger *slog.Logger, interval time.Duration) *Relay {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &Relay{
		pool:     pool,
		producer: producer,
		schemas:  schemas,
		logger:   logger,
		interval: interval,
	}
}

// Run запускает фоновый процесс релея до отмены контекста.
func (r *Relay) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for _, schema := range r.schemas {
				if err := r.processSchema(ctx, schema); err != nil && !isCancellation(err) {
					r.logger.Error("outbox relay error", slog.String("schema", schema), slog.String("error", err.Error()))
				}
			}
		}
	}
}

func (r *Relay) processSchema(ctx context.Context, schema string) error {
	query := fmt.Sprintf(`
		SELECT id, topic, event_key, event_type, payload, created_at
		FROM "%s".outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at ASC
		LIMIT 50
	`, schema)

	rows, err := r.pool.Raw().Query(ctx, query)
	if err != nil {
		return fmt.Errorf("querying unpublished outbox events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Topic, &e.EventKey, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return fmt.Errorf("scanning outbox event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, e := range events {
		if r.producer != nil {
			msg := kafka.Message{
				Topic: e.Topic,
				Key:   []byte(e.EventKey),
				Value: e.Payload,
				Headers: map[string]string{
					"event_type": e.EventType,
					"event_id":   e.ID.String(),
				},
			}
			if err := r.producer.Publish(ctx, msg); err != nil {
				return fmt.Errorf("publishing event %s to kafka: %w", e.ID, err)
			}
		}

		// Помечаем событие как опубликованное
		markQuery := fmt.Sprintf(`UPDATE "%s".outbox_events SET published_at = NOW() WHERE id = $1`, schema)
		if _, err := r.pool.Raw().Exec(ctx, markQuery, e.ID); err != nil {
			return fmt.Errorf("marking outbox event as published: %w", err)
		}
	}

	return nil
}

func isCancellation(err error) bool {
	return err == context.Canceled || err == context.DeadlineExceeded || pgx.ErrTxClosed == err
}

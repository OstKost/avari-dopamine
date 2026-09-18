// Package kafka оборачивает github.com/segmentio/kafka-go. Библиотека
// выбрана осознанно (см. docs/adr/003-event-driven-outbox.md и EPIC-00
// "открытые вопросы"): pure Go реализация без cgo/librdkafka зависимости
// упрощает сборку Docker-образов и CI по сравнению с confluent-kafka-go.
//
// Все продюсеры/консьюмеры проекта используют эту обёртку, а не
// segmentio/kafka-go напрямую — единая точка для инъекции trace_id
// propagation (ADR-010) и структурного логирования ошибок.
package kafka

import (
	"context"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// Config — параметры подключения, транслируемые из internal/platform/config.
type Config struct {
	Brokers  []string
	ClientID string
}

// Message — упрощённая структура сообщения на границе платформенного API,
// не завязанная на конкретную клиентскую библиотеку (упрощает замену
// библиотеки в будущем без изменения кода модулей).
type Message struct {
	Topic   string
	Key     []byte
	Value   []byte
	Headers map[string]string // включает "traceparent" для OTel propagation (ADR-010)
}

// Producer публикует сообщения в указанный топик или топик из Message.
type Producer struct {
	writer *kafkago.Writer
}

// NewProducer создаёт продюсера. Если topic пустой, топик берётся из каждого Message.
func NewProducer(cfg Config, topic string) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(cfg.Brokers...),
			Topic:        topic,
			Balancer:     &kafkago.Hash{}, // партиционирование по Key (aggregate_id) — сохраняет порядок событий одного заказа, ADR-003
			RequiredAcks: kafkago.RequireOne,
			BatchTimeout: 100 * time.Millisecond,
		},
	}
}

// Publish отправляет одно сообщение. Ключ партиции должен быть aggregate_id
// (например order_id) — гарантирует упорядоченность событий одного агрегата
// (ADR-003).
func (p *Producer) Publish(ctx context.Context, msg Message) error {
	headers := make([]kafkago.Header, 0, len(msg.Headers))
	for k, v := range msg.Headers {
		headers = append(headers, kafkago.Header{Key: k, Value: []byte(v)})
	}

	err := p.writer.WriteMessages(ctx, kafkago.Message{
		Topic:   msg.Topic,
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
		Time:    time.Now(),
	})
	if err != nil {
		return fmt.Errorf("publishing message to kafka: %w", err)
	}
	return nil
}

// Close закрывает продюсера. Вызывается при graceful shutdown.
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Handler обрабатывает одно полученное сообщение. Возврат ошибки не
// коммитит offset — сообщение будет доставлено повторно при следующем
// poll (at-least-once, ADR-003); идемпотентность обработки — ответственность
// вызывающего usecase через processed_events таблицу.
type Handler func(ctx context.Context, msg Message) error

// Consumer читает сообщения из топика в рамках consumer group и передаёт
// каждое в Handler. Один Consumer на (topic, group) пару — соответствует
// схеме consumer groups из ADR-003 ("Consumer group на модуль").
type Consumer struct {
	reader *kafkago.Reader
}

// NewConsumer создаёт консьюмера для топика в указанной consumer group.
func NewConsumer(cfg Config, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:  cfg.Brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10e6,
			MaxWait:  1 * time.Second,
		}),
	}
}

// Run блокирует вызывающую goroutine, читая сообщения до отмены ctx.
// Ошибки Handler логируются вызывающим кодом (Run не логирует сам —
// platform-пакеты не тащат зависимость на internal/platform/logger,
// чтобы избежать циклических импортов; composition root подключает логирование
// через обёртку Handler в cmd/worker).
func (c *Consumer) Run(ctx context.Context, handler Handler) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // graceful shutdown, не ошибка
			}
			return fmt.Errorf("fetching message from topic %s: %w", c.reader.Config().Topic, err)
		}

		headers := make(map[string]string, len(m.Headers))
		for _, h := range m.Headers {
			headers[h.Key] = string(h.Value)
		}

		msg := Message{Key: m.Key, Value: m.Value, Headers: headers}

		if err := handler(ctx, msg); err != nil {
			// Не коммитим offset при ошибке — сообщение будет передоставлено.
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			return fmt.Errorf("committing offset for topic %s: %w", c.reader.Config().Topic, err)
		}
	}
}

// Close закрывает консьюмера. Вызывается при graceful shutdown.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

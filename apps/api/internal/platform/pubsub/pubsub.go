package pubsub

import (
	"sync"

	"github.com/google/uuid"
)

// Event представляет абстрактное событие для доставки подписчикам SSE (ADR-009).
type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// Hub — in-process pub/sub диспетчер событий заказов для SSE соединений.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID][]chan Event
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uuid.UUID][]chan Event),
	}
}

// Subscribe регистрирует новый канал для получения событий указанного заказа.
// Возвращает канал событий и функцию отписки.
func (h *Hub) Subscribe(orderID uuid.UUID) (<-chan Event, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Event, 16)
	h.subscribers[orderID] = append(h.subscribers[orderID], ch)

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		subs := h.subscribers[orderID]
		for i, sub := range subs {
			if sub == ch {
				h.subscribers[orderID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
		if len(h.subscribers[orderID]) == 0 {
			delete(h.subscribers, orderID)
		}
	}

	return ch, unsubscribe
}

// Publish рассылает событие всем активным подписчикам данного заказа (non-blocking).
func (h *Hub) Publish(orderID uuid.UUID, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs, ok := h.subscribers[orderID]
	if !ok {
		return
	}

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// Если буфер переполнен, пропускаем во избежание зависания
		}
	}
}

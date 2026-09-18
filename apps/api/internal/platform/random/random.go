// Package random предоставляет инъектируемый источник случайности.
//
// Почему не math/rand напрямую в domain/usecase коде: AGENTS.md раздел 7
// запрещает прямые вызовы math/rand вне этого пакета. Причина —
// тестируемость (ADR-012): любой usecase, зависящий от случайности
// (расстояние ПВЗ, длительность доставки, выбор курьера), должен принимать
// random.Source через конструктор, чтобы тесты могли подставить
// детерминированный источник и проверять точные значения или границы
// диапазонов без flaky-тестов.
package random

import (
	"math/rand"
	"sync"
)

// Source — контракт случайности, которым пользуются usecase-слои модулей.
// Не экспортирует сам rand.Rand, чтобы не тащить лишний API туда, где он
// не нужен, и чтобы можно было легко подменить реализацию (например, на
// криптографически стойкую, если когда-либо потребуется).
type Source interface {
	// Float64Range возвращает случайное число в [min, max).
	Float64Range(min, max float64) float64
	// IntRange возвращает случайное целое в [min, max] (обе границы включительно).
	IntRange(min, max int) int
	// Bool возвращает true с вероятностью p (0.0..1.0).
	Bool(p float64) bool
	// Pick выбирает случайный элемент среза. Паникует на пустом срезе —
	// вызывающий код обязан гарантировать непустой пул значений.
	Pick(items []string) string
}

// mathRandSource — реализация на math/rand с собственным mutex, так как
// стандартный rand.Rand не safe for concurrent use, а платформенные
// сервисы (delivery scheduler, payment mock) обращаются к общему источнику
// из нескольких goroutine.
type mathRandSource struct {
	mu  sync.Mutex
	rnd *rand.Rand
}

// New создаёт источник с явным seed. Для продакшн-кода использовать
// seed на основе time.Now().UnixNano() один раз при старте процесса
// (не на каждый вызов). Для тестов — фиксированный seed для
// воспроизводимости (ADR-012).
func New(seed int64) Source {
	return &mathRandSource{rnd: rand.New(rand.NewSource(seed))}
}

func (s *mathRandSource) Float64Range(min, max float64) float64 {
	if min >= max {
		return min
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return min + s.rnd.Float64()*(max-min)
}

func (s *mathRandSource) IntRange(min, max int) int {
	if min >= max {
		return min
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return min + s.rnd.Intn(max-min+1)
}

func (s *mathRandSource) Bool(p float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rnd.Float64() < p
}

func (s *mathRandSource) Pick(items []string) string {
	if len(items) == 0 {
		panic("random.Pick: empty items slice")
	}
	s.mu.Lock()
	idx := s.rnd.Intn(len(items))
	s.mu.Unlock()
	return items[idx]
}

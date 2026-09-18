package logger

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"
)

// gelfMessage — минимальная структура GELF-сообщения версии 1.1.
// Спецификация: https://docs.graylog.org/docs/gelf
type gelfMessage struct {
	Version      string  `json:"version"`
	Host         string  `json:"host"`
	ShortMessage string  `json:"short_message"`
	Timestamp    float64 `json:"timestamp"`
	Level        int     `json:"level"` // syslog severity: 3=error, 4=warning, 6=info, 7=debug
	// Дополнительные поля GELF требуют префикса "_" в имени ключа.
	Extra map[string]any `json:"-"`
}

// gelfHandler — slog.Handler, отправляющий записи по UDP в формате GELF
// в Graylog (ADR-010). UDP выбран осознанно для минимальной задержки на
// critical path; риск потери отдельных сообщений под нагрузкой принят
// и задокументирован в ADR-010 как trade-off, приемлемый на масштабе проекта.
type gelfHandler struct {
	conn     net.Conn
	host     string
	minLevel slog.Level
	attrs    []slog.Attr
	groups   []string
}

func newGELFHandler(host string, port int, minLevel slog.Level) (slog.Handler, func(), error) {
	// net.JoinHostPort вместо fmt.Sprintf("%s:%d", ...) — корректно
	// обёртывает host в скобки для IPv6-адресов (go vet host:port flag).
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, func() {}, fmt.Errorf("dialing graylog GELF UDP endpoint %s: %w", addr, err)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown-host"
	}

	h := &gelfHandler{conn: conn, host: hostname, minLevel: minLevel}
	cleanup := func() {
		_ = conn.Close()
	}
	return h, cleanup, nil
}

func (h *gelfHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *gelfHandler) Handle(_ context.Context, r slog.Record) error {
	msg := gelfMessage{
		Version:      "1.1",
		Host:         h.host,
		ShortMessage: r.Message,
		Timestamp:    float64(r.Time.UnixNano()) / float64(time.Second),
		Level:        slogLevelToSyslog(r.Level),
	}

	fields := make(map[string]any)
	groupPrefix := groupKeyPrefix(h.groups)
	for _, a := range h.attrs {
		fields[groupPrefix+a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		fields[groupPrefix+a.Key] = a.Value.Any()
		return true
	})

	payload, err := marshalGELF(msg, fields)
	if err != nil {
		return fmt.Errorf("marshaling GELF message: %w", err)
	}

	compressed, err := compressGELF(payload)
	if err != nil {
		return fmt.Errorf("compressing GELF message: %w", err)
	}

	_, err = h.conn.Write(compressed)
	return err
}

func (h *gelfHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := &gelfHandler{conn: h.conn, host: h.host, minLevel: h.minLevel}
	next.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return next
}

func (h *gelfHandler) WithGroup(name string) slog.Handler {
	// Группы не разворачиваются в отдельную вложенность GELF (плоская модель
	// полей), просто префиксуем ключи при следующем WithAttrs — упрощённо,
	// достаточно для нужд проекта (глубокая вложенность slog-групп не
	// используется в кодовой базе).
	next := &gelfHandler{conn: h.conn, host: h.host, minLevel: h.minLevel, attrs: h.attrs}
	next.groups = append(append([]string{}, h.groups...), name)
	return next
}

// groupKeyPrefix строит точечный префикс из цепочки имён групп, например
// ["payment", "webhook"] -> "payment.webhook.". Для пустой цепочки возвращает
// пустую строку (никакого префикса).
func groupKeyPrefix(groups []string) string {
	if len(groups) == 0 {
		return ""
	}
	prefix := ""
	for _, g := range groups {
		prefix += g + "."
	}
	return prefix
}

func marshalGELF(msg gelfMessage, extra map[string]any) ([]byte, error) {
	base := map[string]any{
		"version":       msg.Version,
		"host":          msg.Host,
		"short_message": msg.ShortMessage,
		"timestamp":     msg.Timestamp,
		"level":         msg.Level,
	}
	for k, v := range extra {
		// GELF требует, чтобы кастомные поля были с префиксом "_" и не назывались "_id".
		key := "_" + k
		if key == "_id" {
			key = "_id_field"
		}
		base[key] = v
	}
	return json.Marshal(base)
}

func compressGELF(payload []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(payload); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func slogLevelToSyslog(level slog.Level) int {
	switch {
	case level >= slog.LevelError:
		return 3
	case level >= slog.LevelWarn:
		return 4
	case level >= slog.LevelInfo:
		return 6
	default:
		return 7
	}
}

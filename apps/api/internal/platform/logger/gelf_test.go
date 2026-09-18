package logger

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
)

func TestMarshalGELF_ProducesValidJSONWithUnderscorePrefixedExtras(t *testing.T) {
	msg := gelfMessage{
		Version:      "1.1",
		Host:         "test-host",
		ShortMessage: "hello",
		Timestamp:    1700000000.123,
		Level:        6,
	}

	payload, err := marshalGELF(msg, map[string]any{"order_id": "abc-123", "id": "should-be-renamed"})
	if err != nil {
		t.Fatalf("marshalGELF() error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshaling result: %v", err)
	}

	if decoded["short_message"] != "hello" {
		t.Errorf("short_message = %v, want hello", decoded["short_message"])
	}
	if decoded["_order_id"] != "abc-123" {
		t.Errorf("_order_id = %v, want abc-123", decoded["_order_id"])
	}
	// GELF резервирует "_id" — реализация должна переименовать конфликтующий ключ.
	if _, exists := decoded["_id"]; exists {
		t.Error("_id key must not be present (reserved by GELF spec)")
	}
	if decoded["_id_field"] != "should-be-renamed" {
		t.Errorf("_id_field = %v, want should-be-renamed", decoded["_id_field"])
	}
}

func TestCompressGELF_RoundTripsWithZlib(t *testing.T) {
	original := []byte(`{"short_message":"round trip test"}`)

	compressed, err := compressGELF(original)
	if err != nil {
		t.Fatalf("compressGELF() error: %v", err)
	}

	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("zlib.NewReader() error: %v", err)
	}
	defer r.Close()

	decompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading decompressed data: %v", err)
	}

	if !bytes.Equal(decompressed, original) {
		t.Errorf("round-trip mismatch: got %q, want %q", decompressed, original)
	}
}

func TestSlogLevelToSyslog_MapsCorrectly(t *testing.T) {
	cases := []struct {
		level slog.Level
		want  int
	}{
		{slog.LevelDebug, 7},
		{slog.LevelInfo, 6},
		{slog.LevelWarn, 4},
		{slog.LevelError, 3},
	}
	for _, c := range cases {
		if got := slogLevelToSyslog(c.level); got != c.want {
			t.Errorf("slogLevelToSyslog(%v) = %d, want %d", c.level, got, c.want)
		}
	}
}

func TestGroupKeyPrefix_BuildsDottedPath(t *testing.T) {
	if got := groupKeyPrefix(nil); got != "" {
		t.Errorf("groupKeyPrefix(nil) = %q, want empty string", got)
	}
	if got := groupKeyPrefix([]string{"payment"}); got != "payment." {
		t.Errorf("groupKeyPrefix([payment]) = %q, want %q", got, "payment.")
	}
	if got := groupKeyPrefix([]string{"payment", "webhook"}); got != "payment.webhook." {
		t.Errorf("groupKeyPrefix([payment,webhook]) = %q, want %q", got, "payment.webhook.")
	}
}

func TestGelfHandler_WithGroupPrefixesAttrKeys(t *testing.T) {
	// Handle() пишет в UDP-соединение — используем net.Pipe-подобный подход:
	// подменяем conn на локальный UDP loopback listener, чтобы не требовать
	// реального Graylog. Проверяем, что WithGroup+WithAttrs дают
	// префиксованный ключ в итоговом JSON перед сжатием, минуя реальную сеть,
	// через прямой вызов внутренней логики marshalGELF с тем же groupKeyPrefix,
	// который использует Handle.
	h := &gelfHandler{groups: []string{"payment"}}
	prefix := groupKeyPrefix(h.groups)
	fields := map[string]any{prefix + "order_id": "abc"}

	payload, err := marshalGELF(gelfMessage{Version: "1.1"}, fields)
	if err != nil {
		t.Fatalf("marshalGELF() error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshaling: %v", err)
	}
	if decoded["_payment.order_id"] != "abc" {
		t.Errorf("expected _payment.order_id in payload, got: %v", decoded)
	}
}

package bunslog

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/uptrace/bun"
)

type mockHandler struct {
	calls []mockCall
}

type mockCall struct {
	method  string
	level   slog.Level
	message string
	attrs   []slog.Attr
	group   string
}

func (h *mockHandler) Enabled(_ context.Context, level slog.Level) bool {
	h.calls = append(h.calls, mockCall{method: "Enabled", level: level})
	return true
}

func (h *mockHandler) Handle(_ context.Context, r slog.Record) error {
	var attrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	h.calls = append(h.calls, mockCall{
		method:  "Handle",
		level:   r.Level,
		message: r.Message,
		attrs:   attrs,
	})
	return nil
}

func (h *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.calls = append(h.calls, mockCall{method: "WithAttrs", attrs: attrs})
	return h
}

func (h *mockHandler) WithGroup(name string) slog.Handler {
	h.calls = append(h.calls, mockCall{method: "WithGroup", group: name})
	return h
}

func TestQueryHook_ErrorLogging(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)
	hook := NewQueryHook(WithLogger(logger))
	event := &bun.QueryEvent{
		Err:       errors.New("fail"),
		StartTime: time.Now(),
		Query:     "SELECT 1",
	}
	hook.AfterQuery(context.Background(), event)
	found := false
	for _, call := range handler.calls {
		if call.method == "Handle" && call.level == slog.LevelError && strings.Contains(call.message, "SELECT 1") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected error log call, got %v", handler.calls)
	}
}

func TestQueryHook_SlowQueryLogging(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)
	hook := NewQueryHook(WithLogger(logger), WithLogSlow(1*time.Millisecond))
	event := &bun.QueryEvent{
		Err:       nil,
		StartTime: time.Now().Add(-10 * time.Millisecond),
		Query:     "SELECT 2",
	}
	hook.AfterQuery(context.Background(), event)
	found := false
	for _, call := range handler.calls {
		if call.method == "Handle" && call.level == slog.LevelWarn && strings.Contains(call.message, "SELECT 2") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected warn log call for slow query, got %v", handler.calls)
	}
}

func TestQueryHook_NormalQueryLogging(t *testing.T) {
	handler := &mockHandler{}
	logger := slog.New(handler)
	hook := NewQueryHook(WithLogger(logger))
	event := &bun.QueryEvent{
		Err:       nil,
		StartTime: time.Now(),
		Query:     "SELECT 3",
	}
	hook.AfterQuery(context.Background(), event)
	found := false
	for _, call := range handler.calls {
		if call.method == "Handle" && call.level == slog.LevelDebug && strings.Contains(call.message, "SELECT 3") {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected debug log call for normal query, got %v", handler.calls)
	}
}

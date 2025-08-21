package bunslog

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/uptrace/bun"
)

type Option func(hook *QueryHook)

// WithEnabled enables/disables this hook
func WithEnabled(on bool) Option {
	return func(h *QueryHook) {
		h.enabled = on
	}
}

// FromEnv configures the hook using the environment variable value.
// For example, WithEnv("BUNDEBUG"):
//   - BUNDEBUG=0 - disables the hook.
//   - BUNDEBUG=1 - enables the hook.
func FromEnv(keys ...string) Option {
	if len(keys) == 0 {
		keys = []string{"BUNDEBUG"}
	}
	return func(h *QueryHook) {
		for _, key := range keys {
			if env, ok := os.LookupEnv(key); ok {
				h.enabled = env != "" && env != "0"
				break
			}
		}
	}
}

func WithErrorLevel(level slog.Level) Option {
	return func(h *QueryHook) {
		h.errorLevel = level
	}
}

func WithQueryLevel(level slog.Level) Option {
	return func(h *QueryHook) {
		h.queryLevel = level
	}
}

func WithSlowLevel(level slog.Level) Option {
	return func(h *QueryHook) {
		h.slowLevel = level
	}
}

func WithLogSlow(duration time.Duration) Option {
	return func(h *QueryHook) {
		h.logSlow = duration
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(h *QueryHook) {
		h.logger = logger
	}
}

// QueryHook wraps query hook
type QueryHook struct {
	enabled bool

	queryLevel slog.Level
	slowLevel  slog.Level
	errorLevel slog.Level

	logSlow time.Duration

	logger *slog.Logger
}

func defaultQueryHook() *QueryHook {
	return &QueryHook{
		enabled:    true,
		logger:     slog.Default(),
		queryLevel: slog.LevelDebug,
		slowLevel:  slog.LevelWarn,
		errorLevel: slog.LevelError,
	}
}

// NewQueryHook returns new instance
func NewQueryHook(options ...Option) *QueryHook {
	h := defaultQueryHook()

	for _, opt := range options {
		opt(h)
	}

	return h
}

// BeforeQuery does nothing
func (h *QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

// AfterQuery convert a bun QueryEvent into a slog message
func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if !h.enabled {
		return
	}

	duration := time.Since(event.StartTime)
	logger := h.logger.With(
		slog.String("operation", event.Operation()),
		slog.Int64("operation_duration_ms", duration.Milliseconds()),
	)

	var level slog.Level

	switch event.Err {
	case nil, sql.ErrNoRows:
		if h.logSlow > 0 && duration >= h.logSlow {
			level = h.slowLevel
		} else {
			level = h.queryLevel
		}
	default:
		level = h.errorLevel
		if err := event.Err; err != nil {
			logger = logger.With(slog.Any("error", err))
		}

	}

	switch level {
	case slog.LevelDebug:
		logger.DebugContext(ctx, event.Query)
	case slog.LevelInfo:
		logger.InfoContext(ctx, event.Query)
	case slog.LevelWarn:
		logger.WarnContext(ctx, event.Query)
	case slog.LevelError:
		h.logger.ErrorContext(ctx, event.Query)
	}
}

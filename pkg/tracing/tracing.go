// Package tracing provides wide-event traces with tail sampling support.
package tracing

import (
	"log/slog"
	"os"
	"sync"
)

var (
	mu      sync.Mutex
	log     *slog.Logger
	sampler Sampler
	enabled bool
)

func init() {
	Configure(DefaultConfig())
}

// Configure applies cfg to the global tracer. It is safe to call concurrently.
func Configure(cfg Config) {
	opts := &slog.HandlerOptions{Level: cfg.Level}

	var h slog.Handler
	if cfg.Format == FormatText {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	s := cfg.Sampler
	if s == nil {
		s = AlwaysSample{}
	}

	mu.Lock()
	defer mu.Unlock()

	enabled = cfg.Enabled
	log = slog.New(h)
	sampler = s
}

// AddBaseAttrs adds base attributes to the trace.
func AddBaseAttrs(attrs ...slog.Attr) {
	mu.Lock()
	defer mu.Unlock()
	log = slog.New(log.Handler().WithAttrs(attrs))
}

// SetSampler sets the sampler for the trace.
func SetSampler(s Sampler) {
	mu.Lock()
	defer mu.Unlock()
	sampler = s
}

// Debug emits a debug-level message.
func Debug(msg string, args ...any) {
	if !enabled {
		return
	}
	log.Debug(msg, args...)
}

// Info emits an info-level message.
func Info(msg string, args ...any) {
	if !enabled {
		return
	}
	log.Info(msg, args...)
}

// Warn emits a warning-level message.
func Warn(msg string, args ...any) {
	if !enabled {
		return
	}
	log.Warn(msg, args...)
}

// Error emits an error-level message.
func Error(msg string, args ...any) {
	if !enabled {
		return
	}
	log.Error(msg, args...)
}

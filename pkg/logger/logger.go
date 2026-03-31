// Package logger provides wide-event traces with tail sampling support.
package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	mu      sync.Mutex
	Logger  *slog.Logger
	enabled bool
)

func init() {
	Configure(DefaultConfig())
}

// Configure applies cfg to the global tracer. It is safe to call concurrently.
func Configure(cfg Config) {
	mu.Lock()
	defer mu.Unlock()

	enabled = cfg.Enabled
	opts := &slog.HandlerOptions{Level: cfg.Level}

	var h slog.Handler
	if cfg.Format == FormatText {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	Logger = slog.New(h)
	slog.SetDefault(Logger)
}

// AddBaseAttrs adds base attributes to the trace.
func AddBaseAttrs(attrs ...slog.Attr) {
	mu.Lock()
	defer mu.Unlock()
	Logger = slog.New(Logger.Handler().WithAttrs(attrs))
	slog.SetDefault(Logger)
}

// SetHandler explicitly overrides the underlying slog.Handler.
// Useful for diverting traces to custom UI rendering engines natively like pkg/cli.
func SetHandler(h slog.Handler) {
	mu.Lock()
	defer mu.Unlock()
	Logger = slog.New(h)
}

// Debug emits a debug-level message.
func Debug(msg string, args ...any) {
	if !enabled {
		return
	}
	Logger.Debug(msg, args...)
}

// Info emits an info-level message.
func Info(msg string, args ...any) {
	if !enabled {
		return
	}
	Logger.Info(msg, args...)
}

// Warn emits a warning-level message.
func Warn(msg string, args ...any) {
	if !enabled {
		return
	}
	Logger.Warn(msg, args...)
}

// Error emits an error-level message.
func Error(msg string, args ...any) {
	if !enabled {
		return
	}
	Logger.Error(msg, args...)
}

// Fatal emits an error-level message and immediately exits the program with status 1.
func Fatal(msg string, args ...any) {
	if enabled {
		Logger.Error(msg, args...)
	}
	os.Exit(1)
}

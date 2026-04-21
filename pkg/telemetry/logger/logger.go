// Package logger provides wide-event structured logging with OpenTelemetry
// trace correlation support.
package logger

import (
	"context"
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

// Configure applies cfg to the global logger. It is safe to call concurrently.
//
// Each entry in cfg.Handlers becomes one slog.Handler. All handlers receive
// every log record via slog.NewMultiHandler.
func Configure(cfg *Config) {
	mu.Lock()
	defer mu.Unlock()

	enabled = cfg.Enabled

	opts := &slog.HandlerOptions{Level: cfg.Level}

	var handlers []slog.Handler
	for _, entry := range cfg.Handlers {
		switch entry.Exporter {
		case ExporterStdout:
			if cfg.Format == FormatText {
				handlers = append(handlers, slog.NewTextHandler(os.Stdout, opts))
			} else {
				handlers = append(handlers, slog.NewJSONHandler(os.Stdout, opts))
			}
		case ExporterOTLP:
			if entry.LoggerProvider != nil {
				handlers = append(handlers, newOTELHandler(entry.LoggerProvider))
			}
		}
	}

	// Fallback to stdout JSON if no handlers specified.
	if len(handlers) == 0 {
		handlers = append(handlers, slog.NewJSONHandler(os.Stdout, opts))
	}

	// Always wrap in traceHandler for correlation.
	Logger = slog.New(newTraceHandler(newMultiHandler(handlers...)))
	slog.SetDefault(Logger)
}

// AddBaseAttrs adds base attributes to every subsequent log record.
func AddBaseAttrs(attrs ...slog.Attr) {
	mu.Lock()
	defer mu.Unlock()
	Logger = slog.New(Logger.Handler().WithAttrs(attrs))
	slog.SetDefault(Logger)
}

// SetHandler explicitly overrides the underlying slog.Handler.
func SetHandler(h slog.Handler) {
	mu.Lock()
	defer mu.Unlock()
	Logger = slog.New(h)
	slog.SetDefault(Logger)
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

// Fatal emits an error-level message and exits the program with status 1.
func Fatal(msg string, args ...any) {
	if enabled {
		Logger.Error(msg, args...)
	}
	os.Exit(1)
}

// Log logs a message with the given level and arguments.
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	Logger.Log(ctx, level, msg, args...)
}

// DebugContext emits a debug-level message, passing ctx for trace correlation.
func DebugContext(ctx context.Context, msg string, args ...any) {
	if !enabled {
		return
	}
	Logger.DebugContext(ctx, msg, args...)
}

// InfoContext emits an info-level message, passing ctx for trace correlation.
func InfoContext(ctx context.Context, msg string, args ...any) {
	if !enabled {
		return
	}
	Logger.InfoContext(ctx, msg, args...)
}

// WarnContext emits a warning-level message, passing ctx for trace correlation.
func WarnContext(ctx context.Context, msg string, args ...any) {
	if !enabled {
		return
	}
	Logger.WarnContext(ctx, msg, args...)
}

// ErrorContext emits an error-level message, passing ctx for trace correlation.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	if !enabled {
		return
	}
	Logger.ErrorContext(ctx, msg, args...)
}

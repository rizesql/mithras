// Package cli provides colored, formatted CLI output for user-facing messages.
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/pflag"

	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

var (
	// Output writer (can be changed for testing)
	Out io.Writer = os.Stdout
	Err io.Writer = os.Stderr

	// Colors
	successColor = color.New(color.FgGreen)
	errorColor   = color.New(color.FgRed)
	warnColor    = color.New(color.FgYellow)
	infoColor    = color.New(color.FgBlue)
	subtleColor  = color.New(color.FgHiBlack)
	headerColor  = color.New(color.FgBlue, color.Bold)
	labelColor   = color.New(color.FgCyan)
	pathColor    = color.New(color.FgWhite)

	// Custom log levels for semantic CLI output.
	// We use high values to distinguish UI messages from internal telemetry.
	LevelSubtle  = slog.Level(20)
	LevelRaw     = slog.Level(21)
	LevelInfo    = slog.Level(22)
	LevelHeader  = slog.Level(23)
	LevelSuccess = slog.Level(24)
	LevelLabel   = slog.Level(25)
)

// Output represents the CLI output handler
type Output struct {
	w       io.Writer
	verbose bool
	noColor bool
}

// New creates a new Output instance
func New() *Output {
	return &Output{
		w:       Out,
		verbose: os.Getenv("VERBOSE") != "" || os.Getenv("DEBUG") != "",
		noColor: os.Getenv("NO_COLOR") != "",
	}
}

// SetWriter sets the output writer
func (o *Output) SetWriter(w io.Writer) {
	o.w = w
}

func (o *Output) Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	f.BoolVarP(&o.verbose, "verbose", "v", o.verbose, "enable verbose output")
	f.BoolVarP(&o.noColor, "no-color", "n", o.noColor, "disable color output")

	return f
}

// IsVerbose returns whether verbose mode is enabled
func (o *Output) IsVerbose() bool {
	return o.verbose
}

// Success prints a success message
func (o *Output) Success(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = successColor.Fprintln(o.w, o.prefix("✓")+" "+msg)
}

// Error prints an error message
func (o *Output) Error(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = errorColor.Fprintln(o.w, o.prefix("✗")+" "+msg)
}

// Warn prints a warning message
func (o *Output) Warn(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = warnColor.Fprintln(o.w, o.prefix("!")+" "+msg)
}

// Info prints an info message
func (o *Output) Info(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = infoColor.Fprintln(o.w, o.prefix("•")+" "+msg)
}

// Header prints a header message
func (o *Output) Header(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = headerColor.Fprintln(o.w, "\n"+msg)
}

// Subtle prints a subtle/secondary message
func (o *Output) Subtle(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = subtleColor.Fprintln(o.w, "  "+msg)
}

// Label prints a labeled value
func (o *Output) Label(label, value string) {
	_, _ = labelColor.Fprintf(o.w, "  %-15s ", label+":")
	_, _ = pathColor.Fprintln(o.w, value)
}

// Verbose prints a message only if verbose mode is enabled
func (o *Output) Verbose(format string, args ...any) {
	if !o.verbose {
		return
	}
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = subtleColor.Fprintln(o.w, "  "+msg)
}

// Block prints an indented block of text
func (o *Output) Block(text string) {
	lines := strings.SplitSeq(strings.TrimSpace(text), "\n")
	for line := range lines {
		_, _ = subtleColor.Fprintln(o.w, "  "+line)
	}
}

// Raw prints raw text without formatting
func (o *Output) Raw(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	_, _ = fmt.Fprintln(o.w, msg)
}

// Fatal prints an error and exits
func (o *Output) Fatal(format string, args ...any) {
	o.Error(format, args...)
	os.Exit(1)
}

// prefix returns a colored prefix symbol
func (o *Output) prefix(symbol string) string {
	if o.noColor {
		return symbol
	}

	switch symbol {
	case "✓":
		return successColor.Sprint(symbol)
	case "✗":
		return errorColor.Sprint(symbol)
	case "!":
		return warnColor.Sprint(symbol)
	case "•":
		return infoColor.Sprint(symbol)
	default:
		return symbol
	}
}

// Configure sets up the global logger for CLI output.
func Configure(enabled bool, level slog.Level) {
	if defaultOutput.noColor {
		color.NoColor = true
	}

	logger.Configure(&logger.Config{
		Enabled:  enabled,
		Level:    level,
		Format:   logger.FormatJSON, // Doesn't matter as we override it
		Handlers: []logger.HandlerEntry{{Exporter: logger.ExporterStdout}},
	})
	logger.SetHandler(NewSlogHandler(defaultOutput, level))
}

// Package-level convenience functions (use default output)

var defaultOutput = New()

// Default returns the default output handler
func Default() *Output {
	return defaultOutput
}

// Success prints a success message using the default output
func Success(format string, args ...any) {
	logger.Log(context.Background(), LevelSuccess, fmt.Sprintf(format, args...))
}

// Error prints an error message using the default output
func Error(format string, args ...any) {
	logger.Log(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

// Warn prints a warning message using the default output
func Warn(format string, args ...any) {
	logger.Log(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Info prints an info message using the default output
func Info(format string, args ...any) {
	logger.Log(context.Background(), LevelInfo, fmt.Sprintf(format, args...))
}

// Header prints a header message using the default output
func Header(format string, args ...any) {
	logger.Log(context.Background(), LevelHeader, fmt.Sprintf(format, args...))
}

// Subtle prints a subtle message using the default output
func Subtle(format string, args ...any) {
	logger.Log(context.Background(), LevelSubtle, fmt.Sprintf(format, args...))
}

// Label prints a labeled value using the default output
func Label(label, value string) {
	logger.Log(context.Background(), LevelLabel, label, "value", value)
}

// Verbose prints a verbose message using the default output
func Verbose(format string, args ...any) {
	logger.Log(context.Background(), slog.LevelDebug, fmt.Sprintf(format, args...))
}

// Block prints a block of text using the default output
func Block(text string) {
	defaultOutput.Block(text)
}

// Raw prints raw text using the default output
func Raw(format string, args ...any) {
	logger.Log(context.Background(), LevelRaw, fmt.Sprintf(format, args...))
}

// Fatal prints an error and exits using the default output
func Fatal(format string, args ...any) {
	Error(format, args...)
	os.Exit(1)
}

// SlogHandler implements slog.Handler to route structured traces seamlessly into the CLI
// Output renderer.
type SlogHandler struct {
	out *Output
}

// NewSlogHandler creates a natively mapped handler wrapping the provided CLI interface.
func NewSlogHandler(out *Output, _ slog.Level) *SlogHandler {
	return &SlogHandler{out: out}
}

// Enabled reports whether the handler handles natively structured records at the given
// architectural severity.
func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	// UI levels (high range) are always enabled.
	if level >= LevelSubtle {
		return true
	}

	// Errors and Warnings are always enabled as they are critical user-facing info.
	if level >= slog.LevelWarn {
		return true
	}

	// For standard telemetry (Info, Debug), require verbose mode.
	if h.out.IsVerbose() {
		return level >= slog.LevelDebug
	}

	return false
}

// Handle rigidly formats and colorizes the slog.Record mapping natively to the UI.
//
//nolint:cyclop
func (h *SlogHandler) Handle(_ context.Context, r slog.Record) error {
	// Format the primary message explicitly based on severity level
	switch r.Level {
	case slog.LevelError:
		h.out.Error("%s", r.Message)
	case slog.LevelWarn:
		h.out.Warn("%s", r.Message)
	case slog.LevelInfo:
		h.out.Info("%s", r.Message)
	case LevelInfo:
		h.out.Info("%s", r.Message)
	case LevelHeader:
		h.out.Header("%s", r.Message)
	case LevelSuccess:
		h.out.Success("%s", r.Message)
	case LevelLabel:
		var val string
		r.Attrs(func(a slog.Attr) bool {
			val = fmt.Sprintf("%v", a.Value.Any())
			return false
		})
		h.out.Label(r.Message, val)
		return nil
	case LevelSubtle:
		h.out.Subtle("%s", r.Message)
	case LevelRaw:
		h.out.Raw("%s", r.Message)
	case slog.LevelDebug:
		h.out.Verbose("%s", r.Message)
	default:
		h.out.Raw("%s", r.Message)
	}

	// If there are attached attributes dynamically attached to the trace, gracefully render
	// them natively beneath
	if r.NumAttrs() > 0 && r.Level != LevelLabel {
		r.Attrs(func(a slog.Attr) bool {
			h.out.Subtle("%s=%v", a.Key, a.Value.Any())
			return true
		})
	}

	return nil
}

// WithAttrs returns a safely copied Handler with structural attributes.
func (h *SlogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// For CLI UI implementations we dynamically ignore persisting base attributes natively
	// to prevent console clutter.
	return h
}

// WithGroup seamlessly returns a statically structured group without crashing string bounds.
func (h *SlogHandler) WithGroup(_ string) slog.Handler {
	return h
}

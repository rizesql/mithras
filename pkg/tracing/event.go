package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rizesql/mithras/pkg/errkit"
)

type eventKey struct{}

// Event accumulates attributes and errors throughout a request lifecycle, emitting a
// single log entry when [End] is called.
type Event struct {
	mu        sync.Mutex
	committed atomic.Bool

	timestamp time.Time
	message   string
	attrs     []slog.Attr
	errors    []error
}

// Attr appends attributes to the event.
func (e *Event) Attr(attrs ...slog.Attr) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.attrs = append(e.attrs, attrs...)
}

// SetErr records an error on the event. Nil errors are ignored, so callers do not need to check.
func (e *Event) SetErr(err error) {
	if err == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.errors = append(e.errors, err)
}

// Start creates a new wide event and stores it in the context.
func Start(ctx context.Context, message string, now time.Time, attrs ...slog.Attr) (context.Context, *Event) {
	evt := &Event{
		mu:        sync.Mutex{},
		committed: atomic.Bool{},

		timestamp: now,
		message:   message,
		attrs:     attrs,
		errors:    []error{},
	}
	return context.WithValue(ctx, eventKey{}, evt), evt
}

// Attr appends attributes to the current event stored in the context.
func Attr(ctx context.Context, attrs ...slog.Attr) {
	evt, ok := ctx.Value(eventKey{}).(*Event)
	if !ok {
		return
	}
	evt.Attr(attrs...)
}

// SetErr records an error on the current event stored in the context. Nil errors are
// ignored, so callers do not need to check.
func SetErr(ctx context.Context, err error) {
	evt, ok := ctx.Value(eventKey{}).(*Event)
	if !ok {
		return
	}
	evt.SetErr(err)
}

// End emits the accumulated event as a log entry if the configured [Sampler] allows it.
func (e *Event) End(now time.Time) {
	if !e.committed.CompareAndSwap(false, true) {
		return
	}

	mu.Lock()
	isEnabled := enabled
	mu.Unlock()

	if !isEnabled {
		return
	}

	e.mu.Lock()
	attrs := e.attrs
	errs := e.errors
	timestamp := e.timestamp
	msg := e.message
	e.mu.Unlock()

	snap := EventSnapshot{
		ErrorCount: len(errs),
		Duration:   now.Sub(timestamp),
	}

	if !sampler.Sample(snap) {
		return
	}

	errors := make([]slog.Attr, 0, len(errs))
	for i, err := range errs {
		code := errkit.GetCode(err)
		errors = append(errors, slog.Group(fmt.Sprintf("%d", i),
			slog.String("code", code.String()),
			slog.String("internal", errkit.GetInternal(err)),
			slog.String("public", errkit.GetPublic(err)),
			slog.Any("steps", errkit.Flatten(err)),
		))
	}

	buf := make([]any, 0, len(attrs)+3)
	buf = append(buf,
		slog.GroupAttrs("errors", errors...),
		slog.GroupAttrs("meta",
			slog.Time("timestamp", timestamp),
			slog.Duration("duration", snap.Duration),
		),
	)
	for _, attr := range attrs {
		buf = append(buf, attr)
	}

	if len(errs) > 0 {
		buf = append(buf, slog.String("outcome", "error"))
		log.Error("error", buf...)
	} else {
		buf = append(buf, slog.String("outcome", "success"))
		log.Info(msg, buf...)
	}
}

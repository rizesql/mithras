package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/rizesql/mithras/internal"
	"github.com/rizesql/mithras/internal/errkit"
)

// Start wraps the global tracer to start a new span.
func Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return internal.Trace.Start(ctx, name, opts...)
}

// Attr adds key-value attributes to the canonical wide event.
func Attr(ctx context.Context, attrs ...attribute.KeyValue) {
	span(ctx).SetAttributes(attrs...)
}

// Event adds a discrete event (log point) to the canonical wide event.
func Event(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span(ctx).AddEvent(name, trace.WithAttributes(attrs...))
}

// Err records an error on the canonical wide event and extracts errkit codes.
func Err(ctx context.Context, err error) {
	if err == nil {
		return
	}

	s := span(ctx)

	code := errkit.GetCode(err)
	if !code.IsZero() {
		s.SetAttributes(attribute.String("mithras.error.code", code.String()))
	}

	s.RecordError(err)
	s.SetStatus(codes.Error, err.Error())
}

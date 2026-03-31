package internal

import "go.opentelemetry.io/otel"

var (
	Trace = otel.Tracer("mithras")
	Meter = otel.Meter("mithras")
)

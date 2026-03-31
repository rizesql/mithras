package telemetry

import (
	"context"
	"os"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"

	"github.com/rizesql/mithras/internal"
)

func mithrasResource() (*resource.Resource, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("mithras"),
		attribute.String("id", internal.Identifier),
		attribute.String("version", internal.Version),
		attribute.String("host", host),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
	), nil
}

func noopShutdown() func(context.Context) error {
	return func(context.Context) error { return nil }
}

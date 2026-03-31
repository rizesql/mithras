package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"

	"github.com/rizesql/mithras/pkg/logger"
)

func ConfigureTracing(ctx context.Context, cfg *Tracing) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return noopShutdown(), nil
	}

	logger.Info("tracing.enabled",
		slog.String("exporter", cfg.Exporter),
		slog.String("protocol", string(cfg.Protocol)),
		slog.String("endpoint", cfg.Endpoint),
	)

	switch cfg.Exporter {
	case "otlp":
		exp, err := otlpTracingExporter(ctx, cfg)
		if err != nil {
			return noopShutdown(), err
		}
		return newTracer(exp), nil
	default:
		return noopShutdown(), fmt.Errorf("%s tracing exporter is unsupported", cfg.Exporter)
	}
}

func newTracer(exporter trace.SpanExporter) func(context.Context) error {
	res, err := mithrasResource()
	if err != nil {
		return noopShutdown()
	}

	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(
			trace.NewBatchSpanProcessor(exporter),
		),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown
}

func otlpTracingExporter(ctx context.Context, cfg *Tracing) (trace.SpanExporter, error) {
	switch cfg.Protocol {
	case ProtocolHTTP:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.Endpoint),
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
		}
		if cfg.URLPath != "" {
			opts = append(opts, otlptracehttp.WithURLPath(cfg.URLPath))
		}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		exp, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, err
		}
		return exp, nil

	case ProtocolGRPC:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		} else {
			opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
		}

		exp, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, err
		}
		return exp, nil

	default:
		return nil, fmt.Errorf("%s protocol is unsupported", cfg.Protocol)
	}
}

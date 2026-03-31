package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/host"
	orn "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc/credentials"

	"github.com/rizesql/mithras/pkg/logger"
)

func ConfigureMetrics(ctx context.Context, cfg *Metrics) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return noopShutdown(), nil
	}

	logger.Info("metrics.enabled",
		slog.String("exporter", cfg.Exporter),
		slog.String("protocol", string(cfg.Protocol)),
		slog.String("endpoint", cfg.Endpoint),
		slog.Duration("interval", cfg.Interval),
	)

	switch cfg.Exporter {
	case "otlp":
		exp, err := otlpMetricsExporter(ctx, cfg)
		if err != nil {
			return noopShutdown(), err
		}
		return newMeter(exp, cfg.Interval), nil
	default:
		return noopShutdown(), fmt.Errorf("%s metrics exporter is unsupported", cfg.Exporter)
	}
}

func newMeter(exporter metric.Exporter, interval time.Duration) func(context.Context) error {
	res, err := mithrasResource()
	if err != nil {
		return noopShutdown()
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(exporter, metric.WithInterval(interval)),
		),
		metric.WithResource(res),
	)

	if err = orn.Start(
		orn.WithMinimumReadMemStatsInterval(time.Second),
		orn.WithMeterProvider(mp),
	); err != nil {
		return noopShutdown()
	}

	if err = host.Start(host.WithMeterProvider(mp)); err != nil {
		return noopShutdown()
	}

	otel.SetMeterProvider(mp)
	return mp.Shutdown
}

func otlpMetricsExporter(ctx context.Context, cfg *Metrics) (metric.Exporter, error) {
	switch cfg.Protocol {
	case ProtocolHTTP:
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
			otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Headers))
		}
		if cfg.URLPath != "" {
			opts = append(opts, otlpmetrichttp.WithURLPath(cfg.URLPath))
		}
		if cfg.Insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}

		exp, err := otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return nil, err
		}
		return exp, nil

	case ProtocolGRPC:
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		}
		if len(cfg.Headers) > 0 {
			opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.Headers))
		}
		if cfg.Insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		} else {
			opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
		}

		exp, err := otlpmetricgrpc.New(ctx, opts...)
		if err != nil {
			return nil, err
		}
		return exp, nil
	default:
		return nil, fmt.Errorf("%s protocol is unsupported", cfg.Protocol)
	}
}

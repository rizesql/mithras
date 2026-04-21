package ratelimit

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/rizesql/mithras/pkg/telemetry"
	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

type Store interface {
	Check(ctx context.Context, req *Request) (*Response, error)
	Reset(ctx context.Context, key string) error
}

func New(ctx context.Context, cfg *Config) (Store, error) {
	logger.Info("ratelimit.starting")

	switch cfg.Type {
	case TypeRedis:
		return newRedis(ctx, &cfg.Redis)
	case TypeMemory:
		return nil, fmt.Errorf("ratelimit: memory store not implemented")
	case TypeNoop:
		return nil, fmt.Errorf("ratelimit: noop store not implemented")
	default:
		return nil, fmt.Errorf("ratelimit: unknown type %q", cfg.Type)
	}
}

type Request struct {
	// Policy name (for logging / namespacing)
	Name string

	// Unique subject (ip:1.2.3.4, user:123, etc.)
	Identifier string

	// Max tokens generated over Duration
	Limit int64

	// Refill period for the token bucket
	Duration time.Duration

	// Bucket capacity (burst size)
	Burst int64

	// Optional time override (mainly for tests)
	Time time.Time
}

func (r *Request) Key() string { return "rl:" + r.Name + ":" + r.Identifier }

type Response struct {
	Limit int64

	// Remaining tokens in the bucket
	Remaining int64

	// Time when at least 1 token will be available again
	Reset time.Time

	// Whether request is allowed
	Success bool

	// Tokens currently consumed (derived)
	Used int64
}

type telemetryStore struct {
	inner Store
	name  string
}

func (s *telemetryStore) Check(ctx context.Context, req *Request) (res *Response, err error) {
	ctx, span := telemetry.Start(ctx, "ratelimit.check", trace.WithAttributes(
		attribute.String("ratelimit.policy", s.name),
		attribute.String("ratelimit.identifier", req.Identifier),
		attribute.Int64("ratelimit.limit", req.Limit),
		attribute.String("ratelimit.duration", req.Duration.String()),
	))
	defer telemetry.End(span, &err)

	res, err = s.inner.Check(ctx, req)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.Bool("ratelimit.success", res.Success),
		attribute.Int64("ratelimit.remaining", res.Remaining),
	)

	if !res.Success {
		span.SetStatus(codes.Error, "rate limit exceeded")
		if !res.Reset.IsZero() {
			span.SetAttributes(attribute.String("ratelimit.reset", res.Reset.Format(time.RFC3339)))
		}
	}

	return res, err
}

func (s *telemetryStore) Reset(ctx context.Context, key string) (err error) {
	ctx, span := telemetry.Start(ctx, "ratelimit.reset")
	defer telemetry.End(span, &err)

	span.SetAttributes(attribute.String("ratelimit.key", key))

	return s.inner.Reset(ctx, key)
}

func WithTelemetry(store Store, name string) Store {
	return &telemetryStore{inner: store, name: name}
}

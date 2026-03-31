package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/rizesql/mithras/pkg/logger"
	"github.com/rizesql/mithras/pkg/telemetry"
)

var allow = redis.NewScript(`
local time = redis.call("TIME")
local now = tonumber(time[1]) * 1000 + math.floor(tonumber(time[2]) / 1000)

local key = KEYS[1]
local rate = tonumber(ARGV[1])      -- tokens per millisecond
local capacity = tonumber(ARGV[2])  -- burst

local data = redis.call("HMGET", key, "tokens", "ts")

local tokens = tonumber(data[1])
local ts = tonumber(data[2])

if tokens == nil then
    tokens = capacity
    ts = now
end

local delta = math.max(0, now - ts)
local refill = delta * rate
tokens = math.min(capacity, tokens + refill)

local allowed = 0
local retry_after = 0

if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
else
    allowed = 0
    retry_after = math.ceil((1 - tokens) / rate)
end

redis.call("HSET", key, "tokens", tokens, "ts", now)

local ttl = math.ceil(capacity / rate)
redis.call("PEXPIRE", key, ttl)

local remaining = math.floor(tokens)
return {allowed, retry_after, remaining}
`)

type RedisStore struct {
	rdb *redis.Client
}

func newRedis(ctx context.Context, cfg *RedisConfig) (*RedisStore, error) {
	logger.Info("ratelimit.redis.starting")

	if cfg.URL == "" {
		return nil, fmt.Errorf("ratelimit: redis: empty URL")
	}

	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("ratelimit: redis: invalid URL: %w", err)
	}

	opts.DialTimeout = 1 * time.Second
	opts.ReadTimeout = 500 * time.Millisecond
	opts.WriteTimeout = 500 * time.Millisecond

	rdb := redis.NewClient(opts)
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		return nil, fmt.Errorf("ratelimit: failed to instrument redis tracing: %w", err)
	}
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		return nil, fmt.Errorf("ratelimit: failed to instrument redis metrics: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger.Info("ratelimit.redis.ping")
	if _, err := rdb.Ping(pingCtx).Result(); err != nil {
		logger.Warn("redis ping failed at startup, will reconnect lazily", "error", err)
	}

	logger.Info("ratelimit.redis.connected")
	return &RedisStore{rdb: rdb}, nil
}

var _ Store = (*RedisStore)(nil)

func (s *RedisStore) Check(ctx context.Context, req *Request) (*Response, error) {
	if req.Identifier == "" {
		return nil, fmt.Errorf("ratelimit: empty identifier")
	}
	if req.Limit <= 0 || req.Duration < time.Millisecond {
		return nil, fmt.Errorf("ratelimit: invalid limit/duration")
	}

	burst := req.Burst
	if burst <= 0 {
		burst = req.Limit
	}

	now := req.Time
	if now.IsZero() {
		now = time.Now()
	}

	rate := float64(req.Limit) / float64(req.Duration.Milliseconds())

	res, err := allow.Run(ctx, s.rdb,
		[]string{req.Key()},
		rate,
		burst,
	).Result()
	if err != nil {
		return nil, fmt.Errorf("ratelimit: redis: %w", err)
	}

	arr := res.([]any)

	var reset time.Time
	allowed := arr[0].(int64) == 1
	retryMs := arr[1].(int64)
	remaining := arr[2].(int64)

	if !allowed {
		reset = now.Add(time.Duration(retryMs) * time.Millisecond)
	}

	return &Response{
		Limit:     req.Limit,
		Remaining: max(remaining, 0),
		Reset:     reset,
		Success:   allowed,
		Used:      req.Limit - remaining,
	}, nil
}

func (s *RedisStore) Reset(ctx context.Context, key string) error {
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("ratelimit: redis reset: %w", err)
	}
	return nil
}

type telemetryStore struct {
	inner Store
	name  string
}

func (s *telemetryStore) Check(ctx context.Context, req *Request) (*Response, error) {
	spanCtx, span := telemetry.Start(ctx, "ratelimit.check", trace.WithAttributes(
		attribute.String("ratelimit.policy", s.name),
		attribute.Int64("ratelimit.limit", req.Limit),
	))
	defer span.End()

	res, err := s.inner.Check(ctx, req)
	if err != nil {
		telemetry.Err(spanCtx, err)
	} else {
		telemetry.Attr(spanCtx,
			attribute.Bool("ratelimit.success", res.Success),
			attribute.Int64("ratelimit.remaining", res.Remaining),
		)
	}

	return res, err
}

func (s *telemetryStore) Reset(ctx context.Context, key string) error {
	spanCtx, span := telemetry.Start(ctx, "ratelimit.reset")
	defer span.End()

	telemetry.Attr(spanCtx, attribute.String("ratelimit.key", key))

	err := s.inner.Reset(spanCtx, key)
	if err != nil {
		telemetry.Err(spanCtx, err)
	}
	return err
}

func WithTelemetry(store Store, name string) *telemetryStore {
	return &telemetryStore{inner: store, name: name}
}

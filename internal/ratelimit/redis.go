package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"

	"github.com/rizesql/mithras/pkg/telemetry/logger"
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
		return nil, errors.New("ratelimit: redis: empty URL")
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
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	now := s.getCurrentTime(req)
	rate, burst := s.calculateRateAndBurst(req)

	res, err := allow.Run(ctx, s.rdb,
		[]string{req.Key()},
		rate,
		burst,
	).Result()
	if err != nil {
		return nil, fmt.Errorf("ratelimit: redis: %w", err)
	}

	return s.parseResponse(res, req.Limit, now)
}

func (s *RedisStore) validateRequest(req *Request) error {
	if req.Identifier == "" {
		return errors.New("ratelimit: empty identifier")
	}

	if req.Limit <= 0 || req.Duration < time.Millisecond {
		return errors.New("ratelimit: invalid limit/duration")
	}

	return nil
}

func (s *RedisStore) getCurrentTime(req *Request) time.Time {
	if req.Time.IsZero() {
		return time.Now()
	}

	return req.Time
}

func (s *RedisStore) calculateRateAndBurst(req *Request) (float64, int64) {
	burst := req.Burst
	if burst <= 0 {
		burst = req.Limit
	}

	rate := float64(req.Limit) / float64(req.Duration.Milliseconds())

	return rate, burst
}

func (s *RedisStore) parseResponse(res any, limit int64, now time.Time) (*Response, error) {
	arr, ok := res.([]any)
	if !ok || len(arr) < 3 {
		return nil, fmt.Errorf("ratelimit: redis: unexpected result type or length: %T", res)
	}

	allowed, ok1 := arr[0].(int64)
	retryMs, ok2 := arr[1].(int64)
	remaining, ok3 := arr[2].(int64)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("ratelimit: redis: unexpected result element types")
	}

	var reset time.Time
	isAllowed := allowed == 1

	if !isAllowed {
		reset = now.Add(time.Duration(retryMs) * time.Millisecond)
	}

	return &Response{
		Limit:     limit,
		Remaining: max(remaining, 0),
		Reset:     reset,
		Success:   isAllowed,
		Used:      limit - remaining,
	}, nil
}

func (s *RedisStore) Reset(ctx context.Context, key string) error {
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("ratelimit: redis reset: %w", err)
	}

	return nil
}

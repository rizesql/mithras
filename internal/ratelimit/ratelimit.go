package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/rizesql/mithras/pkg/logger"
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

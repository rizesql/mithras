package ratelimit

import (
	"time"
)

type Policy struct {
	// Name identifies this policy in logs, audit events, and error codes.
	Name string

	// MaxRequests is the number of requests allowed per Window.
	MaxRequests uint64

	// Burst controls token-bucket burst behavior:
	//   1              — strict, no burst (use for per-account / per-token policies)
	//   == MaxRequests — full burst allowed (use for per-IP policies)
	Burst int

	// Window is the duration over which MaxRequests are counted.
	Window time.Duration

	// KeyFunc extracts the rate limit key from the request. If it returns an
	// empty string the policy is skipped for that request.
	KeyFunc KeyFunc

	// Store is the backend counter. If nil, the middleware panics at
	// construction time — always provide a store.
	Store Store

	// FailOpen controls behavior when Store returns an error:
	//   true  — the request is allowed through (availability over security)
	//   false — the request is rejected with 503 (security over availability)
	//
	// Sensitive auth policies (login-per-account, reset-password-per-token,
	// refresh-per-token) should use FailOpen: false.
	FailOpen bool
}

// PolicyOption is a functional option for configuring a Policy.
type PolicyOption func(*Policy)

// WithStore explicitly overrides the token bucket backend.
func WithStore(store Store) PolicyOption {
	return func(p *Policy) {
		p.Store = store
	}
}

// WithBurst enables token accumulation up to the maximum request limit.
// This is the recommended "pit of success" override for basic per-IP policies to
// accommodate NAT spikes.
func WithBurst() PolicyOption {
	// #nosec G115
	return func(p *Policy) {
		p.Burst = int(p.MaxRequests)
	}
}

// WithFailOpen ensures the policy allows requests through if the store errors.
func WithFailOpen() PolicyOption {
	return func(p *Policy) {
		p.FailOpen = true
	}
}

// NewPolicy creates a new rate limit policy with secure "pit of success" defaults:
// - Strict Burst (1) prevents payload spikes. Use WithFullBurst() for IP-based policies.
// - FailClosed (FailOpen = false) prioritizes security over availability.
// Use WithFailOpen(true) for public endpoints.
func NewPolicy(
	name string,
	maxRequests uint64,
	window time.Duration,
	keyFunc KeyFunc,
	opts ...PolicyOption,
) Policy {
	p := Policy{
		Name:        name,
		MaxRequests: maxRequests,
		Window:      window,
		KeyFunc:     keyFunc,
		Burst:       1,
		FailOpen:    false,
	}

	for _, opt := range opts {
		opt(&p)
	}

	return p
}

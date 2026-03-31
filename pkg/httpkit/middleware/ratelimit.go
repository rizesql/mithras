package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/internal/errkit"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/logger"
	"github.com/rizesql/mithras/pkg/telemetry"
)

func WithRateLimit(policies ...ratelimit.Policy) httpkit.Middleware {
	for _, p := range policies {
		if p.Store == nil {
			panic(fmt.Sprintf("ratelimit: policy %q has a nil Store", p.Name))
		}
	}

	wrapped := make([]ratelimit.Policy, len(policies))
	for i, p := range policies {
		p.Store = ratelimit.WithTelemetry(p.Store, p.Name)
		wrapped[i] = p
	}

	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, c *httpkit.Context) error {
			var mostRestrictive *ratelimit.Response

			for _, p := range wrapped {
				key := p.KeyFunc(ctx, c)
				if key == "" {
					logger.Warn("ratelimit.key_empty", "policy", p.Name)
					continue
				}

				// #nosec G115
				res, err := p.Store.Check(ctx, &ratelimit.Request{
					Name:       p.Name,
					Identifier: key,
					Limit:      int64(p.MaxRequests),
					Duration:   p.Window,
					Burst:      int64(p.Burst),
				})

				if err != nil {
					if p.FailOpen {
						logger.Warn("ratelimit.fail_open",
							"policy", p.Name,
							"error", err,
						)
						continue
					}

					return errkit.Wrap(err,
						errkit.Code(errkit.App.Internal.Code("rate_limit_store_error")),
						errkit.Internal(fmt.Sprintf("rate limit store error for policy %q", p.Name)),
						errkit.Public("Service temporarily unavailable. Please try again shortly."),
					)
				}

				if mostRestrictive == nil || res.Remaining < mostRestrictive.Remaining {
					mostRestrictive = res
				}

				if !res.Success {
					telemetry.Attr(ctx,
						attribute.Bool("http.ratelimit.throttled", true),
						attribute.String("http.ratelimit.policy", p.Name),
					)

					setHeaders(c, res)

					return errkit.New("",
						errkit.Code(errkit.User.RateLimit.Code(p.Name)),
						errkit.Internal(fmt.Sprintf("rate limit exceeded for policy %q", p.Name)),
						errkit.Public("Too many requests. Please slow down and try again later."),
					)
				}
			}

			if mostRestrictive != nil {
				setHeaders(c, mostRestrictive)
			}

			return next(ctx, c)
		}
	}
}

func setHeaders(c *httpkit.Context, r *ratelimit.Response) {
	c.Res().SetHeader("X-RateLimit-Limit", strconv.FormatInt(r.Limit, 10))
	c.Res().SetHeader("X-RateLimit-Remaining", strconv.FormatInt(r.Remaining, 10))

	if !r.Reset.IsZero() {
		secs := max(int(time.Until(r.Reset).Seconds()), 1)
		c.Res().SetHeader("Retry-After", strconv.Itoa(secs))
		c.Res().SetHeader("X-RateLimit-Reset", strconv.FormatInt(r.Reset.Unix(), 10))
	}
}

// RetryAfterSeconds is a helper for tests that parses the Retry-After header.
func RetryAfterSeconds(header http.Header) int {
	v := header.Get("Retry-After")
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

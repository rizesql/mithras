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
	"github.com/rizesql/mithras/pkg/telemetry"
	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

func registerPolicies(policies ...ratelimit.Policy) []ratelimit.Policy {
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

	return wrapped
}

func WithRateLimit(policies ...ratelimit.Policy) httpkit.Middleware {
	wrapped := registerPolicies(policies...)

	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, cx *httpkit.Context) error {
			var mostRestrictive *ratelimit.Response
			var restrictivePolicy string

			for _, pol := range wrapped {
				key := pol.KeyFunc(ctx, cx)
				if key == "" {
					logger.Warn("ratelimit.key_empty", "policy", pol.Name)
					continue
				}

				// #nosec G115
				res, err := pol.Store.Check(ctx, &ratelimit.Request{
					Name:       pol.Name,
					Identifier: key,
					Limit:      int64(pol.MaxRequests),
					Duration:   pol.Window,
					Burst:      int64(pol.Burst),
				})
				if err != nil {
					if pol.FailOpen {
						logger.Warn("ratelimit.fail_open",
							"policy", pol.Name,
							"error", err,
						)

						continue
					}

					return errkit.Wrap(err,
						errkit.App.Internal.Code("rate_limit_store_error"),
						errkit.Internal(fmt.Sprintf("rate limit store error for policy %q", pol.Name)),
						errkit.Public("Service temporarily unavailable. Please try again shortly."),
					)
				}

				if mostRestrictive == nil || res.Remaining < mostRestrictive.Remaining {
					mostRestrictive = res
					restrictivePolicy = pol.Name
				}

				if !res.Success {
					telemetry.Attr(ctx,
						attribute.Bool("ratelimit.throttled", true),
						attribute.String("ratelimit.policy", pol.Name),
					)

					setHeaders(cx, res)
					return errkit.New("",
						errkit.User.RateLimit.Code(pol.Name),
						errkit.Internal(fmt.Sprintf("rate limit exceeded for policy %q", pol.Name)),
						errkit.Public("Too many requests. Please slow down and try again later."),
					)
				}
			}

			if mostRestrictive != nil {
				setHeaders(cx, mostRestrictive)
				telemetry.Attr(ctx,
					attribute.String("ratelimit.policy", restrictivePolicy),
					attribute.Int64("ratelimit.remaining", mostRestrictive.Remaining),
				)
			}

			return next(ctx, cx)
		}
	}
}

func setHeaders(cx *httpkit.Context, res *ratelimit.Response) {
	cx.Res().SetHeader("X-RateLimit-Limit", strconv.FormatInt(res.Limit, 10))
	cx.Res().SetHeader("X-RateLimit-Remaining", strconv.FormatInt(res.Remaining, 10))

	if !res.Reset.IsZero() {
		secs := max(int(time.Until(res.Reset).Seconds()), 1)
		cx.Res().SetHeader("Retry-After", strconv.Itoa(secs))
		cx.Res().SetHeader("X-RateLimit-Reset", strconv.FormatInt(res.Reset.Unix(), 10))
	}
}

// RetryAfterSeconds is a helper for tests that parses the Retry-After header.
func RetryAfterSeconds(header http.Header) int {
	v := header.Get("Retry-After")
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}

	return n
}

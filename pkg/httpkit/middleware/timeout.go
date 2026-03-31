package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/rizesql/mithras/internal/errkit"
	"github.com/rizesql/mithras/pkg/httpkit"
)

const (
	// DefaultRequestTimeout is the default timeout for requests.
	DefaultRequestTimeout = 30 * time.Second
)

// WithTimeout returns a middleware that sets a timeout for the request.
func WithTimeout(timeout time.Duration) httpkit.Middleware {
	if timeout <= 0 {
		timeout = DefaultRequestTimeout
	}

	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, c *httpkit.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			err := next(timeoutCtx, c)
			if err == nil {
				return nil
			}

			if errors.Is(err, context.Canceled) {
				return errkit.Wrap(err,
					errkit.Code(errkit.User.Request.Code("client_closed_request")),
					errkit.Internal("The client closed the connection before the request completed"),
					errkit.Public("Client closed request"),
				)
			}

			if errors.Is(err, context.DeadlineExceeded) {
				return errkit.Wrap(err,
					errkit.Code(errkit.System.Timeout.Code("request_timeout")),
					errkit.Internal("The request exceeded the maximum processing time"),
					errkit.Public("Request timeout"),
				)
			}

			return err
		}
	}
}

package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/internal/errkit"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

// WithPanicRecovery returns a middleware that recovers from panics and returns a 500 response.
func WithPanicRecovery() httpkit.Middleware {
	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, cx *httpkit.Context) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					if errors.Is(err, http.ErrAbortHandler) {
						panic(rec)
					}

					stack := debug.Stack()
					telemetry.Attr(ctx, attribute.String("panic_stack", string(stack)))

					err = errkit.Wrap(
						fmt.Errorf("panic: %v", rec),
						errkit.App.Internal.Code("panic"),
						errkit.Internal("an unhandled panic occurred during request processing"),
						errkit.Public("An unexpected error occurred while processing your request."),
					)
				}
			}()

			return next(ctx, cx)
		}
	}
}

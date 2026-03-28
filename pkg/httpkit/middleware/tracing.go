package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/tracing"
)

// WithTracing returns middleware that emits HTTP request details. It records method, path,
// status code, request ID, and errors.
func WithTracing(clk clock.Clock) httpkit.Middleware {
	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, c *httpkit.Context) error {
			req := c.Req().Raw()

			uri := req.URL.RequestURI()
			if uri == "" {
				uri = req.URL.Path
			}

			ctx, evt := tracing.Start(ctx, req.Method+" "+uri, c.Req().Timestamp())
			defer func() { evt.End(clk.Now()) }()

			nextErr := next(ctx, c)
			evt.SetErr(nextErr)

			statusCode := c.Res().StatusCode()
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			evt.Attr(slog.Group("http",
				slog.String("request_id", c.Req().ID()),
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("query_params", req.URL.RawQuery),
				slog.Int("status_code", statusCode),
				slog.String("host", req.Host),
				slog.String("user_agent", req.UserAgent()),
				slog.String("ip_address", c.Req().IP()),
			))

			return nextErr
		}
	}
}

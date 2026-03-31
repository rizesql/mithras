package middleware

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

func WithTelemetry() httpkit.Middleware {
	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, c *httpkit.Context) error {
			req := c.Req().Raw()

			ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))

			uri := req.URL.RequestURI()
			if uri == "" {
				uri = req.URL.Path
			}

			spanCtx, span := telemetry.Start(ctx, req.Method+" "+uri)
			wideCtx := telemetry.InjectMainSpan(spanCtx, span)
			start := c.Req().Timestamp()

			var nextErr error
			defer func() {
				statusCode := c.Res().StatusCode()
				if statusCode == 0 {
					statusCode = http.StatusOK
				}

				outcome := "success"
				if nextErr != nil {
					outcome = "error"
					telemetry.Err(wideCtx, nextErr)
				}

				telemetry.Attr(wideCtx,
					attribute.String("http.outcome", outcome),
					attribute.String("http.request.id", c.Req().ID()),
					semconv.HTTPRequestMethodKey.String(req.Method),
					semconv.URLPathKey.String(req.URL.Path),
					semconv.URLQueryKey.String(req.URL.RawQuery),
					semconv.HTTPResponseStatusCodeKey.Int(statusCode),
					semconv.ServerAddressKey.String(req.Host),
					semconv.UserAgentOriginalKey.String(req.UserAgent()),
					semconv.ClientAddressKey.String(c.Req().IP()),
					semconv.HTTPRequestBodySize(int(req.ContentLength)),
					attribute.Int64("http.duration_ms", time.Since(start).Milliseconds()),
				)

				if nextErr != nil {
					telemetry.Err(wideCtx, nextErr)
				} else {

					span.SetStatus(codes.Ok, outcome)
				}

				span.End()
			}()

			nextErr = next(wideCtx, c)
			return nextErr
		}
	}
}

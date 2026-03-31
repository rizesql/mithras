package middleware

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

// WithValidation returns a middleware that validates the request using the given validator.
func WithValidation(v *validator.Validator) httpkit.Middleware {
	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, c *httpkit.Context) error {
			if err := v.Validate(ctx, c.Req().Raw()); err != nil {
				telemetry.Attr(ctx, attribute.Bool("validation.failed", true))
				if valErrs, ok := err.(validator.ValidationErrors); ok {
					var b strings.Builder
					for _, e := range valErrs {
						b.WriteString(e.Path)
						b.WriteByte(',')
					}

					telemetry.Attr(ctx, attribute.String("validation.failed_fields", b.String()))
				}

				return err
			}

			return next(ctx, c)
		}
	}
}

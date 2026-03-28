package middleware

import (
	"context"

	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/httpkit"
)

// WithValidation returns a middleware that validates the request using the given validator.
func WithValidation(v *validator.Validator) httpkit.Middleware {
	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, c *httpkit.Context) error {
			if err := v.Validate(ctx, c.Req().Raw()); err != nil {
				return err
			}

			return next(ctx, c)
		}
	}
}

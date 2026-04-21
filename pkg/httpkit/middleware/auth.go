package middleware

import (
	"context"
	"net/http"

	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/pkg/httpkit"
)

func WithAuth(verifier *auth.Verifier) httpkit.Middleware {
	return func(next httpkit.HandleFunc) httpkit.HandleFunc {
		return func(ctx context.Context, c *httpkit.Context) error {
			cookie, err := c.Req().Raw().Cookie(auth.AuthSessionCookie)
			if err != nil {
				c.Redirect("/login", http.StatusSeeOther)
				return nil
			}

			user, err := verifier.Verify(ctx, cookie.Value)
			if err != nil {
				c.Redirect("/login", http.StatusSeeOther)
				return nil
			}

			c.SetUser(user)
			return next(ctx, c)
		}
	}
}

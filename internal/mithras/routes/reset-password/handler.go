package resetpassword

import (
	"context"
	"net/http"
	"time"

	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
)

type Req api.ResetPasswordRequest

type handler struct {
	pr *auth.PasswordReset
}

func New(p *platform.Platform) *handler {
	return &handler{pr: p.PasswordReset}
}

func (h *handler) Method() string { return http.MethodPost }
func (h *handler) Path() string   { return "/reset-password" }

func (h *handler) Handle(ctx context.Context, c *httpkit.Context) error {
	req, err := httpkit.BindBody[Req](c)
	if err != nil {
		return err
	}

	if err := h.pr.Reset(ctx, req.Token, req.Password); err != nil {
		return err
	}

	return c.Res().JSON(http.StatusOK, nil)
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("reset-password-per-ip",
			5, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
			ratelimit.WithBurst(),
		),

		ratelimit.NewPolicy("reset-password-per-token",
			3, time.Minute,
			ratelimit.KeyBearerToken(),
			ratelimit.WithStore(p.RateLimit),
		),
	)
}

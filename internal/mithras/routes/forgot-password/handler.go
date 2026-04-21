package forgotpassword

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
)

type Req api.ForgotPasswordRequest

type handler struct {
	pr *auth.PasswordReset
}

func New(p *platform.Platform) *handler {
	return &handler{pr: p.PasswordReset}
}

func (h *handler) Method() string { return http.MethodPost }
func (h *handler) Path() string   { return "/forgot-password" }

func (h *handler) Handle(ctx context.Context, c *httpkit.Context) error {
	req, err := httpkit.BindBody[Req](c)
	if err != nil {
		return err
	}

	if err := h.pr.Request(ctx,
		string(req.Email),
		c.Req().Raw().UserAgent(),
		c.Req().IP(),
	); err != nil {
		return err
	}

	return c.Res().JSON(http.StatusAccepted, nil)
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("forgot-password-per-ip",
			3, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
		),

		ratelimit.NewPolicy("forgot-password-per-account",
			3, time.Hour,
			ratelimit.KeyBodyValue("email", strings.ToLower),
			ratelimit.WithStore(p.RateLimit),
		),
	)
}

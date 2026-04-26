package logout

import (
	"context"
	"net/http"
	"time"

	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
)

type handler struct {
	logout *auth.Logout
}

func New(p *platform.Platform) *handler {
	return &handler{
		logout: auth.NewLogout(p.DB, p.Clock),
	}
}

func (h *handler) Method() string { return http.MethodPost }
func (h *handler) Path() string   { return "/logout" }

func (h *handler) Handle(ctx context.Context, c *httpkit.Context) error {
	tok, err := httpkit.BearerAuth(c)
	if err != nil {
		return err
	}

	if err := h.logout.Logout(ctx, tok); err != nil {
		return err
	}

	return c.Res().Send(http.StatusNoContent, []byte{})
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("logout-per-ip",
			100, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
			ratelimit.WithBurst(),
		),
	)
}

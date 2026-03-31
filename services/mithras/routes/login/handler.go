package login

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
	"github.com/rizesql/mithras/services/mithras/platform"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Method() string { return http.MethodPost }
func (h *handler) Path() string   { return "/login" }

func (h *handler) Handle(_ context.Context, _ *httpkit.Context) error {
	return nil
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("login-per-ip",
			20, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
			ratelimit.WithBurst(),
		),

		ratelimit.NewPolicy("login-per-account",
			10, time.Minute,
			ratelimit.KeyBodyValue("email", strings.ToLower),
			ratelimit.WithStore(p.RateLimit),
		),
	)
}

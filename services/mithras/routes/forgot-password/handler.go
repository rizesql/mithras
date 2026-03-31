package forgotpassword

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
func (h *handler) Path() string   { return "/forgot-password" }

func (h *handler) Handle(_ context.Context, _ *httpkit.Context) error {
	return nil
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

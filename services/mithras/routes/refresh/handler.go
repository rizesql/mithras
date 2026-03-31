package refresh

import (
	"context"
	"net/http"
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
func (h *handler) Path() string   { return "/refresh" }

func (h *handler) Handle(_ context.Context, _ *httpkit.Context) error {
	return nil
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("refresh-per-token",
			6, time.Hour,
			ratelimit.KeyBearerToken(),
			ratelimit.WithStore(p.RateLimit),
		),

		ratelimit.NewPolicy("refresh-per-ip",
			30, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
			ratelimit.WithBurst(),
		),
	)
}

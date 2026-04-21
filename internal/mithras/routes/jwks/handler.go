package jwks

import (
	"context"
	"net/http"
	"time"

	"github.com/rizesql/mithras/internal/jws"
	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
)

type handler struct {
	jws jws.Store
}

func New(p *platform.Platform) *handler {
	return &handler{jws: p.JWS}
}

func (h *handler) Method() string { return http.MethodGet }
func (h *handler) Path() string   { return "/.well-known/jwks.json" }

func (h *handler) Handle(ctx context.Context, c *httpkit.Context) error {
	keys, err := h.jws.PublicKeys(ctx)
	if err != nil {
		return err
	}

	jwks, err := jws.JWKS(keys)
	if err != nil {
		return err
	}

	return c.Res().JSON(http.StatusOK, jwks)
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("jwks-per-ip",
			1000, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
			ratelimit.WithBurst(),
			ratelimit.WithFailOpen(),
		),
	)
}

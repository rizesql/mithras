package oas

import (
	"context"
	"net/http"
	"time"

	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
)

type Res api.OAuthServerMetadata

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Method() string { return http.MethodGet }
func (h *handler) Path() string   { return "/.well-known/oauth-authorization-server" }

func (h *handler) Handle(_ context.Context, c *httpkit.Context) error {
	req := c.Req().Raw()
	scheme := "https"
	if req.TLS == nil && req.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	issuer := scheme + "://" + req.Host

	return c.Res().JSON(http.StatusOK, Res{
		Issuer:                        issuer,
		AuthorizationEndpoint:         issuer + "/authorize",
		TokenEndpoint:                 issuer + "/token",
		RegistrationEndpoint:          issuer + "/register",
		JwksUri:                       issuer + "/.well-known/jwks.json",
		ResponseTypesSupported:        []string{"code"},
		GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported: []string{"S256"},
	})
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("oas-per-ip",
			1000, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
			ratelimit.WithBurst(),
			ratelimit.WithFailOpen(),
		),
	)
}

package login

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

type Req api.LoginRequest

type Res struct {
	RedirectURL string `json:"redirect_url"`
}

type handler struct {
	oauth *auth.OAuth2
	login *auth.Login
}

func New(p *platform.Platform) *handler {
	return &handler{
		oauth: p.OAuth2,
		login: auth.NewLogin(p.DB, p.Clock, p.Issuer, &p.Config.Auth),
	}
}

func (h *handler) Method() string { return http.MethodPost }
func (h *handler) Path() string   { return "/login" }

func (h *handler) Handle(ctx context.Context, c *httpkit.Context) error {
	req, err := httpkit.BindBody[Req](c)
	if err != nil {
		return err
	}

	usr, err := h.login.Authenticate(ctx, string(req.Email), req.Password)
	if err != nil {
		return err
	}

	cookie, err := c.Req().Raw().Cookie("Auth-State")
	if err != nil || cookie.Value == "" {
		return errMissingAuthState
	}

	state, err := h.oauth.DecryptState(cookie.Value)
	if err != nil {
		return err
	}

	code, err := h.oauth.MintCode(ctx, usr.Pk, state)
	if err != nil {
		return err
	}

	http.SetCookie(c.Res().Writer(), h.oauth.ClearStateCookie())

	return c.Res().JSON(http.StatusOK, Res{
		RedirectURL: h.oauth.BuildRedirectURL(state, code),
	})
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("login-form-per-ip",
			20, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
			ratelimit.WithBurst(),
		),
		// ratelimit.NewPolicy("login-form-per-account",
		// 	10, time.Minute,
		// 	ratelimit.KeyFormValue("email", strings.ToLower),
		// 	ratelimit.WithStore(p.RateLimit),
		// ),
	)
}

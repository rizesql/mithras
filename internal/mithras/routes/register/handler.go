package register

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

type Req api.RegisterRequest
type Res struct {
	RedirectURL string `json:"redirect_url"`
}

type handler struct {
	oauth    *auth.OAuth2
	register auth.Register
}

func New(p *platform.Platform) *handler {
	return &handler{
		oauth:    p.OAuth2,
		register: auth.NewRegister(p.DB, p.Clock, p.Issuer, &p.Config.Auth),
	}
}

func (h *handler) Method() string { return http.MethodPost }
func (h *handler) Path() string   { return "/register" }

func (h *handler) Handle(ctx context.Context, c *httpkit.Context) error {
	req, err := httpkit.BindBody[Req](c)
	if err != nil {
		return err
	}

	res, err := h.register.Register(ctx,
		req.Name,
		string(req.Email),
		req.Password,
		c.Req().Raw().UserAgent(),
		c.Req().IP(),
	)

	if err != nil {
		// c.Redirect("/register?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return errRegistrationFailed(err)
	}

	cookie, err := c.Req().Raw().Cookie("Auth-State")
	if err != nil || cookie.Value == "" {
		return errMissingAuthState
	}

	state, err := h.oauth.DecryptState(cookie.Value)
	if err != nil {
		return err
	}

	code, err := h.oauth.MintCode(ctx, res.UserPk, state)
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
		ratelimit.NewPolicy("register-form-per-ip",
			5, time.Hour,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
		),
	)
}

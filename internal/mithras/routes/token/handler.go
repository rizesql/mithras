package token

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/idkit"
)

type (
	Res api.TokenResponse
)

type handler struct {
	db      *db.Database
	oauth   *auth.OAuth2
	login   *auth.Login
	refresh *auth.Refresh
	cfg     *auth.Config
}

func New(p *platform.Platform) *handler {
	return &handler{
		db:      p.DB,
		oauth:   p.OAuth2,
		login:   auth.NewLogin(p.DB, p.Clock, p.Issuer, &p.Config.Auth),
		refresh: auth.NewRefresh(p.DB, p.Clock, p.Issuer, &p.Config.Auth),
		cfg:     &p.Config.Auth,
	}
}

func (h *handler) Method() string { return http.MethodPost }
func (h *handler) Path() string   { return "/token" }

func (h *handler) Handle(ctx context.Context, c *httpkit.Context) error {
	r := c.Req().Raw()
	if err := r.ParseForm(); err != nil {
		return errInvalidFormData(err)
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		return h.handleAuthorizationCode(ctx, c)
	case "refresh_token":
		return h.handleRefreshToken(ctx, c)
	default:
		return errUnsupportedGrantType
	}
}

func (h *handler) handleAuthorizationCode(ctx context.Context, c *httpkit.Context) error {
	r := c.Req().Raw()
	code := r.FormValue("code")
	if code == "" {
		return errMissingCode
	}

	row, err := h.oauth.ConsumeCode(ctx, idkit.AuthorizationCodeID(code))
	if err != nil {
		return err
	}

	if row.ClientID != r.FormValue("client_id") {
		return errClientIDMismatch
	}

	if row.RedirectUri != r.FormValue("redirect_uri") {
		return errRedirectURIMismatch
	}

	if row.Challenge != "" {
		codeVerifier := r.FormValue("code_verifier")
		if codeVerifier == "" {
			return errMissingCodeVerifier
		}

		if !h.validatePKCE(codeVerifier, row.Challenge) {
			return errPKCEVerificationFailed
		}
	}

	usr, err := db.Query.GetUserByPk(ctx, h.db, row.UserPk)
	if err != nil {
		return errUserLookupFailed(err)
	}

	res, err := h.login.CreateSession(ctx, row.UserPk, usr.ID, r.UserAgent(), c.Req().IP())
	if err != nil {
		return err
	}

	return c.Res().JSON(http.StatusOK, Res{
		AccessToken:  res.AccessToken.Raw(),
		RefreshToken: res.RefreshToken.Raw(),
		ExpiresIn:    int(h.cfg.AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
	})
}

func (h *handler) handleRefreshToken(ctx context.Context, c *httpkit.Context) error {
	r := c.Req().Raw()
	refreshToken := r.FormValue("refresh_token")

	if refreshToken == "" {
		return errMissingRefreshToken
	}

	res, err := h.refresh.Refresh(ctx, refreshToken)
	if err != nil {
		return err
	}

	return c.Res().JSON(http.StatusOK, Res{
		AccessToken:  res.AccessToken.Raw(),
		RefreshToken: res.RefreshToken.Raw(),
		ExpiresIn:    int(h.cfg.AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
	})
}

func (h *handler) validatePKCE(verifier, challenge string) bool {
	hash := sha256.Sum256([]byte(verifier))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])
	return subtle.ConstantTimeCompare([]byte(encoded), []byte(challenge)) == 1
}

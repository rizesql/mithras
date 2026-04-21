package authorize

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/httpkit"
)

type handler struct {
	plt *platform.Platform
}

func New(plt *platform.Platform) *handler {
	return &handler{plt: plt}
}

func (h *handler) Method() string { return http.MethodGet }
func (h *handler) Path() string   { return "/authorize" }

type AuthorizeParams struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

func (h *handler) Handle(_ context.Context, c *httpkit.Context) error {
	params, err := httpkit.BindQuery[api.AuthorizeParams](c)
	if err != nil {
		return errInvalidQueryParams(err)
	}

	if params.ResponseType != "code" {
		return errUnsupportedResponseType(params.ResponseType)
	}

	if params.CodeChallengeMethod != "S256" {
		return errUnsupportedCodeChallengeMethod(params.CodeChallengeMethod)
	}

	if params.CodeChallenge == "" {
		return errMissingCodeChallenge
	}

	if err := h.validateRedirectURI(c.Req().Raw(), params.RedirectUri); err != nil {
		return err
	}

	state := auth.AuthorizeState{
		ClientID:        params.ClientId,
		RedirectURI:     params.RedirectUri,
		State:           *params.State,
		Challenge:       params.CodeChallenge,
		ChallengeMethod: string(params.CodeChallengeMethod),
	}

	encrypted, err := h.plt.OAuth2.EncryptState(state)
	if err != nil {
		return err
	}

	http.SetCookie(c.Res().Writer(), &http.Cookie{
		Name:     "Auth-State",
		Value:    encrypted,
		Path:     "/",
		Expires:  h.plt.Clock.Now().Add(5 * time.Minute),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	c.Redirect("/login", http.StatusFound)
	return nil
}

func (h *handler) validateRedirectURI(req *http.Request, redirectURI string) error {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return errInvalidRedirectURI
	}

	host := u.Hostname()
	if host == "localhost" || host == "127.0.0.1" {
		return nil
	}

	requestHost := req.Host
	if strings.Contains(requestHost, ":") {
		requestHost, _, _ = strings.Cut(requestHost, ":")
	}

	if !h.isSameRootDomain(host, requestHost) {
		return errInvalidRedirectURIDomain(host, requestHost)
	}

	return nil
}

func (h *handler) isSameRootDomain(host, target string) bool {
	hostRoot, err1 := publicsuffix.EffectiveTLDPlusOne(host)
	targetRoot, err2 := publicsuffix.EffectiveTLDPlusOne(target)

	if err1 != nil || err2 != nil {
		return host == target
	}

	return hostRoot == targetRoot
}

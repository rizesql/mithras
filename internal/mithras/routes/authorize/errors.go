package authorize

import (
	"fmt"

	"github.com/rizesql/mithras/internal/errkit"
	"github.com/rizesql/mithras/pkg/api"
)

var (
	errInvalidQueryParams = func(err error) error {
		return errkit.Wrap(err,
			errkit.User.Request.Code("invalid_query_params"),
			errkit.Internal("failed to bind query parameters"),
			errkit.Public("The authorization request is missing required parameters."),
		)
	}

	errUnsupportedResponseType = func(resType api.AuthorizeParamsResponseType) error {
		return errkit.New("unsupported response type",
			errkit.User.Request.Code("unsupported_response_type"),
			errkit.Internal(fmt.Sprintf("response_type %q is not supported", resType)),
			errkit.Public("Only response_type=code is supported."),
		)
	}

	errUnsupportedCodeChallengeMethod = func(method api.AuthorizeParamsCodeChallengeMethod) error {
		return errkit.New("unsupported code challenge method",
			errkit.User.Request.Code("unsupported_code_challenge_method"),
			errkit.Internal(fmt.Sprintf("code_challenge_method %q is not supported", method)),
			errkit.Public("Only code_challenge_method=S256 is supported."),
		)
	}

	errMissingCodeChallenge = errkit.New("missing code challenge",
		errkit.User.Request.Code("missing_code_challenge"),
		errkit.Internal("code_challenge is required"),
		errkit.Public("A code challenge is required for PKCE."),
	)

	errInvalidRedirectURI = errkit.New("invalid redirect uri",
		errkit.User.Request.Code("invalid_redirect_uri"),
		errkit.Internal("failed to parse redirect_uri"),
		errkit.Public("The provided redirect_uri is invalid."),
	)

	errInvalidRedirectURIDomain = func(host, requestHost string) error {
		return errkit.New("invalid redirect uri domain",
			errkit.User.Request.Code("invalid_redirect_uri_domain"),
			errkit.Internal(
				fmt.Sprintf("redirect_uri host %q does not match request host %q", host, requestHost),
			),
			errkit.Public("The redirect_uri domain is not allowed."),
		)
	}
)

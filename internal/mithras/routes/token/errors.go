package token

import "github.com/rizesql/mithras/internal/errkit"

var (
	errInvalidFormData = func(err error) error {
		return errkit.Wrap(err,
			errkit.User.Request.Code("invalid_form_data"),
			errkit.Internal("failed to parse x-www-form-urlencoded body"),
			errkit.Public("The request body is not valid form data."),
		)
	}

	errUnsupportedGrantType = errkit.New("unsupported grant type",
		errkit.User.Request.Code("unsupported_grant_type"),
		errkit.Internal("grant_type must be authorization_code or refresh_token"),
		errkit.Public("Only authorization_code and refresh_token grant types are supported."),
	)

	errMissingCode = errkit.New("missing required parameter",
		errkit.User.Request.Code("missing_token_params"),
		errkit.Internal("code is required"),
		errkit.Public("The request is missing the required parameter: code"),
	)

	errClientIDMismatch = errkit.New("client id mismatch",
		errkit.User.Request.Code("client_id_mismatch"),
		errkit.Internal("client_id does not match the one registered with the code"),
		errkit.Public("The client_id provided does not match the authorization request."),
	)

	errRedirectURIMismatch = errkit.New("redirect uri mismatch",
		errkit.User.Request.Code("redirect_uri_mismatch"),
		errkit.Internal("redirect_uri does not match the one registered with the code"),
		errkit.Public("The redirect_uri provided does not match the authorization request."),
	)

	errMissingCodeVerifier = errkit.New("missing code_verifier",
		errkit.User.Request.Code("missing_code_verifier"),
		errkit.Internal("code_verifier is required for PKCE verification"),
		errkit.Public("The request is missing the required parameter: code_verifier"),
	)

	errPKCEVerificationFailed = errkit.New("pkce verification failed",
		errkit.User.Request.Code("pkce_verification_failed"),
		errkit.Internal("code_verifier hash does not match code_challenge"),
		errkit.Public("PKCE verification failed. The code_verifier is invalid."),
	)

	errUserLookupFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("user_lookup_failed"),
			errkit.Internal("failed to lookup user by pk"),
			errkit.Public("Failed to process token exchange."),
		)
	}

	errMissingRefreshToken = errkit.New("missing refresh token",
		errkit.User.Request.Code("missing_refresh_token"),
		errkit.Internal("refresh_token is required for refresh_token grant"),
		errkit.Public("The refresh_token parameter is required."),
	)
)

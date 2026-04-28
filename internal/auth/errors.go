package auth

import (
	"github.com/rizesql/mithras/internal/errkit"
)

var (
	// Login errors
	errUserNotFound = errkit.New("user not found",
		errkit.WithCode(errkit.User.Request.Code("user_not_found")),
		errkit.Internal("user not found in database"),
		errkit.Public("User not found."),
	)

	errWrongPassword = errkit.New("wrong password",
		errkit.WithCode(errkit.User.Request.Code("wrong_password")),
		errkit.Internal("wrong password"),
		errkit.Public("Wrong password."),
	)

	errSessionLookupFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.App.Internal.Code("session_lookup_failed")),
			errkit.Internal("database lookup failed"),
			errkit.Public("Failed to process login."),
		)
	}

	errAccountSuspended = errkit.New("account suspended",
		errkit.WithCode(errkit.User.Forbidden.Code("account_suspended")),
		errkit.Internal("account is suspended"),
		errkit.Public("Your account has been suspended. Please contact support."),
	)

	errAccountLocked = func(lockedUntil string) error {
		return errkit.New("account locked",
			errkit.WithCode(errkit.User.Forbidden.Code("account_locked")),
			errkit.Internal("account is locked until "+lockedUntil),
			errkit.Public("Too many failed attempts. Your account is temporarily locked."),
		)
	}

	errPasswordVerificationFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.App.Internal.Code("password_verification_failed")),
			errkit.Internal("failed to verify password"),
			errkit.Public("Failed to process login."),
		)
	}

	errRolesLookupFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.App.Internal.Code("roles_lookup_failed")),
			errkit.Internal("failed to lookup user roles"),
			errkit.Public("Failed to process login."),
		)
	}

	errTokenSigningFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.App.Internal.Code("token_signing_failed")),
			errkit.Internal("failed to sign access token"),
			errkit.Public("Failed to process login."),
		)
	}

	errTokenGenerationFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.App.Internal.Code("token_generation_failed")),
			errkit.Internal("failed to generate refresh token"),
			errkit.Public("Failed to process login."),
		)
	}

	errSessionInsertFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.App.Internal.Code("session_insert_failed")),
			errkit.Internal("failed to insert session"),
			errkit.Public("Failed to process login."),
		)
	}

	// Password reset errors
	errUserLookupFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("user_lookup_failed"),
			errkit.Internal("database lookup failed"),
			errkit.Public("Failed to process request."),
		)
	}

	errTokenInsertFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("token_insert_failed"),
			errkit.Internal("failed to insert reset token"),
			errkit.Public("Failed to process request."),
		)
	}

	errInvalidResetToken = errkit.New("invalid reset token",
		errkit.User.Request.Code("reset_token_invalid"),
		errkit.Public("The password reset link is invalid."),
	)

	errResetTokenNotFound = errkit.New("reset token not found or expired",
		errkit.User.Request.Code("reset_token_invalid"),
		errkit.Public("The password reset link is invalid or has expired."),
	)

	errTokenLookupFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("token_lookup_failed"),
			errkit.Internal("failed to lookup reset token"),
			errkit.Public("Failed to process request."),
		)
	}

	errPasswordHistoryLookupFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("password_history_lookup_failed"),
			errkit.Internal("failed to lookup password history"),
			errkit.Public("Failed to reset password."),
		)
	}

	errPasswordHistoryVerificationFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("password_history_verification_failed"),
			errkit.Internal("failed to verify against password history"),
			errkit.Public("Failed to reset password."),
		)
	}

	errPasswordRecentlyUsed = errkit.New("password recently used",
		errkit.User.Request.Code("password_recently_used"),
		errkit.Internal("new password matches one of the last 5 passwords"),
		errkit.Public("You cannot reuse one of your last 5 passwords."),
	)

	errPasswordResetTransactionFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("password_reset_transaction_failed"),
			errkit.Internal("failed to update password and invalidate tokens"),
			errkit.Public("Failed to reset password."),
		)
	}

	// Registration errors
	errPasswordHashFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("password_hash_failed"),
			errkit.Internal("failed to hash password"),
			errkit.Public("Failed to register user."),
		)
	}

	errDuplicateEmail = func(email string) error {
		return errkit.New("user with email already exists",
			errkit.User.Request.Code("duplicate_email"),
			errkit.Publicf("A user with email %v already exists", email),
			errkit.Internalf("duplicate user"),
		)
	}

	errRegistrationDatabaseError = func(err error) error {
		return errkit.Wrap(err,
			errkit.System.Code("service_unavailable"),
			errkit.Internal("database error"),
			errkit.Public("Failed to register user."),
		)
	}

	errCredentialInsertFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.System.Code("service_unavailable"),
			errkit.Internal("failed to insert credential"),
			errkit.Public("Failed to register user."),
		)
	}

	errPasswordHistoryInsertFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.System.Code("service_unavailable"),
			errkit.Internal("failed to insert password history"),
			errkit.Public("Failed to register user."),
		)
	}

	errDefaultRoleAssignmentFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.System.Code("service_unavailable"),
			errkit.Internal("failed to assign default role"),
			errkit.Public("Failed to register user."),
		)
	}

	// Refresh errors
	errInvalidRefreshToken = func(internal string) error {
		return errkit.New("invalid refresh token",
			errkit.User.Request.Code("invalid_refresh_token"),
			errkit.Internal(internal),
			errkit.Public("The provided refresh token is invalid or expired."),
		)
	}

	errRefreshTokenGenerationFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("token_generation_failed"),
			errkit.Internal("failed to generate new refresh token"),
			errkit.Public("An internal error occurred."),
		)
	}

	errRefreshTokenSigningFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("token_signing_failed"),
			errkit.Internal("failed to sign access token"),
			errkit.Public("Failed to process refresh."),
		)
	}

	// Verification errors
	errSessionNotFound = errkit.New("session not found", errkit.User.Auth.Code("session_not_found"))
	errSessionExpired  = errkit.New("session expired", errkit.User.Auth.Code("session_expired"))
	errSessionRevoked  = errkit.New("session revoked", errkit.User.Auth.Code("session_revoked"))

	// Logout errors
	errInvalidLogoutToken = func(internal string) error {
		return errkit.New("invalid refresh token",
			errkit.User.Auth.Code("invalid_refresh_token"),
			errkit.Internal(internal),
			errkit.Public("The provided refresh token is invalid or expired."),
		)
	}

	errLogoutSessionLookupFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Unavailable.Code("session_lookup_failed"),
			errkit.Internal("failed to load session by refresh token hash"),
			errkit.Public("Service temporarily unavailable."),
		)
	}

	// OAuth2 errors
	errOAuth2CodeInsertFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("oauth2_code_insert_failed"),
			errkit.Internal("failed to insert authorization code"),
			errkit.Public("Failed to process authorization request."),
		)
	}

	errOAuth2InvalidCode = errkit.New("invalid or expired authorization code",
		errkit.User.Request.Code("invalid_authorization_code"),
		errkit.Internal("authorization code not found, already used, or expired"),
		errkit.Public("The provided authorization code is invalid or has expired."),
	)

	errOAuth2CodeLookupFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("oauth2_code_lookup_failed"),
			errkit.Internal("failed to lookup authorization code"),
			errkit.Public("Failed to process token exchange."),
		)
	}

	errOAuth2StateMarshalFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("oauth2_state_marshal_failed"),
			errkit.Internal("failed to marshal oauth2 state"),
			errkit.Public("Failed to process authorization request."),
		)
	}

	errOAuth2StateEncryptionFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("oauth2_state_encryption_failed"),
			errkit.Internal("failed to encrypt oauth2 state"),
			errkit.Public("Failed to process authorization request."),
		)
	}

	errOAuth2StateDecodeFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.User.Request.Code("oauth2_state_decode_failed"),
			errkit.Internal("failed to decode oauth2 state"),
			errkit.Public("The authorization state is invalid."),
		)
	}

	errOAuth2StateDecryptionFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.User.Request.Code("oauth2_state_decryption_failed"),
			errkit.Internal("failed to decrypt oauth2 state"),
			errkit.Public("The authorization state is invalid or has expired."),
		)
	}

	errOAuth2StateUnmarshalFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.App.Internal.Code("oauth2_state_unmarshal_failed"),
			errkit.Internal("failed to unmarshal oauth2 state"),
			errkit.Public("The authorization state is invalid."),
		)
	}
)

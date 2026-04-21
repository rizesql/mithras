package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/cryptokit"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/idkit"
)

type AuthorizeState struct {
	ClientID        string `json:"cid"`
	RedirectURI     string `json:"ruri"`
	State           string `json:"st"`
	Challenge       string `json:"ch"`
	ChallengeMethod string `json:"cm"`
}

type OAuth2 struct {
	db  *db.Database
	clk clock.Clock
	aes *cryptokit.AESGCM
}

func NewOAuth2(db *db.Database, clk clock.Clock, kek []byte) (*OAuth2, error) {
	aes, err := cryptokit.NewAESGCM(kek[:])
	if err != nil {
		return nil, fmt.Errorf("auth.NewOAuth2: %w", err)
	}

	return &OAuth2{
		db:  db,
		clk: clk,
		aes: aes,
	}, nil
}

func (o *OAuth2) MintCode(
	ctx context.Context,
	userPk int64,
	state AuthorizeState,
) (idkit.AuthorizationCodeID, error) {
	code := idkit.NewAuthorizationCodeID()
	expiresAt := o.clk.Now().Add(5 * time.Minute)

	_, err := db.Query.InsertAuthorizationCode(ctx, o.db, db.InsertAuthorizationCodeParams{
		Code:        code,
		UserPk:      userPk,
		ClientID:    state.ClientID,
		RedirectUri: state.RedirectURI,
		Challenge:   state.Challenge,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return "", errOAuth2CodeInsertFailed(err)
	}

	return code, nil
}

func (o *OAuth2) ConsumeCode(
	ctx context.Context,
	code idkit.AuthorizationCodeID,
) (db.ConsumeAuthorizationCodeRow, error) {
	row, err := db.Query.ConsumeAuthorizationCode(ctx, o.db, code)
	if err != nil {
		if db.IsNotFound(err) {
			return db.ConsumeAuthorizationCodeRow{}, errOAuth2InvalidCode
		}
		return db.ConsumeAuthorizationCodeRow{}, errOAuth2CodeLookupFailed(err)
	}

	return row, nil
}

func (o *OAuth2) EncryptState(state AuthorizeState) (string, error) {
	raw, err := json.Marshal(state)
	if err != nil {
		return "", errOAuth2StateMarshalFailed(err)
	}

	encrypted, err := o.aes.Encrypt(raw)
	if err != nil {
		return "", errOAuth2StateEncryptionFailed(err)
	}

	return base64.URLEncoding.EncodeToString(encrypted), nil
}

// DecryptState decrypts the base64-encoded encrypted state into an AuthorizeState.
func (o *OAuth2) DecryptState(encrypted string) (AuthorizeState, error) {
	raw, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return AuthorizeState{}, errOAuth2StateDecodeFailed(err)
	}

	decrypted, err := o.aes.Decrypt(raw)
	if err != nil {
		return AuthorizeState{}, errOAuth2StateDecryptionFailed(err)
	}

	var state AuthorizeState
	if err := json.Unmarshal(decrypted, &state); err != nil {
		return AuthorizeState{}, errOAuth2StateUnmarshalFailed(err)
	}

	return state, nil
}

func (o *OAuth2) BuildRedirectURL(state AuthorizeState, code idkit.AuthorizationCodeID) string {
	params := url.Values{}
	params.Set("code", code.String())
	if state.State != "" {
		params.Set("state", state.State)
	}

	target := state.RedirectURI
	if strings.Contains(target, "?") {
		return target + "&" + params.Encode()
	}
	return target + "?" + params.Encode()
}

func (o *OAuth2) ClearStateCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "Auth-State",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

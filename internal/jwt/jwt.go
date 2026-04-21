package jwt

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/rizesql/mithras/internal/jws"
)

// Token represents a signed JWS (JSON Web Signature).
type Token string

// String redacts the token to prevent accidental leakage in logs.
func (t Token) String() string { return "[REDACTED]" }

// Raw returns the underlying token string. Use this ONLY for API responses.
func (t Token) Raw() string { return string(t) }

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	KID string `json:"kid,omitempty"`
}

type Claims struct {
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	Roles     []string `json:"roles,omitempty"`
}

func Sign(key jws.Key, claims Claims) (Token, error) {
	header, err := json.Marshal(Header{
		Alg: key.Alg,
		Typ: "JWT",
		KID: key.ID.String(),
	})
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	message := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(payload)

	sig, err := key.Signer.Sign(rand.Reader, []byte(message), crypto.Hash(0))
	if err != nil {
		return "", err
	}

	return Token(message + "." + base64.RawURLEncoding.EncodeToString(sig)), nil
}

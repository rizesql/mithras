package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// Refresh is a secure, random string used to obtain new access tokens.
// It is stored as a SHA-256 hash in the database.
type Refresh string

// String redacts the token to prevent accidental leakage in logs.
func (t Refresh) String() string { return "[REDACTED]" }

// Raw returns the underlying token string. Use this ONLY for API responses.
func (t Refresh) Raw() string { return string(t) }

// GenerateRefresh creates a 32-byte secure random token.
func GenerateRefresh() (Refresh, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return Refresh(base64.RawURLEncoding.EncodeToString(b)), nil
}

// Hash returns the SHA-256 hash of the refresh token for database storage.
func (t Refresh) Hash() []byte {
	h := sha256.Sum256([]byte(t))
	return h[:]
}

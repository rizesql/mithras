package jws

import (
	"context"
)

// Store handles retrieval of keys for signing and verification.
type Store interface {
	// SigningKey returns the currently active key for signing new tokens.
	SigningKey(ctx context.Context) (Key, error)

	// PublicKeys returns all active public keys (including older ones)
	PublicKeys(ctx context.Context) ([]Key, error)
}

package jws

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/idkit"
)

type KeyGenerator func() (Key, error)

// KeyCodec handles algorithm-specific key generation and serialization.
// Implementations own the algorithm name, how to produce a new keypair,
// how to marshal the private key to raw bytes (for encryption at rest),
// and how to unmarshal raw bytes back to a crypto.Signer.
type KeyCodec interface {
	// Alg returns the algorithm name stored on each Key (e.g. "EdDSA").
	Alg() string
	// New generates a fresh key-pair and returns the signer and its
	// serialized private-key bytes ready for encryption.
	New() (signer crypto.Signer, raw []byte, err error)
	// Decode reconstructs a crypto.Signer from raw private-key bytes.
	Decode(raw []byte) (crypto.Signer, error)
}

// Key represents a specific version of a signing key.
type Key struct {
	ID        idkit.KeyID
	Signer    crypto.Signer
	Alg       string
	RotatesAt time.Time
	ExpiresAt time.Time
}

// Public is a convenience so callers don't need to type-assert.
func (k Key) Public() crypto.PublicKey {
	return k.Signer.Public()
}

func (k Key) ToJWK() (api.JWK, error) {
	switch pub := k.Signer.Public().(type) {
	case ed25519.PublicKey:
		var jwk api.JWK

		err := jwk.FromJWKOkp(api.JWKOkp{
			Kid: k.ID.String(),
			Alg: api.EdDSA,
			Kty: api.OKP,
			Crv: api.Ed25519,
			X:   base64.RawURLEncoding.EncodeToString(pub),
		})

		return jwk, err

	default:
		return api.JWK{}, fmt.Errorf("unsupported key type: %T", pub)
	}
}

func JWKS(keys []Key) (api.JWKS, error) {
	jwks := api.JWKS{Keys: make([]api.JWK, 0, len(keys))}
	for _, k := range keys {
		jwk, err := k.ToJWK()
		if err != nil {
			return api.JWKS{}, err
		}

		jwks.Keys = append(jwks.Keys, jwk)
	}

	return jwks, nil
}

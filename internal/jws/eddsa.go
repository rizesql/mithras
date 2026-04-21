package jws

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/rizesql/mithras/pkg/idkit"
)

// Ed25519Key generates a single Ed25519 Key. Satisfies KeyGenerator,
// kept for use with MemoryStore.
func Ed25519Key() (Key, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return Key{}, fmt.Errorf("generate ed25519: %w", err)
	}

	return Key{ID: idkit.NewKeyID(), Signer: priv, Alg: "EdDSA"}, nil
}

// EdDSA is a KeyCodec for Ed25519 / EdDSA keys.
type EdDSA struct{}

func (EdDSA) Alg() string { return "EdDSA" }

func (EdDSA) New() (crypto.Signer, []byte, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ed25519: %w", err)
	}

	return priv, []byte(priv), nil
}

func (EdDSA) Decode(raw []byte) (crypto.Signer, error) {
	if len(raw) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("jws: invalid ed25519 private key size: got %d, want %d",
			len(raw), ed25519.PrivateKeySize)
	}

	return ed25519.PrivateKey(raw), nil
}

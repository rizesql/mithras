package jws

import (
	"github.com/rizesql/mithras/pkg/cryptokit"
)

// encryptor handles encrypting and decrypting JWS private keys at rest.
// It is a wrapper around cryptokit.AESGCM for compatibility.
type encryptor struct {
	*cryptokit.AESGCM
}

// newEncryptor creates a new Encryptor using the provided Key Encryption Key (KEK).
func newEncryptor(kek []byte) (*encryptor, error) {
	aesGCM, err := cryptokit.NewAESGCM(kek)
	if err != nil {
		return nil, err
	}

	return &encryptor{AESGCM: aesGCM}, nil
}

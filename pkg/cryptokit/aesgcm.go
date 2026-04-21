package cryptokit

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// AESGCM handles encrypting and decrypting data using AES-GCM.
type AESGCM struct {
	aead cipher.AEAD
}

// NewAESGCM creates a new AESGCM instance using the provided 32-byte key.
func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("kek must be exactly 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	return &AESGCM{aead: aesGCM}, nil
}

// Encrypt seals raw bytes using AES-GCM, prepending a random nonce.
func (e *AESGCM) Encrypt(raw []byte) ([]byte, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}

	return e.aead.Seal(nonce, nonce, raw, nil), nil
}

// Decrypt opens AES-GCM ciphertext (nonce || ciphertext) and returns the raw bytes.
func (e *AESGCM) Decrypt(encryptedData []byte) ([]byte, error) {
	nonceSize := e.aead.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("cryptokit: ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	return e.aead.Open(nil, nonce, ciphertext, nil)
}

package password

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type params struct {
	time    uint32
	memory  uint32
	threads uint8
}

var argon2id = params{
	time:    uint32(2),
	memory:  uint32(64 * 1024),
	threads: uint8(2),
}

const (
	keyLen  = 32
	saltLen = 16
)

var phcPrefix = fmt.Sprintf(
	"$argon2id$v=%d$m=%d,t=%d,p=%d",
	argon2.Version, argon2id.memory, argon2id.time, argon2id.threads,
)

// dummyHash is used for timing attack mitigation. It matches the Argon2id parameters
// used by this package (m=64MB, t=2, p=2).
//
//nolint:lll
const dummyHash = "$argon2id$v=19$m=65536,t=2,p=2$ZHVtbXktc2FsdC0xNi1ieQ$dGhpcy1pcy1hLWR1bW15LTMyLWJ5dGUtaGFzaC0hISEh"

type Hashed struct {
	value string
}

func (Hashed) String() string { return "[REDACTED]" }

func (r Raw) Hash() (Hashed, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return Hashed{}, err
	}

	hash := argon2.IDKey(
		[]byte(r.value),
		salt,
		argon2id.time,
		argon2id.memory,
		argon2id.threads,
		keyLen,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("%s$%s$%s", phcPrefix, b64Salt, b64Hash)

	return Hashed{value: encoded}, nil
}

func (h Hashed) Verify(raw Raw) (bool, error) {
	target := h.value
	if target == "" {
		target = dummyHash
	}

	dec, err := decode(target)
	if err != nil {
		return false, err
	}

	if dec.algorithm != "argon2id" {
		return false, errors.New("unsupported algorithm")
	}

	if dec.version != argon2.Version {
		return false, errors.New("incompatible version")
	}

	computed := argon2.IDKey(
		[]byte(raw.value),
		dec.salt,
		dec.time,
		dec.memory,
		dec.threads,
		// #nosec G115
		uint32(len(dec.hash)),
	)

	match := subtle.ConstantTimeCompare(dec.hash, computed) == 1

	if h.value == "" {
		return false, nil
	}

	return match, nil
}

func (h Hashed) NeedsRehash() (bool, error) {
	dec, err := decode(h.value)
	if err != nil {
		return true, err
	}

	if dec.algorithm != "argon2id" {
		return true, nil
	}

	if dec.memory != argon2id.memory ||
		dec.time != argon2id.time ||
		dec.threads != argon2id.threads {
		return true, nil
	}

	return false, nil
}

func (h *Hashed) Scan(src any) error {
	switch v := src.(type) {
	case string:
		h.value = v
		return nil
	case []byte:
		h.value = string(v)
		return nil
	default:
		return fmt.Errorf("password.Hashed: cannot scan type %T", src)
	}
}

func (h Hashed) Value() (driver.Value, error) {
	return h.value, nil
}

type decodedHash struct {
	salt      []byte
	hash      []byte
	time      uint32
	memory    uint32
	algorithm string
	version   int
	threads   uint8
}

func decode(value string) (decodedHash, error) {
	parts := strings.Split(value, "$")
	if len(parts) != 6 {
		return decodedHash{}, errors.New("invalid hash format")
	}

	var dec decodedHash
	dec.algorithm = parts[1]

	n, err := fmt.Sscanf(parts[2], "v=%d", &dec.version)
	if err != nil || n != 1 {
		return decodedHash{}, errors.New("invalid version format")
	}

	n, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &dec.memory, &dec.time, &dec.threads)
	if err != nil || n != 3 {
		return decodedHash{}, errors.New("invalid parameter format")
	}

	dec.salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return decodedHash{}, err
	}

	dec.hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return decodedHash{}, err
	}

	return dec, nil
}

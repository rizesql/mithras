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

type Hashed struct {
	value string
}

func (Hashed) String() string { return "[REDACTED]" }

func (r Raw) Hash() (Hashed, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return Hashed{}, err
	}

	hash := argon2.IDKey([]byte(r.value), salt, argon2id.time, argon2id.memory, argon2id.threads, keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("%s$%s$%s", phcPrefix, b64Salt, b64Hash)

	return Hashed{value: encoded}, nil
}

func (h Hashed) Verify(rhs string) (bool, error) {
	d, err := decode(h.value)
	if err != nil {
		return false, err
	}

	if d.algorithm != "argon2id" {
		return false, errors.New("unsupported algorithm")
	}
	if d.version != argon2.Version {
		return false, errors.New("incompatible version")
	}

	if len(d.hash) > 256 {
		return false, errors.New("hash length exceeds reasonable bounds")
	}

	computed := argon2.IDKey(
		[]byte(rhs),
		d.salt,
		d.time,
		d.memory,
		d.threads,
		//nolint:gosec // already checked d.hash is not too long
		// #nosec G115
		uint32(len(d.hash)),
	)

	eq := subtle.ConstantTimeCompare(d.hash, computed) == 1

	return eq, nil
}

func (h Hashed) NeedsRehash() (bool, error) {
	d, err := decode(h.value)
	if err != nil {
		return true, err
	}

	if d.algorithm != "argon2id" {
		return true, nil
	}

	if d.memory != argon2id.memory ||
		d.time != argon2id.time ||
		d.threads != argon2id.threads {
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
	algorithm string
	version   int
	memory    uint32
	time      uint32
	threads   uint8
	salt      []byte
	hash      []byte
}

func decode(value string) (decodedHash, error) {
	parts := strings.Split(value, "$")
	if len(parts) != 6 {
		return decodedHash{}, errors.New("invalid hash format")
	}

	var d decodedHash
	d.algorithm = parts[1]

	n, err := fmt.Sscanf(parts[2], "v=%d", &d.version)
	if err != nil || n != 1 {
		return decodedHash{}, errors.New("invalid version format")
	}

	n, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &d.memory, &d.time, &d.threads)
	if err != nil || n != 3 {
		return decodedHash{}, errors.New("invalid parameter format")
	}

	d.salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return decodedHash{}, err
	}

	d.hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return decodedHash{}, err
	}

	return d, nil
}

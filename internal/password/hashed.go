package password

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
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

type Hashed struct {
	value string
}

func (Hashed) String() string { return "[REDACTED]" }

func (r Raw) Hash() (Hashed, error) {
	return Hashed(r), nil
}

func (h Hashed) Verify(raw Raw) (bool, error) {
	return h.value == raw.value, nil
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

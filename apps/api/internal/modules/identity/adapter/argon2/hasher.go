package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultParams согласно рекомендациям OWASP и ADR-007:
// memory: 64MB, time: 3, parallelism: 2, salt: 16 bytes, key: 32 bytes.
var DefaultParams = Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

type Hasher struct {
	params Params
}

func NewHasher(params ...Params) *Hasher {
	p := DefaultParams
	if len(params) > 0 {
		p = params[0]
	}
	return &Hasher{params: p}
}

func (h *Hasher) HashPassword(password string) (string, error) {
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, h.params.Iterations, h.params.Memory, h.params.Parallelism, h.params.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.params.Memory, h.params.Iterations, h.params.Parallelism, b64Salt, b64Hash)

	return encoded, nil
}

func (h *Hasher) ComparePassword(encodedHash, password string) (bool, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return false, errors.New("invalid hash format")
	}

	var version int
	if _, err := fmt.Sscanf(vals[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parsing hash version: %w", err)
	}
	if version != argon2.Version {
		return false, errors.New("incompatible argon2 version")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("parsing hash parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, fmt.Errorf("decoding salt: %w", err)
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false, fmt.Errorf("decoding hash: %w", err)
	}

	keyLen := uint32(len(decodedHash))
	comparisonHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLen)

	if subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1 {
		return true, nil
	}

	return false, nil
}

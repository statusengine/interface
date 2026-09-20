package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ErrInvalidHash is returned when a stored hash cannot be parsed - a row
// written by a different tool, or a truncated column.
var ErrInvalidHash = errors.New("auth: unrecognised password hash format")

// argon2 parameters. These follow the RFC 9106 second recommended option
// (64 MiB, 3 passes), which is the memory-lean side of the two and a
// reasonable fit for a daemon that also holds a MySQL pool.
const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

// HashPassword returns a PHC-formatted argon2id hash. The parameters are
// encoded into the string, so raising them later does not invalidate hashes
// written today.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: generating salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword reports whether password matches encoded. It compares in
// constant time and re-derives with the parameters stored in the hash
// rather than the current constants.
func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, ErrInvalidHash
	}
	if version != argon2.Version {
		return false, fmt.Errorf("%w: unsupported argon2 version %d", ErrInvalidHash, version)
	}

	var memory, t uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &t, &threads); err != nil {
		return false, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrInvalidHash
	}

	got := argon2.IDKey([]byte(password), salt, t, memory, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// dummyHash is verified against when a login names a user that does not
// exist, so a missing account and a wrong password take the same time and
// the response cannot be used to enumerate usernames.
var dummyHash string

func init() {
	h, err := HashPassword("this password matches nothing")
	if err != nil {
		panic("auth: cannot hash the dummy password: " + err.Error())
	}
	dummyHash = h
}

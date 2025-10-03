package passhash

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Parameters tuned for interactive logins.
const (
	memory      uint32 = 64 * 1024
	iterations  uint32 = 3
	parallelism uint8  = 2
	saltLength  uint32 = 16
	hashLength  uint32 = 32
)

// HashPassword returns a PHC formatted Argon2id hash string for the provided password.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, hashLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memory, iterations, parallelism, b64Salt, b64Hash)
	return encoded, nil
}

// VerifyPassword compares a plaintext password with a PHC formatted Argon2id hash.
func VerifyPassword(encoded, password string) (bool, error) {
	if encoded == "" {
		return false, errors.New("empty hash")
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}
	var m uint32
	var t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	calc := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(decodedHash)))
	if subtleConstantTimeEquals(calc, decodedHash) {
		return true, nil
	}
	return false, nil
}

func subtleConstantTimeEquals(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

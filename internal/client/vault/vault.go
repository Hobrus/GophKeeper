package vault

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
)

// KeyLength defines AES-256 key size.
const KeyLength = 32

// Path returns default vault key path.
func Path() string {
	home, _ := os.UserHomeDir()
	return home + string(os.PathSeparator) + ".gophkeeper_vault_key"
}

// Exists checks if vault key file exists.
func Exists() bool {
	_, err := os.Stat(Path())
	return err == nil
}

// Generate creates and stores a new random key.
func Generate() ([]byte, error) {
	if Exists() {
		return nil, errors.New("vault key already exists")
	}
	key := make([]byte, KeyLength)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := Save(key); err != nil {
		return nil, err
	}
	return key, nil
}

// Save writes key to disk base64 encoded with 0600 perms.
func Save(key []byte) error {
	b64 := base64.StdEncoding.EncodeToString(key)
	return os.WriteFile(Path(), []byte(b64), 0600)
}

// Load reads key from disk.
func Load() ([]byte, error) {
	b, err := os.ReadFile(Path())
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return nil, err
	}
	if len(key) != KeyLength {
		return nil, errors.New("invalid key length")
	}
	return key, nil
}

package cryptohelper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

// EncryptAESGCM encrypts plaintext using AES-256-GCM.
// The returned slice is nonce||ciphertext where nonce has length gcm.NonceSize().
// The aad parameter is used as Additional Authenticated Data.
func EncryptAESGCM(key, plaintext, aad []byte) ([]byte, error) {
	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blk)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, aad)
	return append(nonce, ciphertext...), nil
}

// DecryptAESGCM decrypts data produced by EncryptAESGCM using the same key and aad.
// The ciphertext must be in the format nonce||ciphertext.
func DecryptAESGCM(key, ciphertext, aad []byte) ([]byte, error) {
	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blk)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	ct := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, aad)
}

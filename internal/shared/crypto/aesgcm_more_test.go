package cryptohelper_test

import (
	"testing"

	cryptohelper "gophkeeper/internal/shared/crypto"
)

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	if _, err := cryptohelper.EncryptAESGCM(make([]byte, 15), []byte("x"), nil); err == nil {
		t.Fatalf("expected error for invalid key length")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key := make([]byte, 32)
	ct, err := cryptohelper.EncryptAESGCM(key, []byte("data"), []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	wrong := make([]byte, 32)
	wrong[0] = 1
	if _, err := cryptohelper.DecryptAESGCM(wrong, ct, []byte("aad")); err == nil {
		t.Fatalf("expected auth error with wrong key")
	}
}

func TestDecrypt_ShortCiphertext(t *testing.T) {
	key := make([]byte, 32)
	if _, err := cryptohelper.DecryptAESGCM(key, []byte("short"), nil); err == nil {
		t.Fatalf("expected error for short ciphertext")
	}
}

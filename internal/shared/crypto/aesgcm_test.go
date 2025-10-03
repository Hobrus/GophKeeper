package cryptohelper_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	cryptohelper "gophkeeper/internal/shared/crypto"
)

func TestAESGCM_EncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	plaintxt := []byte("secret data")
	aad := []byte("meta")
	ciphertext, err := cryptohelper.EncryptAESGCM(key, plaintxt, aad)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	if bytes.Equal(ciphertext, plaintxt) {
		t.Fatalf("ciphertext must differ from plaintext")
	}
	decrypted, err := cryptohelper.DecryptAESGCM(key, ciphertext, aad)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if !bytes.Equal(decrypted, plaintxt) {
		t.Fatalf("decrypted mismatch: got %q want %q", decrypted, plaintxt)
	}
}

func TestAESGCM_BadAAD(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	ciphertext, err := cryptohelper.EncryptAESGCM(key, []byte("x"), []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cryptohelper.DecryptAESGCM(key, ciphertext, []byte("b")); err == nil {
		t.Fatalf("expected auth error with wrong AAD")
	}
}

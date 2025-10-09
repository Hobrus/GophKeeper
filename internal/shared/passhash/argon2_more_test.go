package passhash

import "testing"

func TestVerify_Errors(t *testing.T) {
	if _, err := VerifyPassword("", "x"); err == nil {
		t.Fatalf("want error on empty hash")
	}
	if _, err := VerifyPassword("$argon2id$bad", "x"); err == nil {
		t.Fatalf("want error on bad format")
	}
}

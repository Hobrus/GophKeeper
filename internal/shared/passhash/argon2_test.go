package passhash

import "testing"

func TestHashAndVerify(t *testing.T) {
	h, err := HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := VerifyPassword(h, "password123")
	if err != nil || !ok {
		t.Fatalf("verify failed: %v", err)
	}
	ok, err = VerifyPassword(h, "wrong")
	if err != nil || ok {
		t.Fatalf("expected mismatch")
	}
}

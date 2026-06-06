package auth

import "testing"

func TestHashPasswordAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("strong-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "strong-password" {
		t.Fatal("password hash must not equal the plaintext password")
	}
	if !CheckPassword(hash, "strong-password") {
		t.Fatal("expected password to match hash")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("expected wrong password to fail")
	}
}

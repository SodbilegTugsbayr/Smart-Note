package userman

import "testing"

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	password := "StrongPass123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == password {
		t.Fatal("HashPassword() returned the plaintext password")
	}

	if !VerifyPassword(password, hash) {
		t.Fatal("VerifyPassword() should accept the original password")
	}

	if VerifyPassword("WrongPass123!", hash) {
		t.Fatal("VerifyPassword() should reject the wrong password")
	}
}

func TestVerifyPasswordRejectsInvalidHash(t *testing.T) {
	if VerifyPassword("password", "not-a-valid-bcrypt-hash") {
		t.Fatal("VerifyPassword() should reject malformed hashes")
	}
}

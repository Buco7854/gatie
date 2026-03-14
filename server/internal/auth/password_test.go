package auth

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "my-secure-password-123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("hash should not be empty")
	}
	if hash == password {
		t.Fatal("hash should differ from password")
	}

	if !CheckPassword(password, hash) {
		t.Error("password should match its hash")
	}

	if CheckPassword("wrong-password", hash) {
		t.Error("wrong password should not match")
	}
}

func TestHashPasswordProducesDifferentHashes(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")

	if h1 == h2 {
		t.Error("bcrypt should produce different hashes for same password (salt)")
	}
}

func TestHashToken(t *testing.T) {
	token := "my-refresh-token-value"

	hash1 := HashToken(token)
	hash2 := HashToken(token)

	if hash1 != hash2 {
		t.Error("SHA-256 hash should be deterministic")
	}
	if hash1 == token {
		t.Error("hash should differ from input")
	}
	if len(hash1) != 64 {
		t.Errorf("SHA-256 hex should be 64 chars, got %d", len(hash1))
	}

	different := HashToken("different-token")
	if hash1 == different {
		t.Error("different inputs should produce different hashes")
	}
}

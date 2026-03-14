package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	m := NewJWTManager("test-secret-key-123", 15*time.Minute, 7*24*time.Hour)

	token, err := m.GenerateAccessToken("member-id-123", "ADMIN", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := m.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}

	if claims.MemberID != "member-id-123" {
		t.Errorf("expected member ID member-id-123, got %s", claims.MemberID)
	}
	if claims.Role != "ADMIN" {
		t.Errorf("expected role ADMIN, got %s", claims.Role)
	}
	if claims.Username != "admin" {
		t.Errorf("expected username admin, got %s", claims.Username)
	}
	if claims.Issuer != "gatie" {
		t.Errorf("expected issuer gatie, got %s", claims.Issuer)
	}
}

func TestValidateExpiredToken(t *testing.T) {
	m := NewJWTManager("test-secret", -1*time.Second, 7*24*time.Hour)

	token, err := m.GenerateAccessToken("id", "MEMBER", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	m1 := NewJWTManager("secret-1", 15*time.Minute, 7*24*time.Hour)
	m2 := NewJWTManager("secret-2", 15*time.Minute, 7*24*time.Hour)

	token, _ := m1.GenerateAccessToken("id", "MEMBER", "user")

	_, err := m2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestValidateInvalidToken(t *testing.T) {
	m := NewJWTManager("secret", 15*time.Minute, 7*24*time.Hour)

	_, err := m.ValidateAccessToken("not-a-jwt-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected 64 char hex string, got %d chars", len(token1))
	}

	token2, _ := GenerateRefreshToken()
	if token1 == token2 {
		t.Error("refresh tokens should be unique")
	}
}

func TestRefreshDuration(t *testing.T) {
	d := 7 * 24 * time.Hour
	m := NewJWTManager("secret", 15*time.Minute, d)
	if m.RefreshDuration() != d {
		t.Errorf("expected %v, got %v", d, m.RefreshDuration())
	}
}

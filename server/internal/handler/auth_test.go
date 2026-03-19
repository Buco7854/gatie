package handler

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestUUIDToString(t *testing.T) {
	// UUID v4: 550e8400-e29b-41d4-a716-446655440000
	bytes := [16]byte{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b,
		0x41, 0xd4,
		0xa7, 0x16,
		0x44, 0x66, 0x55, 0x44, 0x00, 0x00,
	}

	u := pgtype.UUID{Bytes: bytes, Valid: true}
	b := u.Bytes
	result := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	expected := "550e8400-e29b-41d4-a716-446655440000"

	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestBuildRefreshCookie(t *testing.T) {
	cookie := buildRefreshCookie("test-token", 7*24*time.Hour)

	if cookie == "" {
		t.Fatal("cookie should not be empty")
	}

	checks := []string{
		"refresh_token=test-token",
		"Path=/api/auth",
		"HttpOnly",
		"Secure",
		"SameSite=Strict",
		"Max-Age=604800",
	}

	for _, check := range checks {
		if !strings.Contains(cookie, check) {
			t.Errorf("cookie missing %q, got: %s", check, cookie)
		}
	}
}

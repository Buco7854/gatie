package handler

import (
	"testing"
	"time"
)

func TestFormatSeconds(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0"},
		{1 * time.Second, "1"},
		{60 * time.Second, "60"},
		{7 * 24 * time.Hour, "604800"},
	}

	for _, tt := range tests {
		got := formatSeconds(tt.d)
		if got != tt.want {
			t.Errorf("formatSeconds(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestUUIDBytesToString(t *testing.T) {
	// UUID v4: 550e8400-e29b-41d4-a716-446655440000
	bytes := [16]byte{
		0x55, 0x0e, 0x84, 0x00,
		0xe2, 0x9b,
		0x41, 0xd4,
		0xa7, 0x16,
		0x44, 0x66, 0x55, 0x44, 0x00, 0x00,
	}

	result := uuidBytesToString(bytes)
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

	// Verify key attributes
	expected := "refresh_token=test-token; HttpOnly; Secure; SameSite=Strict; Path=/auth; Max-Age=604800"
	if cookie != expected {
		t.Errorf("got:\n  %s\nwant:\n  %s", cookie, expected)
	}
}

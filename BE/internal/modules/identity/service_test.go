package identity

import (
	"crypto/rand"
	"strings"
	"testing"
)

func TestPasswordHashingAndValidation(t *testing.T) {
	password := rand.Text() + "-test-only"
	a, err := passwordHash(password)
	if err != nil {
		t.Fatal(err)
	}
	b, err := passwordHash(password)
	if err != nil {
		t.Fatal(err)
	}
	if a == b || strings.Contains(a, password) || !verifyPassword(password, a) || verifyPassword("wrong-password", a) {
		t.Fatal("salted password verification failed")
	}
	for _, bad := range []string{"short", strings.Repeat("x", 129)} {
		if _, err := passwordHash(bad); err == nil {
			t.Fatal("invalid password accepted")
		}
	}
	if verifyPassword(password, "invalid") {
		t.Fatal("malformed hash accepted")
	}
}
func TestEmailNormalization(t *testing.T) {
	got, err := normalizeEmail("  PLAYER@EXAMPLE.COM ")
	if err != nil || got != "player@example.com" {
		t.Fatal("email not normalized")
	}
	for _, bad := range []string{"invalid", "a@localhost", "Name <a@example.com>", strings.Repeat("x", 255) + "@example.com"} {
		if _, err := normalizeEmail(bad); err == nil {
			t.Fatal("invalid email accepted")
		}
	}
}

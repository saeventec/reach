package main

import (
	"strings"
	"testing"
)

func TestHashPasswordUsesThePHCFormat(t *testing.T) {
	hash := hashPassword("purple tram winter")

	wantStart := "$argon2id$v=19$m=19456,t=2,p=1$"
	if !strings.HasPrefix(hash, wantStart) {
		t.Errorf("got %q, want it to start with %q", hash, wantStart)
	}

	// $argon2id$v=19$m=...$salt$hash: 5 parts after the first "$".
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Fatalf("got %d parts, want 6: %q", len(parts), hash)
	}
	// Base64 needs 22 characters for 16 bytes, and 43 for 32 bytes.
	salt := parts[4]
	if len(salt) != 22 {
		t.Errorf("salt: got %d characters, want 22", len(salt))
	}
	hashPart := parts[5]
	if len(hashPart) != 43 {
		t.Errorf("hash: got %d characters, want 43", len(hashPart))
	}
}

func TestHashPasswordNeverContainsThePassword(t *testing.T) {
	password := "purple tram winter"

	hash := hashPassword(password)

	if strings.Contains(hash, password) {
		t.Errorf("the hash contains the password: %q", hash)
	}
}

// Each hash gets a new random salt, so the same password gives a different
// hash every time.
func TestHashPasswordUsesANewSaltEachTime(t *testing.T) {
	first := hashPassword("purple tram winter")
	second := hashPassword("purple tram winter")

	if first == second {
		t.Errorf("two hashes of the same password are equal: %q", first)
	}
}

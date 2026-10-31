package main

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
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

func TestVerifyPasswordAcceptsTheRightPassword(t *testing.T) {
	hash := hashPassword("purple tram winter")

	matches, err := verifyPassword("purple tram winter", hash)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matches {
		t.Errorf("the right password was refused")
	}
}

func TestVerifyPasswordRefusesAWrongPassword(t *testing.T) {
	hash := hashPassword("purple tram winter")

	matches, err := verifyPassword("purple tram summer", hash)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches {
		t.Errorf("a wrong password was accepted")
	}
}

// A hash made with other settings must still verify: the settings are read
// from the hash, not from our constants. This is what lets us raise the
// settings later.
func TestVerifyPasswordUsesTheSettingsInTheHash(t *testing.T) {
	// Made with m=12288, t=3, p=1 (another OWASP setting), for "purple tram winter".
	salt := []byte("0123456789abcdef")
	oldHash := argon2.IDKey([]byte("purple tram winter"), salt, 3, 12288, 1, 32)
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(oldHash)
	phcString := "$argon2id$v=19$m=12288,t=3,p=1$" + encodedSalt + "$" + encodedHash

	matches, err := verifyPassword("purple tram winter", phcString)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matches {
		t.Errorf("a hash with older settings was refused")
	}
}

func TestVerifyPasswordReportsMalformedHashes(t *testing.T) {
	malformedHashes := []string{
		"",
		"purple tram winter",
		"$argon2i$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA",  // argon2i, not argon2id
		"$argon2id$v=16$m=19456,t=2,p=1$c2FsdA$aGFzaA", // an old argon2 version
		"$argon2id$v=19$m=lots,t=2,p=1$c2FsdA$aGFzaA",  // memory is not a number
		"$argon2id$v=19$m=19456,t=0,p=1$c2FsdA$aGFzaA", // 0 passes would crash argon2
		"$argon2id$v=19$m=19456,t=2,p=1$not base64!$aGFzaA",
	}

	for _, malformedHash := range malformedHashes {
		_, err := verifyPassword("purple tram winter", malformedHash)
		if !errors.Is(err, errMalformedHash) {
			t.Errorf("verifyPassword(%q): got %v, want %v", malformedHash, err, errMalformedHash)
		}
	}
}

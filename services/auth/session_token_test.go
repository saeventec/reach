package main

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestNewSessionTokenIsRandomAndLongEnough(t *testing.T) {
	first := newSessionToken()
	second := newSessionToken()

	if first == second {
		t.Errorf("two tokens are equal: %q", first)
	}

	tokenBytes, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("the token is not URL base64: %v", err)
	}
	if len(tokenBytes) != sessionTokenLength {
		t.Errorf("token length: got %d bytes, want %d", len(tokenBytes), sessionTokenLength)
	}
}

func TestHashSessionTokenIsStable(t *testing.T) {
	token := newSessionToken()

	first := hashSessionToken(token)
	second := hashSessionToken(token)

	// The same token must always give the same hash, or we could never find
	// the session again.
	if !bytes.Equal(first, second) {
		t.Errorf("the same token gave two different hashes")
	}
	if len(first) != 32 {
		t.Errorf("hash length: got %d bytes, want 32", len(first))
	}
}

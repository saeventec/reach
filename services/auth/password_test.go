package main

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	testCases := []struct {
		password string
		want     error
	}{
		{"purple tram winter", nil},
		{"12345678", nil},
		{strings.Repeat("a", 128), nil},
		{"شب روشن پاییزی", nil}, // any language is fine
		{"1234567", errPasswordLength},
		{strings.Repeat("a", 129), errPasswordLength},
		{"sadra.samadi", errPasswordIsUsername},
		{"Sadra.Samadi", errPasswordIsUsername},
		{"my reach password", errPasswordHasAppName},
		{"REACHme123", errPasswordHasAppName},
	}

	for _, testCase := range testCases {
		got := validatePassword(testCase.password, "sadra.samadi")
		if !errors.Is(got, testCase.want) {
			t.Errorf("validatePassword(%q): got %v, want %v", testCase.password, got, testCase.want)
		}
	}
}

// The length counts characters, not bytes: "é" is 2 bytes in UTF-8, but
// one character.
func TestValidatePasswordCountsCharacters(t *testing.T) {
	password := strings.Repeat("é", 8) // 8 characters, 16 bytes

	err := validatePassword(password, "sadra")

	if err != nil {
		t.Errorf("got %v, want no error", err)
	}
}

// "é" can be one code (U+00E9) or "e" plus a combining accent (U+0065
// U+0301). They look the same, and must become the same password.
func TestNormalizePasswordMakesLookAlikesEqual(t *testing.T) {
	oneCode := "café au lait"
	twoCodes := "café au lait"

	if normalizePassword(oneCode) != normalizePassword(twoCodes) {
		t.Errorf("the two ways to write é give different passwords")
	}
}

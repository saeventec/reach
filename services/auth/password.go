package main

// The password rules (ADR 0025). They follow NIST's advice: a minimum
// length, no "must contain a digit" rules, and every character allowed.
// Learn more: docs/learn/password-hashing.md#the-password-rules

import (
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	minimumPasswordLength = 8
	// A maximum stops someone from sending a huge "password" that takes a
	// long time to hash. NIST asks to allow at least 64.
	maximumPasswordLength = 128
)

// appName may not appear in a password: it is the first word an attacker
// would try for this app.
const appName = "reach"

var (
	errPasswordLength     = errors.New("a password must be 8 to 128 characters long")
	errPasswordIsUsername = errors.New("the password may not be the same as the username")
	errPasswordHasAppName = errors.New(`the password may not contain the word "reach"`)
)

// normalizePassword brings the password into Unicode's NFC form. Some
// letters, like "é", can be written as one code or as two ("e" plus an
// accent), and different devices type them differently. NFC picks one
// form, so the same password always gives the same hash.
// Learn more: docs/learn/password-hashing.md#unicode-normalization
func normalizePassword(typed string) string {
	return norm.NFC.String(typed)
}

// validatePassword checks a normalized password against the rules. The
// username must already be normalized too.
func validatePassword(password string, username string) error {
	length := utf8.RuneCountInString(password)
	if length < minimumPasswordLength || length > maximumPasswordLength {
		return errPasswordLength
	}

	// Compare in lowercase, so "Sadra1" is caught for the user "sadra1".
	lowercasePassword := strings.ToLower(password)
	if lowercasePassword == username {
		return errPasswordIsUsername
	}
	if strings.Contains(lowercasePassword, appName) {
		return errPasswordHasAppName
	}
	return nil
}

package main

// The username rules from ADR 0021. The users table checks the same rules
// (migrations/00001_create_users_table.sql). Checking them here first lets
// us send the user a clear message, instead of a database error.
// Learn more: docs/learn/postgresql.md#ids-and-constraints

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	minimumUsernameLength = 3
	maximumUsernameLength = 32
)

// usernamePattern is the same pattern as the username_characters
// constraint: a-z and 0-9, with single separators (. _ -) between them.
// MustCompile stops the program at startup if the pattern is invalid, so a
// typo here can never go unnoticed.
var usernamePattern = regexp.MustCompile(`^[a-z0-9]([._-]?[a-z0-9])*$`)

var (
	errUsernameLength     = errors.New("a username must be 3 to 32 characters long")
	errUsernameCharacters = errors.New("a username may only use a-z, 0-9 and single . _ - between them, and must start and end with a letter or digit")
)

// normalizeUsername turns what the user typed into the stored form:
// without spaces around it, and in lowercase. So " Sadra " becomes "sadra".
func normalizeUsername(typed string) string {
	trimmed := strings.TrimSpace(typed)
	lowercase := strings.ToLower(trimmed)
	return lowercase
}

// validateUsername checks a normalized username against the rules. The
// error message is written for the user, so it can be shown as it is.
func validateUsername(username string) error {
	// RuneCountInString counts characters, not bytes: "ä" is one
	// character but two bytes.
	length := utf8.RuneCountInString(username)
	if length < minimumUsernameLength || length > maximumUsernameLength {
		return errUsernameLength
	}

	if !usernamePattern.MatchString(username) {
		return errUsernameCharacters
	}
	return nil
}

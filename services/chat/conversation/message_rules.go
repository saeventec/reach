package conversation

// The rules for messages.

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maximumMessageLength = 4000
	defaultPageSize      = 50
	maximumPageSize      = 100
)

// uuidPattern matches a UUID in its standard text form, in lowercase:
// 8-4-4-4-12 hexadecimal digits. regexp.MustCompile builds the pattern once,
// when the program starts.
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// uuidVersion7Pattern is the same, but only for version 7: the 13th digit
// is the version ("7"), and the 17th the variant (8, 9, a or b).
var uuidVersion7Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// isUUID reports whether a text is a UUID. Postgres writes UUIDs in
// lowercase, so we compare in lowercase too.
func isUUID(text string) bool {
	lowercase := strings.ToLower(text)
	return uuidPattern.MatchString(lowercase)
}

// isUUIDVersion7 reports whether a text is a UUIDv7. A UUIDv7 starts with
// the time it was made, so newer ids sort after older ones.
func isUUIDVersion7(text string) bool {
	lowercase := strings.ToLower(text)
	return uuidVersion7Pattern.MatchString(lowercase)
}

// normalizeMessageText removes spaces and empty lines around a message,
// and checks its length. The error message is written for the user.
func normalizeMessageText(typed string) (string, error) {
	text := strings.TrimSpace(typed)
	length := utf8.RuneCountInString(text)
	if length == 0 {
		return "", invalidRequest("the message is empty")
	}
	if length > maximumMessageLength {
		return "", invalidRequest("a message can have at most 4000 characters")
	}
	return text, nil
}

// pageSize turns the requested number of messages into the real one:
// 0 means the default, and more than the maximum is refused.
func pageSize(requested int) (int, error) {
	if requested == 0 {
		return defaultPageSize, nil
	}
	if requested < 0 || requested > maximumPageSize {
		return 0, invalidRequest("list 1 to 100 messages at a time")
	}
	return requested, nil
}

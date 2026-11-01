package conversation

// The rules for conversations, as small functions without side effects,
// so they are easy to read and to test.

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// The group rules (the owner's design, milestone 3).
const (
	maximumGroupNameLength = 64
	// maximumGroupMembers counts everyone, the creator included.
	maximumGroupMembers = 50
)

// normalizeUsername turns a typed username into the stored form: without
// spaces around it, and in lowercase. It is the same rule as in the auth
// service, so chat can match what was typed with what auth answers.
func normalizeUsername(typed string) string {
	trimmed := strings.TrimSpace(typed)
	lowercase := strings.ToLower(trimmed)
	return lowercase
}

// normalizeUsernames normalizes every typed username, and drops empty
// ones and repeats, keeping the order they were typed in.
func normalizeUsernames(typed []string) []string {
	usernames := make([]string, 0, len(typed))
	for _, typedUsername := range typed {
		username := normalizeUsername(typedUsername)
		if username == "" {
			continue
		}
		if slices.Contains(usernames, username) {
			continue
		}
		usernames = append(usernames, username)
	}
	return usernames
}

// normalizeGroupName removes spaces around a typed group name, and checks
// its length. The message of the error is written for the user.
func normalizeGroupName(typed string) (string, error) {
	name := strings.TrimSpace(typed)
	// RuneCountInString counts characters, not bytes: "ä" is one
	// character but two bytes.
	length := utf8.RuneCountInString(name)
	if length == 0 {
		return "", invalidRequest("the group needs a name")
	}
	if length > maximumGroupNameLength {
		return "", invalidRequest("the group name can have at most 64 characters")
	}
	return name, nil
}

// directPairKey is the key of the 1:1 conversation between two people:
// their ids, sorted and joined. Sorting makes it the same key whoever
// starts the chat: alice with bob, or bob with alice.
func directPairKey(firstUserID string, secondUserID string) string {
	ids := []string{firstUserID, secondUserID}
	slices.Sort(ids)
	return strings.Join(ids, ":")
}

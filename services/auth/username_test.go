package main

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeUsernameTrimsAndLowercases(t *testing.T) {
	got := normalizeUsername("  Sadra.Samadi ")

	want := "sadra.samadi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A "table-driven" test: one list of cases, checked by the same loop.
// Adding a case is one line. Learn more: docs/learn/testing.md#table-driven-tests
func TestValidateUsername(t *testing.T) {
	testCases := []struct {
		username string
		want     error
	}{
		{"sadra", nil},
		{"sadra.samadi", nil},
		{"s-2", nil},
		{"sadra_s1", nil},
		{"abc", nil},
		{strings.Repeat("a", 32), nil},
		{"ab", errUsernameLength},
		{strings.Repeat("a", 33), errUsernameLength},
		{"Sadra", errUsernameCharacters},
		{".sadra", errUsernameCharacters},
		{"sadra.", errUsernameCharacters},
		{"sa..dra", errUsernameCharacters},
		{"sa._dra", errUsernameCharacters},
		{"sa dra", errUsernameCharacters},
		{"sädra", errUsernameCharacters},
	}

	for _, testCase := range testCases {
		got := validateUsername(testCase.username)
		if !errors.Is(got, testCase.want) {
			t.Errorf("validateUsername(%q): got %v, want %v", testCase.username, got, testCase.want)
		}
	}
}

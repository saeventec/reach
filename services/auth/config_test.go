package main

import "testing"

func TestReadCookieSecure(t *testing.T) {
	testCases := []struct {
		value     string
		want      bool
		wantError bool
	}{
		{"", true, false}, // not set: the safe default
		{"true", true, false},
		{"false", false, false},
		{"maybe", false, true},
	}

	for _, testCase := range testCases {
		// Setenv sets the variable for this test only, and restores it after.
		t.Setenv("COOKIE_SECURE", testCase.value)

		got, err := readCookieSecure()

		gotError := err != nil
		if got != testCase.want || gotError != testCase.wantError {
			t.Errorf("COOKIE_SECURE=%q: got (%v, error %v), want (%v, error %v)",
				testCase.value, got, err, testCase.want, testCase.wantError)
		}
	}
}

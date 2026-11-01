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

func TestReadIPRateLimitBurst(t *testing.T) {
	testCases := []struct {
		value     string
		want      int
		wantError bool
	}{
		{"", 20, false}, // not set: the safe default
		{"500", 500, false},
		{"0", 0, true},
		{"-3", 0, true},
		{"many", 0, true},
	}

	for _, testCase := range testCases {
		t.Setenv("RATE_LIMIT_BURST_PER_IP", testCase.value)

		got, err := readIPRateLimitBurst()

		gotError := err != nil
		if got != testCase.want || gotError != testCase.wantError {
			t.Errorf("RATE_LIMIT_BURST_PER_IP=%q: got (%d, error %v), want (%d, error %v)",
				testCase.value, got, err, testCase.want, testCase.wantError)
		}
	}
}

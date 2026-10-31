package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Real User-Agent headers, shortened where they repeat.
func TestDescribeUserAgent(t *testing.T) {
	testCases := []struct {
		userAgent string
		want      string
	}{
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			"Chrome on Windows",
		},
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
			"Edge on Windows",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
			"Safari on macOS",
		},
		{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:132.0) Gecko/20100101 Firefox/132.0",
			"Firefox on macOS",
		},
		{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Mobile/15E148 Safari/604.1",
			"Safari on iPhone",
		},
		{
			"Mozilla/5.0 (Linux; Android 15; Pixel 9) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36",
			"Chrome on Android",
		},
		{
			"Mozilla/5.0 (X11; Linux x86_64; rv:132.0) Gecko/20100101 Firefox/132.0",
			"Firefox on Linux",
		},
		{"curl/8.5.0", "Unknown browser on an unknown system"},
		{"", "Unknown browser on an unknown system"}, // sessions from before step 6
	}

	for _, testCase := range testCases {
		got := describeUserAgent(testCase.userAgent)
		if got != testCase.want {
			t.Errorf("describeUserAgent(%q):\ngot  %q\nwant %q", testCase.userAgent, got, testCase.want)
		}
	}
}

func TestLimitUserAgentKeepsShortOnes(t *testing.T) {
	userAgent := "curl/8.5.0"

	got := limitUserAgent(userAgent)

	if got != userAgent {
		t.Errorf("got %q, want %q", got, userAgent)
	}
}

func TestLimitUserAgentShortensLongOnes(t *testing.T) {
	// 511 ASCII bytes, then "é" (2 bytes), which a cut at 512 splits in half.
	userAgent := strings.Repeat("a", 511) + "é" + strings.Repeat("b", 100)

	got := limitUserAgent(userAgent)

	if len(got) > maximumUserAgentLength {
		t.Errorf("length: got %d bytes, want at most %d", len(got), maximumUserAgentLength)
	}
	if !utf8.ValidString(got) {
		t.Errorf("the shortened text is not valid UTF-8")
	}
}

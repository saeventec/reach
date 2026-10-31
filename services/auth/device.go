package main

// Describing devices for "your devices": a readable name like "Chrome on
// Windows", guessed from the browser's User-Agent header.
// Learn more: docs/learn/sessions.md#devices

import "strings"

// maximumUserAgentLength limits how much of the User-Agent header we store.
// Real ones are a few hundred characters; a client could send far more.
const maximumUserAgentLength = 512

// limitUserAgent shortens a User-Agent header to what we store.
func limitUserAgent(userAgent string) string {
	if len(userAgent) <= maximumUserAgentLength {
		return userAgent
	}
	shortened := userAgent[:maximumUserAgentLength]
	// Cutting at a byte count can split a multi-byte character in half.
	// ToValidUTF8 drops such a broken piece, because Postgres only stores
	// valid text.
	return strings.ToValidUTF8(shortened, "")
}

// describeUserAgent turns a User-Agent header into a short, readable name.
// It is a simple guess, not exact: browsers copy each other's names for
// historical reasons.
func describeUserAgent(userAgent string) string {
	browser := guessBrowser(userAgent)
	system := guessSystem(userAgent)
	return browser + " on " + system
}

// guessBrowser picks the browser's name. The order matters: Edge and Opera
// also say "Chrome", and Chrome also says "Safari", so the more specific
// names are checked first.
func guessBrowser(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "Edg/"):
		return "Edge"
	case strings.Contains(userAgent, "OPR/"):
		return "Opera"
	case strings.Contains(userAgent, "Firefox/"):
		return "Firefox"
	case strings.Contains(userAgent, "Chrome/"):
		return "Chrome"
	case strings.Contains(userAgent, "Safari/"):
		return "Safari"
	default:
		return "Unknown browser"
	}
}

// guessSystem picks the operating system's name. Android also says
// "Linux", and iPhones also say "like Mac OS X", so those come first.
func guessSystem(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "Android"):
		return "Android"
	case strings.Contains(userAgent, "iPhone"):
		return "iPhone"
	case strings.Contains(userAgent, "iPad"):
		return "iPad"
	case strings.Contains(userAgent, "Windows"):
		return "Windows"
	case strings.Contains(userAgent, "Mac OS X"):
		return "macOS"
	case strings.Contains(userAgent, "CrOS"):
		return "ChromeOS"
	case strings.Contains(userAgent, "Linux"):
		return "Linux"
	default:
		return "an unknown system"
	}
}

package main

// Settings only the auth service has. The ones every service has (the port,
// the database address) are in packages/go/config.
// Learn more: docs/learn/environment-variables.md

import (
	"fmt"
	"os"
	"strconv"
)

// readCookieSecure says whether cookies get the Secure flag, which tells the
// browser to send them only over HTTPS. It comes from the COOKIE_SECURE
// environment variable, and is true unless it is set to "false". So a real
// server is safe even if someone forgets the setting. On your computer there
// is no HTTPS yet, so .env sets it to false.
// Learn more: docs/learn/sessions.md#the-cookie
func readCookieSecure() (bool, error) {
	value := os.Getenv("COOKIE_SECURE")
	if value == "" {
		return true, nil
	}

	// ParseBool understands "true", "false", "1", "0" and a few more.
	secure, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("COOKIE_SECURE must be true or false, not %q", value)
	}
	return secure, nil
}

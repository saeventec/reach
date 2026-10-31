package main

// Configuration comes from environment variables, so the same program runs
// anywhere. Learn more: docs/learn/environment-variables.md

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// defaultPort is used when the PORT environment variable is empty.
const defaultPort = "8080"

// readPort returns the port to listen on, from the PORT environment variable.
func readPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return defaultPort
	}
	return port
}

// readDatabaseURL returns the database address, from the DATABASE_URL
// environment variable. It has no default: it contains a password, so it
// must always be set on purpose.
func readDatabaseURL() (string, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return "", errors.New("the DATABASE_URL environment variable is not set")
	}
	return databaseURL, nil
}

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

package main

// Settings only the chat service has. The ones every service has (the port,
// the database address) are in packages/go/config.
// Learn more: docs/learn/environment-variables.md

import (
	"errors"
	"os"
)

// readAuthURL returns the auth service's address, like "http://auth:8080",
// from the AUTH_URL environment variable. It has no default, so
// compose.yaml says plainly which service chat depends on.
func readAuthURL() (string, error) {
	authURL := os.Getenv("AUTH_URL")
	if authURL == "" {
		return "", errors.New("the AUTH_URL environment variable is not set")
	}
	return authURL, nil
}

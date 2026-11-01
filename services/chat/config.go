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

// readNATSURL returns the NATS server's address, like "nats://nats:4222",
// from the NATS_URL environment variable. Like AUTH_URL, it has no default.
func readNATSURL() (string, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		return "", errors.New("the NATS_URL environment variable is not set")
	}
	return natsURL, nil
}

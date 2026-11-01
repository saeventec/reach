// Package config reads a service's settings from environment variables, so
// the same program runs anywhere. Learn more: docs/learn/environment-variables.md
package config

import (
	"errors"
	"os"
)

// defaultPort is used when the PORT environment variable is empty.
const defaultPort = "8080"

// Port returns the port to listen on, from the PORT environment variable.
func Port() string {
	port := os.Getenv("PORT")
	if port == "" {
		return defaultPort
	}
	return port
}

// DatabaseURL returns the database address, from the DATABASE_URL
// environment variable. It has no default: it contains a password, so it
// must always be set on purpose.
func DatabaseURL() (string, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return "", errors.New("the DATABASE_URL environment variable is not set")
	}
	return databaseURL, nil
}

// NATSURL returns the NATS server's address, like "nats://nats:4222", from
// the NATS_URL environment variable. It has no default, so compose.yaml
// says plainly which services use NATS.
func NATSURL() (string, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		return "", errors.New("the NATS_URL environment variable is not set")
	}
	return natsURL, nil
}

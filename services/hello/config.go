package main

// Configuration comes from environment variables, so the same program runs
// anywhere. Learn more: docs/learn/environment-variables.md

import (
	"errors"
	"os"
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

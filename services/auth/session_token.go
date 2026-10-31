package main

// Session tokens: the random secret the browser keeps in a cookie (ADR 0022).
// Learn more: docs/learn/sessions.md

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// sessionTokenLength is the token's size in bytes. 32 random bytes (256
// bits) can't be guessed, not even with all the computers in the world.
const sessionTokenLength = 32

// newSessionToken returns a new random token, as text that is safe in a
// cookie.
func newSessionToken() string {
	tokenBytes := make([]byte, sessionTokenLength)
	// rand.Read never fails: if the system can't give random bytes, Go stops
	// the program.
	rand.Read(tokenBytes)

	// "URL" base64 uses only letters, digits, "-" and "_", which are all
	// allowed in a cookie value.
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	return token
}

// hashSessionToken returns the SHA-256 hash of a token. Only the hash is
// stored, so someone who reads the database can't use the tokens in it.
//
// Why SHA-256 here, but slow argon2id for passwords? A password is chosen by
// a person, so an attacker can try likely ones, and each try must be slow.
// A token is 32 random bytes: there is nothing likely to try, so a fast hash
// is enough.
func hashSessionToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	// Sum256 returns a fixed-size array. [:] turns it into a slice ([]byte),
	// the type the database code expects.
	return hash[:]
}

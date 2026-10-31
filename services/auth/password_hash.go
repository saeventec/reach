package main

// Storing passwords safely, with argon2id.
// Learn more: docs/learn/password-hashing.md

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// The argon2id settings: OWASP's recommended minimum (19 MiB of memory,
// 2 passes, 1 thread). Each guess then costs an attacker the same memory
// and time, which makes trying billions of passwords very expensive.
const (
	argonMemoryKiB  = 19 * 1024 // memory in KiB: 19 MiB
	argonPasses     = 2         // how many times argon2id goes over the memory
	argonThreads    = 1         // how many CPU threads it uses
	argonSaltLength = 16        // bytes: 128 bits, as RFC 9106 recommends
	argonHashLength = 32        // bytes: 256 bits, as RFC 9106 recommends
)

// hashPassword returns the argon2id hash of a normalized password, in the
// standard "PHC" text format, which keeps the settings and the salt next to
// the hash:
//
//	$argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>
//
// Because the settings are stored with each hash, they can be raised later
// without breaking the hashes that already exist.
func hashPassword(password string) string {
	// A new random salt for every password. The same password then gives a
	// different hash for every user, so an attacker can't use one list of
	// pre-computed hashes for all of them. rand.Read never fails: if the
	// system can't give random bytes, Go stops the program.
	salt := make([]byte, argonSaltLength)
	rand.Read(salt)

	hash := argon2.IDKey([]byte(password), salt, argonPasses, argonMemoryKiB, argonThreads, argonHashLength)

	// Base64 turns bytes into plain text. "Raw" leaves out the "=" padding,
	// as the PHC format asks.
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	phcString := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemoryKiB, argonPasses, argonThreads, encodedSalt, encodedHash)
	return phcString
}

package main

// Storing passwords safely, with argon2id.
// Learn more: docs/learn/password-hashing.md

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

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

// errMalformedHash means a stored hash is not in the format hashPassword
// writes. That is a bug or damaged data, never the user's fault.
var errMalformedHash = errors.New("the stored password hash is malformed")

// argonSettings are the settings stored inside one PHC string.
type argonSettings struct {
	version   int
	memoryKiB uint32
	passes    uint32
	threads   uint8
	salt      []byte
	hash      []byte
}

// verifyPassword reports whether a normalized password matches a hash made
// by hashPassword. It hashes the password again, with the salt and settings
// stored in the hash, and compares the two results.
func verifyPassword(password string, phcString string) (bool, error) {
	settings, err := parsePHCString(phcString)
	if err != nil {
		return false, err
	}

	hashLength := uint32(len(settings.hash))
	newHash := argon2.IDKey([]byte(password), settings.salt, settings.passes, settings.memoryKiB, settings.threads, hashLength)

	// ConstantTimeCompare takes the same time whether the first byte or the
	// last byte differs. A normal comparison stops at the first difference,
	// and the time it takes could tell an attacker how much was right.
	matches := subtle.ConstantTimeCompare(newHash, settings.hash) == 1
	return matches, nil
}

// parsePHCString reads the parts of "$argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>".
func parsePHCString(phcString string) (argonSettings, error) {
	// Splitting at "$" gives 6 parts. The first one is empty, because the
	// text starts with "$".
	parts := strings.Split(phcString, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return argonSettings{}, errMalformedHash
	}

	var settings argonSettings

	// Sscanf reads values out of text, following a pattern: the opposite
	// of Sprintf.
	_, err := fmt.Sscanf(parts[2], "v=%d", &settings.version)
	if err != nil || settings.version != argon2.Version {
		return argonSettings{}, errMalformedHash
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &settings.memoryKiB, &settings.passes, &settings.threads)
	if err != nil {
		return argonSettings{}, errMalformedHash
	}
	// argon2.IDKey crashes the program when passes or threads is 0. Our
	// own hashes never have 0, but damaged data should give an error.
	if settings.passes == 0 || settings.threads == 0 {
		return argonSettings{}, errMalformedHash
	}

	settings.salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argonSettings{}, errMalformedHash
	}

	settings.hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argonSettings{}, errMalformedHash
	}
	return settings, nil
}

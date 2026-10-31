# 0025. Password rules, and hashing with argon2id

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Sign-up (milestone 2, step 4) needs rules for passwords and a safe way to store them. The brief already chose argon2id. We still need its settings and the rules. We checked the current advice of OWASP (Password Storage Cheat Sheet) and NIST (SP 800-63B-4).

## Decision

- **Hashing:** argon2id with OWASP's minimum settings (19 MiB memory, 2 passes, 1 thread), a 16-byte random salt, a 32-byte hash, stored in the standard PHC string format. Library: `golang.org/x/crypto/argon2`.
- **Rules (the owner's choice):** 8 to 128 characters, counted as Unicode characters. Not the same as the username, and not containing "reach". Every character allowed. No composition rules.
- **Normalization:** the password is converted to Unicode NFC before checking and hashing.

## Why

- argon2id is the first choice of both OWASP and RFC 9106. The settings are stored with each hash, so they can be raised later.
- The owner chose a minimum of 8 so that friends can sign up easily, and a simple check instead of a blocklist, to keep the step small.
- Composition rules and periodic changes are left out, as NIST asks.

## Alternatives considered

- **15 characters minimum:** what NIST requires when the password is the only login factor, as in Reach. Safer, but stricter for friends.
- **A blocklist of common or leaked passwords:** also a NIST requirement. One more file and check.
- **bcrypt, scrypt, PBKDF2:** older or weaker against graphics cards (see [docs/learn/password-hashing.md](../learn/password-hashing.md)).

## Consequences

- **Below NIST's rules:** 8 characters and no blocklist are fine for a few friends on one computer. **Revisit both before Reach runs on a real server (milestone 11)**, or when a second factor is added.
- The minimum can be raised later for new passwords, without touching existing hashes.
- Each sign-up and login spends about 19 MiB and 25 ms on hashing. Rate limiting (step 8) keeps many logins at once from using up the server's memory.

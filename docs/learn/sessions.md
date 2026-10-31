# Sessions and tokens

## What it is

HTTP has no memory: every request stands alone. After you log in, the server needs a way to recognize your next requests without asking for the password again. A **session** is that memory: a row in the `sessions` table that says "this browser is logged in as this user, until this time".

The browser proves which session is its own with a **session token**: a long random secret, kept in a cookie. It's like a cloakroom ticket. Whoever holds the ticket gets the coat, so the ticket must be impossible to guess and must never be shown to anyone.

## Why we use it

It is the simplest safe way to stay logged in on the web, and logging out really works: delete the row, and the token is worthless ([ADR 0022](../decisions/0022-keep-sessions-in-httponly-cookies.md)).

## How a session works in Reach

1. **Log in:** the password is right, so the service creates a token: 32 random bytes from `crypto/rand`, written as text (URL-safe base64).
2. It stores the token's **SHA-256 hash** in `sessions`, with the user and the expiry time, and sends the token itself to the browser in a cookie.
3. **Every later request:** the browser sends the cookie. The service hashes the token again and looks for a session with that hash that hasn't expired.
4. **Renewed on use:** finding the session also moves its expiry 30 days ahead, in the same SQL statement (`RenewSession` in [queries/sessions.sql](../../services/auth/queries/sessions.sql)). So you stay logged in while you use Reach, and are logged out after 30 days of not using it.

## Why only the hash is stored

If someone reads the database (a stolen backup, a bug), they find only hashes. A hash can't be turned back into the token, and the browser must send the token itself, so the hashes are useless to them. It is the same idea as for passwords.

**Why a fast hash (SHA-256) and not argon2id?** argon2id is slow on purpose, because people choose passwords and attackers can try likely ones. A token is 32 random bytes: there are 2²⁵⁶ possible tokens, and none is more likely than another. Even a fast hash can't be searched, so slow hashing would only waste time on every request.

## Tiny example

From [services/auth/session_token.go](../../services/auth/session_token.go):

```go
token := newSessionToken()           // "kQ3v...": goes into the cookie
tokenHash := hashSessionToken(token) // 32 bytes: goes into the database
```

## Foreign keys

```sql
user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
```

`REFERENCES users (id)` is a **foreign key**: the database refuses a session whose `user_id` is not an existing user. `ON DELETE CASCADE` means: when a user is deleted, their sessions are deleted too, so no session can point to a missing user.

## Alternatives

- **JWT (JSON Web Token):** a signed token the server doesn't store. It can't be cancelled before it expires (see ADR 0022).
- **A fixed expiry** (for example, 30 days after login, no renewal): simpler, but active users get logged out. The owner chose renewal.
- **Server-side sessions in memory:** fast, but lost on restart and not shared between copies of the service. The brief asks for stateless services.

## Learn more

- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)

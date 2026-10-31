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

## The cookie

After a successful login, the service answers with a `Set-Cookie` header ([services/auth/session_cookie.go](../../services/auth/session_cookie.go)):

```
Set-Cookie: reach_session=kQ3v...; Path=/; Max-Age=2592000; HttpOnly; SameSite=Strict
```

From then on, the browser adds `Cookie: reach_session=kQ3v...` to every request to our site, by itself. Each part of the `Set-Cookie` line:

| Part | What it means | Why |
| --- | --- | --- |
| `reach_session=kQ3v...` | the name and the token | |
| `Path=/` | send it to every path of the site | the API paths start with `/auth.v1.`, `/hello.v1.`, ... |
| `Max-Age=2592000` | keep it for 30 days (in seconds) | matches the session. `GetCurrentUser` sends the cookie again, so the browser's 30 days restart too. |
| `HttpOnly` | the page's JavaScript can't read it | an attacker's script on our page can't steal the token |
| `SameSite=Strict` | only send it with requests our own site starts | see CSRF below |
| `Secure` (on a real server) | only send it over HTTPS | on public networks, plain HTTP can be read by others |

**CSRF** ("cross-site request forgery"): another website could make your browser send a request to Reach. Without protection, the browser would add your cookie, and the request would run as you. `SameSite=Strict` tells the browser to never add the cookie to requests that another site starts. Our React app runs on our own site, so its requests always get the cookie.

**`COOKIE_SECURE`:** your computer has no HTTPS yet (`http://localhost:8080`). Chrome and Firefox accept `Secure` cookies on `localhost` anyway, but other browsers (like Safari) may not, and we haven't tested them. So `.env.example` sets `COOKIE_SECURE=false`. If the setting is missing, it is **true**: a real server is safe even if someone forgets it. HTTPS comes in milestone 10.

**The token never appears in a response body.** It is only in the cookie, so the page's code never sees it. `LogIn` and `GetCurrentUser` return only the user's id and name.

## Logging in safely

- **The same message for a wrong username and a wrong password:** "wrong username or password". Otherwise, login could be used to test which usernames exist.
- **The same time, too:** checking a password takes about 15 to 25 ms. If an unknown username were answered right away, the speed would tell. So for an unknown username, the service checks the password against a dummy hash anyway (`dummyPasswordHash` in [services/auth/session.go](../../services/auth/session.go)).
- **No hashing of huge inputs:** a password longer than 128 characters can't be right, so it is refused before hashing.

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

## Devices

On the web, each login is one **device**: one row in `sessions`. Logging in on your laptop and on your phone gives two devices. (Later, the native apps will have real device records with their own keys. The brief asks the data model to allow many devices per user from the start, and this is the first part of that.)

Each session remembers:

- **`user_agent`:** the browser's `User-Agent` header, sent with every request. For example: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36`. It is stored as sent (up to 512 bytes).
- **`last_used_at`:** updated by `RenewSession`, so on every use.

"your devices" shows a readable name instead of the raw text: `describeUserAgent` in [services/auth/device.go](../../services/auth/device.go) makes "Chrome on Windows" with a few plain `strings.Contains` checks. The order of the checks matters: for historical reasons, Edge's text also contains "Chrome", and Chrome's also contains "Safari", so the more specific names are checked first. It is a guess, not exact, and a client can send any text it likes. That's fine for a name in a list, but never use it for security decisions.

The name is worked out when the list is shown, not stored. So when the guessing gets better, older devices get better names too.

**An index** (`sessions_user_id_index`) makes "all sessions of this user" fast. Without it, Postgres would read every row of the table to find them. Indexes cost a little space and a little time on each insert, so we add one only when a query needs it.

## Not done yet

- **Old sessions stay in the table** after they expire. They are harmless (`RenewSession` ignores them), but the table grows. Deleting them regularly is a "background job", a later topic.

## Alternatives

- **JWT (JSON Web Token):** a signed token the server doesn't store. It can't be cancelled before it expires (see ADR 0022).
- **A fixed expiry** (for example, 30 days after login, no renewal): simpler, but active users get logged out. The owner chose renewal.
- **Server-side sessions in memory:** fast, but lost on restart and not shared between copies of the service. The brief asks for stateless services.

## Learn more

- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)

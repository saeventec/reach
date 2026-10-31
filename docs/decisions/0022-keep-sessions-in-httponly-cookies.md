# 0022. Keep sessions in HttpOnly cookies

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

After logging in, the browser must prove on every request who it is, without sending the password again. We need to decide what that proof is and where the browser keeps it.

## Decision

At login, the `auth` service creates a long **random session token**. The browser keeps it in a cookie marked **HttpOnly**, and sends it back automatically with every request. The database stores only a **hash** of the token, together with the user and the device. Logging out deletes that row.

## Why

- **HttpOnly** means the page's JavaScript can't read the cookie. If an attacker ever manages to run script on our page, they still can't steal the token.
- **Logging out really works:** once the row is gone, the token is worthless. This is also what "log out another device" needs.
- **Storing only a hash** means that someone who reads the database can't use the tokens in it. It is the same idea as not storing passwords.
- It is simple: no signing keys, no token formats.

## Alternatives considered

- **JWT (JSON Web Token):** a signed token the server doesn't need to store. It can't be cancelled before it expires, so "log out a device" needs an extra block list anyway.
- **Token in `localStorage`, sent in a header by our code:** common in tutorials, but any script on the page can read and steal it.
- **Short access token plus refresh token:** useful when many separate servers check tokens. More moving parts than we need now.

## Consequences

- Cookies are sent automatically, so we must protect against requests from other websites (CSRF). We set the cookie's `SameSite` rule when we build login, and explain it then.
- Every request that needs a logged-in user asks the `auth` database whether the token is valid. That is fine at our size.
- The exact cookie settings (`Secure`, `SameSite`, how long a session lives) are decided and explained in step 5.

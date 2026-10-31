# Rate limiting

## What it is

A **rate limit** caps how many times something may happen in a period of time. For example: at most 5 wrong passwords for one username per 15 minutes. Tries over the limit are refused with "too many attempts, please try again later" (the Connect code `resource_exhausted`, HTTP 429).

## Why we use it

Without a limit, an attacker can try thousands of passwords per minute for one account (**brute force**), or create accounts in bulk. argon2id makes each guess slower, but a limit stops it after a handful of tries ([ADR 0027](../decisions/0027-rate-limit-logins-in-memory.md)).

## The token bucket

We use `golang.org/x/time/rate`, made by the Go team. It works like a bucket of tokens:

- Each key (an IP address, or a username) has a bucket that holds at most a few tokens: the **burst**.
- Every try takes one token out. An empty bucket means: refused.
- Tokens drip back in at a fixed speed, up to the burst.

```
username bucket: holds 5, one token back every 3 minutes
5 wrong passwords in a row → empty → refused
3 minutes later → 1 token → one more try
15 minutes later → full again
```

## Our limits

From [services/auth/rate_limit.go](../../services/auth/rate_limit.go):

| Limit | Burst | Refill | Counts |
| --- | --- | --- | --- |
| Per IP address | 20 | one every 3 s (20/min) | every login and every sign-up |
| Per username | 5 | one every 3 min (5 per 15 min) | only **failed** logins |

- **Per address** slows one computer trying many accounts, or signing up in bulk.
- **Per username** slows guessing one account's password from many computers. It only counts failures, so a person who types the right password is never slowed down.
- An attacker could fill a username's bucket on purpose, so the real owner has to wait a few minutes. That is why it is a slow-down, not a lock: the account is never blocked for long.

## In memory

The buckets live in a Go `map` in the service's memory:

- **A mutex** (`sync.Mutex`, "mutual exclusion") protects the map. Requests run at the same time in separate goroutines, and a Go map breaks if two goroutines change it at once. `Lock()` waits until no other goroutine holds the lock; `defer Unlock()` releases it when the function returns.
- **Cleanup:** every minute, a goroutine removes buckets that are full again. A full bucket behaves exactly like a new one, so nothing is lost, and the map doesn't keep every address ever seen. It stops when the service stops (`ctx.Done()`).
- **The cost:** the counts are lost on a restart, and each copy of the service has its own. The brief wants services that keep no *important* data in memory; a few minutes of counts is not important. When Reach runs several copies (milestone 11), the counts move to a shared store like Valkey.

## The client's address

Every request reaches the service through Caddy, so the connection always comes from Caddy. Caddy puts the real client's address in the `X-Forwarded-For` header, and **replaces** any `X-Forwarded-For` a client sends itself, so a client can't fake its address. We tested it: 21 logins, each with a different fake address, and the 21st was still refused. `clientIP` in [services/auth/client_address.go](../../services/auth/client_address.go) reads the last address in the header, the one our gateway added.

**On your computer**, all requests arrive from `172.18.0.1`: Docker's own gateway address, which is how your computer looks from inside Docker's network. So locally everyone shares one address limit. On a real server (milestone 10), we must make sure Docker passes the real client addresses through, or everyone would share one limit there too.

## Alternatives

- **Counts in Postgres:** shared by all copies and kept on restart, but a database write on every login and a cleanup job.
- **Counts in Valkey (or Redis):** shared and fast, made for this. One more container to run; a good fit once there are several copies.
- **In the gateway (a Caddy plugin):** stops requests before they reach the service, but needs a custom Caddy build.
- **Locking an account after N failures:** simple, but anyone can lock anyone out by typing wrong passwords.

## Learn more

- [The rate package](https://pkg.go.dev/golang.org/x/time/rate)
- [OWASP: Blocking brute-force attacks](https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks)

# Errors and responses

How the program reports success and failure to callers.

## One response shape for everything

- **What:** every response, success or failure, has the same envelope: a short `code`, a human `message`, and a `payload`.
- **Why:** clients handle every response the same way. Even small side projects use the same shape, so it becomes a habit.
- **Example:**

```
{ "code": "ok",          "message": "",                 "payload": { ... } }
{ "code": "bad-request", "message": "invalid username", "payload": null }
```

## A small set of named codes

- **What:** response codes are a short, named list. Each pairs a readable key with a status.
- **Why:** the list documents every outcome the API can return.
- **Example:**

```
ok           → 200
bad-request  → 400
unauthorized → 401
forbidden    → 403
not-found    → 404
internal     → 500
```

## Named errors per feature

- **What:** each feature declares its expected failures as named values in its `error` file, each with a code and a message written for people.
- **Why:** callers can check for a specific error, and messages are consistent.
- **Example:**

```
ErrDuplicateUser  = failure(bad-request,  "duplicate username")
ErrWrongLogin     = failure(unauthorized, "wrong credentials")
ErrNotActive      = failure(forbidden,    "not active")
```

## One error type that wraps the cause

- **What:** one failure type carries a code, a message and the original error underneath. It supports "unwrapping", so code can still check what caused it.
- **Why:** the response is built from the code and message, while the cause is kept for logs.

## Send responses through one helper

- **What:** handlers never build responses by hand. They call `sendOk(payload)` or `sendError(err)`. The error helper turns any error into the response shape, and unknown errors become `internal`.
- **Why:** one place decides how errors look. Handlers stay short.
- **Note:** for `internal` errors, log the details but send the client only a general message. Details can reveal how the system works inside.

## A central error handler

- **What:** a middleware runs after every request, logs every error that was recorded, and sends the last one as the response.
- **Why:** no error is silently lost.

## Same metrics for every request

- **What:** a middleware counts every request and measures its duration, labeled by method, route and status, and exposes them on a protected metrics endpoint.
- **Why:** you can see traffic and slow routes without touching feature code.

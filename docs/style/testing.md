# Testing

How the code is checked.

## Tests next to the code

- **What:** each package has its tests in the same folder.
- **Why:** you find a part's tests where the part is, and tests can reach its internals when needed.

## Test the building blocks alone

- **What:** every building block (config, log, cache, auth, web, database helpers, metrics) has its own small tests that don't need the rest of the app.
- **Why:** when one breaks, you know exactly where.

## Hand-written fakes

- **What:** tests replace real dependencies with small hand-written fakes that implement the same interface: a fake repository, a fake user, a fake auth guard that always lets a test user in.
- **Why:** tests are fast, need no database, and the fake is short enough to read.
- **Example:**

```
fakeRepository:
  save(record)  → give it id 1, return ok
  find(id)      → return the prepared record
  delete(record) → return ok
```

## Ready-made test helpers in each block

- **What:** each building block ships its own test helpers in a `mock` file: a test server, a fake auth, a request helper.
- **Why:** writing a new test takes a few lines. Every test uses the same tools.

## Test through the real request path

- **What:** feature tests send a real request to an in-memory server (with the real routes and middleware) and check the response.
- **Why:** the test checks what a client sees: status, code, message, payload.
- **Example:**

```
server   = testServer()
response = request(server, POST "/tags/", { name: "Tag" })
check status   == 200
check code     == "ok"
check payload.name == "Tag"
```

## Plug test-only parts in through the same interfaces

- **What:** tests create tiny parts that implement the app's plug-in interfaces (a fake router, a fake interceptor) to test the building block that uses them.
- **Why:** proves the plug-in points work, without any real feature.

## Every check says what it checks

- **What:** each assertion has a short label, like `"Status"` or `"Cache should be empty"`.
- **Why:** a failing test tells you what went wrong without reading the code.

## Test the edges

- **What:** check the limits as well as the normal case: a page size that's too big, a negative page number, a missing config file, a cache miss.
- **Why:** bugs live at the edges.

## One test per action

- **What:** one test per operation (create, list, update, delete), each named after it.
- **Why:** the test list reads like a feature list.

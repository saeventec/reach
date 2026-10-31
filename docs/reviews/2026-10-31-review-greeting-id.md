# Review: Adding an id to the greeting response

- **Date:** 2026-10-31
- **Commit or change:** the owner's exercise from step 4: add `id` to `GetGreetingResponse`, reviewed by the agent

## Feedback

### 1. Field numbers changed

- **Question or comment:** the new `id` field took number 1, and `message` moved to number 2. Field numbers are what travel over the network, so an old client would read the `id` as if it were the message. `buf breaking --against '.git#branch=main'` reports it.
- **Answer or fix:** kept `message = 1` and gave the new field the next free number: `id = 2`. Nothing used the API yet, so nothing broke, but this is the habit to build now.

### 2. Three return values

- **Question or comment:** `NewestGreeting` returned `(int64, string, error)`. Each new column would add one more return value.
- **Answer or fix:** added a small `greeting` struct (`id`, `message`). `NewestGreeting` now returns `(greeting, error)`, and the fake store and tests use it too.

### 3. The HTTP test didn't check the id

- **Question or comment:** `TestGetGreetingOverHTTP` only checked the message.
- **Answer or fix:** it now checks the id too. In JSON, Protobuf sends `int64` as a string (`"id":"2"`), so the test reads it with the `json:"id,string"` tag.

### 4. Everything else was right

- **Question or comment:** the query, `Scan`, error handling, the fake store, the test, and regenerating the code with `task generate`.
- **Answer or fix:** no changes needed. One comment was updated to say the query reads the whole row.

## What I learned

- Never change or reuse a Protobuf field number. New fields get the next free number, and `buf breaking` checks it.
- When a function returns several related values, group them in a struct.
- Protobuf JSON sends 64-bit numbers as strings, because JavaScript can't hold them exactly.

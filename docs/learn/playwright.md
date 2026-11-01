# Playwright: end-to-end tests

## What it is

Playwright is a tool that drives a real browser from a test: it opens pages, types, clicks, and checks what is on screen, like a very fast and patient person. An **end-to-end test** uses the whole system the way a user does: the browser, the gateway, every service, NATS and the database. Nothing is faked.

## Why we use it

Our other tests each check one part: unit tests (one function, with fakes), integration tests (a store with a real Postgres). Only an end-to-end test checks that the parts work *together*: that alice's message really travels browser → gateway → chat → outbox → NATS → realtime → WebSocket → bob's browser. Real-time chat needs two people at once, and Playwright can open several separate browser windows in one test. See [ADR 0034](../decisions/0034-use-playwright-for-end-to-end-tests.md).

They are slow (seconds, not milliseconds) and need everything running, so we keep them few: the main user stories. The many small cases stay in unit tests. That's the **test pyramid**: many fast unit tests at the bottom, fewer integration tests, a few end-to-end tests at the top.

## Running them

```
task test:e2e
```

It needs Docker running. The first run downloads Chromium (the open-source base of Chrome), about 150 MB, into a folder outside the repo. On WSL2 (Linux), the browser also needs some system libraries; install them once with `npx --prefix apps/web playwright install-deps chromium` (it asks for your password, because it uses `sudo apt`).

### The test system

`task test:e2e` (in [Taskfile.yml](../../Taskfile.yml)) does four things:

1. Downloads the browser, if it isn't there yet.
2. Starts a separate copy of the whole system: `docker compose -p reach-e2e … up --build --wait`.
   - `-p reach-e2e` is its own Compose **project**: its own containers and its own volumes, so an empty database. Your `task dev` data is never touched.
   - [apps/web/e2e/compose.e2e.yaml](../../apps/web/e2e/compose.e2e.yaml) changes `compose.yaml` a little: no published ports, except the gateway on **8180**. So the tests can run while `task dev` uses 8080.
   - `--wait` returns once every service is healthy and the migrations have finished.
3. Runs the tests: `playwright test`, with the settings in [apps/web/playwright.config.ts](../../apps/web/playwright.config.ts).
4. Stops the copy and deletes its volumes (`defer`: this runs even when a test fails).

CI runs the same task on every push, and keeps the traces of failed tests for 7 days (see [ci.md](ci.md)).

## Tiny example

From [apps/web/e2e/chat.e2e.ts](../../apps/web/e2e/chat.e2e.ts), shortened:

```ts
test("alice and bob chat live", async ({ browser }) => {
  const alice = await signUp(browser, aliceName); // a browser window each
  const bob = await signUp(browser, bobName);

  await startChat(alice, bobName);
  await openChat(alice, bobName);
  await send(alice, "Hi bob!");

  // bob's list updates live; alice sees ✓✓ when bob's browser got it.
  await expect(bob.getByText(`${aliceName}: Hi bob!`)).toBeVisible();
  await expect(alice.getByLabel("Delivered")).toBeVisible();
});
```

- **`browser.newContext()`** is a separate browser window with its own cookies, so alice and bob are logged in separately.
- **Locators** find things the way a person does: `getByRole("button", { name: "Send" })`, `getByLabel("Message")`, `getByText("Hi bob!")`. They don't depend on CSS classes, which change often. This also checks that the page works with screen readers, which use the same roles and labels.
- **`expect(…).toBeVisible()` waits** (up to 5 seconds) until it is true. So a test never needs `sleep`: it waits exactly as long as the live update takes.
- **Strict mode:** a locator that matches two elements is an error, not a guess. Our first try failed this way: `getByLabel("Username")` also matched "Members (usernames, …)". `{ exact: true }` fixed it.

## Our tests

| Test | Checks |
| --- | --- |
| alice and bob chat live | a new chat appears in bob's list live; the message arrives; alice sees ✓✓ then "Read"; bob's answer arrives on alice's open page |
| a group of three | a group appears for everyone live; a message reaches both others; "Read" needs everyone |
| a reload | after reloading, each message is shown once |
| the gateway ([gateway.e2e.ts](../../apps/web/e2e/gateway.e2e.ts)) | without a session, a forged `X-User-Id` gets 401, for chat and for `/realtime`; with a session, the gateway replaces a forged header with the real user's id |

The gateway tests use Playwright's **`request` fixture**: plain HTTP calls, without a page, that keep cookies like a browser. They are proven too: with the session check taken off the `/chat.v1.*` route, both forged-header tests fail.

Each test makes its own usernames with a random suffix (like `alice-3f9a01c2`), so tests never share accounts, also when run again.

## Things we learned

- **The tests found nothing wrong in the app**, but the first tries found two mistakes in the tests: the "Username" locator, and reusing usernames when a test is repeated (`--repeat-each`).
- **The rate limit counts us too.** auth allows 20 sign-ups and logins at once per address ([rate-limiting.md](rate-limiting.md)), and each person in a test uses 2. The 8 people in today's tests use 16. More tests need a plan, for example: a setting for the limit that the test system raises; or signing up a few people once and reusing their logins (Playwright's "storage state").
- **Checking for flaky tests:** a **flaky** test sometimes passes and sometimes fails with the same code. We ran the suite 5 times in a row (restarting auth in between, for the limit): 15 out of 15 passed.

## Alternatives

- **Cypress:** popular and friendly, but one browser tab per test, so two users at once is awkward.
- **Selenium / WebDriver:** the oldest standard, works with every browser, slower and more setup.
- **Testing by hand:** no setup, but slow, and easy to forget a case.

## Learn more

- https://playwright.dev/docs/intro
- https://playwright.dev/docs/locators
- https://playwright.dev/docs/browser-contexts
- https://playwright.dev/docs/ci

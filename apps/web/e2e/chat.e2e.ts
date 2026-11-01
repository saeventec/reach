// End-to-end tests for chat: real browsers, against the whole system
// (gateway, auth, chat, NATS, realtime, Postgres), started by "task
// test:e2e". Each person gets their own "browser context": like a separate
// browser window with its own cookies, so they are logged in separately.
// Learn more: docs/learn/playwright.md
import { expect, test } from "@playwright/test";
import type { Browser, Page } from "@playwright/test";

// newSuffix returns 8 random hexadecimal digits, like "3f9a01c2". Each
// test makes its own, so no two tests (or two runs, or a test repeated
// with --repeat-each) ever sign up the same username.
function newSuffix(): string {
  return crypto.randomUUID().slice(0, 8);
}

// Careful: auth allows 20 sign-ups and logins at once per address, then
// one more every 3 seconds (services/auth/rate_limit.go). Each person here
// uses 2 (the sign-up, then the login it does by itself), so the 7 people
// in this file use 14. More tests need a plan: see docs/learn/playwright.md.

// signUp opens a new browser window for a new person, creates their
// account, and waits for their chats page.
async function signUp(browser: Browser, username: string): Promise<Page> {
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.goto("/signup");
  await page.getByLabel("Username", { exact: true }).fill(username);
  await page.getByLabel("Password").fill("a long test password");
  await page.getByRole("button", { name: "Create account" }).click();
  await expect(page.getByRole("heading", { name: "Chats" })).toBeVisible();
  return page;
}

// startChat starts a 1:1 chat from the chats page. exact: the label must be
// exactly "Username"; without it, the group form's "Members (usernames, …)"
// matches too, because getByLabel looks for the text inside the label.
async function startChat(page: Page, username: string): Promise<void> {
  await page.getByLabel("Username", { exact: true }).fill(username);
  await page.getByRole("button", { name: "Start chat" }).click();
}

// openChat clicks a conversation in the list.
async function openChat(page: Page, title: string): Promise<void> {
  await page.getByRole("link", { name: new RegExp(title) }).click();
  await expect(page.getByRole("heading", { name: title })).toBeVisible();
}

// send types a message and sends it.
async function send(page: Page, text: string): Promise<void> {
  await page.getByLabel("Message").fill(text);
  await page.getByRole("button", { name: "Send" }).click();
}

test("alice and bob chat live, and alice sees delivered, then read", async ({
  browser,
}) => {
  const suffix = newSuffix();
  const aliceName = `alice-${suffix}`;
  const bobName = `bob-${suffix}`;
  const alice = await signUp(browser, aliceName);
  const bob = await signUp(browser, bobName);

  // alice starts a chat with bob, and opens it.
  await startChat(alice, bobName);
  await openChat(alice, bobName);

  // bob's list shows the new chat without a reload (a live event).
  await expect(
    bob.getByRole("link", { name: new RegExp(aliceName) }),
  ).toBeVisible();

  await send(alice, "Hi bob!");

  // bob's list shows the message live. It reached bob's browser, so it
  // counts as delivered: alice sees ✓✓.
  await expect(bob.getByText(`${aliceName}: Hi bob!`)).toBeVisible();
  await expect(alice.getByLabel("Delivered")).toBeVisible();

  // bob opens the chat: the message is read, and alice sees it live.
  await openChat(bob, aliceName);
  await expect(bob.getByText("Hi bob!")).toBeVisible();
  await expect(alice.getByLabel("Read")).toBeVisible();

  // bob answers; it arrives on alice's open page.
  await send(bob, "Hi alice!");
  await expect(alice.getByText("Hi alice!")).toBeVisible();
});

test("a group of three: everyone sees a message live", async ({ browser }) => {
  const suffix = newSuffix();
  const aliceName = `alice-${suffix}`;
  const bobName = `bob-${suffix}`;
  const carolName = `carol-${suffix}`;
  const alice = await signUp(browser, aliceName);
  const bob = await signUp(browser, bobName);
  const carol = await signUp(browser, carolName);
  const groupName = `Friends ${suffix}`;

  // carol creates the group with alice and bob.
  await carol.getByLabel("Group name").fill(groupName);
  await carol
    .getByLabel("Members (usernames, separated by commas)")
    .fill(`${aliceName}, ${bobName}`);
  await carol.getByRole("button", { name: "Create group" }).click();

  // alice and bob see the new group without a reload, and open it.
  await openChat(alice, groupName);
  await openChat(bob, groupName);
  await openChat(carol, groupName);

  await send(carol, "Pizza tonight?");

  await expect(alice.getByText("Pizza tonight?")).toBeVisible();
  await expect(bob.getByText("Pizza tonight?")).toBeVisible();
  // Everyone has the group open, so carol's message is read by all.
  await expect(carol.getByLabel("Read")).toBeVisible();
});

test("a reload shows the same messages, each once", async ({ browser }) => {
  const suffix = newSuffix();
  const aliceName = `alice-${suffix}`;
  const bobName = `bob-${suffix}`;
  const alice = await signUp(browser, aliceName);
  await signUp(browser, bobName);

  await startChat(alice, bobName);
  await openChat(alice, bobName);
  await send(alice, "first");
  await send(alice, "second");
  await expect(alice.getByText("second")).toBeVisible();

  await alice.reload();

  await expect(alice.getByText("first")).toHaveCount(1);
  await expect(alice.getByText("second")).toHaveCount(1);
});

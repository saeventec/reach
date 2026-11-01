// End-to-end tests for the gateway's session check (ADR 0029): chat and
// realtime trust the X-User-Id header, so only the gateway may set it.
// These tests send plain HTTP requests, without a browser page, through
// Playwright's "request" fixture: a small HTTP client that keeps cookies,
// like a browser. Learn more: docs/learn/caddy.md#checking-the-session-at-the-gateway
import { expect, test } from "@playwright/test";

// A made-up user id that someone could try to pretend to be.
const someoneElsesId = "01a00000-0000-7000-8000-0000000000ff";

// chatPath is the address of one chat API call. A Connect call is a JSON
// request sent with HTTP POST to /package.Service/Method.
function chatPath(service: string, method: string): string {
  return `/chat.v1.${service}/${method}`;
}

test("without a session, a forged user header is refused", async ({
  request,
}) => {
  const response = await request.post(
    chatPath("ConversationService", "ListConversations"),
    {
      headers: { "X-User-Id": someoneElsesId },
      data: {},
    },
  );

  // 401: the gateway's session check answered, and chat never saw it.
  expect(response.status()).toBe(401);
});

test("without a session, the live updates are refused", async ({ request }) => {
  const response = await request.get("/realtime", {
    headers: { "X-User-Id": someoneElsesId },
  });

  expect(response.status()).toBe(401);
});

test("with a session, a forged user header is replaced by the real one", async ({
  request,
}) => {
  // Sign up and log in; the request fixture keeps the session cookie.
  // (One person: 2 of auth's 20 sign-ups and logins per run.)
  const username = `mallory-${crypto.randomUUID().slice(0, 8)}`;
  const account = { username, password: "a long test password" };
  const signUp = await request.post("/auth.v1.AccountService/SignUp", {
    data: account,
  });
  expect(signUp.ok()).toBe(true);
  const logIn = await request.post("/auth.v1.SessionService/LogIn", {
    data: account,
  });
  expect(logIn.ok()).toBe(true);

  // mallory ("the attacker" in security examples) claims to be someone
  // else, and starts a chat with mallory's own username.
  const response = await request.post(
    chatPath("ConversationService", "StartDirectConversation"),
    {
      headers: { "X-User-Id": someoneElsesId },
      data: { username },
    },
  );

  // If chat believed the forged header, the caller would be "someone else",
  // and a chat with mallory would be created. The gateway put mallory's own
  // id in its place, so chat sees mallory starting a chat with mallory.
  expect(response.status()).toBe(400);
  const body = await response.json();
  expect(body.message).toBe("you can't start a chat with yourself");
});

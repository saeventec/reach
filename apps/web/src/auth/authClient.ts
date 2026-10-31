// Talks to the auth service (AccountService and SessionService), with code
// generated from proto/auth/v1/. Learn more: docs/learn/connect-rpc.md
//
// The session token travels in an HttpOnly cookie that this code can't
// read. The browser adds it to every call by itself, because the calls go
// to the same address as the page. Learn more: docs/learn/sessions.md
import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AccountService } from "../gen/auth/v1/account_pb.ts";
import { SessionService } from "../gen/auth/v1/session_pb.ts";

// A user, as the rest of the web app sees it.
export type User = {
  id: string;
  username: string;
};

// All the calls the pages need. The app passes the real ones (authApi
// below); tests pass fakes. Learn more: docs/learn/testing.md#fakes
export type AuthApi = {
  signUp: (username: string, password: string) => Promise<void>;
  logIn: (username: string, password: string) => Promise<User>;
  // Returns null when nobody is logged in.
  getCurrentUser: () => Promise<User | null>;
  logOut: () => Promise<void>;
};

const transport = createConnectTransport({ baseUrl: window.location.origin });
const accountClient = createClient(AccountService, transport);
const sessionClient = createClient(SessionService, transport);

async function signUp(username: string, password: string): Promise<void> {
  await accountClient.signUp({ username, password });
}

async function logIn(username: string, password: string): Promise<User> {
  const response = await sessionClient.logIn({ username, password });
  return { id: response.userId, username: response.username };
}

async function getCurrentUser(): Promise<User | null> {
  try {
    const response = await sessionClient.getCurrentUser({});
    return { id: response.userId, username: response.username };
  } catch (error) {
    // "Not logged in" is a normal answer here, not a failure.
    const connectError = ConnectError.from(error);
    if (connectError.code === Code.Unauthenticated) {
      return null;
    }
    throw error;
  }
}

async function logOut(): Promise<void> {
  await sessionClient.logOut({});
}

// The real API calls, for the app.
export const authApi: AuthApi = {
  signUp,
  logIn,
  getCurrentUser,
  logOut,
};

// The error codes whose messages the auth service writes for people, like
// "the username is already taken". Other errors (like "internal") only get a
// general message, because their text is not meant for users.
const codesWithReadableMessages = [
  Code.InvalidArgument,
  Code.AlreadyExists,
  Code.Unauthenticated,
  Code.NotFound,
];

// errorMessage turns an error from an API call into a sentence to show.
export function errorMessage(error: unknown): string {
  const connectError = ConnectError.from(error);
  if (codesWithReadableMessages.includes(connectError.code)) {
    // rawMessage is the service's message without the "[code]" prefix.
    return capitalize(connectError.rawMessage);
  }
  return "Something went wrong. Please try again.";
}

// capitalize makes the first letter uppercase: "wrong password" becomes
// "Wrong password".
function capitalize(text: string): string {
  const firstLetter = text.charAt(0).toUpperCase();
  const rest = text.slice(1);
  return firstLetter + rest;
}

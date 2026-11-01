// Turning an error from an API call into a sentence to show. Every feature
// (auth, chat) uses it. Learn more: docs/learn/connect-rpc.md
import { Code, ConnectError } from "@connectrpc/connect";

// The error codes whose messages our services write for people, like
// "the username is already taken". Other errors (like "internal") only get a
// general message, because their text is not meant for users.
const codesWithReadableMessages = [
  Code.InvalidArgument,
  Code.AlreadyExists,
  Code.Unauthenticated,
  Code.NotFound,
  Code.ResourceExhausted,
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

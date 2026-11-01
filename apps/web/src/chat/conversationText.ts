// Small functions that turn conversations into text for the screen, and
// the "new group" form's text into usernames. Kept apart from the pages,
// so they are easy to test.
import type { Conversation } from "./chatClient.ts";

// conversationTitle is the name to show: the group's name, or, for a 1:1
// chat, the other person's username.
export function conversationTitle(
  conversation: Conversation,
  myUserId: string,
): string {
  if (conversation.kind === "group") {
    return conversation.groupName;
  }
  const otherPerson = conversation.members.find(
    (member) => member.userId !== myUserId,
  );
  // Only if the data is broken: a 1:1 chat always has someone else.
  if (otherPerson === undefined) {
    return "Unknown";
  }
  return otherPerson.username;
}

// lastMessagePreview is the line under the title: who wrote the newest
// message, and what. Like "You: see you at 8" or "bob: hi!".
export function lastMessagePreview(
  conversation: Conversation,
  myUserId: string,
): string {
  const lastMessage = conversation.lastMessage;
  if (lastMessage === null) {
    return "No messages yet";
  }
  if (lastMessage.senderId === myUserId) {
    return `You: ${lastMessage.text}`;
  }
  const sender = conversation.members.find(
    (member) => member.userId === lastMessage.senderId,
  );
  const senderName = sender === undefined ? "Someone" : sender.username;
  return `${senderName}: ${lastMessage.text}`;
}

// parseUsernames splits what was typed in the "members" field, like
// "bob, carol  dave", into ["bob", "carol", "dave"]. Commas and spaces both
// separate names; empty parts are dropped.
export function parseUsernames(typed: string): string[] {
  // A regular expression: one or more commas or whitespace characters.
  const separators = /[,\s]+/;
  const parts = typed.split(separators);
  const usernames = parts.filter((part) => part !== "");
  return usernames;
}

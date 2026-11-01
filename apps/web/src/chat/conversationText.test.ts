import { expect, test } from "vitest";
import type { Conversation } from "./chatClient.ts";
import {
  conversationTitle,
  lastMessagePreview,
  parseUsernames,
} from "./conversationText.ts";

const aliceId = "01a00000-0000-7000-8000-00000000000a";
const bobId = "01a00000-0000-7000-8000-00000000000b";

// aliceAndBob is a 1:1 chat. A test changes the fields it cares about.
function aliceAndBob(changes: Partial<Conversation> = {}): Conversation {
  const defaults: Conversation = {
    id: "01a19641-0000-7000-8000-0000000000c1",
    kind: "direct",
    groupName: "",
    members: [
      { userId: aliceId, username: "alice", deliveredUpTo: 0n, readUpTo: 0n },
      { userId: bobId, username: "bob", deliveredUpTo: 0n, readUpTo: 0n },
    ],
    lastActivityAt: new Date("2026-11-01T12:00:00Z"),
    lastMessage: null,
  };
  return { ...defaults, ...changes };
}

function messageFrom(senderId: string, text: string) {
  return {
    id: "01a19641-0000-7000-8000-000000000001",
    conversationId: "01a19641-0000-7000-8000-0000000000c1",
    sequence: 1n,
    senderId,
    text,
    sentAt: new Date("2026-11-01T12:00:00Z"),
  };
}

test("a 1:1 chat is called after the other person", () => {
  expect(conversationTitle(aliceAndBob(), aliceId)).toBe("bob");
  expect(conversationTitle(aliceAndBob(), bobId)).toBe("alice");
});

test("a group is called by its name", () => {
  const group = aliceAndBob({ kind: "group", groupName: "Friends" });

  expect(conversationTitle(group, aliceId)).toBe("Friends");
});

test("the preview says who wrote the newest message", () => {
  const fromBob = aliceAndBob({ lastMessage: messageFrom(bobId, "hi!") });
  const fromAlice = aliceAndBob({ lastMessage: messageFrom(aliceId, "hey") });

  expect(lastMessagePreview(fromBob, aliceId)).toBe("bob: hi!");
  expect(lastMessagePreview(fromAlice, aliceId)).toBe("You: hey");
  expect(lastMessagePreview(aliceAndBob(), aliceId)).toBe("No messages yet");
});

test("usernames can be separated by commas and spaces", () => {
  expect(parseUsernames(" bob, carol  dave,,")).toEqual([
    "bob",
    "carol",
    "dave",
  ]);
  expect(parseUsernames("   ")).toEqual([]);
});

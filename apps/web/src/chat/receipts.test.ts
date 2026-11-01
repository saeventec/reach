import { expect, test } from "vitest";
import type { Member, Message } from "./chatClient.ts";
import { applyReceipt, messageStatus } from "./receipts.ts";

const aliceId = "01a00000-0000-7000-8000-00000000000a";
const bobId = "01a00000-0000-7000-8000-00000000000b";
const carolId = "01a00000-0000-7000-8000-00000000000c";

function member(
  userId: string,
  deliveredUpTo: bigint,
  readUpTo: bigint,
): Member {
  return { userId, username: userId, deliveredUpTo, readUpTo };
}

// fromAlice is alice's message number 5.
const fromAlice: Message = {
  id: "01a19641-0000-7000-8000-000000000005",
  conversationId: "01a19641-0000-7000-8000-0000000000c1",
  sequence: 5n,
  senderId: aliceId,
  text: "hi",
  sentAt: new Date("2026-11-01T12:00:00Z"),
};

test("in a 1:1 chat, the status follows the other person's watermarks", () => {
  // alice's own watermarks (0) don't count: she wrote it.
  const sent = [member(aliceId, 0n, 0n), member(bobId, 4n, 4n)];
  const delivered = [member(aliceId, 0n, 0n), member(bobId, 5n, 4n)];
  const read = [member(aliceId, 0n, 0n), member(bobId, 9n, 5n)];

  expect(messageStatus(fromAlice, sent)).toBe("sent");
  expect(messageStatus(fromAlice, delivered)).toBe("delivered");
  expect(messageStatus(fromAlice, read)).toBe("read");
});

test("in a group, it needs everyone", () => {
  const bobReadCarolNot = [
    member(aliceId, 0n, 0n),
    member(bobId, 5n, 5n),
    member(carolId, 5n, 2n),
  ];
  const carolHasNothing = [
    member(aliceId, 0n, 0n),
    member(bobId, 5n, 5n),
    member(carolId, 2n, 2n),
  ];

  expect(messageStatus(fromAlice, bobReadCarolNot)).toBe("delivered");
  expect(messageStatus(fromAlice, carolHasNothing)).toBe("sent");
});

test("a receipt moves one member's watermarks, and only up", () => {
  const members = [member(aliceId, 3n, 3n), member(bobId, 9n, 6n)];
  const receipt = { userId: bobId, deliveredUpTo: 8n, readUpTo: 8n };

  const updated = applyReceipt(members, receipt);

  // delivered stays 9 (8 is lower: an older event); read goes up to 8.
  expect(updated[1]).toEqual(member(bobId, 9n, 8n));
  // The other members are not touched.
  expect(updated[0]).toBe(members[0]);
});

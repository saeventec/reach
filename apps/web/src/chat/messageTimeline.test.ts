import { expect, test } from "vitest";
import type { Message } from "./chatClient.ts";
import { mergeMessages, newestSequence } from "./messageTimeline.ts";

// message makes a message with this sequence number.
function message(sequence: bigint): Message {
  return {
    id: `01a19641-0000-7000-8000-${sequence.toString().padStart(12, "0")}`,
    conversationId: "01a19641-0000-7000-8000-0000000000c1",
    sequence,
    senderId: "01a00000-0000-7000-8000-00000000000a",
    text: `message ${sequence}`,
    sentAt: new Date("2026-11-01T12:00:00Z"),
  };
}

// sequences lists the sequence numbers, to compare easily.
function sequences(messages: Message[]): bigint[] {
  return messages.map((item) => item.sequence);
}

test("keeps the messages in order, whatever order they arrive in", () => {
  const merged = mergeMessages([message(2n)], [message(3n), message(1n)]);

  expect(sequences(merged)).toEqual([1n, 2n, 3n]);
});

test("adds a message that arrives twice only once", () => {
  const current = [message(1n), message(2n)];

  const merged = mergeMessages(current, [message(2n), message(2n)]);

  expect(sequences(merged)).toEqual([1n, 2n]);
});

test("knows the newest sequence number, or 0 for none", () => {
  expect(newestSequence([message(1n), message(7n)])).toBe(7n);
  expect(newestSequence([])).toBe(0n);
});

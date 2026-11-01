// The messages of one conversation on screen: always sorted by sequence
// number, and each message only once, however it arrived (the first page,
// older pages, a live event, a catch-up, or the answer to our own send).
// Learn more: docs/learn/message-ordering.md
import type { Message } from "./chatClient.ts";

// mergeMessages adds the incoming messages to the current ones. A message
// that is already there (the same sequence number) is not added twice:
// that is how copies from "at least once" delivery are ignored.
export function mergeMessages(
  current: Message[],
  incoming: Message[],
): Message[] {
  // A Map from sequence number to message: setting a key twice keeps one.
  const bySequence = new Map<bigint, Message>();
  for (const message of current) {
    bySequence.set(message.sequence, message);
  }
  for (const message of incoming) {
    bySequence.set(message.sequence, message);
  }

  const merged = Array.from(bySequence.values());
  merged.sort(compareBySequence);
  return merged;
}

// newestSequence is the sequence number of the newest message, or 0 when
// there is none: "everything after 0" means from the first message.
export function newestSequence(messages: Message[]): bigint {
  const newest = messages.at(-1);
  if (newest === undefined) {
    return 0n;
  }
  return newest.sequence;
}

// newestSequenceFromOthers is the sequence number of the newest message
// someone else sent, or 0 when there is none. Your own messages don't need
// marking as read: you wrote them.
export function newestSequenceFromOthers(
  messages: Message[],
  myUserId: string,
): bigint {
  const fromOthers = messages.filter(
    (message) => message.senderId !== myUserId,
  );
  return newestSequence(fromOthers);
}

// compareBySequence sorts the oldest first. sort() wants a negative number,
// 0, or a positive number; bigints can't be subtracted into a plain number,
// so compare them.
function compareBySequence(first: Message, second: Message): number {
  if (first.sequence < second.sequence) {
    return -1;
  }
  if (first.sequence > second.sequence) {
    return 1;
  }
  return 0;
}

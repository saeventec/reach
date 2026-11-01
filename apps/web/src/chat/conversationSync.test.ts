import { expect, test, vi } from "vitest";
import type { Message, MessagePage, MessagePosition } from "./chatClient.ts";
import { ConversationSync } from "./conversationSync.ts";
import type { TimelineView } from "./conversationSync.ts";
import { fakeChatApi } from "./fakeChatApi.ts";

const chatId = "01a19641-0000-7000-8000-0000000000c1";

function message(sequence: bigint): Message {
  return {
    id: `01a19641-0000-7000-8000-${sequence.toString().padStart(12, "0")}`,
    conversationId: chatId,
    sequence,
    senderId: "01a00000-0000-7000-8000-00000000000b",
    text: `message ${sequence}`,
    sentAt: new Date("2026-11-01T12:00:00Z"),
  };
}

// messagesUpTo makes the messages from 1 to last.
function messagesUpTo(last: number): Message[] {
  const messages: Message[] = [];
  for (let sequence = 1; sequence <= last; sequence++) {
    messages.push(message(BigInt(sequence)));
  }
  return messages;
}

// fakeServer answers listMessages like the chat service, from a list of
// stored messages, with pages of pageSize.
function fakeServer(stored: Message[], pageSize = 50) {
  return async (
    _conversationId: string,
    position: MessagePosition,
  ): Promise<MessagePage> => {
    let matching = stored;
    if (position.kind === "after") {
      matching = stored.filter((item) => item.sequence > position.sequence);
      return {
        messages: matching.slice(0, pageSize),
        hasMore: matching.length > pageSize,
      };
    }
    if (position.kind === "before") {
      matching = stored.filter((item) => item.sequence < position.sequence);
    }
    return {
      messages: matching.slice(-pageSize),
      hasMore: matching.length > pageSize,
    };
  };
}

// start makes a sync, and keeps the last view it gave the page.
function start(listMessages: ReturnType<typeof fakeServer>) {
  const spy = vi.fn(listMessages);
  const views: TimelineView[] = [];
  const sync = new ConversationSync(
    fakeChatApi({ listMessages: spy }),
    chatId,
    (view) => views.push(view),
  );
  const shown = () => views.at(-1)?.messages.map((item) => item.sequence);
  return { sync, spy, shown, views };
}

test("opens with the newest page, and loads older ones on request", async () => {
  const { sync, shown, views } = start(fakeServer(messagesUpTo(60)));

  await sync.loadNewest();
  expect(shown()).toHaveLength(50);
  expect(views.at(-1)?.hasOlder).toBe(true);

  await sync.loadOlder();
  expect(shown()).toHaveLength(60);
  expect(views.at(-1)?.hasOlder).toBe(false);
});

test("a live message that follows the newest one is simply added", async () => {
  const { sync, spy, shown } = start(fakeServer(messagesUpTo(3)));
  await sync.loadNewest();

  sync.add(message(4n));

  expect(shown()).toEqual([1n, 2n, 3n, 4n]);
  expect(spy).toHaveBeenCalledTimes(1); // only the first page
});

test("a copy of a message is ignored", async () => {
  const { sync, shown } = start(fakeServer(messagesUpTo(3)));
  await sync.loadNewest();

  sync.add(message(3n));

  expect(shown()).toEqual([1n, 2n, 3n]);
});

test("a gap is filled by catching up", async () => {
  const stored = messagesUpTo(5);
  const { sync, shown } = start(fakeServer(stored));
  await sync.loadNewest();
  // While we weren't looking, 4 and 5 were stored; then 6 arrives live.
  stored.push(message(6n));
  const shownBefore = shown();

  sync.add(message(6n));

  expect(shownBefore).toEqual([1n, 2n, 3n, 4n, 5n]);
  await vi.waitFor(() => expect(shown()).toEqual([1n, 2n, 3n, 4n, 5n, 6n]));
});

test("catching up loads every missed page", async () => {
  const stored = messagesUpTo(3);
  const { sync, shown } = start(fakeServer(stored, 2));
  await sync.loadNewest();
  // Offline for a while: 4 to 8 were sent.
  stored.push(...messagesUpTo(8).slice(3));

  await sync.catchUp();

  // Pages of 2: the first page was 2 and 3 (1 is "older"). Catching up
  // took three more pages: 4-5, 6-7, and 8.
  expect(shown()).toEqual([2n, 3n, 4n, 5n, 6n, 7n, 8n]);
});

test("does not catch up before the first page", async () => {
  const { sync, spy } = start(fakeServer(messagesUpTo(3)));

  await sync.catchUp();

  expect(spy).not.toHaveBeenCalled();
});

test("a live message that arrives before the first page is kept", async () => {
  const stored = messagesUpTo(3);
  const { sync, shown } = start(fakeServer(stored));

  sync.add(message(4n));
  stored.push(message(4n));
  await sync.loadNewest();

  expect(shown()).toEqual([1n, 2n, 3n, 4n]);
});

test("a catch-up asked for during another one runs once more after it", async () => {
  const stored = messagesUpTo(3);
  const server = fakeServer(stored);
  // The first catch-up's answer waits until the test lets it go.
  let releaseFirst: () => void = () => {};
  const firstAnswerHeld = new Promise<void>((resolve) => {
    releaseFirst = resolve;
  });
  let afterCalls = 0;
  const listMessages = async (
    conversationId: string,
    position: MessagePosition,
  ) => {
    if (position.kind === "after") {
      afterCalls++;
      if (afterCalls === 1) {
        const answer = await server(conversationId, position);
        await firstAnswerHeld;
        return answer;
      }
    }
    return server(conversationId, position);
  };
  const { sync, shown } = start(listMessages);
  await sync.loadNewest();

  const first = sync.catchUp(); // starts, and waits for its answer
  stored.push(message(4n)); // a new message, after the first question
  const second = sync.catchUp(); // asked for during the first one
  releaseFirst();
  await Promise.all([first, second]);

  // The first answer had nothing new; the extra round found 4.
  expect(shown()).toEqual([1n, 2n, 3n, 4n]);
});

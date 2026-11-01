// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { ConnectError, Code } from "@connectrpc/connect";
import { MemoryRouter } from "react-router";
import { afterEach, expect, test, vi } from "vitest";
import type { User } from "../auth/authClient.ts";
import { ServerEventSchema } from "../gen/realtime/v1/events_pb.ts";
import type { ChatApi, Conversation, Message } from "./chatClient.ts";
import { ConversationPage } from "./ConversationPage.tsx";
import { fakeChatApi } from "./fakeChatApi.ts";
import type { LiveUpdateHandlers, Subscribe } from "./liveUpdates.ts";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

const alice: User = {
  id: "01a00000-0000-7000-8000-00000000000a",
  username: "alice",
};
const bobId = "01a00000-0000-7000-8000-00000000000b";
const chatId = "01a19641-0000-7000-8000-0000000000c1";

const chatWithBob: Conversation = {
  id: chatId,
  kind: "direct",
  groupName: "",
  members: [
    { userId: alice.id, username: "alice", deliveredUpTo: 0n, readUpTo: 0n },
    { userId: bobId, username: "bob", deliveredUpTo: 0n, readUpTo: 0n },
  ],
  lastActivityAt: new Date("2026-11-01T12:00:00Z"),
  lastMessage: null,
};

// message makes a message in the chat with bob.
function message(sequence: bigint, senderId: string, text: string): Message {
  return {
    id: `01a19641-0000-7000-8000-${sequence.toString().padStart(12, "0")}`,
    conversationId: chatId,
    sequence,
    senderId,
    text,
    sentAt: new Date("2026-11-01T12:00:00Z"),
  };
}

// chatApi returns a ChatApi with the chat with bob and two messages. A test
// replaces the calls it cares about.
function chatApi(changes: Partial<ChatApi> = {}): ChatApi {
  return fakeChatApi({
    getConversation: async () => chatWithBob,
    listMessages: async () => ({
      messages: [message(1n, bobId, "hi alice"), message(2n, alice.id, "hey")],
      hasMore: false,
    }),
    markDelivered: async () => {},
    markRead: async () => {},
    ...changes,
  });
}

// No live updates: subscribing does nothing.
const noLiveUpdates: Subscribe = () => () => {};

function renderPage(api: ChatApi, subscribe: Subscribe = noLiveUpdates) {
  render(
    <MemoryRouter>
      <ConversationPage
        conversationId={chatId}
        user={alice}
        chatApi={api}
        subscribe={subscribe}
      />
    </MemoryRouter>,
  );
}

// fakeLiveUpdates keeps the page's handlers, so a test can play the server.
function fakeLiveUpdates() {
  const live: { handlers: LiveUpdateHandlers | null } = { handlers: null };
  const subscribe: Subscribe = (handlers) => {
    live.handlers = handlers;
    return () => {};
  };
  return { live, subscribe };
}

// messageSentEvent is what realtime sends for a new message.
function messageSentEvent(sent: Message) {
  return create(ServerEventSchema, {
    event: {
      case: "messageSent",
      value: {
        message: { ...sent, sentAt: timestampFromDate(sent.sentAt) },
        memberIds: [alice.id, bobId],
      },
    },
  });
}

function typeMessage(text: string) {
  const input = screen.getByLabelText("Message");
  fireEvent.change(input, { target: { value: text } });
  fireEvent.click(screen.getByRole("button", { name: "Send" }));
}

test("shows the chat's name and its messages", async () => {
  renderPage(chatApi());

  await screen.findByRole("heading", { name: "bob" });
  screen.getByText("hi alice");
  screen.getByText("hey");
});

test("says so for a chat you are not in", async () => {
  renderPage(chatApi({ getConversation: async () => null }));

  await screen.findByText("You have no chat with this address.");
});

test("sends a message with a new UUIDv7, and shows it", async () => {
  // A name starting with "_" tells TypeScript the parameter is unused on
  // purpose (it must be there, so the next ones are in their places).
  const sendMessage = vi.fn(
    async (_conversationId: string, messageId: string, text: string) => ({
      ...message(3n, alice.id, text),
      id: messageId,
    }),
  );
  renderPage(chatApi({ sendMessage }));
  await screen.findByText("hey");

  typeMessage("see you at 8");

  await screen.findByText("see you at 8");
  const [conversationId, messageId, text] = sendMessage.mock.calls[0];
  expect(conversationId).toBe(chatId);
  expect(messageId).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-7/);
  expect(text).toBe("see you at 8");
  // Confirmed: no "Sending…" left.
  await vi.waitFor(() => expect(screen.queryByText("Sending…")).toBeNull());
});

test("a failed message can be sent again, with the same id", async () => {
  const sendMessage = vi
    .fn()
    .mockRejectedValueOnce(new ConnectError("offline", Code.Unavailable))
    .mockImplementationOnce(
      async (_conversationId: string, messageId: string, text: string) => ({
        ...message(3n, alice.id, text),
        id: messageId,
      }),
    );
  renderPage(chatApi({ sendMessage }));
  await screen.findByText("hey");

  typeMessage("are you there?");
  await screen.findByText(/Not sent/);
  fireEvent.click(screen.getByRole("button", { name: "Retry" }));

  await vi.waitFor(() => expect(screen.queryByText(/Not sent/)).toBeNull());
  expect(sendMessage).toHaveBeenCalledTimes(2);
  const firstId = sendMessage.mock.calls[0][1];
  const secondId = sendMessage.mock.calls[1][1];
  expect(secondId).toBe(firstId);
});

test("shows older messages on request", async () => {
  const listMessages = vi
    .fn()
    .mockResolvedValueOnce({
      messages: [message(51n, bobId, "the newest")],
      hasMore: true,
    })
    .mockResolvedValueOnce({
      messages: [message(50n, bobId, "an older one")],
      hasMore: false,
    });
  renderPage(chatApi({ listMessages }));
  await screen.findByText("the newest");

  fireEvent.click(screen.getByRole("button", { name: "Show older messages" }));

  await screen.findByText("an older one");
  expect(listMessages).toHaveBeenLastCalledWith(chatId, {
    kind: "before",
    sequence: 51n,
  });
  // No more older ones: the button is gone.
  expect(
    screen.queryByRole("button", { name: "Show older messages" }),
  ).toBeNull();
});

test("shows a live message for this chat, and ignores other chats", async () => {
  const { live, subscribe } = fakeLiveUpdates();
  renderPage(chatApi(), subscribe);
  await screen.findByText("hey");

  const otherChat = {
    ...message(1n, bobId, "in another chat"),
    conversationId: "01a19641-0000-7000-8000-0000000000c9",
  };
  live.handlers?.onEvent(messageSentEvent(otherChat));
  live.handlers?.onEvent(messageSentEvent(message(3n, bobId, "you there?")));

  await screen.findByText("you there?");
  expect(screen.queryByText("in another chat")).toBeNull();
});

test("catches up after a reconnect", async () => {
  const listMessages = vi
    .fn()
    .mockResolvedValueOnce({
      messages: [message(1n, bobId, "hi alice")],
      hasMore: false,
    })
    .mockResolvedValueOnce({
      messages: [message(2n, bobId, "sent while you were offline")],
      hasMore: false,
    });
  const { live, subscribe } = fakeLiveUpdates();
  renderPage(chatApi({ listMessages }), subscribe);
  await screen.findByText("hi alice");

  live.handlers?.onConnected();

  await screen.findByText("sent while you were offline");
  expect(listMessages).toHaveBeenLastCalledWith(chatId, {
    kind: "after",
    sequence: 1n,
  });
});

// setTabVisible pretends the browser tab is in front (true) or in the
// background (false). vi.spyOn replaces the property's getter for the test.
function setTabVisible(isVisible: boolean) {
  const state = isVisible ? "visible" : "hidden";
  vi.spyOn(document, "visibilityState", "get").mockReturnValue(state);
}

test("marks the messages read when the page is visible", async () => {
  setTabVisible(true);
  const markRead = vi.fn(async () => {});
  renderPage(chatApi({ markRead }));
  await screen.findByText("hey");

  await vi.waitFor(() => expect(markRead).toHaveBeenCalledWith(chatId, 2n));
});

test("only marks them delivered while the tab is in the background", async () => {
  setTabVisible(false);
  const markRead = vi.fn(async () => {});
  const markDelivered = vi.fn(async () => {});
  renderPage(chatApi({ markRead, markDelivered }));
  await screen.findByText("hey");
  await vi.waitFor(() =>
    expect(markDelivered).toHaveBeenCalledWith(chatId, 2n),
  );
  expect(markRead).not.toHaveBeenCalled();

  // Coming back to the tab: now they are read.
  setTabVisible(true);
  document.dispatchEvent(new Event("visibilitychange"));

  await vi.waitFor(() => expect(markRead).toHaveBeenCalledWith(chatId, 2n));
});

test("shows delivered, then read, as bob's watermarks move", async () => {
  // bob has received alice's message 2, but not read it.
  const withBobsWatermarks = {
    ...chatWithBob,
    members: [
      { userId: alice.id, username: "alice", deliveredUpTo: 2n, readUpTo: 2n },
      { userId: bobId, username: "bob", deliveredUpTo: 2n, readUpTo: 1n },
    ],
  };
  const { live, subscribe } = fakeLiveUpdates();
  renderPage(
    chatApi({ getConversation: async () => withBobsWatermarks }),
    subscribe,
  );
  await screen.findByLabelText("Delivered");

  live.handlers?.onEvent(
    create(ServerEventSchema, {
      event: {
        case: "receiptUpdated",
        value: {
          conversationId: chatId,
          userId: bobId,
          deliveredUpTo: 2n,
          readUpTo: 2n,
          memberIds: [alice.id, bobId],
        },
      },
    }),
  );

  await screen.findByLabelText("Read");
});

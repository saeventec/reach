// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { ConnectError, Code } from "@connectrpc/connect";
import { MemoryRouter } from "react-router";
import { afterEach, expect, test, vi } from "vitest";
import type { User } from "../auth/authClient.ts";
import type { ChatApi, Conversation, Message } from "./chatClient.ts";
import { ConversationPage } from "./ConversationPage.tsx";
import { fakeChatApi } from "./fakeChatApi.ts";

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
    listConversations: async () => [chatWithBob],
    listMessages: async () => ({
      messages: [message(1n, bobId, "hi alice"), message(2n, alice.id, "hey")],
      hasMore: false,
    }),
    ...changes,
  });
}

function renderPage(api: ChatApi) {
  render(
    <MemoryRouter>
      <ConversationPage conversationId={chatId} user={alice} chatApi={api} />
    </MemoryRouter>,
  );
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
  renderPage(chatApi({ listConversations: async () => [] }));

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

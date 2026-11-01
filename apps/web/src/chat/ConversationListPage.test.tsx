// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { ConnectError, Code } from "@connectrpc/connect";
import { afterEach, expect, test, vi } from "vitest";
import type { User } from "../auth/authClient.ts";
import type { ChatApi, Conversation } from "./chatClient.ts";
import { ConversationListPage } from "./ConversationListPage.tsx";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

const alice: User = {
  id: "01a00000-0000-7000-8000-00000000000a",
  username: "alice",
};
const bobId = "01a00000-0000-7000-8000-00000000000b";

const chatWithBob: Conversation = {
  id: "01a19641-0000-7000-8000-0000000000c1",
  kind: "direct",
  groupName: "",
  members: [
    { userId: alice.id, username: "alice", deliveredUpTo: 0n, readUpTo: 0n },
    { userId: bobId, username: "bob", deliveredUpTo: 1n, readUpTo: 1n },
  ],
  lastActivityAt: new Date("2026-11-01T12:00:00Z"),
  lastMessage: {
    id: "01a19641-0000-7000-8000-000000000001",
    conversationId: "01a19641-0000-7000-8000-0000000000c1",
    sequence: 1n,
    senderId: bobId,
    text: "see you at 8",
    sentAt: new Date("2026-11-01T12:00:00Z"),
  },
};

const friends: Conversation = {
  ...chatWithBob,
  id: "01a19641-0000-7000-8000-0000000000c2",
  kind: "group",
  groupName: "Friends",
  lastMessage: null,
};

// fakeChatApi returns a ChatApi with the chat with bob. A test replaces
// the calls it cares about.
function fakeChatApi(changes: Partial<ChatApi> = {}): ChatApi {
  const defaults: ChatApi = {
    listConversations: async () => [chatWithBob],
    startDirectConversation: async () => chatWithBob,
    createGroup: async () => friends,
  };
  return { ...defaults, ...changes };
}

// typeInto types a value into the input with this label.
function typeInto(label: string, value: string) {
  const input = screen.getByLabelText(label);
  fireEvent.change(input, { target: { value } });
}

test("lists the chats with their newest message", async () => {
  render(<ConversationListPage user={alice} chatApi={fakeChatApi()} />);

  await screen.findByText("bob");
  screen.getByText("bob: see you at 8");
});

test("says so when there are no chats yet", async () => {
  const listConversations = async () => [];

  render(
    <ConversationListPage
      user={alice}
      chatApi={fakeChatApi({ listConversations })}
    />,
  );

  await screen.findByText("No chats yet. Start one with a username below.");
});

test("starts a chat and loads the list again", async () => {
  const startDirectConversation = vi.fn(async () => chatWithBob);
  const listConversations = vi.fn(async () => [chatWithBob]);
  render(
    <ConversationListPage
      user={alice}
      chatApi={fakeChatApi({ startDirectConversation, listConversations })}
    />,
  );
  await screen.findByText("bob");

  typeInto("Username", "bob");
  fireEvent.click(screen.getByRole("button", { name: "Start chat" }));

  await vi.waitFor(() =>
    expect(startDirectConversation).toHaveBeenCalledWith("bob"),
  );
  // Once when the page opened, once after the new chat.
  await vi.waitFor(() => expect(listConversations).toHaveBeenCalledTimes(2));
});

test("creates a group from a name and a list of usernames", async () => {
  const createGroup = vi.fn(async () => friends);
  render(
    <ConversationListPage
      user={alice}
      chatApi={fakeChatApi({ createGroup })}
    />,
  );
  await screen.findByText("bob");

  typeInto("Group name", "Friends");
  typeInto("Members (usernames, separated by commas)", "bob, carol");
  fireEvent.click(screen.getByRole("button", { name: "Create group" }));

  await vi.waitFor(() =>
    expect(createGroup).toHaveBeenCalledWith("Friends", ["bob", "carol"]),
  );
});

test("shows the service's message when the username is unknown", async () => {
  const startDirectConversation = async () => {
    throw new ConnectError("nobody has the username dave", Code.NotFound);
  };
  render(
    <ConversationListPage
      user={alice}
      chatApi={fakeChatApi({ startDirectConversation })}
    />,
  );
  await screen.findByText("bob");

  typeInto("Username", "dave");
  fireEvent.click(screen.getByRole("button", { name: "Start chat" }));

  await screen.findByText("Nobody has the username dave");
});

test("shows an error when the chats can't be loaded", async () => {
  const listConversations = async () => {
    throw new Error("network down");
  };
  vi.spyOn(console, "error").mockImplementation(() => {});

  render(
    <ConversationListPage
      user={alice}
      chatApi={fakeChatApi({ listConversations })}
    />,
  );

  await screen.findByText("Could not load your chats.");
});

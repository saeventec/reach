// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { create } from "@bufbuild/protobuf";
import { ConnectError, Code } from "@connectrpc/connect";
import type { ReactElement } from "react";
import { MemoryRouter } from "react-router";
import { afterEach, expect, test, vi } from "vitest";
import type { User } from "../auth/authClient.ts";
import { ServerEventSchema } from "../gen/realtime/v1/events_pb.ts";
import type { ChatApi, Conversation } from "./chatClient.ts";
import { ConversationListPage } from "./ConversationListPage.tsx";
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

// chatApi returns a ChatApi with the chat with bob. A test replaces the
// calls it cares about.
function chatApi(changes: Partial<ChatApi> = {}): ChatApi {
  return fakeChatApi({
    listConversations: async () => [chatWithBob],
    startDirectConversation: async () => chatWithBob,
    createGroup: async () => friends,
    ...changes,
  });
}

// No live updates: subscribing does nothing.
const noLiveUpdates: Subscribe = () => () => {};

// fakeLiveUpdates lets a test play the server: it keeps the page's
// handlers, so the test can call them.
function fakeLiveUpdates() {
  const live: { handlers: LiveUpdateHandlers | null; stopped: boolean } = {
    handlers: null,
    stopped: false,
  };
  const subscribe: Subscribe = (handlers) => {
    live.handlers = handlers;
    return () => {
      live.stopped = true;
    };
  };
  return { live, subscribe };
}

// renderPage draws the page inside a router, because its items are links.
function renderPage(page: ReactElement) {
  return render(<MemoryRouter>{page}</MemoryRouter>);
}

// typeInto types a value into the input with this label.
function typeInto(label: string, value: string) {
  const input = screen.getByLabelText(label);
  fireEvent.change(input, { target: { value } });
}

test("lists the chats with their newest message, as links", async () => {
  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi()}
      subscribe={noLiveUpdates}
    />,
  );

  await screen.findByText("bob");
  screen.getByText("bob: see you at 8");
  const link = screen.getByRole("link", { name: /bob/ });
  expect(link.getAttribute("href")).toBe(`/chats/${chatWithBob.id}`);
});

test("says so when there are no chats yet", async () => {
  const listConversations = async () => [];

  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ listConversations })}
      subscribe={noLiveUpdates}
    />,
  );

  await screen.findByText("No chats yet. Start one with a username below.");
});

test("starts a chat and loads the list again", async () => {
  const startDirectConversation = vi.fn(async () => chatWithBob);
  const listConversations = vi.fn(async () => [chatWithBob]);
  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ startDirectConversation, listConversations })}
      subscribe={noLiveUpdates}
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
  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ createGroup })}
      subscribe={noLiveUpdates}
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
  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ startDirectConversation })}
      subscribe={noLiveUpdates}
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

  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ listConversations })}
      subscribe={noLiveUpdates}
    />,
  );

  await screen.findByText("Could not load your chats.");
});

test("loads the list again on a new message, a new chat, and a reconnect", async () => {
  const listConversations = vi.fn(async () => [chatWithBob]);
  const { live, subscribe } = fakeLiveUpdates();
  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ listConversations })}
      subscribe={subscribe}
    />,
  );
  await screen.findByText("bob");
  expect(listConversations).toHaveBeenCalledTimes(1);

  // create() builds a Protobuf message from plain values: here, what
  // realtime sends when bob starts a chat with alice.
  const conversationCreated = create(ServerEventSchema, {
    event: {
      case: "conversationCreated",
      value: { conversationId: "chat-2", memberIds: [alice.id, bobId] },
    },
  });
  live.handlers?.onEvent(conversationCreated);
  await vi.waitFor(() => expect(listConversations).toHaveBeenCalledTimes(2));
});

test("creates a group from a name and a list of usernames", async () => {
  const createGroup = vi.fn(async () => friends);
  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ createGroup })}
      subscribe={noLiveUpdates}
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
  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ startDirectConversation })}
      subscribe={noLiveUpdates}
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

  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ listConversations })}
      subscribe={noLiveUpdates}
    />,
  );

  await screen.findByText("Could not load your chats.");
});

test("loads the list again on a new message, a new chat, and a reconnect", async () => {
  const listConversations = vi.fn(async () => [chatWithBob]);
  const { live, subscribe } = fakeLiveUpdates();
  renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi({ listConversations })}
      subscribe={subscribe}
    />,
  );
  await screen.findByText("bob");
  expect(listConversations).toHaveBeenCalledTimes(1);

  live.handlers?.onEvent({
    $typeName: "realtime.v1.ServerEvent",
    event: {
      case: "conversationCreated",
      value: {
        $typeName: "chat.v1.ConversationCreated",
        conversationId: "chat-2",
        memberIds: [alice.id],
      },
    },
  });
  await vi.waitFor(() => expect(listConversations).toHaveBeenCalledTimes(2));

  live.handlers?.onConnected();
  await vi.waitFor(() => expect(listConversations).toHaveBeenCalledTimes(3));
});

test("closes the live updates when the page closes", async () => {
  const { live, subscribe } = fakeLiveUpdates();
  const page = renderPage(
    <ConversationListPage
      user={alice}
      chatApi={chatApi()}
      subscribe={subscribe}
    />,
  );
  await screen.findByText("bob");

  page.unmount();

  expect(live.stopped).toBe(true);
});

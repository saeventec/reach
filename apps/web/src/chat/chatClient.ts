// Talks to the chat service (ConversationService and MessageService), with code generated
// from proto/chat/v1/. The gateway checks the session cookie first, and
// tells chat who you are (ADR 0029). Learn more: docs/learn/connect-rpc.md
import { timestampDate } from "@bufbuild/protobuf/wkt";
import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import {
  ConversationKind,
  ConversationService,
} from "../gen/chat/v1/conversation_pb.ts";
import type {
  Conversation as ConversationMessage,
  Member as MemberMessage,
} from "../gen/chat/v1/conversation_pb.ts";
import { MessageService } from "../gen/chat/v1/message_pb.ts";
import type {
  ListMessagesRequest,
  Message as MessageMessage,
} from "../gen/chat/v1/message_pb.ts";

// A message, as the rest of the web app sees it.
export type Message = {
  id: string;
  conversationId: string;
  // The message's place in its conversation: 1, 2, 3 ... A bigint, because
  // it is an int64 in the API. Learn more: docs/learn/protobuf.md#int64-in-json
  sequence: bigint;
  senderId: string;
  text: string;
  sentAt: Date;
};

// A person in a conversation, with their delivered and read watermarks.
export type Member = {
  userId: string;
  username: string;
  deliveredUpTo: bigint;
  readUpTo: bigint;
};

// A 1:1 chat or a group.
export type Conversation = {
  id: string;
  kind: "direct" | "group";
  // Empty for a 1:1 chat.
  groupName: string;
  members: Member[];
  lastActivityAt: Date;
  // null when there is no message yet.
  lastMessage: Message | null;
};

// Which messages ListMessages returns: the newest ones (to open a
// conversation), the ones after a sequence number (to catch up), or the
// ones before it (to scroll back).
export type MessagePosition =
  | { kind: "newest" }
  | { kind: "after"; sequence: bigint }
  | { kind: "before"; sequence: bigint };

// One page of messages, the oldest first. hasMore: there are more in the
// same direction.
export type MessagePage = {
  messages: Message[];
  hasMore: boolean;
};

// All the calls the pages need. The app passes the real ones (chatApi
// below); tests pass fakes. Learn more: docs/learn/testing.md#fakes
export type ChatApi = {
  listConversations: () => Promise<Conversation[]>;
  // Returns null for a conversation you are not in (or that doesn't exist).
  getConversation: (conversationId: string) => Promise<Conversation | null>;
  startDirectConversation: (username: string) => Promise<Conversation>;
  createGroup: (name: string, usernames: string[]) => Promise<Conversation>;
  // messageId is made by the app (newUUIDv7), so a retry sends the same
  // one, and the server stores the message once.
  sendMessage: (
    conversationId: string,
    messageId: string,
    text: string,
  ) => Promise<Message>;
  listMessages: (
    conversationId: string,
    position: MessagePosition,
  ) => Promise<MessagePage>;
  markDelivered: (conversationId: string, sequence: bigint) => Promise<void>;
  markRead: (conversationId: string, sequence: bigint) => Promise<void>;
};

const transport = createConnectTransport({ baseUrl: window.location.origin });
const conversationClient = createClient(ConversationService, transport);
const messageClient = createClient(MessageService, transport);

async function listConversations(): Promise<Conversation[]> {
  const response = await conversationClient.listConversations({});
  const conversations = response.conversations.map(toConversation);
  return conversations;
}

async function getConversation(
  conversationId: string,
): Promise<Conversation | null> {
  try {
    const response = await conversationClient.getConversation({
      conversationId,
    });
    return toConversation(required(response.conversation));
  } catch (error) {
    // "Not found" is a normal answer here, not a failure.
    const connectError = ConnectError.from(error);
    if (connectError.code === Code.NotFound) {
      return null;
    }
    throw error;
  }
}

async function startDirectConversation(
  username: string,
): Promise<Conversation> {
  const response = await conversationClient.startDirectConversation({
    username,
  });
  return toConversation(required(response.conversation));
}

async function createGroup(
  name: string,
  usernames: string[],
): Promise<Conversation> {
  const response = await conversationClient.createGroup({ name, usernames });
  return toConversation(required(response.conversation));
}

async function sendMessage(
  conversationId: string,
  messageId: string,
  text: string,
): Promise<Message> {
  const response = await messageClient.sendMessage({
    conversationId,
    messageId,
    text,
  });
  if (response.message === undefined) {
    throw new Error("the chat service sent no message");
  }
  return toMessage(response.message);
}

async function listMessages(
  conversationId: string,
  position: MessagePosition,
): Promise<MessagePage> {
  const response = await messageClient.listMessages({
    conversationId,
    position: toPositionField(position),
  });
  const messages = response.messages.map(toMessage);
  return { messages, hasMore: response.hasMore };
}

async function markDelivered(
  conversationId: string,
  sequence: bigint,
): Promise<void> {
  await messageClient.markDelivered({ conversationId, sequence });
}

async function markRead(
  conversationId: string,
  sequence: bigint,
): Promise<void> {
  await messageClient.markRead({ conversationId, sequence });
}

// The real API calls, for the app.
export const chatApi: ChatApi = {
  listConversations,
  getConversation,
  startDirectConversation,
  createGroup,
  sendMessage,
  listMessages,
  markDelivered,
  markRead,
};

// toPositionField turns a MessagePosition into the request's "position"
// oneof. In the generated TypeScript, a oneof is an object with "case"
// (which field is set) and "value". No case means: the newest messages.
function toPositionField(
  position: MessagePosition,
): ListMessagesRequest["position"] {
  switch (position.kind) {
    case "after":
      return { case: "afterSequence", value: position.sequence };
    case "before":
      return { case: "beforeSequence", value: position.sequence };
    case "newest":
      return { case: undefined };
  }
}

// toConversation turns a Conversation message from the API into the app's
// Conversation type.
function toConversation(message: ConversationMessage): Conversation {
  const kind = message.kind === ConversationKind.GROUP ? "group" : "direct";
  const members = message.members.map(toMember);
  // A missing last message means: no message yet.
  const lastMessage =
    message.lastMessage === undefined ? null : toMessage(message.lastMessage);

  return {
    id: message.id,
    kind,
    groupName: message.groupName,
    members,
    lastActivityAt: toDate(message.lastActivityAt),
    lastMessage,
  };
}

function toMember(message: MemberMessage): Member {
  return {
    userId: message.userId,
    username: message.username,
    deliveredUpTo: message.deliveredUpTo,
    readUpTo: message.readUpTo,
  };
}

// toMessage is also used for the live events (ConversationPage.tsx),
// which carry the same Message type.
export function toMessage(message: MessageMessage): Message {
  return {
    id: message.id,
    conversationId: message.conversationId,
    sequence: message.sequence,
    senderId: message.senderId,
    text: message.text,
    sentAt: toDate(message.sentAt),
  };
}

// toDate turns a Protobuf Timestamp into a JavaScript Date. The service
// always sends the dates, so a missing one is a bug.
function toDate(timestamp: Timestamp | undefined): Date {
  if (timestamp === undefined) {
    throw new Error("the chat service sent a date-less value");
  }
  return timestampDate(timestamp);
}

// required returns the conversation from a response. A message field can
// always be missing in Protobuf, but the service always sends this one.
function required(
  conversation: ConversationMessage | undefined,
): ConversationMessage {
  if (conversation === undefined) {
    throw new Error("the chat service sent no conversation");
  }
  return conversation;
}

// Talks to the chat service (ConversationService), with code generated
// from proto/chat/v1/. The gateway checks the session cookie first, and
// tells chat who you are (ADR 0029). Learn more: docs/learn/connect-rpc.md
import { timestampDate } from "@bufbuild/protobuf/wkt";
import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import {
  ConversationKind,
  ConversationService,
} from "../gen/chat/v1/conversation_pb.ts";
import type {
  Conversation as ConversationMessage,
  Member as MemberMessage,
} from "../gen/chat/v1/conversation_pb.ts";
import type { Message as MessageMessage } from "../gen/chat/v1/message_pb.ts";

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

// All the calls the pages need. The app passes the real ones (chatApi
// below); tests pass fakes. Learn more: docs/learn/testing.md#fakes
export type ChatApi = {
  listConversations: () => Promise<Conversation[]>;
  startDirectConversation: (username: string) => Promise<Conversation>;
  createGroup: (name: string, usernames: string[]) => Promise<Conversation>;
};

const transport = createConnectTransport({ baseUrl: window.location.origin });
const conversationClient = createClient(ConversationService, transport);

async function listConversations(): Promise<Conversation[]> {
  const response = await conversationClient.listConversations({});
  const conversations = response.conversations.map(toConversation);
  return conversations;
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

// The real API calls, for the app.
export const chatApi: ChatApi = {
  listConversations,
  startDirectConversation,
  createGroup,
};

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

function toMessage(message: MessageMessage): Message {
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

// "Delivered" means: one of your devices has received the message. The app
// tells the server as soon as a message reaches it, live or in the list,
// so the sender sees ✓✓ even before you open the chat.
// Learn more: docs/learn/message-ordering.md#delivered-and-read
import type { ServerEvent } from "../gen/realtime/v1/events_pb.ts";
import type { ChatApi, Conversation } from "./chatClient.ts";

// markLiveMessageDelivered marks a live message from someone else as
// delivered: the conversation page uses it for the other conversations.
// (The conversation on screen is marked read by useMarkAsSeen, which counts
// as delivered too; the list page marks new messages when it reloads.)
export function markLiveMessageDelivered(
  chatApi: ChatApi,
  myUserId: string,
  serverEvent: ServerEvent,
  openConversationId: string,
): void {
  if (serverEvent.event.case !== "messageSent") {
    return;
  }
  const message = serverEvent.event.value.message;
  if (message === undefined || message.senderId === myUserId) {
    return;
  }
  if (message.conversationId === openConversationId) {
    return;
  }
  chatApi
    .markDelivered(message.conversationId, message.sequence)
    .catch(logError);
}

// markListDelivered marks the newest message of each conversation in the
// list as delivered, if it is from someone else and newer than what this
// user's devices have already reported.
export function markListDelivered(
  chatApi: ChatApi,
  conversations: Conversation[],
  myUserId: string,
): void {
  for (const conversation of conversations) {
    const lastMessage = conversation.lastMessage;
    if (lastMessage === null || lastMessage.senderId === myUserId) {
      continue;
    }
    const me = conversation.members.find(
      (member) => member.userId === myUserId,
    );
    const alreadyDelivered =
      me !== undefined && me.deliveredUpTo >= lastMessage.sequence;
    if (alreadyDelivered) {
      continue;
    }
    chatApi
      .markDelivered(conversation.id, lastMessage.sequence)
      .catch(logError);
  }
}

// A failed mark only means the sender's ✓✓ is late; it is logged, not
// shown.
function logError(markError: unknown) {
  console.error("could not mark the messages delivered", markError);
}

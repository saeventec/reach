// The messages of a conversation, the oldest at the top. Your own
// messages are on the right, in blue, with ✓ (sent), ✓✓ (delivered) or
// "Read"; the others on the left, with the sender's name.
import type { Member, Message } from "./chatClient.ts";
import { messageStatus } from "./receipts.ts";
import type { MessageStatus } from "./receipts.ts";

// A message you sent that the server hasn't confirmed yet.
export type PendingMessage = {
  id: string;
  text: string;
  status: "sending" | "failed";
  // Why it failed, for the screen. Empty while sending.
  error: string;
};

type MessageListProps = {
  // The members with their watermarks.
  members: Member[];
  myUserId: string;
  messages: Message[];
  pending: PendingMessage[];
  onRetry: (pendingMessage: PendingMessage) => void;
};

export function MessageList({
  members,
  myUserId,
  messages,
  pending,
  onRetry,
}: MessageListProps) {
  if (messages.length === 0 && pending.length === 0) {
    return <p className="text-slate-500">No messages yet. Say hello!</p>;
  }

  return (
    <ol className="flex flex-col gap-2">
      {messages.map((message) => (
        <MessageItem
          key={message.id}
          message={message}
          isMine={message.senderId === myUserId}
          senderName={senderName(members, message.senderId)}
          status={messageStatus(message, members)}
        />
      ))}
      {pending.map((pendingMessage) => (
        <PendingItem
          key={pendingMessage.id}
          pendingMessage={pendingMessage}
          onRetry={() => onRetry(pendingMessage)}
        />
      ))}
    </ol>
  );
}

type MessageItemProps = {
  message: Message;
  isMine: boolean;
  senderName: string;
  status: MessageStatus;
};

// The marks for each status. aria-label lets screen readers say a word
// instead of "check mark check mark".
const statusMarks: Record<MessageStatus, { mark: string; label: string }> = {
  sent: { mark: "✓", label: "Sent" },
  delivered: { mark: "✓✓", label: "Delivered" },
  read: { mark: "Read", label: "Read" },
};

function MessageItem({
  message,
  isMine,
  senderName,
  status,
}: MessageItemProps) {
  const time = formatTime(message.sentAt);

  if (isMine) {
    const { mark, label } = statusMarks[status];
    return (
      <li className="self-end rounded bg-blue-600 px-3 py-2 text-white">
        <p className="whitespace-pre-wrap">{message.text}</p>
        <p className="text-right text-xs text-blue-100">
          {time} <span aria-label={label}>{mark}</span>
        </p>
      </li>
    );
  }

  return (
    <li className="self-start rounded border border-slate-200 bg-white px-3 py-2">
      <p className="text-xs font-medium text-slate-500">{senderName}</p>
      <p className="whitespace-pre-wrap text-slate-900">{message.text}</p>
      <p className="text-xs text-slate-400">{time}</p>
    </li>
  );
}

type PendingItemProps = {
  pendingMessage: PendingMessage;
  onRetry: () => void;
};

function PendingItem({ pendingMessage, onRetry }: PendingItemProps) {
  return (
    <li className="self-end rounded bg-blue-300 px-3 py-2 text-white">
      <p className="whitespace-pre-wrap">{pendingMessage.text}</p>
      {pendingMessage.status === "sending" ? (
        <p className="text-right text-xs">Sending…</p>
      ) : (
        <p className="text-right text-xs">
          Not sent: {pendingMessage.error}{" "}
          <button type="button" onClick={onRetry} className="underline">
            Retry
          </button>
        </p>
      )}
    </li>
  );
}

// senderName is the username of the member with this id.
function senderName(members: Member[], senderId: string): string {
  const sender = members.find((member) => member.userId === senderId);
  return sender === undefined ? "Someone" : sender.username;
}

// formatTime writes the time the way the browser's language does, like
// "12:04" or "12:04 PM".
function formatTime(date: Date): string {
  const format = new Intl.DateTimeFormat(undefined, { timeStyle: "short" });
  return format.format(date);
}

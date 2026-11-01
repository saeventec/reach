// One conversation: its messages, older ones on request, and sending.
// Learn more: docs/learn/message-ordering.md
import { useEffect, useState } from "react";
import { Link } from "react-router";
import type { User } from "../auth/authClient.ts";
import { errorMessage } from "../shared/errorMessage.ts";
import type { ChatApi, Conversation, Message } from "./chatClient.ts";
import { conversationTitle } from "./conversationText.ts";
import { MessageComposer } from "./MessageComposer.tsx";
import { MessageList } from "./MessageList.tsx";
import type { PendingMessage } from "./MessageList.tsx";
import { mergeMessages } from "./messageTimeline.ts";
import { newUUIDv7 } from "./uuidv7.ts";

type PageState =
  | { status: "loading" }
  | { status: "loaded"; conversation: Conversation }
  | { status: "notFound" }
  | { status: "failed" };

type ConversationPageProps = {
  conversationId: string;
  user: User;
  chatApi: ChatApi;
};

export function ConversationPage({
  conversationId,
  user,
  chatApi,
}: ConversationPageProps) {
  const [pageState, setPageState] = useState<PageState>({ status: "loading" });
  const [messages, setMessages] = useState<Message[]>([]);
  // True if there are messages before the oldest one on screen.
  const [hasOlder, setHasOlder] = useState(false);
  const [pending, setPending] = useState<PendingMessage[]>([]);
  const [olderError, setOlderError] = useState("");

  useEffect(() => {
    // See NewestGreeting.tsx for why "ignore" is needed.
    let ignore = false;

    async function load() {
      try {
        // The conversation's name and members come from the list: chat has
        // no "get one conversation" call yet, and the list is small.
        const conversations = await chatApi.listConversations();
        const found = conversations.find(
          (conversation) => conversation.id === conversationId,
        );
        if (found === undefined) {
          if (!ignore) {
            setPageState({ status: "notFound" });
          }
          return;
        }

        const page = await chatApi.listMessages(conversationId, {
          kind: "newest",
        });
        if (!ignore) {
          setPageState({ status: "loaded", conversation: found });
          setMessages(page.messages);
          setHasOlder(page.hasMore);
        }
      } catch (loadError) {
        console.error("could not load the conversation", loadError);
        if (!ignore) {
          setPageState({ status: "failed" });
        }
      }
    }

    load();

    return () => {
      ignore = true;
    };
  }, [chatApi, conversationId]);

  // send shows the message as "Sending…" at once, then asks the server.
  // A retry calls it again with the same id, so the server stores the
  // message once, even if the first try did arrive.
  async function send(messageId: string, text: string) {
    const sending: PendingMessage = {
      id: messageId,
      text,
      status: "sending",
      error: "",
    };
    setPending((current) => replacePending(current, sending));

    try {
      const stored = await chatApi.sendMessage(conversationId, messageId, text);
      setPending((current) => removePending(current, messageId));
      setMessages((current) => mergeMessages(current, [stored]));
    } catch (sendError) {
      const failed: PendingMessage = {
        ...sending,
        status: "failed",
        error: errorMessage(sendError),
      };
      setPending((current) => replacePending(current, failed));
    }
  }

  function handleSend(text: string) {
    const messageId = newUUIDv7();
    send(messageId, text);
  }

  function handleRetry(pendingMessage: PendingMessage) {
    send(pendingMessage.id, pendingMessage.text);
  }

  async function handleShowOlder() {
    const oldest = messages[0];
    if (oldest === undefined) {
      return;
    }
    setOlderError("");
    try {
      const page = await chatApi.listMessages(conversationId, {
        kind: "before",
        sequence: oldest.sequence,
      });
      setMessages((current) => mergeMessages(current, page.messages));
      setHasOlder(page.hasMore);
    } catch (olderLoadError) {
      setOlderError(errorMessage(olderLoadError));
    }
  }

  return (
    <section className="flex w-full max-w-md flex-col gap-4">
      <Link to="/chats" className="text-sm text-blue-600 hover:underline">
        ← All chats
      </Link>

      {pageState.status === "loading" && (
        <p className="text-slate-500">Loading the chat…</p>
      )}
      {pageState.status === "notFound" && (
        <p className="text-red-600">You have no chat with this address.</p>
      )}
      {pageState.status === "failed" && (
        <p className="text-red-600">Could not load the chat.</p>
      )}

      {pageState.status === "loaded" && (
        <>
          {/* <>...</> is a "fragment": it groups elements without adding
              an element to the page. */}
          <h2 className="text-2xl font-semibold text-slate-900">
            {conversationTitle(pageState.conversation, user.id)}
          </h2>
          {hasOlder && (
            <button
              type="button"
              onClick={handleShowOlder}
              className="text-sm text-blue-600 hover:underline"
            >
              Show older messages
            </button>
          )}
          {olderError !== "" && <p className="text-red-600">{olderError}</p>}
          <MessageList
            conversation={pageState.conversation}
            myUserId={user.id}
            messages={messages}
            pending={pending}
            onRetry={handleRetry}
          />
          <MessageComposer onSend={handleSend} />
        </>
      )}
    </section>
  );
}

// replacePending puts a pending message in the list, in place of the one
// with the same id, or at the end.
function replacePending(
  current: PendingMessage[],
  pendingMessage: PendingMessage,
): PendingMessage[] {
  const others = current.filter((item) => item.id !== pendingMessage.id);
  const isRetry = others.length < current.length;
  if (!isRetry) {
    return [...current, pendingMessage];
  }
  return current.map((item) =>
    item.id === pendingMessage.id ? pendingMessage : item,
  );
}

// removePending takes the confirmed message out of the pending list.
function removePending(
  current: PendingMessage[],
  messageId: string,
): PendingMessage[] {
  return current.filter((item) => item.id !== messageId);
}

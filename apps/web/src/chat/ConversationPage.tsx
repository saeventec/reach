// One conversation: its messages, older ones on request, sending, live
// updates, and delivered and read. ConversationSync keeps the messages;
// this page shows them.
// Learn more: docs/learn/message-ordering.md
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router";
import type { User } from "../auth/authClient.ts";
import { errorMessage } from "../shared/errorMessage.ts";
import { toMessage } from "./chatClient.ts";
import type { ChatApi, Conversation, Member } from "./chatClient.ts";
import { ConversationSync } from "./conversationSync.ts";
import type { TimelineView } from "./conversationSync.ts";
import { conversationTitle } from "./conversationText.ts";
import type { Subscribe } from "./liveUpdates.ts";
import { MessageComposer } from "./MessageComposer.tsx";
import { MessageList } from "./MessageList.tsx";
import type { PendingMessage } from "./MessageList.tsx";
import { markLiveMessageDelivered } from "./markDelivered.ts";
import { newestSequence, newestSequenceFromOthers } from "./messageTimeline.ts";
import { applyReceipt, mergeMembers } from "./receipts.ts";
import { useMarkAsSeen } from "./useMarkAsSeen.ts";
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
  subscribe: Subscribe;
};

export function ConversationPage({
  conversationId,
  user,
  chatApi,
  subscribe,
}: ConversationPageProps) {
  const [pageState, setPageState] = useState<PageState>({ status: "loading" });
  const [view, setView] = useState<TimelineView>({
    messages: [],
    hasOlder: false,
  });
  // One ConversationSync for the page's whole life. A function passed to
  // useState runs only once, when the page first appears. (The page gets a
  // new "key" for each conversation, see App.tsx, so a new one each time.)
  const [sync] = useState(
    () => new ConversationSync(chatApi, conversationId, setView),
  );
  const [pending, setPending] = useState<PendingMessage[]>([]);
  const [olderError, setOlderError] = useState("");
  // The members with their watermarks: from the list at first, then moved
  // by live receipt events. They decide the ✓ marks.
  const [members, setMembers] = useState<Member[]>([]);

  const newest = newestSequence(view.messages);
  const newestFromOthers = newestSequenceFromOthers(view.messages, user.id);
  useMarkAsSeen(chatApi, conversationId, newestFromOthers);

  // A ref to an element on the page: the empty <div> at the bottom. When
  // a newer message arrives, or you send one, scroll it into view, so the
  // newest message and the form are on screen. Older messages (added at
  // the top) don't change "newest", so reading old ones isn't disturbed.
  const bottom = useRef<HTMLDivElement>(null);
  useEffect(() => {
    // "?." does nothing while the page is still loading (no element yet).
    bottom.current?.scrollIntoView({ block: "end" });
  }, [newest, pending.length]);

  useEffect(() => {
    // See NewestGreeting.tsx for why "ignore" is needed.
    let ignore = false;

    async function load() {
      try {
        const found = await chatApi.getConversation(conversationId);
        if (found === null) {
          if (!ignore) {
            setPageState({ status: "notFound" });
          }
          return;
        }

        await sync.loadNewest();
        if (!ignore) {
          setPageState({ status: "loaded", conversation: found });
          setMembers((current) => mergeMembers(current, found.members));
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
  }, [chatApi, conversationId, sync]);

  // Live updates: a new message in this conversation is added (a copy is
  // ignored, a gap is filled), a receipt moves a member's watermarks, and
  // after every (re)connect the page catches up on what it missed.
  useEffect(() => {
    // refreshMembers loads the members' watermarks again: receipt events
    // sent while the WebSocket was closed are lost, like messages.
    async function refreshMembers() {
      try {
        const found = await chatApi.getConversation(conversationId);
        if (found !== null) {
          setMembers((current) => mergeMembers(current, found.members));
        }
      } catch (refreshError) {
        console.error("could not reload the members", refreshError);
      }
    }

    const stop = subscribe({
      onEvent: (serverEvent) => {
        // Messages in your other conversations count as delivered.
        markLiveMessageDelivered(chatApi, user.id, serverEvent, conversationId);
        const event = serverEvent.event;
        if (event.case === "messageSent") {
          const eventMessage = event.value.message;
          if (eventMessage?.conversationId === conversationId) {
            sync.add(toMessage(eventMessage));
          }
        }
        if (event.case === "receiptUpdated") {
          const receipt = event.value;
          if (receipt.conversationId === conversationId) {
            setMembers((current) => applyReceipt(current, receipt));
          }
        }
      },
      onConnected: () => {
        sync.catchUp();
        refreshMembers();
      },
    });
    return stop;
  }, [subscribe, conversationId, sync, chatApi, user.id]);

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
      sync.add(stored);
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
    setOlderError("");
    try {
      await sync.loadOlder();
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
          {view.hasOlder && (
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
            members={members}
            myUserId={user.id}
            messages={view.messages}
            pending={pending}
            onRetry={handleRetry}
          />
          <MessageComposer onSend={handleSend} />
          <div ref={bottom} />
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

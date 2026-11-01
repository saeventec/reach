// The start page of a logged-in user: their conversations, the most
// recently active first, and the forms to start a new chat or group.
import { useEffect, useState } from "react";
import type { User } from "../auth/authClient.ts";
import type { ChatApi, Conversation } from "./chatClient.ts";
import { conversationTitle, lastMessagePreview } from "./conversationText.ts";
import type { Subscribe } from "./liveUpdates.ts";
import { NewChatForm } from "./NewChatForm.tsx";
import { NewGroupForm } from "./NewGroupForm.tsx";

type ConversationsState =
  | { status: "loading" }
  | { status: "loaded"; conversations: Conversation[] }
  | { status: "failed" };

type ConversationListPageProps = {
  user: User;
  chatApi: ChatApi;
  subscribe: Subscribe;
};

export function ConversationListPage({
  user,
  chatApi,
  subscribe,
}: ConversationListPageProps) {
  const [state, setState] = useState<ConversationsState>({
    status: "loading",
  });
  // Changing this number loads the list again: the effect below runs
  // whenever one of the values in its list ([chatApi, reloadCount]) changes.
  const [reloadCount, setReloadCount] = useState(0);

  useEffect(() => {
    // See NewestGreeting.tsx for why "ignore" is needed.
    let ignore = false;

    async function loadConversations() {
      try {
        const conversations = await chatApi.listConversations();
        if (!ignore) {
          setState({ status: "loaded", conversations });
        }
      } catch (loadError) {
        console.error("could not load the conversations", loadError);
        if (!ignore) {
          setState({ status: "failed" });
        }
      }
    }

    loadConversations();

    return () => {
      ignore = true;
    };
  }, [chatApi, reloadCount]);

  // Live updates: a new message changes a preview and the order, and a new
  // conversation (maybe started by someone else) joins the list. The list
  // simply loads again: one small call, and always right. Delivered and
  // read don't change the list, so they are ignored here.
  useEffect(() => {
    const stop = subscribe({
      onEvent: (serverEvent) => {
        const kind = serverEvent.event.case;
        if (kind === "messageSent" || kind === "conversationCreated") {
          setReloadCount((count) => count + 1);
        }
      },
      // Connected (again): events may have been missed while it was
      // closed, so load the list again ("catching up").
      onConnected: () => {
        setReloadCount((count) => count + 1);
      },
    });
    // The cleanup: when the page closes, close the WebSocket too.
    return stop;
  }, [subscribe]);

  // After a new chat or group, load the list again: the server puts it in
  // the right place (the newest first). The old list stays on screen until
  // the new one arrives.
  function reload() {
    setReloadCount((count) => count + 1);
  }

  return (
    <section className="flex w-80 flex-col gap-6">
      <h2 className="text-2xl font-semibold text-slate-900">Chats</h2>
      <ConversationList state={state} myUserId={user.id} />
      <NewChatForm chatApi={chatApi} onStarted={reload} />
      <NewGroupForm chatApi={chatApi} onCreated={reload} />
    </section>
  );
}

type ConversationListProps = {
  state: ConversationsState;
  myUserId: string;
};

function ConversationList({ state, myUserId }: ConversationListProps) {
  if (state.status === "loading") {
    return <p className="text-slate-500">Loading your chats…</p>;
  }

  if (state.status === "failed") {
    return <p className="text-red-600">Could not load your chats.</p>;
  }

  if (state.conversations.length === 0) {
    return (
      <p className="text-slate-500">
        No chats yet. Start one with a username below.
      </p>
    );
  }

  return (
    <ul className="flex flex-col gap-2">
      {state.conversations.map((conversation) => (
        <li
          key={conversation.id}
          className="rounded border border-slate-200 bg-white px-3 py-2"
        >
          <p className="font-medium text-slate-900">
            {conversationTitle(conversation, myUserId)}
          </p>
          {/* "truncate" cuts a long text at the end of the line, with "…". */}
          <p className="truncate text-sm text-slate-500">
            {lastMessagePreview(conversation, myUserId)}
          </p>
        </li>
      ))}
    </ul>
  );
}

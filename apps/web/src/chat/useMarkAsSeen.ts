// useMarkAsSeen tells the server how far you have got in a conversation:
// "read" up to the newest message while the page is visible, or only
// "delivered" while the browser tab is in the background. It is a custom
// hook: a function that uses React's hooks, so a page can share or move
// this work in one line. Learn more: docs/learn/react.md#custom-hooks
import { useEffect, useRef } from "react";
import type { ChatApi } from "./chatClient.ts";

// newestSequence is the newest message from someone else (see
// newestSequenceFromOthers). A failed mark only means the ✓✓ is late for
// the sender, so it is logged, not shown, and tried again later.
export function useMarkAsSeen(
  chatApi: ChatApi,
  conversationId: string,
  newestSequence: bigint,
): void {
  // What we already told the server, so we don't repeat it. useRef keeps a
  // value between renders without drawing the page again when it changes.
  const told = useRef({ delivered: 0n, read: 0n });

  useEffect(() => {
    // No messages yet: nothing to mark.
    if (newestSequence === 0n) {
      return;
    }

    function mark() {
      const isVisible = document.visibilityState === "visible";
      const before = told.current;
      if (isVisible && newestSequence > before.read) {
        // Read also counts as delivered, on the server too.
        told.current = { delivered: newestSequence, read: newestSequence };
        const marking = chatApi.markRead(conversationId, newestSequence);
        marking.catch((markError) => forget(before, markError));
        return;
      }
      if (!isVisible && newestSequence > before.delivered) {
        told.current = { ...before, delivered: newestSequence };
        const marking = chatApi.markDelivered(conversationId, newestSequence);
        marking.catch((markError) => forget(before, markError));
      }
    }

    // forget undoes what we noted as told, after a failed call, so the
    // next chance (a new message, or coming back to the tab) tries again.
    function forget(
      before: { delivered: bigint; read: bigint },
      markError: unknown,
    ) {
      told.current = before;
      console.error("could not mark the messages", markError);
    }

    mark();
    // When you come back to the tab, the messages count as read.
    document.addEventListener("visibilitychange", mark);
    return () => {
      document.removeEventListener("visibilitychange", mark);
    };
  }, [chatApi, conversationId, newestSequence]);
}

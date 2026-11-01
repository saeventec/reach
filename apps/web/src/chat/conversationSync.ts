// ConversationSync keeps one conversation's messages up to date: the first
// page, older pages, live messages, and catching up on what was missed.
// It is plain TypeScript, not React, so its rules are easy to test. The
// page gives it a function to call whenever the messages change.
// Learn more: docs/learn/message-ordering.md#catching-up-in-the-app
import type { ChatApi, Message } from "./chatClient.ts";
import { mergeMessages, newestSequence } from "./messageTimeline.ts";

// What the page shows.
export type TimelineView = {
  messages: Message[];
  // True if there are messages before the oldest one shown.
  hasOlder: boolean;
};

export class ConversationSync {
  private readonly chatApi: ChatApi;
  private readonly conversationId: string;
  private readonly onChange: (view: TimelineView) => void;

  private messages: Message[] = [];
  private hasOlder = false;
  // False until the first page arrived. Until then, there is nothing to
  // catch up from.
  private isLoaded = false;
  // Only one catch-up runs at a time. A request during one sets
  // catchUpAgain, and the running one goes once more at the end.
  private isCatchingUp = false;
  private catchUpAgain = false;

  constructor(
    chatApi: ChatApi,
    conversationId: string,
    onChange: (view: TimelineView) => void,
  ) {
    this.chatApi = chatApi;
    this.conversationId = conversationId;
    this.onChange = onChange;
  }

  // loadNewest loads the newest page: to open the conversation.
  async loadNewest(): Promise<void> {
    const page = await this.chatApi.listMessages(this.conversationId, {
      kind: "newest",
    });
    // Merged, not replaced: live messages may have arrived meanwhile.
    this.messages = mergeMessages(this.messages, page.messages);
    this.hasOlder = page.hasMore;
    this.isLoaded = true;
    this.notify();
  }

  // loadOlder loads the page before the oldest message shown.
  async loadOlder(): Promise<void> {
    const oldest = this.messages[0];
    if (oldest === undefined) {
      return;
    }
    const page = await this.chatApi.listMessages(this.conversationId, {
      kind: "before",
      sequence: oldest.sequence,
    });
    this.messages = mergeMessages(this.messages, page.messages);
    this.hasOlder = page.hasMore;
    this.notify();
  }

  // add puts in a message we know is stored: the answer to our own send,
  // or a live one from the WebSocket. A copy is ignored. If the message
  // leaves a gap (we have up to 42, and 45 arrives), it catches up on the
  // missing ones.
  add(message: Message): void {
    const expected = newestSequence(this.messages) + 1n;
    this.messages = mergeMessages(this.messages, [message]);
    this.notify();

    const leavesGap = message.sequence > expected;
    if (this.isLoaded && leavesGap) {
      this.catchUp();
    }
  }

  // catchUp loads every message after the newest one shown, page by page:
  // after a reconnect, or after a gap. Errors are logged; the next
  // reconnect tries again.
  async catchUp(): Promise<void> {
    if (!this.isLoaded) {
      return;
    }
    if (this.isCatchingUp) {
      this.catchUpAgain = true;
      return;
    }

    this.isCatchingUp = true;
    try {
      do {
        this.catchUpAgain = false;
        await this.loadAllAfterNewest();
      } while (this.catchUpAgain);
    } catch (catchUpError) {
      console.error("could not catch up on the messages", catchUpError);
    } finally {
      // "finally" runs whether the work succeeded or threw.
      this.isCatchingUp = false;
    }
  }

  // loadAllAfterNewest asks for the messages after the newest one, until
  // the server says there are no more.
  private async loadAllAfterNewest(): Promise<void> {
    let hasMore = true;
    while (hasMore) {
      const page = await this.chatApi.listMessages(this.conversationId, {
        kind: "after",
        sequence: newestSequence(this.messages),
      });
      this.messages = mergeMessages(this.messages, page.messages);
      this.notify();
      hasMore = page.hasMore;
    }
  }

  // notify gives the page a fresh view. A new array each time, so React
  // sees that something changed.
  private notify(): void {
    const view = { messages: [...this.messages], hasOlder: this.hasOlder };
    this.onChange(view);
  }
}

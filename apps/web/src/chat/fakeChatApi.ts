// A fake ChatApi for tests: every call fails with "not used in this test",
// unless the test gives its own. Only tests import this file, so it is not
// part of the app the browser loads. Learn more: docs/learn/testing.md#fakes
import type { ChatApi } from "./chatClient.ts";

export function fakeChatApi(changes: Partial<ChatApi> = {}): ChatApi {
  const notUsedHere = async () => {
    throw new Error("not used in this test");
  };
  const defaults: ChatApi = {
    listConversations: notUsedHere,
    getConversation: notUsedHere,
    startDirectConversation: notUsedHere,
    createGroup: notUsedHere,
    sendMessage: notUsedHere,
    listMessages: notUsedHere,
    markDelivered: notUsedHere,
    markRead: notUsedHere,
  };
  // "...": copy all fields of defaults, then overwrite them with changes.
  return { ...defaults, ...changes };
}

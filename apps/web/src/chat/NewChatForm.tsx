// The form that starts a 1:1 chat with someone, by their username.
// Learn more: docs/learn/react.md#forms
import { useState } from "react";
import type { FormEvent } from "react";
import { errorMessage } from "../shared/errorMessage.ts";
import { TextField } from "../shared/FormFields.tsx";
import type { ChatApi, Conversation } from "./chatClient.ts";

type NewChatFormProps = {
  chatApi: ChatApi;
  // Called with the conversation: a new one, or the one you already had
  // with this person (there is only one per pair).
  onStarted: (conversation: Conversation) => void;
};

export function NewChatForm({ chatApi, onStarted }: NewChatFormProps) {
  const [username, setUsername] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const conversation = await chatApi.startDirectConversation(username);
      setUsername("");
      onStarted(conversation);
    } catch (startError) {
      setError(errorMessage(startError));
    }
    setIsSubmitting(false);
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-2">
      <h3 className="text-lg font-semibold text-slate-900">New chat</h3>
      <TextField
        label="Username"
        value={username}
        onChange={setUsername}
        autoComplete="off"
      />
      {error !== "" && <p className="text-red-600">{error}</p>}
      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {isSubmitting ? "Starting…" : "Start chat"}
      </button>
    </form>
  );
}

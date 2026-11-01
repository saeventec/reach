// The form that creates a group: a name, and the other members'
// usernames. You are added on your own.
import { useState } from "react";
import type { FormEvent } from "react";
import { errorMessage } from "../shared/errorMessage.ts";
import { TextField } from "../shared/FormFields.tsx";
import type { ChatApi, Conversation } from "./chatClient.ts";
import { parseUsernames } from "./conversationText.ts";

type NewGroupFormProps = {
  chatApi: ChatApi;
  onCreated: (conversation: Conversation) => void;
};

export function NewGroupForm({ chatApi, onCreated }: NewGroupFormProps) {
  const [name, setName] = useState("");
  const [members, setMembers] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const usernames = parseUsernames(members);
      const conversation = await chatApi.createGroup(name, usernames);
      setName("");
      setMembers("");
      onCreated(conversation);
    } catch (createError) {
      // The service checks the rules (a name of 1 to 64 characters, 1 to
      // 49 other members) and says which one was broken.
      setError(errorMessage(createError));
    }
    setIsSubmitting(false);
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-2">
      <h3 className="text-lg font-semibold text-slate-900">New group</h3>
      <TextField
        label="Group name"
        value={name}
        onChange={setName}
        autoComplete="off"
      />
      <TextField
        label="Members (usernames, separated by commas)"
        value={members}
        onChange={setMembers}
        autoComplete="off"
      />
      {error !== "" && <p className="text-red-600">{error}</p>}
      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {isSubmitting ? "Creating…" : "Create group"}
      </button>
    </form>
  );
}

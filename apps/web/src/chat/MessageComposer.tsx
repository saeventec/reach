// The "write a message" form at the bottom of a conversation.
import { useState } from "react";
import type { FormEvent } from "react";

type MessageComposerProps = {
  // Called with the typed text. Sending happens in the page; the form only
  // clears itself, so you can type the next message at once.
  onSend: (text: string) => void;
};

export function MessageComposer({ onSend }: MessageComposerProps) {
  const [text, setText] = useState("");

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    // Only spaces: nothing to send. (The server would refuse it too.)
    if (text.trim() === "") {
      return;
    }
    onSend(text);
    setText("");
  }

  return (
    <form onSubmit={handleSubmit} className="flex gap-2">
      <label className="sr-only" htmlFor="message-text">
        Message
      </label>
      {/* "sr-only" hides the label on screen, but screen readers still
          read it for the input. */}
      <input
        id="message-text"
        type="text"
        value={text}
        onChange={(event) => setText(event.target.value)}
        placeholder="Write a message"
        autoComplete="off"
        className="flex-1 rounded border border-slate-300 px-3 py-2 text-slate-900"
      />
      <button
        type="submit"
        className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
      >
        Send
      </button>
    </form>
  );
}

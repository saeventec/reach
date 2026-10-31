// Shows the newest greeting, loaded from the API when the component appears.
// Learn more: docs/learn/react.md#state-and-effects
import { useEffect, useState } from "react";
import type { Greeting } from "./greetingClient.ts";

// The three situations the component can be in. TypeScript makes sure
// "greeting" can only be used once the status is "loaded".
type LoadState =
  | { status: "loading" }
  | { status: "loaded"; greeting: Greeting }
  | { status: "failed" };

type NewestGreetingProps = {
  // How to load the greeting. The app passes the real API call; tests pass
  // a fake. Learn more: docs/learn/testing.md#fakes
  loadGreeting: () => Promise<Greeting>;
};

export function NewestGreeting({ loadGreeting }: NewestGreetingProps) {
  const [state, setState] = useState<LoadState>({ status: "loading" });

  useEffect(() => {
    // In development, React runs effects twice to find bugs, and a
    // component can disappear before loading ends. "ignore" makes sure an
    // old, unneeded result never changes the state.
    let ignore = false;

    async function load() {
      try {
        const greeting = await loadGreeting();
        if (!ignore) {
          setState({ status: "loaded", greeting });
        }
      } catch (error) {
        console.error("could not load the greeting", error);
        if (!ignore) {
          setState({ status: "failed" });
        }
      }
    }

    load();

    // React calls this "cleanup" function when the effect is no longer needed.
    return () => {
      ignore = true;
    };
  }, [loadGreeting]);

  if (state.status === "loading") {
    return <p className="text-slate-500">Loading the greeting…</p>;
  }

  if (state.status === "failed") {
    return <p className="text-red-600">Could not load the greeting.</p>;
  }

  return (
    <div className="text-center">
      <p className="text-2xl text-slate-700">{state.greeting.message}</p>
      <p className="text-sm text-blue-600">
        Greeting #{state.greeting.id.toString()}
      </p>
    </div>
  );
}

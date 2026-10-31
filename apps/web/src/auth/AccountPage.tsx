// The page of a logged-in user: who you are, and logging out.
import { useState } from "react";
import { errorMessage } from "./authClient.ts";
import type { AuthApi, User } from "./authClient.ts";

type AccountPageProps = {
  user: User;
  authApi: AuthApi;
  // Called after this device logged out.
  onLoggedOut: () => void;
};

export function AccountPage({ user, authApi, onLoggedOut }: AccountPageProps) {
  const [error, setError] = useState("");

  async function handleLogOut() {
    setError("");
    try {
      await authApi.logOut();
      onLoggedOut();
    } catch (logOutError) {
      setError(errorMessage(logOutError));
    }
  }

  return (
    <section className="flex w-80 flex-col gap-4">
      <h2 className="text-2xl font-semibold text-slate-900">Your account</h2>
      <p className="text-slate-700">
        Logged in as <strong>{user.username}</strong>
      </p>

      {error !== "" && <p className="text-red-600">{error}</p>}

      <button
        type="button"
        onClick={handleLogOut}
        className="rounded border border-slate-300 px-4 py-2 text-slate-700 hover:bg-slate-100"
      >
        Log out
      </button>
    </section>
  );
}

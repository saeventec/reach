// The login form. Learn more: docs/learn/react.md#forms
import { useState } from "react";
import type { FormEvent } from "react";
import { Link } from "react-router";
import { errorMessage } from "./authClient.ts";
import type { AuthApi, User } from "./authClient.ts";
import { PasswordField, TextField } from "./FormFields.tsx";

type LogInPageProps = {
  authApi: AuthApi;
  // Called with the user after a successful login.
  onLoggedIn: (user: User) => void;
};

export function LogInPage({ authApi, onLoggedIn }: LogInPageProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    // A form normally reloads the page when sent. We send the data
    // ourselves instead.
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const user = await authApi.logIn(username, password);
      onLoggedIn(user);
    } catch (logInError) {
      setError(errorMessage(logInError));
      setIsSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex w-80 flex-col gap-4">
      <h2 className="text-2xl font-semibold text-slate-900">Log in</h2>

      <TextField
        label="Username"
        value={username}
        onChange={setUsername}
        autoComplete="username"
      />
      <PasswordField
        label="Password"
        value={password}
        onChange={setPassword}
        autoComplete="current-password"
      />

      {error !== "" && <p className="text-red-600">{error}</p>}

      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {isSubmitting ? "Logging in…" : "Log in"}
      </button>

      <p className="text-sm text-slate-600">
        New here?{" "}
        <Link to="/signup" className="text-blue-600 hover:underline">
          Create an account
        </Link>
      </p>
    </form>
  );
}

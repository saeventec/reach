// The sign-up form. After creating the account, it logs the new user in,
// so they don't have to type everything again.
import { useState } from "react";
import type { FormEvent } from "react";
import { Link } from "react-router";
import { errorMessage } from "../shared/errorMessage.ts";
import type { AuthApi, User } from "./authClient.ts";
import { PasswordField, TextField } from "../shared/FormFields.tsx";

type SignUpPageProps = {
  authApi: AuthApi;
  // Called with the user after the account was created and logged in.
  onLoggedIn: (user: User) => void;
};

export function SignUpPage({ authApi, onLoggedIn }: SignUpPageProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      await authApi.signUp(username, password);
      const user = await authApi.logIn(username, password);
      onLoggedIn(user);
    } catch (signUpError) {
      setError(errorMessage(signUpError));
      setIsSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex w-80 flex-col gap-4">
      <h2 className="text-2xl font-semibold text-slate-900">
        Create an account
      </h2>

      <TextField
        label="Username"
        value={username}
        onChange={setUsername}
        autoComplete="username"
      />
      <p className="-mt-3 text-xs text-slate-500">
        3 to 32 characters: a-z, 0-9, and . _ - between them.
      </p>

      <PasswordField
        label="Password"
        value={password}
        onChange={setPassword}
        autoComplete="new-password"
      />
      <p className="-mt-3 text-xs text-slate-500">
        8 to 128 characters. A few words make a good password.
      </p>

      {error !== "" && <p className="text-red-600">{error}</p>}

      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {isSubmitting ? "Creating your account…" : "Create account"}
      </button>

      <p className="text-sm text-slate-600">
        Already have an account?{" "}
        <Link to="/login" className="text-blue-600 hover:underline">
          Log in
        </Link>
      </p>
    </form>
  );
}

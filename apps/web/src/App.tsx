// The main component of the web app: it finds out who is logged in, and
// shows the page for the current address. Learn more: docs/learn/react.md
// and docs/learn/react-router.md
// The class names are Tailwind CSS. Learn more: docs/learn/tailwind.md
import { useEffect, useState } from "react";
import { NavLink, Navigate, Route, Routes } from "react-router";
import { AccountPage } from "./auth/AccountPage.tsx";
import type { AuthApi, User } from "./auth/authClient.ts";
import { LogInPage } from "./auth/LogInPage.tsx";
import { SignUpPage } from "./auth/SignUpPage.tsx";
import type { ChatApi } from "./chat/chatClient.ts";
import { ConversationListPage } from "./chat/ConversationListPage.tsx";
import type { Subscribe } from "./chat/liveUpdates.ts";
import type { Greeting } from "./greeting/greetingClient.ts";
import { NewestGreeting } from "./greeting/NewestGreeting.tsx";

// What we know about the logged-in user. "user: null" means: we asked, and
// nobody is logged in.
type UserState =
  | { status: "loading" }
  | { status: "loaded"; user: User | null }
  | { status: "failed" };

type AppProps = {
  // The API calls. main.tsx passes the real ones; tests pass fakes.
  authApi: AuthApi;
  chatApi: ChatApi;
  // Opens the live updates (a WebSocket). Tests pass a fake.
  subscribe: Subscribe;
  loadGreeting: () => Promise<Greeting>;
};

function App({ authApi, chatApi, subscribe, loadGreeting }: AppProps) {
  const [userState, setUserState] = useState<UserState>({ status: "loading" });

  // Ask the auth service once, when the app starts, who is logged in. The
  // browser sends the session cookie with the call.
  useEffect(() => {
    // See NewestGreeting.tsx for why "ignore" is needed.
    let ignore = false;

    async function loadUser() {
      try {
        const user = await authApi.getCurrentUser();
        if (!ignore) {
          setUserState({ status: "loaded", user });
        }
      } catch (error) {
        console.error("could not check who is logged in", error);
        if (!ignore) {
          setUserState({ status: "failed" });
        }
      }
    }

    loadUser();

    return () => {
      ignore = true;
    };
  }, [authApi]);

  function handleLoggedIn(user: User) {
    setUserState({ status: "loaded", user });
  }

  function handleLoggedOut() {
    setUserState({ status: "loaded", user: null });
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-8 bg-slate-50">
      <h1 className="text-4xl font-bold text-slate-900">Reach</h1>

      {userState.status === "loading" && (
        <p className="text-slate-500">Loading…</p>
      )}

      {userState.status === "failed" && (
        <p className="text-red-600">
          Could not reach Reach. Please reload the page.
        </p>
      )}

      {userState.status === "loaded" && userState.user !== null && <MainMenu />}

      {userState.status === "loaded" && (
        <AppRoutes
          user={userState.user}
          authApi={authApi}
          chatApi={chatApi}
          subscribe={subscribe}
          onLoggedIn={handleLoggedIn}
          onLoggedOut={handleLoggedOut}
        />
      )}

      <footer>
        <NewestGreeting loadGreeting={loadGreeting} />
      </footer>
    </main>
  );
}

// MainMenu links the pages of a logged-in user. NavLink is a Link that
// knows whether its page is the one showing ("isActive"), to highlight it.
function MainMenu() {
  function linkClass({ isActive }: { isActive: boolean }): string {
    return isActive ? "font-semibold text-slate-900" : "text-blue-600";
  }

  return (
    <nav className="flex gap-6">
      <NavLink to="/chats" className={linkClass}>
        Chats
      </NavLink>
      <NavLink to="/account" className={linkClass}>
        Account
      </NavLink>
    </nav>
  );
}

type AppRoutesProps = {
  user: User | null;
  authApi: AuthApi;
  chatApi: ChatApi;
  subscribe: Subscribe;
  onLoggedIn: (user: User) => void;
  onLoggedOut: () => void;
};

// The pages, one per address. Each one checks whether someone is logged
// in, and sends you to the right page if not. <Navigate> changes the
// address; "replace" means the Back button skips the page that redirected.
function AppRoutes({
  user,
  authApi,
  chatApi,
  subscribe,
  onLoggedIn,
  onLoggedOut,
}: AppRoutesProps) {
  const isLoggedIn = user !== null;

  const logInPage = isLoggedIn ? (
    <Navigate to="/chats" replace />
  ) : (
    <LogInPage authApi={authApi} onLoggedIn={onLoggedIn} />
  );

  const signUpPage = isLoggedIn ? (
    <Navigate to="/chats" replace />
  ) : (
    <SignUpPage authApi={authApi} onLoggedIn={onLoggedIn} />
  );

  const accountPage = isLoggedIn ? (
    <AccountPage user={user} authApi={authApi} onLoggedOut={onLoggedOut} />
  ) : (
    <Navigate to="/login" replace />
  );

  const chatsPage = isLoggedIn ? (
    <ConversationListPage user={user} chatApi={chatApi} subscribe={subscribe} />
  ) : (
    <Navigate to="/login" replace />
  );

  // Any other address, like "/", goes to the page that fits.
  const startPage = isLoggedIn ? "/chats" : "/login";

  return (
    <Routes>
      <Route path="/login" element={logInPage} />
      <Route path="/signup" element={signUpPage} />
      <Route path="/account" element={accountPage} />
      <Route path="/chats" element={chatsPage} />
      <Route path="*" element={<Navigate to={startPage} replace />} />
    </Routes>
  );
}

export default App;

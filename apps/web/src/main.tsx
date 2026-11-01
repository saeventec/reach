// The entry point of the web app: it puts the App component into the page.
// Learn more: docs/learn/react.md
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";
import "./index.css";
import App from "./App.tsx";
import { authApi } from "./auth/authClient.ts";
import { chatApi } from "./chat/chatClient.ts";
import { subscribeToLiveUpdates } from "./chat/liveUpdates.ts";
import { fetchNewestGreeting } from "./greeting/greetingClient.ts";

// BrowserRouter connects React Router to the browser's address bar and its
// Back and Forward buttons. Learn more: docs/learn/react-router.md
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <App
        authApi={authApi}
        chatApi={chatApi}
        subscribe={subscribeToLiveUpdates}
        loadGreeting={fetchNewestGreeting}
      />
    </BrowserRouter>
  </StrictMode>,
);

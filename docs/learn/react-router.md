# React Router

## What it is

A **router** shows a different page for each address in the browser's address bar (`/login`, `/signup`, `/account`), without loading a new HTML page from the server. The Back and Forward buttons and bookmarks work as people expect.

Our web app is a **single-page app** (SPA): the server always sends the same `index.html`, and React draws the right page in the browser. React Router connects the address bar to React.

## Why we use it

The owner chose real addresses over switching screens with plain React state ([ADR 0026](../decisions/0026-use-react-router.md)), so pages can be bookmarked and the Back button works. React Router is the most used router for React. Version 8 has three "modes"; we use the simplest one, **declarative mode**: routes are React components, nothing else.

## Tiny example

From [apps/web/src/main.tsx](../../apps/web/src/main.tsx) and [App.tsx](../../apps/web/src/App.tsx):

```tsx
import { BrowserRouter, Navigate, Route, Routes } from "react-router";

// main.tsx: connect the router to the address bar, once, around the app.
<BrowserRouter>
  <App />
</BrowserRouter>

// App.tsx: one Route per address. Only the one that matches is drawn.
<Routes>
  <Route path="/login" element={logInPage} />
  <Route path="/account" element={accountPage} />
  <Route path="*" element={<Navigate to={startPage} replace />} />   {/* any other address */}
</Routes>
```

- **`<Link to="/signup">`** is a link that changes the address without reloading the page. (A plain `<a href>` would reload the whole app.)
- **`<Navigate to="/login" replace />`** changes the address as soon as it is drawn: a redirect. `replace` swaps the current entry in the browser's history, so Back doesn't return to the page that redirected you.

**`<NavLink>`** is a `<Link>` that knows whether its page is the one showing. Its `className` can be a function that gets `{ isActive }`, so the menu (`MainMenu` in [App.tsx](../../apps/web/src/App.tsx)) highlights the current page.

**URL parameters.** `<Route path="/chats/:conversationId">` matches any address like `/chats/01a19641-…`, and `useParams()` gives the part in that place: `{ conversationId: "01a19641-…" }`. One route serves every conversation.

**`key` to start fresh.** `ConversationRoute` in [App.tsx](../../apps/web/src/App.tsx) passes the id as the page's `key`. When you open another chat, the key changes, so React throws the old page (with its messages in state) away and makes a new one. Without it, React would reuse the page and its old state.

## Pages that depend on being logged in

`App` keeps the logged-in user in state, and each route decides what to draw from it:

```tsx
const accountPage = isLoggedIn ? (
  <AccountPage user={user} ... />
) : (
  <Navigate to="/login" replace />
);
```

So after a login, the app only sets the user. On the next draw, `/login` sees a user and redirects to `/account` by itself. No code has to "go to the account page" by hand: the address follows from the state. This is React's **declarative** style: you describe what the page should look like for each state, and React does the rest.

This is a convenience for the user, not security. Anyone can change the code in their browser. The real check is on the server: every call that needs a login checks the session cookie.

## How the address reaches the app

When you open `http://localhost:8080/account` directly, the browser asks the server for `/account`. Vite's development server answers every address that isn't a file with `index.html`, so the app starts, and React Router reads the address. (When we serve the built app ourselves in a later milestone, the server must do the same.)

## Testing

Tests use **`MemoryRouter`** instead of `BrowserRouter`: it keeps the address in memory, so a test can start at any address without a real address bar:

```tsx
render(
  <MemoryRouter initialEntries={["/account"]}>
    <App authApi={fakeAuthApi()} ... />
  </MemoryRouter>,
);
await screen.findByRole("heading", { name: "Log in" });   // redirected
```

## Alternatives

- **Plain React state:** a `screen` variable decides what to show. No library, but no real addresses, and the Back button leaves the app.
- **TanStack Router:** fully type-safe routes, newer, more to learn.
- **wouter:** a very small router with a similar API, less used.
- **A framework (Next.js, React Router framework mode):** routing plus server rendering. Much more than we need.

## Learn more

- [React Router: declarative mode](https://reactrouter.com/start/declarative/installation)
- [Routing](https://reactrouter.com/start/declarative/routing) and [navigating](https://reactrouter.com/start/declarative/navigating)

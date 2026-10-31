# React

## What it is

React is a library for building user interfaces from **components**. A component is a function that returns what should appear on the screen. You combine small components into bigger ones, like building blocks.

Components are written in **JSX**: HTML-like tags inside TypeScript. JSX is turned into normal JavaScript before the browser runs it.

## Why we use it

The most used UI library, with the most learning material. See [ADR 0004](../decisions/0004-react-typescript-vite-for-web.md).

## Tiny example

`apps/web/src/App.tsx`:

```tsx
function App() {
  return (
    <main>
      <h1>Reach</h1>
    </main>
  );
}

export default App;   // so other files can import it
```

`apps/web/src/main.tsx` puts it into the page:

```tsx
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
```

- `document.getElementById("root")` finds the empty `<div id="root">` in `index.html`. React draws the app inside it.
- The `!` tells TypeScript "this is never null here". The `<div>` is always in `index.html`.
- `StrictMode` runs extra checks while developing, to warn about common mistakes. It does nothing in the final build.
- Files with JSX end in `.tsx` (TypeScript + JSX).

## Props

Props are a component's inputs, like a function's parameters. TypeScript describes them:

```tsx
type NewestGreetingProps = {
  loadGreeting: () => Promise<Greeting>;   // a function that returns a greeting, later
};

export function NewestGreeting({ loadGreeting }: NewestGreetingProps) { ... }

<NewestGreeting loadGreeting={fetchNewestGreeting} />   // using it
```

## State and effects

- **State** (`useState`) is data a component remembers. When it changes, React draws the component again.
- **Effect** (`useEffect`) is code that runs after the component appears on the screen, for work outside React, like loading data from the network.

From `apps/web/src/greeting/NewestGreeting.tsx`:

```tsx
const [state, setState] = useState<LoadState>({ status: "loading" });

useEffect(() => {
  let ignore = false;
  async function load() {
    const greeting = await loadGreeting();
    if (!ignore) {
      setState({ status: "loaded", greeting });
    }
  }
  load();
  return () => { ignore = true; };    // "cleanup": runs when the effect is no longer needed
}, [loadGreeting]);                   // run again only if loadGreeting changes
```

- The `ignore` flag: in development, `StrictMode` runs every effect twice to find bugs, and a component can disappear before loading ends. The flag makes sure an old result never changes the state.
- `LoadState` is a **union type**: the state is either loading, loaded (with a greeting), or failed. TypeScript only lets you use `state.greeting` after checking that `state.status === "loaded"`.
- Functions whose names start with `use` are called **hooks**. Call them only at the top of a component, never inside an `if`.

Later, TanStack Query can do this loading for us, with caching and retries. Writing it by hand first shows what such a library saves (the owner chose "later" in milestone 2).

## Forms

From [apps/web/src/auth/LogInPage.tsx](../../apps/web/src/auth/LogInPage.tsx):

```tsx
const [username, setUsername] = useState("");

async function handleSubmit(event: FormEvent<HTMLFormElement>) {
  event.preventDefault();          // don't let the browser reload the page
  const user = await authApi.logIn(username, password);
  onLoggedIn(user);
}

<form onSubmit={handleSubmit}>
  <TextField label="Username" value={username} onChange={setUsername} autoComplete="username" />
  <button type="submit">Log in</button>
</form>
```

- **Controlled inputs:** the input shows `value` from state, and `onChange` puts every keystroke back into state. So the state always holds what is on the screen.
- **`<form onSubmit>`** runs on a click on the submit button and on Enter in a field. `preventDefault()` stops the browser's old behavior of sending the form and loading a new page.
- **`autoComplete`** tells the browser and password managers what a field is: `username`, `current-password` (login) or `new-password` (sign-up, so they can suggest a strong one).
- **`<label>` around the input** connects the two: clicking the label focuses the input, and screen readers read the label. Tests find inputs the same way: `screen.getByLabelText("Username")`.
- While the call runs, the button is `disabled`, so a second click can't send the form twice.
- **Errors:** `errorMessage` in [authClient.ts](../../apps/web/src/auth/authClient.ts) shows the service's message for errors written for people ("the username is already taken"), and a general message for everything else, because an `internal` error's text is not meant for users.

## Folders by feature

Code is grouped by what it does: `src/greeting/` holds the component, its API client and its test, and `src/auth/` holds the login, sign-up and account pages with their API client. `src/gen/` holds generated code (never edit it).

## Alternatives

- **Vue:** templates closer to plain HTML, popular, a gentler start.
- **Svelte:** compiles components away, so there is less code in the browser.
- **Solid:** looks like React, faster updates, smaller community.

## Learn more

- https://react.dev/learn

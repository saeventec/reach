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

Later (milestone 2), TanStack Query can do this loading for us, with caching and retries. Writing it by hand once shows what such a library saves.

## Folders by feature

Code is grouped by what it does: `src/greeting/` holds the component, its API client and its test. `src/gen/` holds generated code (never edit it).

## Alternatives

- **Vue:** templates closer to plain HTML, popular, a gentler start.
- **Svelte:** compiles components away, so there is less code in the browser.
- **Solid:** looks like React, faster updates, smaller community.

## Learn more

- https://react.dev/learn

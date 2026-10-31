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
  )
}

export default App   // so other files can import it
```

`apps/web/src/main.tsx` puts it into the page:

```tsx
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
```

- `document.getElementById('root')` finds the empty `<div id="root">` in `index.html`. React draws the app inside it.
- The `!` tells TypeScript "this is never null here". The `<div>` is always in `index.html`.
- `StrictMode` runs extra checks while developing, to warn about common mistakes. It does nothing in the final build.
- Files with JSX end in `.tsx` (TypeScript + JSX).

## Alternatives

- **Vue:** templates closer to plain HTML, popular, a gentler start.
- **Svelte:** compiles components away, so there is less code in the browser.
- **Solid:** looks like React, faster updates, smaller community.

## Learn more

- https://react.dev/learn

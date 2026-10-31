# TypeScript

## What it is

TypeScript is JavaScript with **types**. You say what kind of value each thing is, and the TypeScript compiler (`tsc`) checks the code before it runs. Browsers only understand JavaScript, so the types are removed when the code is built.

```ts
function greet(name: string): string {
  return "Hello, " + name;
}

greet(42); // error before the code even runs: 42 is not a string
```

## Why we use it

It catches mistakes early, and editors can autocomplete better. See [ADR 0004](../decisions/0004-react-typescript-vite-for-web.md). We use TypeScript 7, a new and much faster version of the compiler.

## Our config: the tsconfig files

They came with Vite's starter and are unchanged:

- `tsconfig.json`: only points to the two files below.
- `tsconfig.app.json`: rules for the app code in `src/` (it runs in the browser).
- `tsconfig.node.json`: rules for `vite.config.ts` (it runs in Node.js).

Some settings worth knowing:

- `"noEmit": true`: `tsc` only checks the code. Vite does the actual building.
- `"jsx": "react-jsx"`: understand JSX in `.tsx` files.
- `"noUnusedLocals"`, `"noUnusedParameters"`: report variables and parameters that are never used.
- `"target": "es2023"`: the JavaScript version to check against. Modern browsers support it.

`npm run build` runs `tsc -b` (check all the code) and then `vite build`.

## Alternatives

- **Plain JavaScript:** no types, so mistakes show up only when the code runs.
- **JSDoc comments:** types written in comments, checked by TypeScript. Less clear.
- **Dart or Elm:** other typed languages that compile to JavaScript, far less common.

## Learn more

- https://www.typescriptlang.org/docs/handbook/typescript-in-5-minutes.html

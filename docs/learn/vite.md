# Vite

## What it is

Vite is the build tool for the web app. It has two jobs:

- **Development server** (`npm run dev`): serves the app on port 5173. With `task dev`, you open it through the gateway at http://localhost:8080. When you save a file, the page updates within a moment, without a full reload. This is called **hot module replacement (HMR)**.
- **Build** (`npm run build`): bundles all files into a few small, optimized files in `dist/`, ready for a real web server.

## Why we use it

It needs almost no configuration and is very fast. See [ADR 0004](../decisions/0004-react-typescript-vite-for-web.md).

## How we set it up

We started from Vite's official starter (`npm create vite@latest -- --template react-ts`) and kept its files. We removed the demo page, its images and CSS, and its linter (we choose one in step 7).

`apps/web/vite.config.ts`:

```ts
export default defineConfig({
  plugins: [react()],   // understand JSX and support HMR for React
  server: {
    host: true,         // listen on all network interfaces (needed in a container)
  },
})
```

- `index.html`: the one HTML page. It loads `src/main.tsx`.
- `src/`: the app's code.
- `Dockerfile`: runs the development server in a container, with `task dev`. It is only for development. Later, a real server will serve the built `dist/` files instead.

## Alternatives

- **webpack:** the older standard, much more configuration.
- **Parcel:** also zero-config, less used.
- **Next.js:** a full framework with its own server. We don't need it, because our backend is Go.

## Learn more

- https://vite.dev/guide/

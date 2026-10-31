# npm and package.json

## What it is

npm is Node.js's package manager: it downloads the JavaScript libraries ("packages") a project uses into a `node_modules/` folder. npm comes with Node.js, which mise installs.

- **`package.json`:** the project's name, its commands (`scripts`), and the libraries it needs.
  - `dependencies`: needed by the app in the browser (React).
  - `devDependencies`: only needed while developing and building (Vite, TypeScript).
- **`package-lock.json`:** the exact version of every library, including the libraries our libraries use. It is committed, so everyone gets the same versions.
- **Version ranges:** `^19.3.1` means "19.3.1 or a newer 19.x". `~7.0.2` means "7.0.2 or a newer 7.0.x". The lockfile pins the exact one.

## Why we use it

It is the standard for JavaScript projects, and it comes with Node.js, so there is no extra tool to install.

## Tiny example

```sh
npm ci                   # install exactly what package-lock.json says (setup, Docker, CI)
npm install some-package # add a new library (updates package.json and the lockfile)
npm run dev              # run the "dev" script from package.json
```

`task setup` runs `npm ci --prefix apps/web`, so your editor can check the web app's code. The container installs its own Linux copy.

## Install scripts

A package can include an **install script**: code that runs on your computer during `npm install`. That's a common way to sneak in malware. So npm blocks these scripts unless the project approves them, in `package.json`:

```json
"allowScripts": {
  "fsevents": false
}
```

`fsevents` (a Mac-only file watcher) is denied: we don't need it, because Vite runs in Docker. Commands: `npm install-scripts ls` (what's waiting for a decision), `npm install-scripts approve <name>`, `npm install-scripts deny <name>`.

## Alternatives

- **pnpm:** faster, and saves disk space by sharing packages between projects. One more tool to install.
- **Yarn:** an older alternative to npm, similar features today.
- **Bun:** a whole new JavaScript runtime with a very fast package manager, younger.

## Learn more

- https://docs.npmjs.com/about-npm
- https://docs.npmjs.com/cli/commands/npm-ci

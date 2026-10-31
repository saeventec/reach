# Docker Compose

## What it is

Docker Compose starts several containers together, described in one file: `compose.yaml`. Instead of typing one long `docker run ...` command per service, you describe them all once and run one command.

Compose also creates a private network, so containers can reach each other by their service name (for example `http://hello:8080`).

## Why we use it

Reach is made of several services (soon: hello, Postgres, the web app, Caddy). Compose starts all of them with `task dev`. Each Compose service maps to one Kubernetes deployment later.

## Tiny example

Our `compose.yaml`:

```yaml
services:
  hello:                          # the service name
    build: ./services/hello       # build the image from this folder's Dockerfile
    ports:
      - "127.0.0.1:8080:8080"     # your computer's port 8080 → the container's port 8080
    healthcheck:                  # how Docker checks that it is alive
      test: ["CMD", "wget", "--quiet", "--spider", "http://localhost:8080/healthz"]
```

- **ports:** without this line, the service is only reachable inside the Compose network. `127.0.0.1` keeps it private to your computer.
- **healthcheck:** Docker runs the `test` command regularly. `wget --spider` asks for the page without saving it. The timing is explained in "Start order" below.

## Start order: depends_on

`depends_on` makes a service wait for another one. The `condition` says what to wait for:

```yaml
hello-migrate:
  depends_on:
    postgres:
      condition: service_healthy                 # wait until its health check passes

hello:
  depends_on:
    hello-migrate:
      condition: service_completed_successfully  # wait until it exited with code 0
```

So `task dev` starts: postgres (healthy) → hello-migrate (finished) → hello.

`start_period` and `start_interval` in a health check mean: while the container is starting (the first 30 seconds), check every second. After that, Docker checks every 30 seconds (its default). Both postgres and hello use them, so they show `(healthy)` within seconds and the next containers don't wait long.

## Settings from .env

Compose reads the `.env` file next to `compose.yaml` and fills in `${NAME}`:

```yaml
POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:?Set POSTGRES_PASSWORD in .env (run "task setup")}
```

The `:?` part stops Compose with that message if the setting is missing or empty, instead of starting with a blank password.

## Volumes

A container's own files disappear when the container is removed. A **named volume** is storage that Docker keeps separately, so data survives:

```yaml
services:
  postgres:
    volumes:
      - postgres-data:/var/lib/postgresql   # volume name : folder inside the container

volumes:
  postgres-data:                            # declare the volume
```

## Compose Watch

An image holds a copy of the code from when it was built. While you edit the web app, Compose Watch keeps the running container up to date:

```yaml
web:
  develop:
    watch:
      - action: sync                       # copy changed files into the container
        path: ./apps/web
        target: /app
      - action: rebuild                    # build a new image and replace the container
        path: ./apps/web/package-lock.json
```

- `sync`: when you save `src/App.tsx`, Compose copies it into the container, and Vite updates the page in your browser. Files listed in `.dockerignore` (like `node_modules`) are never copied.
- `rebuild`: new libraries mean a new `npm ci`, so Compose rebuilds the image.
- `task dev` turns it on with `docker compose up --watch`.

The hello service uses `rebuild` for all its files, because Go code must be compiled. Saving a `.go` file rebuilds the image and replaces the container within about 15 seconds. The `migrations/` folder is ignored: migrations are applied by `hello-migrate`, not by the service.

Known glitch: once, in testing, a rebuild of `hello` also recreated `postgres` and left it stopped. `hello` then logged `lookup postgres ... no such host`. Your data is safe (it lives in the volume). To fix it, press Ctrl+C and run `task dev` again.

Why not a **bind mount** (sharing the folder with the container)? It is the older way, but `node_modules` would then be shared too. Its Mac or Windows versions of some libraries don't run in the Linux container.

## Useful commands

```sh
task dev                  # = docker compose up --build --watch: build and start everything
docker compose ps         # list the running services and their health
docker compose ps -a      # also list services that have exited, like hello-migrate
docker compose logs hello # show one service's logs
docker compose down       # remove the stopped containers and the network (volumes are kept)
docker compose down -v    # also delete the volumes: all database data is lost!
```

Ctrl+C in the `task dev` terminal stops the containers. They still exist, stopped, until `docker compose down`.

## Alternatives

- **Plain `docker run` commands:** fine for one container, messy for many.
- **Kubernetes (k3s, kind):** made for many servers. We move there in milestone 11.
- **Podman Compose:** the same idea for Podman.

## Learn more

- https://docs.docker.com/compose/
- https://docs.docker.com/reference/compose-file/services/

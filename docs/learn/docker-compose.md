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
- **healthcheck:** we keep Docker's default timing. It checks every 30 seconds, so a new container shows `(health: starting)` for up to 30 seconds before `(healthy)`. `wget --spider` asks for the page without saving it.

## Useful commands

```sh
task dev                  # = docker compose up --build: build and start everything
docker compose ps         # list the services and their health
docker compose logs hello # show one service's logs
docker compose down       # remove the stopped containers and the network
```

Ctrl+C in the `task dev` terminal stops the containers. They still exist, stopped, until `docker compose down`.

## Alternatives

- **Plain `docker run` commands:** fine for one container, messy for many.
- **Kubernetes (k3s, kind):** made for many servers. We move there in milestone 11.
- **Podman Compose:** the same idea for Podman.

## Learn more

- https://docs.docker.com/compose/
- https://docs.docker.com/reference/compose-file/services/

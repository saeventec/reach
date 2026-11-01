# Docker

## What it is

Docker packs a program together with everything it needs (files, libraries, settings) into an **image**. Running an image gives a **container**: an isolated box that runs the same way on every computer.

- **Image:** a read-only template, like a recipe plus its ingredients.
- **Container:** a running copy of an image, like a dish cooked from the recipe. One image can run as many containers.
- **Dockerfile:** a text file with the steps to build an image.
- **Base image:** the image you start from (`FROM ...`), such as a small Linux system.

On Mac and Windows, Docker Desktop runs a small Linux virtual machine, because containers need a Linux kernel.

## Why we use it

The whole system must start on one computer with one command, on Mac and Windows. Containers give every service the same environment everywhere, and the same images can move to Kubernetes later. See [ADR 0002](../decisions/0002-microservices-in-docker-compose.md).

## Tiny example

`services/hello/Dockerfile`, shortened. It uses a **multi-stage build**: the first stage compiles, and the second stage only keeps the result.

```dockerfile
FROM golang:1.27.3-alpine3.24 AS build   # stage 1: has the Go compiler
WORKDIR /src                             # work inside /src
COPY . .                                 # copy our source code in
RUN CGO_ENABLED=0 go build -o /hello .   # compile to one file: /hello

FROM alpine:3.24.3                       # stage 2: small Linux, no compiler
COPY --from=build /hello /hello          # take only the compiled program
USER nobody                              # don't run as root
ENTRYPOINT ["/hello"]                    # start the program
```

The final image is about 28 MB. With the compiler inside, it would be several hundred MB.

Image tags like `alpine:3.24.3` pin an exact version, so a build gives the same result next month. We update them on purpose, by hand. (A bot like Renovate could do it later.)

## The build context

The **build context** is the set of files Docker can see while it builds an image. `COPY` can only copy files from inside it. With `docker build ./services/hello`, the context is that folder, so the Dockerfile can't reach `packages/go`, which is outside it.

So our Go services are built with the **whole repo** as their context ([compose.yaml](../../compose.yaml): `context: .` and `dockerfile: services/hello/Dockerfile`), and each `COPY` names its folder from the repo's top: `COPY packages/go /src/packages/go`.

The whole repo would be a lot to send (`node_modules`, `.git`), and it holds a secret (`.env`). A **`.dockerignore`** file at the top of the context says what to leave out. Ours leaves out everything (`*`), then lets the Go folders back in (`!packages/go`, `!services`). Listing what to keep, instead of what to skip, means a new folder is never sent by accident.

## Useful commands

```sh
docker image ls        # list images on your computer
docker ps              # list running containers
docker logs <name>     # show what a container printed
```

## Alternatives

- **Podman:** works like Docker, without a background service. Very compatible.
- **OrbStack:** a faster, lighter replacement for Docker Desktop on Mac. Same commands.
- **Runtime base images:** see [ADR 0015](../decisions/0015-use-alpine-as-the-runtime-image.md).

## Learn more

- https://docs.docker.com/get-started/docker-overview/
- https://docs.docker.com/build/building/multi-stage/

# Builds the container image for the realtime service.
# It has two stages: one to compile the program, and one small image to run it.
# Learn more: docs/learn/docker.md

# Stage 1: compile. This image has the full Go toolchain.
FROM golang:1.27.3-alpine3.24 AS build

# The build context is the whole repo (see compose.yaml), because the
# service needs the shared code in packages/go too. So every COPY below
# names its folder from the repo's top. The repo's .dockerignore keeps
# everything else out. Learn more: docs/learn/docker.md#the-build-context

# Download the libraries first, in their own layer. Docker reuses a layer
# while its input files don't change, so the download only runs again when
# a go.mod or go.sum changes, not on every code change. The shared module's
# go.mod is needed too: it lists the libraries the shared code uses.
COPY packages/go/go.mod packages/go/go.sum /src/packages/go/
COPY services/realtime/go.mod services/realtime/go.sum /src/services/realtime/

# All the next commands run inside this folder of the image. The replace
# line in go.mod ("../../packages/go") points from here to the shared code.
WORKDIR /src/services/realtime
RUN go mod download

# Copy the source code: the shared code, and this service.
COPY packages/go /src/packages/go
COPY services/realtime /src/services/realtime

# CGO_ENABLED=0 builds a program that needs no C libraries from the system,
# so it runs in any small image.
RUN CGO_ENABLED=0 go build -o /realtime .

# Stage 2: run. A small Linux image, without the Go toolchain.
# Only the compiled program is copied in, so the final image stays small.
FROM alpine:3.24.3

COPY --from=build /realtime /realtime

# Run as the unprivileged "nobody" user, not as root, for safety.
USER nobody

# Documents which port the service listens on. Publishing it happens in compose.yaml.
EXPOSE 8080

# The command that runs when the container starts.
ENTRYPOINT ["/realtime"]

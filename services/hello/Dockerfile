# Builds the container image for the hello service.
# It has two stages: one to compile the program, and one small image to run it.
# Learn more: docs/learn/docker.md

# Stage 1: compile. This image has the full Go toolchain.
FROM golang:1.27.3-alpine3.24 AS build

# All the next commands run inside this folder of the image.
WORKDIR /src

# Copy the source code from services/hello into the image.
COPY . .

# CGO_ENABLED=0 builds a program that needs no C libraries from the system,
# so it runs in any small image.
RUN CGO_ENABLED=0 go build -o /hello .

# Stage 2: run. A small Linux image, without the Go toolchain.
# Only the compiled program is copied in, so the final image stays small.
FROM alpine:3.24.3

COPY --from=build /hello /hello

# Run as the unprivileged "nobody" user, not as root, for safety.
USER nobody

# Documents which port the service listens on. Publishing it happens in compose.yaml.
EXPOSE 8080

# The command that runs when the container starts.
ENTRYPOINT ["/hello"]

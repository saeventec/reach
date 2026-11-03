# Delivery and operations

How the software is built, shipped and run.

## Supporting services in one compose file

- **What:** databases, caches and proxies for local development start with one compose command, with named volumes so data survives restarts.
- **Why:** a new machine is ready in minutes, with the same versions for everyone.

## Small images built in two stages

- **What:** a container image is built in two stages: a big "build" stage with the tools, and a tiny "run" stage that only holds the finished program. Dependencies are downloaded before the source is copied in.
- **Why:** small, safer images. Rebuilds are fast, because the dependency step is reused when only the code changes.
- **Example:**

```
stage build: copy dependency list → download → copy code → compile
stage run:   tiny base image → copy only the compiled program
```

## A reverse proxy routes by host name

- **What:** one proxy in front of everything sends each host name to the right app (`api.localhost`, `app.localhost`, `dash.localhost`). Apps opt in with labels.
- **Why:** real-looking addresses locally, one entry point, and automatic HTTPS certificates in production.

## Server setup as code

- **What:** preparing a server (updates, firewall, container runtime, cluster install, fetching the access file) is a set of small automation scripts, one per job. Each one is safe to run again ("idempotent").
- **Why:** a new server is rebuilt from scratch with a few commands, not from memory.
- **Example:** `system-update`, `common-config`, `install-runtime`, `install-cluster`, `fetch-access-config`.

## Step-by-step setup guides

- **What:** the few manual steps that can't be automated (like creating the first admin user on a server) are written as a numbered guide, with the command and the expected output for each step.
- **Why:** anyone can follow it, and you know each step worked.

## GitOps: git is the source of truth

- **What:** the cluster watches the repository and applies whatever is there. Nobody changes the cluster by hand.
- **Why:** every change is reviewed, recorded and reversible through git.

## Environments share one description

- **What:** dev and prod use the same component descriptions, and each environment only patches what differs: host names, connection settings, which components it runs.
- **Why:** environments can't drift apart by accident.

## Explicit order between layers

- **What:** infrastructure (proxy, database operator) is applied first; shared services and environments wait for it ("depends on").
- **Why:** nothing starts before what it needs.

## Setup jobs separate from the server

- **What:** database migrations and starter data run as a one-off job, separate from the long-running server.
- **Why:** the server starts fast, and a failed migration doesn't crash-loop the app.

## Automatic image updates

- **What:** CI builds and pushes an image on every push to the main branch (only for the app whose folder changed). The cluster notices the new image and updates itself, recording the exact image fingerprint (digest) in git.
- **Why:** shipping is "merge and wait", and git always shows exactly what runs.
- **Note:** use version tags as well as `latest`, so rolling back is easy.

## Mirrors when the network needs them

- **What:** package and image downloads go through configurable mirrors, set in one variable.
- **Why:** builds keep working where the default registries are slow or blocked.

## Pin versions

- **What:** base images, charts and packages use explicit versions (`postgres:18`, chart `0.27.0`).
- **Why:** the same input gives the same result tomorrow.

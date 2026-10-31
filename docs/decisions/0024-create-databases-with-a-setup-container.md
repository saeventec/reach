# 0024. Create each service's database and user with a setup container

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Each service owns its data (see the brief, "Microservices"). Until now, the hello service logged in as `postgres`, the superuser, which can read and change everything. With `auth` as the second service, each service should log in as its own user that can only use its own database. Something has to create these users and databases.

## Decision

A small `postgres-setup` container runs [postgres/setup.sql](../../postgres/setup.sql) with `psql` on every start, before the migrations. The file creates each role and database only if it is missing, makes the role the database's owner, and takes away other roles' right to connect. Each service's password comes from `.env`.

## Why

- **It works on existing databases.** The owner's local data (on the Mac and on Windows) is kept, and the next service (`chat`) is added the same way, with no need to delete data.
- **Desired state:** the file says what should exist, and every start makes it true. A changed password in `.env` takes effect on the next start.
- **Everything as code:** no manual `psql` commands.

## Alternatives considered

- **The image's first-start scripts (`/docker-entrypoint-initdb.d/`):** less code, but they run only when the data folder is empty, so every new service would mean deleting the local database.
- **Create the users inside each service's migrations:** the migrations would need superuser rights, which is what we are trying to avoid.
- **One database with one schema per service:** also allowed by the brief, but separate databases separate the data more clearly, and are easier to move to separate servers later.

## Consequences

- One more container that runs and exits at each start (about one second).
- Adding a service means adding a few lines to `setup.sql` and one password to `.env.example`.
- `.env` files created before milestone 2 need the two new passwords. Compose stops with a clear message if one is missing.
- The superuser password is now only used by `postgres-setup`.

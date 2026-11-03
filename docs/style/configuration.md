# Configuration

How the program gets its settings.

## Settings come from the environment

- **What:** every setting is an environment variable. The same build runs locally, in a container, and in every environment. Only the variables change.
- **Why:** no rebuild per environment, and no settings hidden in code. This follows the "Twelve-Factor App" guidelines.

## One settings struct, read once

- **What:** all settings live in one type, with the variable name, default and "required" flag next to each field. It is read once at startup and passed to the parts that need it.
- **Why:** every setting is visible in one place. A missing required value stops the program at startup with a clear error.
- **Example:**

```
Config:
  debug        ← DEBUG          default true
  logMode      ← LOG_MODE       default "debug"
  port         ← WEB_PORT       default 8080
  databaseURL  ← POSTGRES_URL   required
  jwtSecret    ← JWT_SECRET     required
```

## Defaults for harmless settings, none for secrets

- **What:** safe settings (port, log level) have defaults. Addresses with passwords and secret keys are required, with no default.
- **Why:** it runs with little setup, but a secret can never be used by accident.

## A local file for development

- **What:** a git-ignored `.env` file holds local values, loaded only if present. A committed `.env.example` lists every variable with harmless sample values.
- **Why:** a new developer copies one file and starts. Real secrets never reach git.

## Settings as URLs

- **What:** connections to databases and caches are a single URL (`scheme://user:password@host:port/name?options`).
- **Why:** one value per connection, in a standard format every tool understands.

## Log modes by name

- **What:** the log level is a named mode (`none`, `error`, `warn`, `info`, `debug`). `none` turns logging off, which is useful in tests.
- **Why:** readable settings, and quiet tests.

## Named loggers per part

- **What:** a logger factory creates one logger per part, each tagged with the part's name.
- **Why:** every log line says where it came from.
- **Example:**

```
level=INFO msg=migrated name=Post
level=INFO msg=serving  name=App
```

## Front-end settings loaded at runtime

- **What:** the front end reads its settings (like the API address) from a small JSON file at startup, instead of building them in. Each environment provides its own file.
- **Why:** one front-end build works in every environment.

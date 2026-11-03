# Organization

How folders and files are arranged, so anyone can find anything.

## One shape, repeated everywhere

- **What:** every feature has the same set of files with the same names, in the same roles. A new feature is a copy of the shape, not a new invention.
- **Why:** learn one feature and you know them all. Reviews are faster because you know where to look. Anything missing or out of place stands out.
- **Example:**

```
post/                  tag/                  customer/
  model                  model                 model
  repository             repository            repository
  service                service               service
  controller             controller            controller
  module                 module                module
  test                   test                  test
```

## Group by feature, not by file type

- **What:** a folder holds everything about one feature (`post/`), not everything of one kind (`controllers/`).
- **Why:** a change to one feature stays in one folder. Deleting a feature means deleting a folder.

## Separate "my app" from "general tools"

- **What:** two areas. One for business features (posts, users, tags). One for reusable building blocks (database, cache, logging, web server, auth, config, metrics). Building blocks never know about business features.
- **Why:** the building blocks stay simple and reusable, and can be tested alone. Business code reads as business code.
- **Example:**

```
features/     post, tag, user         ← knows about the blog
building/     db, cache, log, web     ← knows nothing about the blog
```

## Nested features for sub-kinds

- **What:** when a feature has clear sub-kinds, give each its own folder inside the parent, with the same shape.
- **Why:** shared parts stay in the parent; each sub-kind stays small.
- **Example:** `user/` holds the shared parts, `user/admin/` and `user/customer/` hold the parts that differ.

## Small files with one topic each

- **What:** split a building block into small files by topic instead of one big file.
- **Why:** the file name tells you what's inside. Files stay short enough to read in one go.
- **Example:** an auth block split into `password`, `token`, `user`, `service`, `helper`, `error`, `mock`.

## A fixed place for each kind of supporting file

- **What:** errors in an `error` file, test doubles in a `mock` file, shared helpers in a `helper` file, inside each package.
- **Why:** you always know where named errors and fakes live.

## The same layering for infrastructure

- **What:** deployment config uses three layers: **components** (one folder per app or tool, same files in each), **environments** (dev, prod, shared: each picks components and changes only what differs), and **cluster** (which environments run, and in what order).
- **Why:** an app is described once. Environments hold only their differences. Adding an environment is small.
- **Example:**

```
components/  backend, frontend, cache, database
environments/ dev → uses components + dev addresses
              prod → uses components + prod addresses
cluster/     infra first, then shared, then dev and prod
```

## Separate entry points by command

- **What:** one program with clear commands for different jobs, like `sync` (prepare the database) and `serve` (run the server).
- **Why:** setup jobs and the long-running server are started, retried and scaled differently, but share the same code and settings.

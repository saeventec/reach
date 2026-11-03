# Design

How the parts of a program fit together.

## Clear layers with one job each

- **What:** each feature has the same layers, and each layer does one thing:
  - **model:** the data shapes (stored record, request, response).
  - **repository:** talks to storage. No business rules.
  - **service:** the business rules. No web or storage details.
  - **controller:** reads the request, calls the service, sends the response. No rules.
  - **module:** builds the feature's parts and plugs them into the app.
- **Why:** you know where every kind of code belongs. Rules can be tested without a web server or a database.

## Separate stored data from API data

- **What:** the stored record, the incoming request and the outgoing response are three different types. Converting between them is written out field by field, in two small functions (one each way).
- **Why:** the database can change without changing the API, and the other way round. Private fields (like a password hash) can't leak into a response by accident. The mapping is visible, with no magic.
- **Example:**

```
toResponse(record) → response:  id, title, createdAt   (no password hash)
fromRequest(request) → record:  only the fields a caller may set
```

## Small interfaces with one method

- **What:** define "plug-in points" as tiny interfaces, then let each feature implement the ones it needs.
- **Why:** the app can treat all features the same way. Adding a feature means implementing a method, not editing central code.
- **Example:**

```
Router       → route(server)       a feature adds its URLs
Interceptor  → intercept(server)   a feature adds middleware
Migrator     → migrate(database)   a feature creates its tables
Seeder       → seed(database)      a feature adds starter data

server.route(customer, tag, post)
database.migrate(customer, tag, post)
```

## Each feature registers itself

- **What:** a feature's module knows its own URLs, permissions, tables and starter data. The app only lists the features.
- **Why:** everything about a feature is in its folder. The startup code stays a short list.
- **Example:**

```
post.route(server):
  POST   /posts      admin only → create
  GET    /posts                 → list
  GET    /posts/:id             → find
```

## Wire everything by hand, in one place

- **What:** one startup function builds every part in order and passes each part what it needs (its "dependencies"). No framework does this automatically.
- **Why:** you can read the startup order top to bottom. Nothing is hidden. This place is called the **composition root**.
- **Example:**

```
config   = loadConfig()
log      = newLogger(config.logMode)
database = connectDatabase(config.databaseURL)
auth     = newAuth(config.secrets)
tags     = newTagModule(database, auth)
posts    = newPostModule(database, auth, tags)
```

## A clear app lifecycle

- **What:** startup has named phases: **setup** (build parts), **configure** (connect them), **mount** (add middleware and routes), **serve** (run). Setup jobs like **sync** reuse setup and skip serving.
- **Why:** each phase has one job, and commands can share the phases they need.

## Thin wrappers around libraries

- **What:** wrap each outside library (database, web framework, cache, logger) in a small type of our own, and let the rest of the code use the wrapper.
- **Why:** one place to configure the library, and the library can be swapped without touching features.

## Interfaces with more than one implementation

- **What:** when a part could reasonably change, give it an interface and at least two implementations, like a real cache and an in-memory cache.
- **Why:** proves the part is swappable, and the simple version is free to use in tests.
- **Example:**

```
Store: get(key), set(key, value)
  ├─ RedisStore     (real)
  └─ MemoryStore    (tests, local runs)
```

## Public interface, private implementation

- **What:** each service or repository has an exported interface and an unexported struct behind it. The constructor returns the interface.
- **Why:** callers depend on the promise, not the details. Tests can hand in a fake.

## Constructors for everything

- **What:** every type is created with a `New…` function that takes its dependencies.
- **Why:** one obvious way to build things. Required inputs can't be forgotten.

## Auth as a plug-in

- **What:** the auth building block checks tokens, but asks a **provider** to turn the token's subject into a user. The app supplies the provider. Route guards take a simple "who may enter" expression (like a role name) that the provider interprets.
- **Why:** the auth block stays general, and the business code decides what users and roles mean.

## Cache what is read often

- **What:** keep small, often-read data (like the current user's status) in a cache, and fall back to storage on a miss.
- **Why:** fewer slow lookups on every request.
- **Note:** always give cached data an expiry, or clear it when the data changes. Otherwise old data lives forever.

## Generic helpers for repeated shapes

- **What:** shared shapes like a page of results (`size`, `page`, `total`, `items`) are defined once and reused for every list.
- **Why:** every list endpoint behaves the same, with the same limits.

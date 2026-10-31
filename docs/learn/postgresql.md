# PostgreSQL

## What it is

PostgreSQL ("Postgres") is a database: a program that stores data safely and lets you ask questions about it in SQL.

- Data lives in **tables**, like spreadsheets: each **row** is one item, each **column** is one field.
- **SQL** is the language to create tables and to read and change data.
- One Postgres server can hold many **databases**. Each Reach service gets its own (the hello service uses `hello`), and no service reads another service's tables.

## Why we use it

Reliable, free, powerful and very popular. See [ADR 0006](../decisions/0006-use-postgresql.md).

## Tiny example

```sql
CREATE TABLE greetings (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    message text NOT NULL
);

INSERT INTO greetings (message) VALUES ('Hello from Postgres!');

SELECT message FROM greetings;
```

- `bigint`: a big whole number. `text`: a string of any length.
- `GENERATED ALWAYS AS IDENTITY`: Postgres picks the number for each new row (1, 2, 3...).
- `PRIMARY KEY`: the column that identifies each row. It must be unique.
- `NOT NULL`: the column must always have a value.

## In our setup

Postgres runs in Docker Compose (the `postgres` service), from the official image `postgres:18.6`:

- `POSTGRES_PASSWORD` sets the password of the main user, `postgres`. It comes from your `.env` file.
- The data is kept in the named volume `postgres-data`, so it survives restarts.
- The port is not published to your computer: only containers in Compose can reach it. That is safer, and you can still look inside:

```sh
docker compose exec postgres psql --username=postgres --dbname=hello
```

`psql` is Postgres's command-line client. Type SQL ending with `;`, `\dt` lists tables, `\q` quits.

## One user per service

`postgres` is the **superuser**: it may read, change and delete everything. A service that logs in as `postgres` could, after a bug or an attack, damage every other service's data. So each service gets its own **role** (Postgres's word for a user) and its own database, and the role owns only that database:

| Service | Role | Database |
| --- | --- | --- |
| hello | `hello` | `hello` |
| auth | `auth` | `auth` |

This is the **principle of least privilege**: give each part only the rights it needs. Each service's password is in `.env` (`HELLO_DATABASE_PASSWORD`, `AUTH_DATABASE_PASSWORD`).

The `postgres-setup` container creates the roles and databases. It runs [postgres/setup.sql](../../postgres/setup.sql) with `psql` on every `task dev`, before the migrations. ([ADR 0024](../decisions/0024-create-databases-with-a-setup-container.md) explains why.) Because it runs every time, each step must be **idempotent**: safe to repeat, with the same result as running it once. Postgres has no `CREATE DATABASE IF NOT EXISTS`, so the file uses a psql pattern:

```sql
SELECT 'CREATE DATABASE auth OWNER auth'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth') \gexec
```

The `SELECT` returns the text of a command, but only if the database is missing. `\gexec` is a psql command: "run each result as SQL". So the `CREATE` runs once, on the first start, and never again.

Two more rules in the file:

- **Owner:** the owner of a database may create tables in it and do everything with them. That is all a service needs.
- **`REVOKE CONNECT ... FROM PUBLIC`:** by default, every role may *connect* to every database (`PUBLIC` means "all roles"). We take that away, so `auth` can't even connect to `hello`. Try it:

```sh
docker compose exec -e PGPASSWORD=localauthpassword postgres psql --host=localhost --username=auth --dbname=hello
# → User does not have CONNECT privilege.
```

## IDs and constraints

The `users` table ([services/auth/migrations/00001_create_users_table.sql](../../services/auth/migrations/00001_create_users_table.sql)) shows two new ideas.

**UUIDv7 ids.** `greetings` numbers its rows 1, 2, 3. `users` uses a **UUID**: a 128-bit random-looking id like `019a3f2c-7b1e-7c4d-9a2b-3c4d5e6f7a8b`. Why:

- Anyone can create one without asking the database, so later the apps can create ids for messages themselves, even offline (see the brief, "Ready for offline sync").
- Ids don't reveal how many users there are, or let anyone guess the next one.
- **Version 7** puts the creation time first, so new ids sort after old ones. That keeps the database's index fast, like plain numbers. Ids made at about the same moment start with the same characters.

`DEFAULT uuidv7()` makes Postgres create one for each new row. The function is built into Postgres 18.

**Constraints** are rules the database checks on every insert and update. A row that breaks one is refused with an error:

```sql
username text NOT NULL UNIQUE,                                     -- must exist; no two the same
CONSTRAINT username_length CHECK (char_length(username) BETWEEN 3 AND 32)
```

The Go code will check the same rules first, to give the user a friendly message. The database check is the safety net: it holds even if the Go code has a bug.

The username rule uses a **regular expression** (a small pattern language for text). `~` means "matches the pattern":

```
^[a-z0-9]([._-]?[a-z0-9])*$
```

- `^` and `$`: the start and the end of the text, so the whole name must match.
- `[a-z0-9]`: one lowercase letter or digit. So the name starts with one.
- `( ... )*`: the part in brackets, repeated zero or more times.
- `[._-]?`: one `.`, `_` or `-`, or none (`?` means "optional").
- So after the first character, each step adds "maybe one separator, then a letter or digit". The name always ends with a letter or digit, and two separators can never touch.

| Username | Result |
| --- | --- |
| `sadra`, `sadra.samadi`, `s-2` | allowed |
| `Sadra` | refused: uppercase |
| `.sadra`, `sadra.` | refused: starts or ends with a separator |
| `sa..dra`, `sa._dra` | refused: two separators in a row |
| `ab` | refused: too short |

## Alternatives

- **MySQL / MariaDB:** also popular, slightly fewer features.
- **SQLite:** a database in one file, no server. Great for apps on a phone (we use it in phase 2), not for a shared server.
- **MongoDB:** stores JSON-like documents instead of tables, with less structure.

## Learn more

- https://www.postgresql.org/docs/18/tutorial.html
- https://hub.docker.com/_/postgres

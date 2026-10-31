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
- `POSTGRES_DB: hello` creates the `hello` database on the very first start.
- The data is kept in the named volume `postgres-data`, so it survives restarts.
- The port is not published to your computer: only containers in Compose can reach it. That is safer, and you can still look inside:

```sh
docker compose exec postgres psql --username=postgres --dbname=hello
```

`psql` is Postgres's command-line client. Type SQL ending with `;`, `\dt` lists tables, `\q` quits.

For now, the hello service connects as the main user `postgres`. Milestone 2 gives each service its own user with only the rights it needs.

## Alternatives

- **MySQL / MariaDB:** also popular, slightly fewer features.
- **SQLite:** a database in one file, no server. Great for apps on a phone (we use it in phase 2), not for a shared server.
- **MongoDB:** stores JSON-like documents instead of tables, with less structure.

## Learn more

- https://www.postgresql.org/docs/18/tutorial.html
- https://hub.docker.com/_/postgres

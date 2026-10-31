# Database migrations (goose)

## What it is

A migration is a numbered SQL file that changes the database one step: create a table, add a column, and so on. Together they are the history of the database structure, saved in git like code.

Analogy: building instructions for furniture. Step 1, step 2, step 3. A new database follows all the steps. An existing one only follows the steps it hasn't done yet.

**goose** is the tool that runs them. It records which migrations are done in its own table, `goose_db_version`, so each one runs exactly once.

## Why we use it

Every computer (yours, CI, later a server) gets exactly the same database structure, automatically. Nobody changes a database by hand. See [ADR 0016](../decisions/0016-run-migrations-in-a-separate-container.md).

## Tiny example

`services/hello/migrations/00001_create_greetings_table.sql`:

```sql
-- +goose Up
CREATE TABLE greetings (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    message text NOT NULL
);

-- +goose Down
DROP TABLE greetings;
```

- `-- +goose Up`: what to do to apply this step.
- `-- +goose Down`: how to undo it.
- The number at the start of the file name (`00001`) sets the order. The next one is `00002_...sql`.

Golden rule: **never edit a migration that has already run.** goose won't run it again. To change something, add a new migration.

## In our setup

The `hello-migrate` container runs `goose up` and exits:

```
OK   00001_create_greetings_table.sql (1.66ms)
goose: successfully migrated database to version: 1
```

On the next start, goose sees everything is done:

```
goose: no migrations to run. current version: 1
```

goose reads its settings from environment variables: `GOOSE_DRIVER` (the database type), `GOOSE_DBSTRING` (the database address) and `GOOSE_MIGRATION_DIR` (where the SQL files are).

## Alternatives

- **golang-migrate:** similar, with separate `up` and `down` files for each step.
- **Atlas:** you describe the wanted end state, and it works out the steps. Powerful, more magic.
- **ORM migrations (GORM AutoMigrate):** generated from code, so you don't see the SQL.

## Learn more

- https://github.com/pressly/goose
- https://pressly.github.io/goose/

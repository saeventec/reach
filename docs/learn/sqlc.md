# sqlc

## What it is

sqlc is a code generator. You write plain SQL queries in `.sql` files, and sqlc writes the Go functions that run them. It reads the migrations to learn the tables, so it knows each column's type and checks every query **before** the program runs.

An analogy: buf turns `.proto` files into Go code. sqlc does the same for SQL.

## Why we use it

- With pgx alone (like `hello`), every query needs hand-written Go: the SQL text, the arguments, and a `Scan` that copies each column into the right variable, in the right order. With many queries (users, sessions, devices), that is easy to get wrong.
- With sqlc, a typo in a column name, or a missing argument, is an error from `task generate`, not a crash at runtime.
- You still write and read real SQL. Nothing is hidden, as an ORM would hide it.
- The generated code uses pgx, which we already know ([ADR 0012](../decisions/0012-use-pgx-for-database-access.md)).

## Tiny example

The query, in [services/auth/queries/users.sql](../../services/auth/queries/users.sql):

```sql
-- name: CreateUser :one
INSERT INTO users (username, password_hash)
VALUES ($1, $2)
RETURNING id, username, created_at;
```

- `-- name: CreateUser` is the name of the Go function.
- `:one` means "returns exactly one row". Others: `:many` (a list of rows), `:exec` (no rows).
- `$1`, `$2` are placeholders. The values are sent separately from the SQL text, so they can never change the query (this prevents **SQL injection**, where an attacker types SQL into a form field).

`task generate` writes this into `gen/queries/users.sql.go` (shortened):

```go
type CreateUserParams struct {
	Username     string
	PasswordHash string
}

type CreateUserRow struct {
	ID        pgtype.UUID
	Username  string
	CreatedAt pgtype.Timestamptz
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error)
```

And this is how the service will call it (step 4):

```go
userQueries := queries.New(pool)
newUser, err := userQueries.CreateUser(ctx, queries.CreateUserParams{
	Username:     "sadra",
	PasswordHash: passwordHash,
})
```

Things to know:

- **`gen/queries/` is generated:** never edit it by hand. Change the `.sql` file and run `task generate`. Like the buf code, it is committed, and CI checks it is up to date ([ADR 0017](../decisions/0017-use-buf-with-local-generators.md)).
- **`pgtype.UUID`, `pgtype.Timestamptz`:** pgx's own Go types for Postgres's `uuid` and `timestamptz`. They are sqlc's default with pgx, so we need no extra library. If they get awkward, sqlc can map them to other types with "overrides" in `sqlc.yaml`.
- **`queries.New(pool)`** accepts anything with `Exec`, `Query` and `QueryRow` (the `DBTX` interface): a pool, one connection, or a transaction.
- The settings are in [services/auth/sqlc.yaml](../../services/auth/sqlc.yaml).

## Alternatives

- **pgx by hand:** no tool, full control. Fine for one query (`hello`), tiring for many.
- **GORM:** an ORM. You write Go, it writes the SQL. Quick to start, but the SQL is hidden.
- **Ent:** a typed ORM from Go code. Powerful, a lot to learn.
- **Jet or Squirrel:** build SQL with Go functions. The queries no longer look like SQL.

## Learn more

- [sqlc documentation](https://docs.sqlc.dev/en/latest/)
- [Using sqlc with pgx](https://docs.sqlc.dev/en/latest/guides/using-go-and-pgx.html)

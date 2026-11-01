# pgx: talking to Postgres from Go

## What it is

pgx is the standard Go library (a "driver") for PostgreSQL. It opens connections to the database, sends SQL, and turns the results into Go values.

- **Connection pool** (`pgxpool.Pool`): opening a database connection is slow, so the pool keeps a few open and lends them out, one per query. The service creates one pool at startup and shares it.
- **Connection string** (`DATABASE_URL`): the database's address, as a URL: `postgres://user:password@host:port/database?sslmode=disable`. `sslmode=disable` means no encryption. That's fine inside Compose on your own computer, but not on a real server.
- **Context** (`ctx`): every query gets one, so it can be cancelled, for example when the client goes away.

## Why we use it

It is fast, well maintained, and the most used Postgres driver for Go. See [ADR 0012](../decisions/0012-use-pgx-for-database-access.md).

## Tiny example

Connecting is in `packages/go/database/database.go` (shared by all services), the query in `services/hello/database.go`:

```go
pool, err := pgxpool.New(ctx, databaseURL) // create the pool (no connection yet)
err = pool.Ping(ctx)                       // connect now, to find mistakes at startup

row := pool.QueryRow(ctx, `SELECT message FROM greetings ORDER BY id DESC LIMIT 1`)

var message string
err = row.Scan(&message)       // copy the column into the variable
if errors.Is(err, pgx.ErrNoRows) {
	// the query found no row
}
```

- `QueryRow` is for queries that return at most one row. `Query` returns many rows.
- `Scan(&message)` gets the **address** of `message` (the `&`), so it can write into it.
- Values from users are never pasted into the SQL string. They go in as `$1, $2` arguments, which prevents **SQL injection** (an attack where input is run as SQL): `pool.QueryRow(ctx, "SELECT ... WHERE id = $1", id)`.

## Alternatives

- **`database/sql` + a driver:** Go's general database interface, one extra layer.
- **sqlc:** generates this kind of code from `.sql` files. The auth service uses it (see [sqlc.md](sqlc.md)); `hello` keeps plain pgx as an example.
- **GORM, Ent (ORMs):** you work with Go objects, and they write the SQL for you. Less to type, but you see less.

## Learn more

- https://pkg.go.dev/github.com/jackc/pgx/v5
- https://github.com/jackc/pgx/wiki/Getting-started-with-pgx

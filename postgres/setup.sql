-- Creates one database and one user per service. The postgres-setup
-- container in compose.yaml runs this file with psql before every start,
-- logged in as the "postgres" superuser.
--
-- Each service logs in as its own user, which owns only its own database.
-- So a bug in one service can't read or change another service's data.
-- Learn more: docs/learn/postgresql.md#one-user-per-service
--
-- The file runs on every start, so every step must be safe to repeat:
-- "create it if it is missing", never "create it" alone.

-- Stop at the first error, so a problem is never hidden.
\set ON_ERROR_STOP on

-- Read the passwords from the environment variables that compose.yaml sets.
\getenv hello_password HELLO_DATABASE_PASSWORD
\getenv auth_password AUTH_DATABASE_PASSWORD
\getenv chat_password CHAT_DATABASE_PASSWORD

-- Postgres has no "CREATE ROLE IF NOT EXISTS" or "CREATE DATABASE IF NOT
-- EXISTS". So each SELECT below builds the CREATE command only when the
-- thing is missing, and \gexec runs the command the SELECT produced.
-- format() with %L puts the password in quotes safely.

-- The hello service --------------------------------------------------------

SELECT 'CREATE ROLE hello LOGIN'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'hello') \gexec

-- Always set the password, so a changed .env takes effect on the next start.
SELECT format('ALTER ROLE hello PASSWORD %L', :'hello_password') \gexec

SELECT 'CREATE DATABASE hello OWNER hello'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'hello') \gexec

-- The owner may do everything in a database. Before milestone 2, the hello
-- database was owned by "postgres", so we set the owner explicitly.
ALTER DATABASE hello OWNER TO hello;

-- By default, every user may connect to every database. Only allow the
-- owner (and superusers).
REVOKE CONNECT ON DATABASE hello FROM PUBLIC;

-- The auth service ---------------------------------------------------------

SELECT 'CREATE ROLE auth LOGIN'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'auth') \gexec

SELECT format('ALTER ROLE auth PASSWORD %L', :'auth_password') \gexec

SELECT 'CREATE DATABASE auth OWNER auth'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth') \gexec

REVOKE CONNECT ON DATABASE auth FROM PUBLIC;

-- The chat service ---------------------------------------------------------

SELECT 'CREATE ROLE chat LOGIN'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'chat') \gexec

SELECT format('ALTER ROLE chat PASSWORD %L', :'chat_password') \gexec

SELECT 'CREATE DATABASE chat OWNER chat'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'chat') \gexec

REVOKE CONNECT ON DATABASE chat FROM PUBLIC;

-- Tables created before milestone 2 -----------------------------------------

-- Before milestone 2, the hello migrations ran as "postgres", so these two
-- tables belong to "postgres", and the hello user could not use them.
-- Give them to hello. On a new database they don't exist yet, so nothing
-- happens. They are created later by hello-migrate, already owned by hello.
\connect hello
ALTER TABLE IF EXISTS greetings OWNER TO hello;
ALTER TABLE IF EXISTS goose_db_version OWNER TO hello;

# Protobuf, buf and code generation

## What it is

**Protocol Buffers (Protobuf)** is a way to describe the data your programs send to each other, in `.proto` files. You write the description once. A tool then **generates code** from it for each language: Go for the services, TypeScript for the web app.

Analogy: a contract both sides sign. The web app and the service can't disagree about what a `GetGreetingResponse` looks like, because both use code made from the same file.

- A **message** is a data type with numbered fields.
- A **service** is a group of **RPCs** (remote procedure calls): functions you call over the network as if they were local.

**buf** is the tool that checks (lints) `.proto` files and runs the code generators. The generators are small programs called plugins:

- `protoc-gen-go` generates Go types for the messages.
- `protoc-gen-connect-go` generates Go code for the service: an interface to implement, and a client.

## Why we use it

One source of truth for the API. See [ADR 0005](../decisions/0005-protobuf-and-connect-for-the-api.md) and [ADR 0017](../decisions/0017-use-buf-with-local-generators.md).

## Tiny example

`proto/hello/v1/greeting.proto`:

```proto
syntax = "proto3";
package hello.v1;               // the API's name and version

service GreetingService {
  rpc GetGreeting(GetGreetingRequest) returns (GetGreetingResponse) {}
}

message GetGreetingRequest {}

message GetGreetingResponse {
  string message = 1;           // field number 1
}
```

`task generate` runs `buf lint`, then `buf generate`, and writes:

```
packages/go/gen/hello/v1/greeting.pb.go                       # the message types
packages/go/gen/hello/v1/hellov1connect/greeting.connect.go   # the service interface and client
```

Never edit generated files by hand (they say `DO NOT EDIT`). Change the `.proto` file and run `task generate` again.

## Rules that keep old clients working

- **Field numbers are forever.** The number (not the name) is what travels over the network. Never change or reuse one.
- **Adding** a field is safe. **Removing or renaming** one can break clients.
- A change that must break things goes into a new version: `hello.v2`.

buf checks these rules for you. This compares your files with the last commit on `main`:

```sh
buf breaking --against '.git#branch=main'
```

For example, moving `message` from field 1 to field 2, and giving number 1 to a new `id` field, gives:

```
Field "1" on message "GetGreetingResponse" changed name from "message" to "id".
Field "1" with name "id" on message "GetGreetingResponse" changed type from "string" to "int64".
```

The fix: keep `message = 1`, and give the new field the next free number, `id = 2`. The order of the fields in the file doesn't matter, only the numbers. CI runs this check on every change (see [ci.md](ci.md)).

## Lists, dates and default values

From [proto/auth/v1/session.proto](../../proto/auth/v1/session.proto):

```protobuf
import "google/protobuf/timestamp.proto";

message ListDevicesResponse {
  repeated Device devices = 1;                  // a list of zero or more devices
}

message Device {
  google.protobuf.Timestamp last_used_at = 4;   // a moment in time
  bool is_current_device = 5;
}
```

- **`repeated`** makes a field a list. In Go it is a slice (`[]*Device`), in TypeScript an array.
- **Well-known types** are standard messages that come with Protobuf, like `google.protobuf.Timestamp`. Each language maps them to its own type: `timestamppb.Timestamp` in Go (`timestamppb.New(time)` makes one), and `Timestamp` from `@bufbuild/protobuf/wkt` in TypeScript. In JSON, a Timestamp is text in the RFC 3339 format, always in UTC: `"2026-10-31T09:00:00Z"`.
- **Default values are left out of JSON.** A `bool` that is `false`, a number that is `0` or an empty text is not sent at all. So a device that isn't the current one has no `isCurrentDevice` in the JSON. The generated code fills in the default again, so your code always sees `false`.

## Our config files

- `buf.yaml`: where the `.proto` files are (`proto/`) and which lint rules to use.
- `buf.gen.yaml`: which generators to run and where their output goes. "Managed mode" sets the Go import path there, so the `.proto` files contain no Go-specific settings.

### All Go code in one shared place

All generated Go code goes into the shared module `packages/go`, in `packages/go/gen/`. Every service imports what it needs from there:

```go
import authv1 "github.com/saeventec/reach/packages/go/gen/auth/v1"
```

- **Why one place:** services use each other's APIs. `chat` calls `auth`'s internal API, and `realtime` reads `chat`'s events. Generated in one place, there is exactly one copy of each API's Go code. ([ADR 0017](../decisions/0017-use-buf-with-local-generators.md) planned this move for when a second service needed it.)
- **`go_package_prefix`** in managed mode gives every file its Go import path: the prefix plus the file's folder in `proto/`.
- A service's program only contains the packages it really imports, so `hello` doesn't grow because `chat`'s API exists.
- sqlc's code is different: it belongs to one service only, so it stays in that service (`services/auth/gen/queries`).
- Until milestone 3, each service had its own copy, limited with buf's `types` option. That stopped working once one API was needed by two Go modules: a `.proto` file can have only one Go import path.

## TypeScript

A third generator, `protoc-gen-es`, writes TypeScript for the web app: `apps/web/src/gen/hello/v1/greeting_pb.ts` (and `auth/v1/account_pb.ts`). It contains the message types (`GetGreetingResponse`) and the service (`GreetingService`). `int64` fields become `bigint`, JavaScript's type for big whole numbers.

This generator is an npm package of the web app, not a mise tool. That way its version always matches the `@bufbuild/protobuf` library the generated code needs. So `task setup` (which runs `npm ci`) must run before `task generate`.

## Alternatives

- **REST + OpenAPI:** describe HTTP endpoints in YAML. Very common, but the contract drifts from the code more easily.
- **JSON Schema:** describes JSON data only, not the API calls.
- **GraphQL:** a query language with its own schema, more complex on the server.
- **protoc (instead of buf):** the original tool, with long commands and no lint.

## Learn more

- https://protobuf.dev/programming-guides/proto3/
- https://buf.build/docs/

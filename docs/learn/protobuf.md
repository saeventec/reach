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
services/hello/gen/hello/v1/greeting.pb.go                       # the message types
services/hello/gen/hello/v1/hellov1connect/greeting.connect.go   # the service interface and client
```

Never edit generated files by hand (they say `DO NOT EDIT`). Change the `.proto` file and run `task generate` again.

## Rules that keep old clients working

- **Field numbers are forever.** The number (not the name) is what travels over the network. Never change or reuse one.
- **Adding** a field is safe. **Removing or renaming** one can break clients.
- A change that must break things goes into a new version: `hello.v2`.

buf can check these rules automatically (`buf breaking`). We'll use that in CI later.

## Our config files

- `buf.yaml`: where the `.proto` files are (`proto/`) and which lint rules to use.
- `buf.gen.yaml`: which generators to run and where their output goes. "Managed mode" sets the Go import path there, so the `.proto` files contain no Go-specific settings.

## Alternatives

- **REST + OpenAPI:** describe HTTP endpoints in YAML. Very common, but the contract drifts from the code more easily.
- **JSON Schema:** describes JSON data only, not the API calls.
- **GraphQL:** a query language with its own schema, more complex on the server.
- **protoc (instead of buf):** the original tool, with long commands and no lint.

## Learn more

- https://protobuf.dev/programming-guides/proto3/
- https://buf.build/docs/

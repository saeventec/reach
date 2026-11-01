# APIs and Connect RPC

## What it is

An **API** (Application Programming Interface) is the list of things a program lets other programs ask it to do. Our hello service's API has one call: `GetGreeting`.

**RPC** (Remote Procedure Call) means calling a function in another program, over the network, as if it were a normal function call.

**Connect** is a library (by the makers of buf) that turns the services in our `.proto` files into real HTTP endpoints. Each call is a `POST` to the path `/<package>.<Service>/<Method>`, with the request as JSON (easy to read) or binary Protobuf (smaller and faster).

## Why we use it

The web app calls the Go services with type-safe generated code, and the same calls can also be made with plain `curl`. See [ADR 0005](../decisions/0005-protobuf-and-connect-for-the-api.md). We use connect-go **v2** (released in 2026). Older tutorials show v1, which looks different: for example, v1 wraps every message in `connect.Request[...]`.

## Tiny example

The server side, in `services/hello`:

```go
// 1. Implement the generated interface: one method per RPC.
func (server *greetingServer) GetGreeting(ctx context.Context, request *hellov1.GetGreetingRequest) (*hellov1.GetGreetingResponse, error) {
	response := &hellov1.GetGreetingResponse{Message: "Hello!"}
	return response, nil
}

// 2. Register it on a Connect server and mount that on our router.
connectServer := connect.NewServer()
hellov1connect.RegisterGreetingServiceHandler(connectServer, greetingService)
connecthttp.Mount(router, connectServer)
```

Calling it with curl, through the gateway:

```sh
curl --header "Content-Type: application/json" --data '{}' \
  http://localhost:8080/hello.v1.GreetingService/GetGreeting
```

```json
{"message":"Hello, World!", "id":"2"}
```

The `id` is an `int64`, and Protobuf sends 64-bit numbers as **strings** in JSON. JavaScript numbers can't hold such big numbers exactly, so a string keeps them safe. The generated TypeScript code turns it back into a number type (`bigint`).

## Errors

A Connect error has a **code** (what kind of problem) and a **message**:

```go
return nil, connect.NewError(connect.CodeNotFound, "there is no greeting yet")
```

```json
{"code":"not_found","message":"there is no greeting yet"}
```

Common codes: `not_found`, `invalid_argument` (bad input), `already_exists` (for example, a taken username), `unauthenticated` (not logged in), `internal` (a bug or outage on our side). For `internal` errors we log the details and send only a general message, because details can reveal how the system works inside.

### A limit on message size

The auth service reads at most 16 KiB per message (`connecthttp.WithReadMaxBytes` in [services/auth/main.go](../../services/auth/main.go)). Without a limit, a client could send a message of many megabytes, and the service would read it all into memory before our own rules (like "at most 128 characters") ever run. A bigger message gets the code `resource_exhausted`. The real messages are far smaller: a sign-up is under 1 KiB.

## Internal APIs

Some APIs are for other services only, never for browsers. `auth`'s `UserDirectoryService` ([proto/authinternal/v1/user_directory.proto](../../proto/authinternal/v1/user_directory.proto)) turns usernames into user ids and back, for `chat`. Each service owns its data, so `chat` stores only user ids and asks `auth` for names.

- **Who can call it:** the gateway forwards only `/auth.v1.*` to `auth`. A browser asking for `/authinternal.v1.…` gets a 404 from the web app's server. Inside Compose, every container can call `http://auth:8080/authinternal.v1.…`.
- **Why a separate package (`authinternal.v1`, not `auth.v1`):** the path decides who can reach it, so internal calls can never end up in the public routes by accident.
- **Not `auth.internal.v1`:** that would put the Go code in a folder named `internal`. Go has a rule for such folders: only code next to them may import them. `chat` couldn't use it.
- **The trust model:** like `X-User-Id` ([caddy.md](caddy.md#checking-the-session-at-the-gateway)), it relies on a closed network. Real companies add service-to-service authentication (mutual TLS, or signed service tokens). A good topic once there is a real server or Kubernetes.

Example: what `chat` sends, and what it gets back:

```
POST http://auth:8080/authinternal.v1.UserDirectoryService/FindUsersByUsername
{"usernames": [" Bob ", "nobody"]}

{"users": [{"id": "01a1…", "username": "bob"}]}
```

A username with no account is just missing from the answer, so the caller can say "nobody has the username nobody".

## The TypeScript client

The web app calls the same API with generated TypeScript code (`apps/web/src/gen/`) and two libraries: `@connectrpc/connect` and `@connectrpc/connect-web` (which sends the calls with the browser's `fetch`).

From `apps/web/src/greeting/greetingClient.ts`:

```ts
const transport = createConnectTransport({ baseUrl: window.location.origin });
const client = createClient(GreetingService, transport);

const response = await client.getGreeting({});   // typed: response.message is a string, response.id a bigint
```

`baseUrl: window.location.origin` means "the same address as the page". The browser only talks to one address, the gateway (http://localhost:8080), and the gateway forwards API calls to the hello service. See [caddy.md](caddy.md).

## Alternatives

- **REST + JSON by hand:** you design URLs and JSON yourself, and must keep both sides in sync by hand.
- **gRPC:** the same idea with Protobuf, but browsers can't call it directly.
- **GraphQL:** clients pick exactly the fields they want; more complex servers.

## Learn more

- https://connectrpc.com/docs/go/getting-started
- https://connectrpc.com/docs/protocol

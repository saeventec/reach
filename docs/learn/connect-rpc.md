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

Calling it with curl:

```sh
curl --header "Content-Type: application/json" --data '{}' \
  http://localhost:8080/hello.v1.GreetingService/GetGreeting
```

```json
{"message":"Hello, World!"}
```

## Errors

A Connect error has a **code** (what kind of problem) and a **message**:

```go
return nil, connect.NewError(connect.CodeNotFound, "there is no greeting yet")
```

```json
{"code":"not_found","message":"there is no greeting yet"}
```

Common codes: `not_found`, `invalid_argument` (bad input), `unauthenticated` (not logged in), `internal` (a bug or outage on our side). For `internal` errors we log the details and send only a general message, because details can reveal how the system works inside.

## Alternatives

- **REST + JSON by hand:** you design URLs and JSON yourself, and must keep both sides in sync by hand.
- **gRPC:** the same idea with Protobuf, but browsers can't call it directly.
- **GraphQL:** clients pick exactly the fields they want; more complex servers.

## Learn more

- https://connectrpc.com/docs/go/getting-started
- https://connectrpc.com/docs/protocol

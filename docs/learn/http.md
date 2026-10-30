# HTTP and web servers

## What it is

HTTP is the language browsers and servers use to talk. The client sends a **request**, and the server sends back a **response**.

- A **request** has a **method** (what to do: `GET` reads, `POST` sends data), a **path** (which thing: `/healthz`), headers (extra information) and sometimes a body.
- A **response** has a **status code** (how it went), headers, and a body (the content).

Common status codes: `200 OK`, `404 Not Found`, `405 Method Not Allowed`, `500 Internal Server Error`.

A server **listens** on a **port**: a number that tells the computer which program gets the traffic. `localhost:8080` means "port 8080 on this computer".

## How the hello service does it

Go's built-in `net/http` package has everything we need:

- **Handler:** a function that receives the request and writes the response. Its two inputs are the `http.ResponseWriter` (where you write the answer) and the `*http.Request` (what the client asked).
- **Router** (`http.ServeMux`): connects patterns like `"GET /healthz"` to handlers. Because the pattern names `GET`, a `POST` to `/healthz` gets `405 Method Not Allowed` automatically.
- **Server** (`http.Server`): listens on the port and passes each request to the router. `ReadHeaderTimeout` closes connections that are too slow to send their request headers, so slow clients can't keep the server busy forever.

## Health checks

A health check is a small endpoint that answers "I'm alive". Our services use `GET /healthz`, which answers `200 OK` with the body `ok`. Docker Compose, and later Kubernetes, call it regularly. If it stops answering, they restart the service.

The `z` at the end is an old Google convention, so the name doesn't clash with a real page called `/health`.

## Tiny example

```sh
curl -i http://localhost:8080/healthz
```

```
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 2

ok
```

`curl` is a command-line tool that sends HTTP requests. `-i` also shows the status and headers.

## Alternatives

- **chi:** a small router library. We don't need it, because Go's router supports methods and path wildcards since Go 1.22.
- **Gin or Echo:** bigger web frameworks with more features and more to learn.

## Learn more

- https://developer.mozilla.org/en-US/docs/Web/HTTP/Overview
- https://pkg.go.dev/net/http

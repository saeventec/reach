# Caddy and the gateway

## What it is

A **gateway** (or **reverse proxy**) is a server that stands in front of other servers. The browser sends every request to the gateway, and the gateway forwards it to the right service, then passes the answer back.

Analogy: a building's front desk. Visitors only know the front desk, and the desk sends each one to the right office.

**Caddy** is the web server we use as our gateway.

```
Browser → Caddy (localhost:8080) ─┬─ /hello.v1.* → hello:8080   (API calls)
                                  └─ everything else → web:5173  (the web app)
```

## Why we use it

- The browser only knows one address. No CORS problems (the browser's rule that blocks calls between different addresses), and the services stay private inside Compose.
- One place for things every request needs: HTTPS, compression, and later rate limiting.
- Caddy's config is tiny, and it can get HTTPS certificates automatically on a real server. See [ADR 0008](../decisions/0008-use-caddy-as-gateway.md).

## Our config

`gateway/Caddyfile`:

```
:80 {
	handle /hello.v1.* {
		reverse_proxy hello:8080
	}

	handle /auth.v1.* {
		reverse_proxy auth:8080
	}

	handle {
		reverse_proxy web:5173
	}
}
```

- `:80 { ... }`: a **site**. Any hostname, on port 80, plain HTTP. With no hostname, Caddy doesn't turn on automatic HTTPS. That's right on your own computer; HTTPS comes in milestone 11.
- `handle /hello.v1.* { ... }`: requests whose path starts with `/hello.v1.` (the `*` at the end means "starts with"). All Connect calls of the hello API look like `/hello.v1.GreetingService/GetGreeting`.
- `handle /auth.v1.* { ... }`: the same for the auth service, like `/auth.v1.AccountService/SignUp`.
- `handle { ... }` with no path: the **fallback**. Only one `handle` runs per request, and this one is used when no other matches.
- `reverse_proxy hello:8080`: forward the request to the hello service, found by its Compose name. WebSockets (used by Vite's live page updates) are forwarded too, automatically.
- Caddy indents its config with **tabs**. `.editorconfig` says so for the Caddyfile.

In `compose.yaml`, the gateway uses the official `caddy` image and shares the `gateway/` folder with it, read-only. It is the only service with `ports`.

## Checking the session at the gateway

From milestone 3, some services (`chat`, `realtime`) only serve logged-in users. Instead of each one reading the session cookie, the gateway asks `auth` first ([ADR 0029](../decisions/0029-authenticate-at-the-gateway.md)):

```
Browser ── POST /chat.v1.… (cookie) ──▶ Caddy
                                         │ 1. GET /internal/check-session (same cookie) ──▶ auth
                                         │    ◀── 200 + X-User-Id: 01a1…   (or 401: stop here)
                                         │ 2. the request, with X-User-Id ──▶ chat
```

Big companies do the same in their API gateways. Envoy calls it "external authorization".

The snippet `(logged_in)` in `gateway/Caddyfile` does it. A route uses it with `import logged_in`:

- **`forward_auth auth:8080`** sends a GET copy of the request, with its headers and cookie, to `uri /internal/check-session`. On a 2xx answer, **`copy_headers X-User-Id`** copies that header from auth's answer onto the real request. Any other answer is sent back to the browser as it is: auth answers `401` with a Connect error body, so the web app sees a normal "unauthenticated" error.
- **`request_header -X-User-Id`** removes the header if the browser sent one. The services trust `X-User-Id`, so a browser must never be able to set it.
- **`route { … }`** keeps these two in the written order. Caddy normally runs directives in [its own fixed order](https://caddyserver.com/docs/caddyfile/directives#directive-order), whatever order you write them in, and that order puts `forward_auth` *before* `request_header`. Without `route`, Caddy added auth's header first and then removed it again, so every logged-in request reached the service with no user. We found it by testing; the docs only show it in that order list.

**WebSockets too:** `/realtime` uses the same snippet. Opening a WebSocket starts as a normal HTTP request with the cookie, so the session check works the same way, and `reverse_proxy` passes the WebSocket through on its own. See [websocket.md](websocket.md).

**Why only the gateway can reach `/internal/check-session`:** Caddy forwards only `/auth.v1.*` from the browser to `auth`. A browser asking for `/internal/check-session` gets the web app's page.

**Tested automatically:** [apps/web/e2e/gateway.e2e.ts](../../apps/web/e2e/gateway.e2e.ts) sends forged headers through the real gateway, with and without a session (see [playwright.md](playwright.md#our-tests)).

**The danger of a trusted header:** if a service ever got a published port, anyone could call it directly with a made-up `X-User-Id`. That's why `chat` and `realtime` have no `ports` in `compose.yaml`.

## Useful commands

```sh
docker compose restart gateway   # load an edited Caddyfile
docker compose logs gateway      # see Caddy's logs
docker run --rm -v "$PWD/gateway:/etc/caddy:ro" caddy:2.11.7-alpine caddy validate --config /etc/caddy/Caddyfile   # check the config
```

Caddy's Docker docs ask for volumes for `/data` and `/config`. They hold HTTPS certificates and keys, which we don't use yet. We add them together with HTTPS.

## Alternatives

- **Traefik:** configures itself from Docker labels; popular with containers, more concepts.
- **Nginx:** the long-time standard, very fast, longer config, HTTPS set up by hand.
- **Envoy:** very powerful, made for large systems, complex config.

## Learn more

- https://caddyserver.com/docs/caddyfile/concepts
- https://caddyserver.com/docs/caddyfile/directives/reverse_proxy
- https://caddyserver.com/docs/caddyfile/directives/forward_auth
- https://caddyserver.com/docs/caddyfile/concepts#snippets

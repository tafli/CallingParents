# ADR-006: CORS Handling — Go Reverse Proxy

## Status

Accepted

## Date

2026-02-20

## Context

The PWA runs in a browser on the Android phone and needs to make HTTP requests to the ProPresenter API. However:

1. The PWA is served from `http://<server>:8080`.
2. The ProPresenter API runs on `http://<propresenter>:50001`.
3. These are different origins (different host and/or port), so the browser enforces **CORS (Cross-Origin Resource Sharing)** restrictions.
4. ProPresenter's API does not guarantee `Access-Control-Allow-Origin` headers, so direct browser-to-ProPresenter requests will likely be blocked.

### Options Considered

1. **Direct browser requests + CORS browser extension** — unreliable, requires manual setup on each device, security risk.
2. **ProPresenter CORS configuration** — not available; ProPresenter does not expose CORS settings.
3. **Go server as reverse proxy** — the server proxies `/api/*` requests to ProPresenter, making all requests same-origin from the browser's perspective.

## Decision

The Go server acts as a **transparent reverse proxy** for all ProPresenter API requests.

### URL Mapping

| Browser Request | Proxied To |
|-----------------|------------|
| `GET /api/v1/messages` | `GET http://<PP_HOST>:<PP_PORT>/v1/messages` |
| `POST /api/v1/message/{id}/trigger` | `POST http://<PP_HOST>:<PP_PORT>/v1/message/{id}/trigger` |
| `GET /api/v1/message/{id}/clear` | `GET http://<PP_HOST>:<PP_PORT>/v1/message/{id}/clear` |
| `GET /api/v1/clear/layer/messages` | `GET http://<PP_HOST>:<PP_PORT>/v1/clear/layer/messages` |

The `/api/` prefix is stripped before forwarding to ProPresenter. All other paths serve static PWA files.

### Implementation

Using Go's `net/http/httputil.ReverseProxy` with a custom `Director` function that:

1. Strips the `/api/` prefix from the request URL path.
2. Sets the target scheme, host, and port to the ProPresenter address.
3. Forwards all headers, query parameters, and request body unchanged.

## Consequences

- **No CORS issues**: all requests from the browser go to the same origin (the Go server), so no cross-origin restrictions apply.
- **Transparency**: the proxy forwards requests and responses without modification (except the URL rewrite), so any ProPresenter API endpoint is accessible.
- **Single point of access**: the Android phone only needs to know the Go server's address, not the ProPresenter machine's address directly.
- **Latency**: adds a small hop through the Go proxy. On a local network, this is negligible (sub-millisecond).
- **Coupling**: if ProPresenter's API changes, only the PWA frontend needs updating; the proxy is generic and forwards any path under `/api/`.

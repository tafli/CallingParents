# ADR-006: CORS Handling — Go Backend Proxy

## Status

Accepted (updated)

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

The Go server acts as a **backend proxy** for ProPresenter API requests using purpose-built endpoints.

### URL Mapping

| Browser Request | Server Action |
|-----------------|---------------|
| `POST /message/send` (`{"name":"Paul"}`) | `POST http://<PP_HOST>:<PP_PORT>/v1/message/<MESSAGE_NAME>/trigger` |
| `POST /message/clear` | `GET http://<PP_HOST>:<PP_PORT>/v1/message/<MESSAGE_NAME>/clear` |
| `GET /message/test` | `GET http://<PP_HOST>:<PP_PORT>/v1/messages` |
| `GET /message/config` | Returns server config (e.g., `autoClearSeconds`) as JSON — no ProPresenter call |
| `GET /version` | Returns build version info as JSON — no ProPresenter call, no auth required |

The PWA sends only the child's name; the server resolves the ProPresenter message template name from the `MESSAGE_NAME` environment variable. All other paths serve static PWA files.

**Note**: `POST /message/clear` exists for the server-side auto-clear feature. The PWA does not expose a manual clear button to the user (see ADR-003).

### Implementation

Using purpose-built HTTP handlers in `internal/message/` that:

1. Accept simplified requests from the PWA (just the child's name for send, nothing for clear/test).
2. Construct the appropriate ProPresenter API request using the configured message template name.
3. Forward the request to ProPresenter and return the result.

## Consequences

- **No CORS issues**: all requests from the browser go to the same origin (the Go server), so no cross-origin restrictions apply.
- **Encapsulation**: the PWA does not need to know the ProPresenter message template name or API structure. Only the child's name is sent.
- **Single point of access**: the Android phone only needs to know the Go server's address, not the ProPresenter machine's address directly.
- **Latency**: adds a small hop through the Go server. On a local network, this is negligible (sub-millisecond).
- **Tight contract**: the server exposes only the three operations needed (send, clear, test), not the full ProPresenter API.

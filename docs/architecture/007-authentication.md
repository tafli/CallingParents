# ADR-007: Authentication — Bearer Token via QR Code

## Status

Accepted

## Date

2026-02-20

## Context

The application runs on a local network in a church setting. Without authentication, anyone on the same network could send messages to the ProPresenter audience screens. A lightweight auth mechanism is needed — one that doesn't require user accounts or a login screen.

### Options Considered

1. **No authentication** — relies on network isolation only. Too easy to abuse.
2. **Username/password** — requires a login screen, password management. Overkill for this setting.
3. **Bearer token via QR code** — token embedded in the URL hash fragment; phone extracts and stores it on first load; all API requests include it in the `Authorization` header.

## Decision

Use a **random bearer token** embedded in the QR code URL. The QR code is the "key" — only devices that scan it can interact with the backend.

### How It Works

1. On startup, the server generates a random 32-byte hex token (or reads `AUTH_TOKEN` from environment for a stable token across restarts).
2. The QR code URL includes the token in the hash fragment: `http://<ip>:<port>#token=<hex>`.
3. The PWA extracts the token from `window.location.hash`, stores it in `localStorage`, and removes it from the URL bar.
4. Every `fetch()` call to protected endpoints includes the `Authorization: Bearer <token>` header.
5. The Go server validates the token via middleware on all protected paths.

### Protected vs. Unprotected

| Path | Protected | Reason |
|------|-----------|--------|
| `/api/*` | Yes | ProPresenter proxy — must not be publicly accessible |
| `/children` | Yes | Children data — read and write |
| `/` (static files) | No | PWA shell must load so the JS can extract the token |

### Token Comparison

Uses `crypto/subtle.ConstantTimeCompare` to prevent timing attacks.

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_TOKEN` | (random) | Bearer token for API auth. If not set, a random token is generated on each startup. |

## Consequences

- **QR code = access key**: scanning the QR code grants full access. Keep it visible only to authorized workers.
- **No login screen**: zero friction for church workers — scan and go.
- **Random token by default**: each restart generates a new token, requiring a new QR code scan. Set `AUTH_TOKEN` for persistence.
- **Hash fragment security**: the token in `#token=...` is never sent to the server in HTTP requests (only via `Authorization` header), and is not logged by proxies.
- **Static files unprotected**: the PWA HTML/JS/CSS loads without auth. This is necessary so the JavaScript can parse the token from the URL hash. The static files contain no sensitive data.

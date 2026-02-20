# AGENTS.md

This file provides instructions for AI coding agents working on the calling-parents codebase. For human-oriented documentation, see `README.md`.

## Architecture decisions

All architecture decisions are recorded as ADRs in `docs/architecture/`. **Always read the relevant ADRs before making changes.** Any change that affects architecture, API design, authentication, configuration, deployment, or observability must conform to the existing ADRs. If a change contradicts an existing ADR, update the ADR first before implementing.

| ADR | Title                        |
|-----|------------------------------|
| [001](docs/architecture/001-project-overview.md) | Project Overview |
| [002](docs/architecture/002-application-type.md) | Application Type — Progressive Web App |
| [003](docs/architecture/003-propresenter-integration.md) | ProPresenter Integration |
| [004](docs/architecture/004-data-management.md) | Data Management — Browser localStorage |
| [005](docs/architecture/005-deployment.md) | Deployment — Go Binary with Embedded PWA |
| [006](docs/architecture/006-cors-api-proxy.md) | CORS Handling — Go Reverse Proxy |

---

## Project overview

A PWA served by a Go binary that lets church child care workers send "Eltern von {Name}" messages to ProPresenter audience screens. The Go server embeds the static PWA files and reverse-proxies `/api/*` requests to ProPresenter's HTTP API to avoid CORS issues. Data (child names, settings) is stored in browser localStorage — the server is stateless.

### Key components

- **`cmd/server/main.go`** — HTTP server entry point; embeds `web/` via `embed.FS`, mounts the reverse proxy on `/api/` and the file server on `/`.
- **`internal/config/`** — Reads `PROPRESENTER_HOST`, `PROPRESENTER_PORT`, `LISTEN_ADDR` from environment variables.
- **`internal/proxy/`** — `httputil.ReverseProxy` that strips `/api` prefix and forwards to ProPresenter.
- **`cmd/server/web/`** — PWA frontend: `index.html`, `app.js`, `style.css`, `manifest.json`, `sw.js`, icons.

### ProPresenter API endpoints used

- `GET /v1/messages` — list messages (connection test)
- `POST /v1/message/{id}/trigger` — show message with `[{"name":"Name","text":{"text":"..."}}]`
- `GET /v1/message/{id}/clear` — hide message
- `GET /v1/clear/layer/messages` — clear all messages


---

## General Guidance

- Always set required env vars first
- Only run tests, when code has changed. Documentation changes do not need to be tested.
- Always run `gofmt -w .` before building or testing.
- Always run `go test ./...` before committing. All tests must pass.
- Always run `go vet ./...` after making changes. Fix any warnings before committing.

---

## Build and test commands

- Always set required env vars first
- Only run tests, when code has changed. Documentation changes do not need to be tested.
- Always run `gofmt -w .` before building or testing.
- Always run `go test ./...` before committing. All tests must pass.
- Always run `go vet ./...` after making changes. Fix any warnings before committing.

---

## Code style guidelines

- Follow standard Go conventions. Always format with `gofmt`.
- Keep dependencies minimal — prefer the Go standard library. Do not add new dependencies without strong justification.
- Always pass `context.Context` as the first parameter for functions that do I/O.
- Wrap errors with `fmt.Errorf("doing thing: %w", err)` — always include context.
- Use `log/slog` for all logging. Never use `fmt.Println` or `log.Println`.
- Use structured key-value pairs: `logger.Info("event", "key", value)`.
- Never expose internal error details to clients.
- `cmd/server/main.go` is the entrypoint only — never put business logic here.

---

## Testing instructions

- Test business logic with unit tests, HTTP layers with `httptest`.
- Tests must never require a real Shelly device — use interfaces or mocks.
- Place test files alongside the code they test: `config_test.go` next to `config.go`.
- Use table-driven tests for functions with multiple input/output cases.
- Use `t.Helper()` in test helpers and `t.Parallel()` where tests do not share mutable state.
- Never depend on host environment variables. Use `t.Setenv` within tests.
- Add or update tests for any code you change, even if not explicitly asked.

---

## Security considerations

- Never log token values. Never expose secrets in error responses.
- Never commit `.env` files to version control.
- Do not add TLS handling to the Go service — TLS is terminated by a reverse proxy (see ADR-006).
- Do not introduce endpoints that forward arbitrary user input to the Shelly device.
- Keep dependencies at the bare minimum. Run `go mod verify` to validate module integrity.
- See ADR-004 for authentication rules and ADR-007 for audit logging requirements.

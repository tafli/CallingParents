# ADR-004: Data Management — Browser localStorage

## Status

Accepted

## Date

2026-02-20

## Context

The application needs to store:

1. A list of children's names (for quick selection buttons).
2. Connection settings (ProPresenter message ID).

The data set is small (tens of names, a few settings), device-local, and does not need to be shared across devices or persisted on a server.

### Options Considered

1. **Server-side database** (SQLite, PostgreSQL) — provides persistence and multi-device access, but adds complexity, a database dependency, and server state management.
2. **Server-side file storage** (JSON file) — simpler than a DB but still requires server-side state and API endpoints for CRUD.
3. **Browser localStorage** — built into every browser, synchronous key-value storage, persists across sessions, no server dependency.

## Decision

Store all application data in the browser's **localStorage**.

### Data Model

```json
{
  "children": ["Anna", "Ben", "Clara", "David", "Emma"],
  "settings": {
    "messageId": "Eltern rufen"
  }
}
```

- `children` — array of strings, sorted alphabetically for display.
- `settings.messageId` — the ProPresenter message name, UUID, or index to trigger.

### Storage Keys

| Key | Type | Description |
|-----|------|-------------|
| `calling_parents_children` | JSON string[] | List of registered children's names |
| `calling_parents_settings` | JSON object | App configuration (message ID) |

## Consequences

- **No server-side state**: the Go server is stateless, serving only static files and proxying API calls. This simplifies deployment and eliminates database maintenance.
- **Device-local**: data is tied to the specific browser on the specific device. If the phone is replaced, settings and child names must be re-entered. This is acceptable given the small data set.
- **No sync**: if multiple devices are used, each maintains its own child list. This is acceptable for the single-phone setup at child care.
- **Storage limits**: localStorage provides ~5–10 MB, far more than needed for this use case.
- **Data loss risk**: clearing browser data erases the child list. This is mitigated by the list being easy to re-create and by using Chrome's "installed PWA" mode which has separate storage.

# ADR-004: Data Management — Server-Side Children File + Browser localStorage

## Status

Accepted (updated)

## Date

2026-02-20

## Context

The application needs to store:

1. A list of children's names (for quick selection buttons).
2. Connection settings (ProPresenter message ID).

The children list should be centrally managed so that all phones receive the same predefined set of names without manual entry.

### Options Considered

1. **QR code with embedded data** — limited to ~4 KB, requires re-scanning on every change, no auto-sync.
2. **Server-side database** — overkill for a simple name list.
3. **Server-side JSON file + browser localStorage** — admin edits a `children.json` file next to the binary; the server serves it via `GET /children`; the PWA fetches and merges on startup, using localStorage as a local cache.

## Decision

Use a **server-side `children.json` file** as the source of truth for the children list. The PWA merges the server list into localStorage on every load. Settings remain in localStorage only.

### Data Flow

1. Admin creates/edits `children.json` next to the server binary (a JSON array of strings).
2. On startup, the server loads and sorts the file.
3. The PWA calls `GET /children` on every load.
4. Server names not already in localStorage are added (merge, not replace).
5. Workers can still add names locally via the settings screen for ad-hoc children.

### `children.json` Format

```json
[
    "Anna",
    "Ben",
    "Clara"
]
```

If the file does not exist, the server starts with an empty list and the PWA falls back to localStorage only.

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CHILDREN_FILE` | `children.json` | Path to the children names JSON file |

### localStorage Keys

| Key | Type | Description |
|-----|------|-------------|
| `calling_parents_children` | JSON string[] | Merged children list (server + local additions) |
| `calling_parents_settings` | JSON object | App configuration (message ID) |

## Consequences

- **Central management**: the admin edits one file; all phones auto-sync on next load.
- **Merge strategy**: server names are additive — they never remove locally-added names. This lets workers add ad-hoc children while keeping the server list as the base.
- **Offline resilience**: if the server is unreachable, the PWA uses the cached localStorage list.
- **Near-stateless server**: the only server-side state is a read-only JSON file. No database, no write endpoints.
- **Easy reset**: deleting localStorage on the phone reverts to the server-provided list on next load.

# Calling Parents

A Progressive Web App for church child care workers to call parents via ProPresenter audience screens.

When a child in the care area needs their parents, the worker opens the app on an Android phone, taps the child's name, and a message like **"Eltern von Paul"** appears on the main projector/TV in the hall — powered by ProPresenter's Messages API.

## Architecture

All architecture decisions are documented as ADRs in [`docs/architecture/`](docs/architecture/):

| ADR | Title |
|-----|-------|
| [001](docs/architecture/001-project-overview.md) | Project Overview |
| [002](docs/architecture/002-application-type.md) | Application Type — Progressive Web App |
| [003](docs/architecture/003-propresenter-integration.md) | ProPresenter Integration |
| [004](docs/architecture/004-data-management.md) | Data Management — Browser localStorage |
| [005](docs/architecture/005-deployment.md) | Deployment — Go Binary with Embedded PWA |
| [006](docs/architecture/006-cors-api-proxy.md) | CORS Handling — Go Backend Proxy |
| [007](docs/architecture/007-authentication.md) | Authentication — Bearer Token via QR Code |

## Prerequisites

- **ProPresenter** running with the API enabled (Settings → Network)
- A **message template** created in ProPresenter:
  - Name: `Eltern rufen` (must match the `MESSAGE_NAME` environment variable)
  - Template text: `Eltern von {Name}`
  - Token: text token named `Name`
  - Theme: your preferred slide design
- The **Messages layer** enabled in the active Look for the audience screen(s)
- **Go 1.22+** installed for building
- A **`children.json`** file with the children's names (see `children.json.example`):
  ```json
  ["Anna", "Ben", "Clara"]
  ```

## Build

The build script cross-compiles for Linux and Windows:

```bash
./build.sh
```

This produces binaries in `dist/`:

| File | Platform |
|------|----------|
| `dist/calling_parents-linux-amd64` | Linux |
| `dist/calling_parents-windows-amd64.exe` | Windows (ProPresenter host) |

## Run

### On Linux (development)

```bash
cp .env-template .env   # adjust values
./run.sh                # builds, loads .env, starts server
```

### On Windows (ProPresenter machine)

1. Copy `calling_parents-windows-amd64.exe`, `.env-template`, `children.json.example`, and `run.bat` to a folder.
2. Rename `.env-template` to `.env` and adjust values.
3. Copy `children.json.example` to `children.json` and edit the names.
4. Double-click `run.bat` — or run the `.exe` directly.

```bat
run.bat
```

| Variable | Default | Description |
|----------|---------|-------------|
| `PROPRESENTER_HOST` | `localhost` | ProPresenter machine hostname/IP |
| `PROPRESENTER_PORT` | `50001` | ProPresenter API port |
| `LISTEN_ADDR` | `:8080` | Server listen address |
| `CHILDREN_FILE` | `children.json` | Path to children names JSON file |
| `MESSAGE_NAME` | `Eltern rufen` | ProPresenter message template name |
| `AUTO_CLEAR_SECONDS` | `30` | Auto-clear message after N seconds (0 = disabled) |
| `ACTIVITY_LOG` | *(empty)* | Path to JSONL activity log file (empty = disabled) |

## Usage

1. Start the server on the ProPresenter Windows machine (or any machine on the church network).
2. On the Android phone, open Chrome and go to `http://<server-ip>:8080`.
3. When prompted, tap "Add to Home Screen" to install the PWA.
4. Open the app → tap ⚙ (Settings):
   - Tap "Verbindung testen" to verify connectivity.
   - Add children's names for quick selection.
5. To call parents: tap a child's name (or type one) → tap **Senden**.
6. To clear the message: tap **Löschen**.

## Development

```bash
# Run tests
go test ./...

# Format code
gofmt -w .

# Vet
go vet ./...
```

## Project Structure

```
cmd/server/
  main.go              — Server entrypoint, embeds web/ and starts HTTP server
  web/                 — PWA static files (embedded into binary)
    index.html         — App shell
    app.js             — Application logic
    style.css          — Mobile-first styles
    manifest.json      — PWA manifest
    sw.js              — Service worker
    icons/             — App icons
internal/
  children/            — Loads children names from JSON file, serves GET /children
  config/              — Configuration from environment variables
  network/             — LAN IP detection for QR code
  message/             — ProPresenter message send/clear/test handlers
  activitylog/         — Append-only JSONL activity logger
docs/architecture/     — Architecture Decision Records
children.json.example  — Example children names file
```

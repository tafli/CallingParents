# Calling Parents — Church Child Care Parent Notification via ProPresenter

> **Notify parents on the big screen.** A lightweight app for church child care teams to instantly display parent call messages on ProPresenter audience screens — from any phone, no app store needed.

When a child in the nursery or kids' area needs their parents, tap their name and a message like **"Parents of Paul"** appears on the main projector or TV in the worship hall — powered by [ProPresenter](https://renewedvision.com/propresenter/)'s Messages API.

---

## Features

- **One-tap parent calling** — select a child's name from a scrollable grid and send instantly
- **ProPresenter integration** — triggers messages directly on audience screens via ProPresenter's HTTP API
- **Progressive Web App (PWA)** — installable from the browser, works on any phone or tablet (Android, iOS, desktop)
- **No app store required** — open a URL, tap "Add to Home Screen", done
- **QR code setup** — scan the terminal QR code to connect and authenticate in one step
- **Offline app shell** — the interface loads even when the server is momentarily unreachable
- **Auto-clear timer** — messages automatically disappear after a configurable timeout
- **Live connection status** — real-time indicator shows whether ProPresenter is reachable
- **Fixed action bar** — input field and send button stay pinned at the bottom, visible regardless of list size
- **Multi-language support** — German and English included, easily extensible (just add a JSON file)
- **Single binary deployment** — one self-contained executable embeds everything, no dependencies to install
- **Auto-upgrading config** — new configuration options are automatically appended to existing `config.toml` files on startup (with a `.bak` backup), so upgrades never break your setup
- **Cross-platform server** — pre-built for Linux and Windows (runs on the ProPresenter machine or any PC on the network)
- **Bearer token authentication** — simple but effective, prevents unauthorized access on the local network
- **Activity logging** — optional JSONL log of all send/clear events with timestamps
- **Server-side children list** — manage names via a JSON file, synced to all connected devices; manual edits are picked up instantly without restart
- **Haptic feedback** — vibration on send for tactile confirmation
- **Zero external dependencies in the frontend** — no frameworks, no build tools, just HTML/CSS/JS

## How It Works

```
┌──────────────┐          ┌──────────────────┐          ┌──────────────────┐
│   Phone /    │  HTTP    │   Go Server      │  HTTP    │  ProPresenter    │
│   Tablet     │ ──────▸  │   (reverse proxy) │ ──────▸  │  (Messages API)  │
│   (PWA)      │          │   + embedded PWA  │          │  → Audience Screen│
└──────────────┘          └──────────────────┘          └──────────────────┘
```

1. The Go server embeds the PWA and proxies API requests to ProPresenter (solving CORS)
2. On startup, a QR code with an auth token is displayed in the terminal
3. The child care worker scans the QR code → the PWA opens, authenticated
4. Tap a name → the server triggers a ProPresenter message → it appears on screen
5. The message auto-clears after a configurable timeout (or the operator clears it manually)

## Screenshots

<!-- TODO: Add screenshots of PWA main view, settings, and QR code terminal -->

## Quick Start

### Prerequisites

- **[ProPresenter](https://renewedvision.com/propresenter/)** with the HTTP API enabled (Settings → Network)
- A **message template** in ProPresenter:
  - Name: `Eltern rufen` (or any name — set it in `MESSAGE_NAME`)
  - Template text containing a token named `Name` (e.g., `Eltern von {Name}` or `Parents of {Name}`)
  - The **Messages layer** enabled in the active Look for audience screen(s)
- **Go 1.22+** for building from source (or use a pre-built binary from [Releases](https://github.com/tafli/CallingParents/releases))

### 1. Build

```bash
./build.sh
```

Produces binaries in `dist/`:

| File | Platform |
|------|----------|
| `dist/calling_parents-linux-amd64` | Linux |
| `dist/calling_parents-windows-amd64.exe` | Windows |

### 2. Configure

On first run, if no `config.toml` exists, one is created automatically with default values. On subsequent updates, any **new configuration keys** are automatically appended to your existing file (a `config.toml.bak` backup is created first) — your customized values are never overwritten. Edit it to customize:

```bash
# Or copy the example manually:
cp config.toml.example config.toml
```

Edit `config.toml` — at minimum, set the ProPresenter host if the server runs on a different machine:

| TOML key | Env override | Default | Description |
|----------|-------------|---------|-------------|
| `propresenter_host` | `PROPRESENTER_HOST` | `localhost` | ProPresenter machine hostname/IP |
| `propresenter_port` | `PROPRESENTER_PORT` | `50001` | ProPresenter API port |
| `listen_addr` | `LISTEN_ADDR` | `:8080` | Server listen address |
| `children_file` | `CHILDREN_FILE` | `children.json` | Path to children names JSON file |
| `message_name` | `MESSAGE_NAME` | `Eltern rufen` | ProPresenter message template name |
| `auto_clear_seconds` | `AUTO_CLEAR_SECONDS` | `30` | Auto-clear after N seconds (0 = disabled) |
| `activity_log` | `ACTIVITY_LOG` | *(empty)* | Path to JSONL activity log (empty = disabled) |
| `auth_token` | `AUTH_TOKEN` | *(random)* | Fixed auth token (empty = generate on each startup) |

Environment variables override TOML values when both are set (useful for Docker/CI).

### 3. Add Children

```bash
cp children.json.example children.json
```

Edit `children.json` with the children's names:

```json
["Anna", "Ben", "Clara", "David", "Emma"]
```

Names can also be managed in the PWA's settings view and are synced bidirectionally.

### 4. Run

**Linux:**
```bash
./run.sh          # builds and starts the server
```

**Windows (ProPresenter machine):**
1. Place the `.exe`, `children.json`, and `run.bat` in a folder (a `config.toml` is created on first run)
2. Double-click `run.bat`

The server prints a QR code in the terminal — scan it with the phone's camera to open and authenticate the PWA in one step.

### 5. Install the PWA

On the phone, Chrome will prompt **"Add to Home Screen"** — tap it for a full-screen, app-like experience. The QR code only needs to be scanned once; the token is stored locally.

## Adding a Language

Translations live in `cmd/server/web/lang/` as simple JSON files:

1. Copy `lang/en.json` to `lang/xx.json`
2. Translate all values
3. Add `{ code: "xx", label: "Language Name" }` to `availableLanguages` in `js/i18n.js`
4. Rebuild and deploy

## Project Structure

```
cmd/server/
  main.go              — Server entry point, embeds web/ and starts HTTP server
  web/                 — PWA static files (embedded into binary)
    index.html         — App shell
    manifest.json      — PWA manifest
    sw.js              — Service worker for offline caching
    js/
      app.js           — Application logic
      i18n.js          — Internationalization module
    css/
      style.css        — Mobile-first responsive styles
    lang/
      de.json          — German translations
      en.json          — English translations
    icons/             — App icons (192×192, 512×512)
internal/
  auth/                — Bearer token validation middleware
  children/            — Children names file I/O and HTTP handlers
  config/              — TOML configuration loading with auto-merge
  message/             — ProPresenter message send/clear/test handlers
  network/             — LAN IP detection for QR code URL
  activitylog/         — Append-only JSONL activity logger
  version/             — Build-time version info and /version endpoint
docs/architecture/     — Architecture Decision Records (ADRs)
```

## Architecture Decision Records

| ADR | Title |
|-----|-------|
| [001](docs/architecture/001-project-overview.md) | Project Overview |
| [002](docs/architecture/002-application-type.md) | Application Type — Progressive Web App |
| [003](docs/architecture/003-propresenter-integration.md) | ProPresenter Integration |
| [004](docs/architecture/004-data-management.md) | Data Management — Browser localStorage |
| [005](docs/architecture/005-deployment.md) | Deployment — Go Binary with Embedded PWA |
| [006](docs/architecture/006-cors-api-proxy.md) | CORS Handling — Go Backend Proxy |
| [007](docs/architecture/007-authentication.md) | Authentication — Bearer Token via QR Code |

## Releasing

```bash
./release.sh v1.0.0
```

This validates the version, checks for a clean working tree, creates an annotated Git tag, and pushes it. The [GitHub Actions workflow](.github/workflows/release.yml) then builds, tests, and publishes a GitHub Release with binaries and config files attached.

## Development

```bash
gofmt -w .         # format
go vet ./...       # lint
go test ./...      # test
```

## License

This project is licensed under the [MIT License](LICENSE).

## Contributing

Contributions are welcome! Please read the architecture decision records in `docs/architecture/` before making changes. Any modification that affects architecture, API design, authentication, or deployment must conform to (or update) the relevant ADRs.

## Built With

This application was built entirely with [GitHub Copilot](https://github.com/features/copilot) (Claude Opus 4.6) — from architecture design and Go backend to the PWA frontend, service worker, and CI/CD pipeline.

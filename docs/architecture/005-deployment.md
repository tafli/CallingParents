# ADR-005: Deployment — Go Binary with Embedded PWA

## Status

Accepted (updated)

## Date

2026-02-20

## Context

The PWA must be served over HTTP to the Android phone. The Go server also acts as an API proxy (see ADR-006). We need a simple deployment strategy suitable for a church environment where technical expertise may be limited.

### Options Considered

1. **Separate web server (nginx/caddy) + Go backend** — more moving parts, harder to set up.
2. **Go binary serving static files from disk** — requires file management alongside the binary.
3. **Single Go binary with embedded static files** — self-contained, single-file deployment.

## Decision

Build a **single Go binary** that embeds the PWA static files using Go's `embed` package and serves them directly. The binary also proxies API requests to ProPresenter.

The binary is **cross-compiled** for both Linux and Windows (amd64). The primary deployment target is the **Windows machine running ProPresenter**, but a Linux build is also produced for flexibility.

### Cross-Compilation

Go supports cross-compilation natively via `GOOS` and `GOARCH` environment variables. The `build.sh` script produces both binaries in the `dist/` directory:

| Target | Output |
|--------|--------|
| Linux amd64 | `dist/calling_parents-linux-amd64` |
| Windows amd64 | `dist/calling_parents-windows-amd64.exe` |

### Versioning

Version information is injected at build time via Go's `-ldflags` mechanism. The `build.sh` script automatically reads:

- **Version**: from `git describe --tags --always --dirty` (e.g., `v1.0.0`, `v1.0.0-3-gabc1234`, or `abc1234-dirty`)
- **Commit**: short Git SHA from `git rev-parse --short HEAD`
- **Date**: UTC build timestamp

These values are injected into `internal/version/` package variables:

```
go build -ldflags "-X .../version.Version=v1.0.0 -X .../version.Commit=abc1234 -X .../version.Date=2026-02-20T12:00:00Z"
```

All three can be overridden via environment variables (`VERSION`, `COMMIT`, `DATE`) for CI builds.

Version is:
- **Logged at startup**: `calling-parents v1.0.0 (abc1234) built 2026-02-20T12:00:00Z`
- **Exposed via `/version` endpoint**: returns JSON `{"version":"...","commit":"...","date":"..."}` (unauthenticated)
- **Displayed in the PWA**: shown in the settings view footer; full details (commit, date) in a tooltip

Without Git tags, the version defaults to the commit hash. Without ldflags (e.g., `go run`), it shows `dev (unknown)`.

### Configuration

All configuration is via a `config.toml` file (TOML format), with optional environment variable overrides for Docker/CI. The Go binary reads the file directly — no shell sourcing needed.

| Variable | TOML key | Default | Description |
|----------|----------|---------|-------------|
| `PROPRESENTER_HOST` | `propresenter_host` | `localhost` | Hostname/IP of the ProPresenter machine |
| `PROPRESENTER_PORT` | `propresenter_port` | `50001` | ProPresenter API port |
| `LISTEN_ADDR` | `listen_addr` | `:8080` | Address and port the server listens on |

A `config.toml.example` file is provided as a reference. If no `config.toml` exists when the server starts, a default one is created automatically with sensible defaults and comments — no manual copying needed.

When the application is updated with new configuration options, existing `config.toml` files are **automatically upgraded**: on startup, the server detects missing keys and appends them (with comments and defaults) to the end of the file. Before modifying the file, a backup is saved as `config.toml.bak`. Keys that the user has already set (or that exist as commented-out entries) are never overwritten. The default config content is generated from a single source of truth (`allConfigBlocks` in `config.go`), ensuring the auto-created file and the merge logic always stay in sync.

### Release Process

Releases are handled by a combination of a local `release.sh` script and a GitHub Actions workflow:

1. **`release.sh <version>`** — local script that validates the version (semver `vX.Y.Z`), checks for a clean working tree, creates an annotated Git tag, and pushes it to the remote.
2. **`.github/workflows/release.yml`** — triggered on `v*` tag pushes. Runs format check, vet, tests, cross-compiles both binaries, and creates a GitHub Release with all distribution assets attached.

To create a release:

```bash
./release.sh v1.0.0
```

The script performs pre-flight checks (clean tree, correct branch, tag doesn't exist, remote reachable, branch up to date) and prompts for confirmation before pushing.

The GitHub Actions workflow attaches the following files to each release:

| File | Description |
|------|-------------|
| `calling_parents-linux-amd64` | Linux binary |
| `calling_parents-windows-amd64.exe` | Windows binary |
| `config.toml.example` | Configuration template |
| `children.json.example` | Sample children list |
| `run.bat` | Windows startup script |

### Deployment Steps (Windows)

1. Download the latest release from [GitHub Releases](https://github.com/tafli/calling-parents/releases) (or run `./build.sh` locally).
2. Copy `calling_parents-windows-amd64.exe` and `config.toml.example` to the ProPresenter Windows machine.
3. Rename `config.toml.example` to `config.toml` and adjust values (typically `propresenter_host = "localhost"`).
4. Optionally copy `run.bat` to the same directory for convenient startup.
5. Run `calling_parents-windows-amd64.exe` (or `run.bat`).
6. On the Android phone, open Chrome and navigate to `http://<server-ip>:8080`.
7. Install the PWA to the home screen when prompted.

### Running as a System Service

On Windows, the binary can be set up as a **scheduled task** (trigger: "At log on") or wrapped as a Windows service using tools like NSSM. On Linux, systemd is the natural choice. A simple startup script (`run.bat` / `run.sh`) is sufficient for most church setups.

## Consequences

- **Single artifact per platform**: one `.exe` to deploy on Windows, no external dependencies (no runtime, no web server, no database).
- **Easy updates**: replace the binary and restart.
- **Runs alongside ProPresenter**: deployed on the same Windows machine, `PROPRESENTER_HOST=localhost` and no network hops for API calls.
- **Cross-compiled from Linux/macOS**: development and CI can happen on any platform; the Windows binary is produced without needing a Windows build machine.
- **No HTTPS**: on a local network, HTTP is acceptable. If HTTPS is needed in the future, a reverse proxy (caddy) can be added in front.

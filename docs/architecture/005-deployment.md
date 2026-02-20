# ADR-005: Deployment — Go Binary with Embedded PWA

## Status

Accepted

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

### Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PROPRESENTER_HOST` | `localhost` | Hostname/IP of the ProPresenter machine |
| `PROPRESENTER_PORT` | `50001` | ProPresenter API port |
| `LISTEN_ADDR` | `:8080` | Address and port the server listens on |

A `.env-template` file is provided. Copy it to `.env` and adjust values. Both `run.sh` (Linux) and `run.bat` (Windows) load `.env` automatically.

### Deployment Steps (Windows)

1. On the development machine, run `./build.sh` to produce both binaries.
2. Copy `dist/calling_parents-windows-amd64.exe` and `.env-template` to the ProPresenter Windows machine.
3. Rename `.env-template` to `.env` and adjust values (typically `PROPRESENTER_HOST=localhost`).
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

# ADR-002: Application Type — Progressive Web App

## Status

Accepted

## Date

2026-02-20

## Context

The child care team uses an Android phone to send parent notification messages. We need to decide what type of application to build.

### Options Considered

1. **Native Android app (Kotlin/Java)** — full native experience, but requires Android SDK toolchain, Play Store deployment or APK sideloading, and ongoing maintenance for OS updates.
2. **Flutter cross-platform app** — native look and feel on multiple platforms, but heavier toolchain and unnecessary cross-platform capability (only Android is needed).
3. **Progressive Web App (PWA)** — runs in Chrome on Android, can be "installed" to the home screen, no app store needed, standard web technologies.

### Evaluation Criteria

- Ease of development and maintenance
- Deployment simplicity (no app store, no sideloading)
- Usability on Android phone
- Offline capability for the app shell

## Decision

Build a **Progressive Web App (PWA)**.

## Rationale

- **No app store deployment**: the PWA is accessed via a URL and installed to the home screen directly from Chrome. No Google Play account or APK distribution needed.
- **Simple tech stack**: HTML, CSS, JavaScript — no build toolchain, no framework dependencies.
- **Installable**: with a proper manifest and service worker, Chrome prompts the user to "Add to Home Screen", providing an app-like experience with full-screen display.
- **Offline app shell**: the service worker caches the HTML/CSS/JS so the app loads instantly even if the server is momentarily unreachable. API calls still require network access (which is needed anyway to reach ProPresenter).
- **Easy updates**: updating the PWA is a server-side file change; the service worker detects the update and refreshes automatically.
- **Sufficient capability**: the app's requirements (HTTP requests, local storage, touch UI) are fully supported by modern mobile browsers.

## Consequences

- The app cannot use raw TCP sockets from the browser; it must use the HTTP API (this is fine — ProPresenter exposes both TCP and HTTP on the same port).
- A Go server is needed to serve the PWA files and proxy API requests (see ADR-006).
- The app depends on Chrome (or another modern browser) being installed on the Android phone.

### PWA UI Design Decisions

- **No manual clear button**: the ProPresenter operator is responsible for clearing messages from the audience screen. The PWA only sends messages. Server-side auto-clear (if configured) handles automatic removal after a timeout.
- **Auth error screen**: if the PWA loads without an authentication token (no QR code scanned), a full-screen overlay blocks all interaction and instructs the user to scan the QR code. No buttons or features are accessible without a valid token.
- **401 handling**: if any API call returns HTTP 401, the PWA clears the stored token and reloads the page, showing the auth error screen. This handles token expiry on server restart.
- **Input clear button**: the name input field includes an inline `×` button to quickly clear entered text, visible only when the field is non-empty.
- **Version display**: the settings view shows the application version in a footer, fetched from the unauthenticated `/version` endpoint. Full version details (commit, build date) are shown as a tooltip.
- **Connection status dot**: the header shows a colored dot indicating ProPresenter connectivity, polled every 10 seconds.
- **Disconnected state UX**: when ProPresenter is unreachable, the PWA provides clear visual feedback through four mechanisms: (1) the Send button is disabled so users cannot attempt futile sends, (2) a warning banner appears below the header reading "⚠ ProPresenter nicht erreichbar", (3) the children name grid is dimmed and non-interactive, and (4) the header background changes from blue to orange. All four indicators revert automatically when connectivity is restored. The theme-color meta tag also updates so the browser chrome reflects the connection state.
- **Toast notifications**: brief feedback messages appear for send/clear results and errors, auto-dismissing after 3 seconds.

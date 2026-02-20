# ADR-003: ProPresenter Integration

## Status

Accepted

## Date

2026-02-20

## Context

ProPresenter provides an HTTP/TCP API on a configurable port (default `50001`). There are two distinct messaging mechanisms:

1. **Stage Messages** (`/v1/stage/message`) — plain text shown only on stage/confidence monitors. Not visible to the audience.
2. **Presentation Messages** (`/v1/messages`, `/v1/message/{id}/trigger`) — rich messages with token placeholders, shown on the **audience screens** via a dedicated messages layer. These support themes (visual templates) designed in ProPresenter.

Our goal is to display "Eltern von {Name}" on the main audience screens (projector/TV) visible to the congregation.

## Decision

Use the **Presentation Messages API** to trigger messages on the audience screens.

### API Endpoints Used

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/v1/messages` | Discover configured messages on startup / in settings |
| `POST` | `/v1/message/{id}/trigger` | Show the notification message with the child's name |
| `GET` | `/v1/message/{id}/clear` | Hide the notification message |
| `GET` | `/v1/clear/layer/messages` | Emergency: clear all messages |

### Message Template Setup (in ProPresenter)

A message must be pre-created in ProPresenter with the following configuration:

- **Name**: `Eltern rufen` (must match the `MESSAGE_NAME` environment variable on the server)
- **Template text**: `Eltern von {Name}`
- **Token**: a text token named `Name`
- **Theme**: a suitable slide design (font size, colors, positioning) created in ProPresenter

### Triggering a Message

To show "Eltern von Paul" on the audience screens:

```
POST /v1/message/{id}/trigger
Content-Type: application/json

[
  {
    "name": "Name",
    "text": {
      "text": "Paul"
    }
  }
]
```

### Clearing a Message

```
GET /v1/message/{id}/clear
```

### Prerequisites in ProPresenter

1. The message template with a `{Name}` token must exist.
2. The active **Look** must have the messages layer enabled for the audience screen(s).
3. The API must be enabled in ProPresenter > Settings > Network.

## Consequences

- Visual design of the notification (fonts, colors, animation) is fully controlled in ProPresenter, keeping the app simple.
- The message template name is configured on the **server** via the `MESSAGE_NAME` environment variable (default `Eltern rufen`). The PWA does not need to know this value.
- ProPresenter's API has **no authentication**; security relies on the local network being trusted.
- If the message template is deleted or renamed in ProPresenter, the app must be reconfigured.

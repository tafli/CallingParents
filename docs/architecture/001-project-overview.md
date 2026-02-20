# ADR-001: Project Overview

## Status

Accepted

## Date

2026-02-20

## Context

Our church uses ProPresenter as presentation software for services. During services, children are looked after in a separate child care area. When a child needs their parents (e.g., the child is unwell or upset), the child care team currently has no efficient way to notify the parents sitting in the main hall.

We need an application that allows child care workers to send a visible message to the audience screens (projector/TV) in the main hall, displaying text such as "Eltern von Paul" (Parents of Paul).

### Users

- **Child care workers**: use an Android phone to select a child and trigger a message.

### Environment

- Church local Wi-Fi network connecting the Android phone and the ProPresenter computer.
- ProPresenter running on a dedicated computer controlling audience and stage screens.
- An Android phone available at the child care area.

### Constraints

- ProPresenter is the sole presentation engine; all visual output must go through it.
- The solution must be simple enough for non-technical volunteers to use.
- The solution must work on the church's local network (no internet dependency for core functionality).

## Decision

Build a lightweight application named **calling_parents** that connects to ProPresenter's HTTP API to trigger parent notification messages on the audience screens.

## Consequences

- All visual design of the notification message is managed inside ProPresenter (themes, positioning, fonts).
- The application only needs to trigger and clear messages, keeping its scope minimal.
- Requires network connectivity between the Android phone and the ProPresenter machine.

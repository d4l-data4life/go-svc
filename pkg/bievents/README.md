# bievents

Library to instrument BI (Business Intelligence) events in Go services.

BI events capture product-analytics activity (registrations, logins, service-specific
actions) as one JSON object per line on `stdout`, with `event-type: bi-event`. The
platform's Data Collection Rule routes them into the dedicated `bievents_CL` table of the
environment's Log Analytics Workspace, separate from operational logs.

## Usage

Create one emitter at service startup:

```go
emitter := bievents.NewEventEmitter(
    "my-service",
    config.Version,
    os.Getenv("HOSTNAME"),
)
```

Emit events with `Log` (exactly what you pass) or `LogCtx` (fills `user-id` from the
request context under `log.UserIDContextKey` when left empty):

```go
err := emitter.LogCtx(r.Context(), bievents.Event{
    ActivityType: bievents.LoginComplete,
    Data: bievents.LoginCompleteData{
        SessionID: sessionID,
        ClientID:  clientID,
    },
})
```

Emitting a BI event should never break the request — log the returned error and carry on.

## Activity types

`types.go` defines the generic, cross-service activity types (`register`, `email-verify`,
`login-email`, `login-complete`, `logout`, `token-refresh`, `account-delete`) with matching
data structs. Services define their own additional activity types and `json`-tagged data
structs for service-specific events.

## Emitted format

```json
{
  "service-name": "my-service",
  "service-version": "v1.2.0",
  "hostname": "my-service-6d9c7f-abcde",
  "event-type": "bi-event",
  "event-id": "7c1f0e2a-9a4d-4c7e-8f21-6d0a1b2c3d4e",
  "activity-type": "login-complete",
  "user-id": "2541c632-7cfa-4772-9b51-4c74a7618b23",
  "tenant-id": "",
  "consent-document-key": "",
  "session-id": "",
  "event-source": "/pkg/handlers/loginHandler.go:42",
  "timestamp": "2026-07-12T09:31:22Z",
  "data": { "session-id": "9f2b...", "client-id": "web-app" }
}
```

The envelope (`service-name`, `service-version`, `hostname`, `event-type`, `event-id`,
`timestamp` when unset) is filled by the emitter; the rest comes from the `Event`.

## Privacy

BI events must not carry raw personal data. Use `Hash()` (SHA-256) for identifiers that
shouldn't leak, keep `data` minimal, and set `consent-document-key` for events that require
a specific consent to be retained.

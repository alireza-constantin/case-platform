# M0 wire contract — issue #2 gate

This freezes the defaults in [issue #1](https://github.com/alireza-constantin/case-platform/issues/1). Changes to neutral messages require coordination between both tracks. This directory is documentation and synthetic examples, not a runtime package, schema service, or generated SDK. Each side owns its language-specific types and builds independently.

All seven operations below are implemented in M0. Frontend and backend assemblies register `phone-demo` / `m0-v1` and `terminal-demo` / `m0-v1`; the launcher verifies their opaque catalogue metadata at startup. Phone and Terminal keep their gameplay types and private definitions inside their own case modules. Independent frontend/backend builds require no cross-language runtime imports. [Local startup](../README.md) and [real system verification](../docs/verification/ticket-7-system.md) describe the assembled path; synthetic examples remain development aids.

## HTTP operations

All routes share the browser origin. JSON requests use `Content-Type: application/json`; JSON responses use the same media type. Bootstrap/mutations use non-GET methods and reject a supplied Origin different from the configured browser origin. A missing Origin permits direct local API tooling; cookies use SameSite=Lax. Invalid requests return 400. Local bootstrap bounds the request to 1 KiB. No CORS authentication, rotating tokens, or automatic retries.

| Operation | Request | Success |
| --- | --- | --- |
| Establish/reuse guest | `POST /api/guest`, `{}` | 204, no body. Reuse a valid stored guest or create one. Persistent opaque HttpOnly cookie; Secure on HTTPS, usable on local HTTP. Never echo credentials in JSON. |
| Safe catalogue | `GET /api/cases` | 200, `{ "cases": [...] }`; each item has only `case_id`, `case_version`, `title`. Public covers only if the launcher uses them. No private definitions or protected inventories. Public endpoint. |
| Own runs | `GET /api/playthroughs` | 200, `{ "playthroughs": [...] }`, summaries belonging to the guest, newest `created_at` first. |
| Create/restart | `POST /api/playthroughs`, `{ "case_id": "...", "case_version": "..." }` | 201 snapshot with fresh server-initialized state. Restart creates a new identity and retains the previous run. |
| Current view | `GET /api/playthroughs/{playthrough_id}` | 200 snapshot for an owned run and available pinned implementation. |
| Action | `POST /api/playthroughs/{playthrough_id}/actions`, `{ "request_id": "...", "base_revision": 0, "action_type": "...", "payload": {} }` | 200 action result. Case owns tag, payload and outcome semantics. |
| Protected asset | `GET /api/playthroughs/{playthrough_id}/assets/{asset_id}` | 200 bytes with their content type and `Cache-Control: no-store`. Verify guest ownership, pinned case/version and case authorization. Never public/static delivery or service-worker caching. |

Run operations require a valid guest cookie. IDs and versions are opaque strings. Asset IDs are identifiers, never storage paths. The route's playthrough ID is authoritative: action bodies cannot choose a different case/version or assign private state. Do not log secret-bearing payloads or private state.

## Neutral envelopes

A summary contains exactly the baseline metadata: `playthrough_id`, `case_id`, `case_version`, nonnegative integer `revision`, nullable `completed_at`, `created_at`, `updated_at`, and `availability` (`available` or `unavailable`). Times are UTC RFC3339/ISO-8601 strings (for example `2026-01-01T00:00:00Z`). Availability is resolved from registration; missing implementations do not delete state or permanently change lifecycle. Restoring a compatible implementation can restore availability.

A snapshot is `{ "playthrough": <summary>, "view": <opaque JSON> }`. An action result adds `request_id` and opaque JSON `outcome` to that snapshot. Payloads, views and outcomes have no shared gameplay fields or Phone/Terminal union. Private state and server-only definitions are never wire fields.

## Errors

Errors use `{ "error": { "code": "...", "message": "safe explanation" } }`, including asset denials. No private state, credentials, storage paths or internal traces.

| Status | Code | Meaning |
| --- | --- | --- |
| 400 | `invalid_request` | Malformed JSON, invalid envelope, excessive body, or cross-origin mutation. |
| 401 | `missing_guest` | No valid guest session. |
| 404 | `not_found` | Unknown or not-owned resource; protected asset denied. Do not distinguish another guest's resource. |
| 409 | `revision_conflict` | Base revision is stale. |
| 409 | `request_id_conflict` | Committed request ID reused with different input. |
| 409 | `case_version_unavailable` | Pinned implementation is absent. |
| 422 | `action_rejected` | Case action validation rejected. A wrong answer may instead be a successful case-owned outcome. |
| 500 | `internal_error` | Unexpected failure, safe message only. |

## Revision, retry and completion agreement

One atomic durable transition advances revision once, including first completion alone or together with changed case state. No-op/log-only processing need not advance it; a persisted attempt counter does. First completion is monotonic; cases decide post-completion gameplay.

Successful authoritative mutations atomically retain state, revision, first completion if any, relevant events, request/input identity and the safe committed result. Check committed identity before stale revision rejection. Identical ID/input retries (including concurrent duplicates) return the recorded result without a second transition. Different input under a committed ID yields `request_id_conflict`. Input identity covers base revision, tag and payload within the routed run. Malformed, unauthorized, stale and ordinary pre-commit failures need no permanent receipts.

While delivery is uncertain, retain the original request ID and input. A replay's revision belongs to its recorded snapshot, not necessarily the latest state. Never install an older replay over a newer known revision; fetch current view to reconcile. For a revision conflict fetch current view and submit a reconciled action with a new ID; never automatically retry gameplay. No offline queue or fake success. Creation/bootstrap need no distributed idempotency system: avoid blind retries; reconcile uncertain run creation using the own-run list.

[examples.json](examples.json) provides synthetic catalogue, creation, view, action, replay and errors. Protected-byte examples are HTTP behavior above, not a public protected fixture. Case semantics stay in the [Phone agreement](../docs/cases/phone-wire.md) and [Terminal agreement](../docs/cases/terminal-wire.md).

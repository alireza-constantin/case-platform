# Ticket #4: authoritative Phone API

Verified locally on 2026-10-07 with Go and native PostgreSQL 18. Tests use a dedicated `m0_ticket4` database at `127.0.0.1:54329`; they do not touch the running primary API or require a frontend build.

Apply `migrations/0001_guests.sql`, `0002_playthroughs.sql`, `0003_action_events.sql`, and `0004_action_receipts.sql` in order with `psql -v ON_ERROR_STOP=1`. All are transactional and repeatable. Then, from `server/`:

```powershell
$env:TEST_DATABASE_URL = 'postgres://case_platform:local_only@127.0.0.1:54329/m0_ticket4?sslmode=disable'
go test ./... -count=1
go build -o bin/api.exe ./cmd/api
go vet ./...
```

All passed against real PostgreSQL. Without `TEST_DATABASE_URL`, the integration tests explicitly skip and provide no acceptance evidence. `git diff --check` and kernel import inspection passed; the kernel imports pgx and standard-library capabilities, with no case-package imports or gameplay branches. Backend assembly explicitly registers Phone through the neutral `Case` interface. Catalogue metadata comes from those same registrations.

Seven grouped HTTP scenarios in `server/internal/assembly/phone_test.go` cover:

- Guest-owned create/list/read, newest-first own runs, nondestructive restart and API-instance restart. Missing guests receive 401; another guest receives the same safe 404 as an unknown resource.
- Public inferable messages, locked projection without protected inventory/private state, rejected forged flags, wrong attempts, successful unlock, atomic gallery opening plus first completion, durable resume and monotonic repeated opening. Wrong attempts persist the private attempt counter; already-unlocked attempts and repeated opening do not advance revision.
- Exact recorded result replay before stale checking, conflicts for changed payload/base revision/action tag, old replay retaining its committed revision, and rejected request IDs remaining usable for corrected submissions.
- Concurrent identical submissions returning one committed result; concurrent distinct submissions sharing a base revision yielding one success and three revision conflicts.
- Protected SVG bytes delivered only for an owned unlocked run, with `Cache-Control: no-store` and `image/svg+xml`. Copied URLs and locked second runs remain denied. Missing registrations retain unavailable summaries; restoring the compatible registration resumes its original state/version.
- Malformed/missing envelopes, forged private state, oversized/trailing JSON, invalid case payloads, unknown tags, cross-origin requests and incorrect media types leaving progress unchanged. Errors use the safe neutral JSON envelope and never echo guest credentials.
- Valid JSON password strings containing U+0000 remaining ordinary case-owned wrong attempts. Input identity canonicalization happens in Go so PostgreSQL JSONB restrictions do not leak into the action payload agreement.

TDD failures were observed before each working vertical slice: create returned 404, actions returned 404, committed retry returned 409, protected content returned 404, and the valid U+0000 password returned 500. Each was followed by its minimal working implementation and a passing focused real-database check.

Actions lock the owned run in an ordinary PostgreSQL transaction. JSONB equality detects actual private-state changes; first completion alone also constitutes a durable change. State, revision, first completion, relevant events and hashed request/input identity plus the safe replay snapshot commit together. No raw password inputs are stored in receipts or operational logs. Case callbacks are synchronous over supplied JSON and server-owned definitions, with no database/network/filesystem capability.

Phone owns the real answer and protected image exclusively in its server package. Its public projection uses the frozen `messages`, `gallery_unlocked` and authorized `gallery_assets` shapes; actions remain `gallery.attempt` and `gallery.open`, with `accepted` and `opened` outcomes. No Terminal implementation, generic gameplay engine or external deployment is included. Browser integration and the first-completion-alone contract regression remain later integration-ticket evidence.

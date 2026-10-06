# Ticket #6: Terminal through the existing authoritative API

Verified locally on 2026-10-07 with native PostgreSQL 18 and Go. The isolated worktree started from integration `f2991fc`, merged integration `7290da7` before implementation, and uses its own `m0_ticket6` database at `127.0.0.1:54329`. No frontend build or running primary API was used.

Apply the existing migrations `0001_guests.sql` through `0004_action_receipts.sql` in order with `psql -v ON_ERROR_STOP=1`. From `server/`:

```powershell
$env:TEST_DATABASE_URL = 'postgres://case_platform:local_only@127.0.0.1:54329/m0_ticket6?sslmode=disable'
go test ./... -count=1
go build -o bin/api.exe ./cmd/api
go vet ./...
```

All passed against real PostgreSQL, including every existing Phone scenario and the bootstrap suite. `git diff --check` passed. Missing `TEST_DATABASE_URL` explicitly skips the integration checks and provides no acceptance evidence.

Four grouped HTTP scenarios in `server/internal/assembly/terminal_test.go` verify:

- A guest creates Terminal, sees a safe root with public instructions, reads help/location/listing, navigates the simulated environment and resumes the persisted directory. Unknown paths and forged private fields cannot change progression.
- Direct and out-of-order reads/mount operations remain denied. A short discovered prerequisite sequence durably changes workstation state, exposes authorized protected-file metadata, and permits no-store text delivery. Another guest and a locked second run cannot use the unlocked run's permission. Reaching the report commits its private read milestone and first completion with one revision; repeated reads preserve that revision/completion. API-instance restart retains the full state and protected access.
- Exact committed replay, changed-input ID conflicts, stale revision rejection, old replay carrying its committed revision, repeated prerequisite no-ops, nondestructive restart and untouched Phone progress for the same guest. Recorded Terminal results survive API-instance restart.
- Missing/invalid/oversized command fields, forged state, unknown commands, traversal-shaped inputs, shell chaining and real-network-shaped inputs remaining invalid or harmless simulated outcomes. No actual shell, filesystem, network or asynchronous capability is present in Terminal's imports or callbacks.

TDD observed a failing Terminal create (409 before registration) and a failing prerequisite-flow scenario (power did not change state) before the corresponding implementation slices passed. Replay and Phone regressions use the already-working neutral path unchanged.

Terminal implements the exact existing `kernel.Case` interface: metadata, initialization, projection, synchronous action computation and asset permission/content. Backend assembly adds its registration. **No kernel, neutral contract, migration or Phone implementation changes were needed.** Command/state/file concepts stay inside `server/internal/cases/terminal`; protected report bytes stay server-owned and are absent from initial projections, public contract examples and frontend fixtures. The case-owned wire notes describe frozen shapes and discoverable server instructions without publishing the actual sequence or protected report.

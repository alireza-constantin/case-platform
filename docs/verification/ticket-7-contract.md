# Ticket #7: backend contract regressions

Executed on 2026-10-07 in an isolated worktree based on integration `45ea0c7078eccb3bda8da96131162371ffbde932`. This evidence covers the narrow backend-contract portion of #7. Browser journeys, startup documentation, registration mismatch and frontend replay evidence are delivered separately by the system-integration track.

Used native PostgreSQL 18 at `127.0.0.1:54329` with dedicated database `m0_ticket7-contract`, applying existing migrations `0001_guests.sql` through `0004_action_receipts.sql` in order. All HTTP servers are `httptest` instances on isolated ephemeral ports; none binds shared API port 8080. From `server/`:

```powershell
$env:TEST_DATABASE_URL = 'postgres://case_platform:local_only@127.0.0.1:54329/m0_ticket7-contract?sslmode=disable'
go test ./internal/assembly -run 'Test(CompletionOnly|CompletedPin)' -count=1
go test ./... -count=1
go build -o bin/api.exe ./cmd/api
go vet ./...
go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/kernel/...
```

Focused regressions, full HTTP/PostgreSQL suite, backend build and vet passed. `git diff --check` passed. Independent kernel dependency inspection found only neutral standard-library/pgx dependencies and no case or assembly imports, directly or transitively. Missing `TEST_DATABASE_URL` skips the integration tests and provides no database evidence.

Two grouped scenarios in `server/internal/assembly/contract_test.go` exercise the public HTTP boundary and the approved read-only durable-outcome inspection:

- **First completion alone, concurrent duplicates, replay and continuation.** A test-only wrapper delegates to the existing Phone module under its unchanged identity, overriding only the proposed JSON state for the already-valid gallery-opening transition. Four identical concurrent completion requests return exactly one recorded result. Case JSON stays unchanged while revision advances once, first completion persists, and exactly one additional relevant event and receipt commit. Identical retry is accepted despite its now-stale base; an older unlock replay keeps its original snapshot without rewriting current progress. Changed input under a committed ID conflicts. Repeated completion and case-allowed post-completion attempts leave revision/event/receipt counts and original timestamp unchanged, including an action that emits no completion signal. The completion-only receipt survives API-instance restart.
- **Completed pin removal and restoration.** A real Phone run completes normally before its implementation is omitted from registration over the same database. Its own-run summary retains ID, case/version, revision and original first-completion time while reporting unavailable. Read, action and protected-asset requests fail safely rather than changing state/history or falling back. Restoring the compatible production registration resumes its original unlocked completed view and protected content. Its exact completion receipt remains replayable, and allowed post-completion exploration is a no-op. Read-only inspection confirms the private JSON and event/receipt counts remain unchanged throughout removal, restoration and replay.

Database inspection observes durable case JSON and relevant event/receipt totals for these runs; assertions do not inspect private helper names, lock choices or SQL text. No guest credential, solution input or private state is printed in evidence.

Both new grouped regressions passed their initial real-database execution. No missing behavior required a red-to-green product change: the existing neutral action transaction already handles first completion without a JSON change and reversible version availability. No third runnable demo, kernel capability, gameplay rule, migration, production registration, frontend or contract change was added. Existing Phone and Terminal HTTP suites remain green.

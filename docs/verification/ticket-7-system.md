# Ticket #7 — real assembled M0 verification

Executed on 2026-10-07 from an isolated checkout based on `45ea0c7`, incorporating backend contract regressions from integration `c7daddf`. The accepted path is a real browser → Vite same-origin proxy → Go → PostgreSQL. Synthetic frontend fixtures were not used for final authority or persistence evidence.

## Environment and startup

Used Bun 1.4.2, Node 24.19.0, Go 1.27.1 and native PostgreSQL 18.6 (Compose CLI 2.40.3). Created fresh database `m0_ticket7-system` in the existing local native cluster at `127.0.0.1:54329`; applied `0001_guests.sql`, `0002_playthroughs.sql`, `0003_action_events.sql` and `0004_action_receipts.sql` in order with `psql -v ON_ERROR_STOP=1`. Installed frontend dependencies from the frozen lockfile. Go bound owned port 8080 with:

```powershell
$env:DATABASE_URL = 'postgres://case_platform:local_only@127.0.0.1:54329/m0_ticket7-system?sslmode=disable'
$env:PUBLIC_ORIGIN = 'http://localhost:5187'
```

Vite ran from this checkout's `app/` at private port 5187. The primary checkout's Vite at 5173 and existing PostgreSQL service were left running. Browser session `ticket7real` used ignored persistent profile `output/playwright/ticket7real/profile`. All API request helpers use absolute same-origin URLs; no credentials are printed or copied into evidence.

The updated [README](../../README.md) describes both native and Compose startup, all four migrations, exact-origin configuration, case usage and independent builds. Compose configuration validates with `docker compose config --quiet`. Its PostgreSQL 18 volume was corrected to `/var/lib/postgresql`, following the [official image layout](https://github.com/docker-library/docs/blob/master/postgres/README.md#pgdata). Container-engine startup was not executed; the native PostgreSQL path supplied live acceptance evidence.

## Grouped real evidence

| Group | Executed behavior and result |
| --- | --- |
| Both journeys and durable lifecycle | A fresh guest played Phone from projected messages through a persisted wrong attempt (revision 1), correct server-validated unlock (revision 2), and gallery opening plus first completion (revision 3). Reload restored unlock before opening. Terminal displayed real help/notes, denied premature read without mutation, followed the privately loaded sequence discoverable in server notes, restored ready state after reload, navigated the authorized archive, and read the actual protected report at revision 7. Authorized image/text responses were no-store and the reader matched actual API bytes. |
| Full process restart, switching and Restart | Closed browser process 4688 and reopened process 33896 using the same profile; also stopped/restarted the owned Go process. Both original ID/pin/revision/completion snapshots restored from the server. Phone → Terminal → Phone canceled a pending protected-file delivery after real Go authorization, with no retained reader/output or React lifecycle errors. Restart created two new locked revision-0 identities, and all four original/fresh runs remained listed. Old completed runs retained their first completion timestamps and revisions. |
| Authority and disclosure | A separate real guest could not read, mutate, or use copied protected URLs for either case (safe 404), and a missing guest received 401. Locked second runs could not inherit an unlocked run's asset permission. Forged case flags received 422; malformed, oversized and clearly cross-origin actions received 400 and left projected state/revision unchanged. Forged localStorage flags and `/gallery` or `/archive` suffixes could not reveal protected content. Initial Phone view omitted the actual answer/inventory; initial Terminal view omitted protected file metadata. A runtime private-definition check found actual answers, maintenance commands and protected bytes absent from frontend source and production output. |
| State, delivery and connection | Intercepted a real Phone request, let `route.fetch()` obtain its committed revision-1 Go response, and aborted only delivery to the browser. No automatic retry occurred. A real second writer unlocked/opened to revision 3; refresh made that newer completion known; explicit identical original retry returned the real recorded older locked response, while SDK/case UI kept the newer unlocked completed state. Four distinct concurrent requests at one base produced one commit and three revision conflicts. Four identical concurrent requests shared revision 1 and the same recorded result; changed input under that ID received `request_id_conflict`. Repeated completion preserved first completion/revision. Separately stopping Go during a loaded Terminal action produced uncertainty, one request and no fake completion/offline queue; after Go restart, explicit original-envelope retry returned the real unchanged locked revision-0 outcome. |
| Registration, contracts and boundaries | With the real API unchanged, a temporary frontend assembly version mismatch initially failed to block startup. Neutral ID/version-set validation was added to the launcher; the same mismatch then produced a clear error with no case execution controls. Restoring assembly metadata started both real registrations normally. Independent frontend typecheck/build/import enforcement and Go build/vet/full real PostgreSQL HTTP suite passed after merging the contract portion. Backend completion-only, event/receipt counts and removed/restored completed-pin guarantees pass in the [contract evidence](ticket-7-contract.md). |

Safe original-run evidence:

| Case | Playthrough ID | Revision | First completion (UTC) | Relevant events / receipts |
| --- | --- | --- | --- | --- |
| Phone | `3887304b09845523eed1506b2094c3f1` | 3 | `2026-10-06T22:36:36.649724Z` | 3 / 3 |
| Terminal | `91729fde56343bd02bbe0dca624cd9e2` | 7 | `2026-10-06T22:36:37.452932Z` | 4 / 7 |

Read-only PostgreSQL inspection confirmed those durable envelope values and event/receipt totals after process restarts and the full suite. It did not print credentials or private case JSON. Direct Vite `/@fs/` probes for server case source and the ignored private verification-input file returned 403; a normal server-shaped frontend URL returned only SPA HTML. New Restart IDs were `cf2274b7a7b8473d7f2911d85cd95ec0` and `67bf6d51a6f094ead952ed213dff3094`, both initially revision 0 and incomplete. Later delivery/race scenarios use separate real runs.

## Reproduce the grouped browser checks

Use a fresh database/profile for the first journey. Start owned Go and Vite with the origin above, then from the repository root:

```powershell
node app/verification/load-private-inputs.mjs
npx --yes --package @playwright/cli playwright-cli -s=ticket7real open http://localhost:5187 --persistent --profile=output/playwright/ticket7real/profile
npx --yes --package @playwright/cli playwright-cli -s=ticket7real run-code --filename app/verification/real-journeys.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket7real close
```

Stop/restart only the owned Go process against the same database, then reopen the **same** browser profile:

```powershell
npx --yes --package @playwright/cli playwright-cli -s=ticket7real open http://localhost:5187 --persistent --profile=output/playwright/ticket7real/profile
npx --yes --package @playwright/cli playwright-cli -s=ticket7real run-code --filename app/verification/real-lifecycle.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket7real run-code --filename app/verification/real-authority.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket7real run-code --filename app/verification/real-delivery.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket7real goto http://localhost:5187
npx --yes --package @playwright/cli playwright-cli -s=ticket7real run-code --filename app/verification/real-registration.cjs
```

For the registration negative probe, temporarily change only the assembly's Terminal registration to `{ ...terminal, case_version: 'verification-mismatch' }`, open `http://localhost:5187/?expectMismatch`, and run `real-registration.cjs`. It must fail execution clearly. Restore the assembly list `[phone, terminal]`, return to the origin and run the positive check. The temporary mismatch is not committed.

For connection loss, first navigate to a retained locked Terminal run while Go is running and wait for its command input. Stop only the owned Go process, run `real-connection.cjs`, restart Go against the same database/origin, then run that script again to check explicit original retry. No browser reload is needed between these two phases.

The private input loader reads existing server definitions at runtime and creates ignored `output/ticket7-private-inputs.json`; it never prints values. The browser checks load it through a temporary file input that is removed immediately. Scripts and public fixtures contain no actual answer, unlocking sequence or protected bytes. Test-only localStorage keys hold safe run IDs or a public failed-command envelope for process/retry comparison; the application never consumes them as authority. Screenshots live under ignored `output/playwright/ticket7real` and contain legitimately revealed test-run material. No credential/storage dumps or secret-bearing traces are exported.

## Independent build and contract commands

```powershell
bun run --cwd app typecheck
bun run --cwd app build
node app/verification/load-private-inputs.mjs --verify-bundle
docker compose config --quiet
cd server
$env:TEST_DATABASE_URL = 'postgres://case_platform:local_only@127.0.0.1:54329/m0_ticket7-system?sslmode=disable'
go test ./... -count=1
go build -o bin/api.exe ./cmd/api
go vet ./...
go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/kernel/...
```

All passed with real PostgreSQL tests enabled, including the merged first-completion-only and completed-pin restoration regressions. A missing `TEST_DATABASE_URL` skips database tests and is not evidence. Frontend platform imports remain neutral; backend kernel imports only standard-library/pgx dependencies, with no case/assembly imports or gameplay parsers. Adding Terminal required its case modules and assembly registrations, with no common gameplay union, terminal engine, shell execution, or neutral SDK expansion.

The product changes found by integration are limited to neutral startup registration validation and the PostgreSQL 18 Compose volume correction. The remaining changes update implementation-status/startup documentation and add ordinary grouped verification scripts. No production deployment, extra demo, schema/migration, manifest service, generated SDK or specialized test framework was introduced. Own API/Vite/browser processes are stopped before handoff; the existing shared PostgreSQL and primary Vite remain running.

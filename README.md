# Case platform

M0 runs two distinct investigations through one React/Vite app, Go API and PostgreSQL database. **Phone Demo — The Last Photograph** presents messages and a protected gallery. **Terminal Demo — The Harbour Workstation** presents a simulated workstation and protected file. Each case owns its gameplay, visuals, server transitions and completion conditions; the platform provides neutral guest/playthrough transport and persistence.

## Local startup

Prerequisites: Bun 1.4+, Node.js 22.12+ for Vite tooling, Go 1.25+, and either a running Docker Compose engine or native PostgreSQL 17+. No cloud credentials. The frontend lockfile and Go module files pin dependencies. Database credentials below are disposable local defaults.

From the repository root, start PostgreSQL with Compose:

```sh
docker compose up -d --wait
for migration in 0001_guests.sql 0002_playthroughs.sql 0003_action_events.sql 0004_action_receipts.sql; do
  docker compose exec -T db psql -U case_platform -d case_platform -v ON_ERROR_STOP=1 -f "/migrations/$migration" || exit 1
done
```

PowerShell uses the same migrations:

```powershell
docker compose up -d --wait
foreach ($migration in @('0001_guests.sql','0002_playthroughs.sql','0003_action_events.sql','0004_action_receipts.sql')) {
  docker compose exec -T db psql -U case_platform -d case_platform -v ON_ERROR_STOP=1 -f "/migrations/$migration"
  if ($LASTEXITCODE -ne 0) { throw 'Migration failed' }
}
```

Apply all four numbered migrations in order before starting Go. Each is transactional and repeatable. The named volume retains guests, runs, events and committed receipts; `docker compose down` stops the database without deleting it. The PostgreSQL 18 image mounts `/var/lib/postgresql`, matching its [official data-directory layout](https://github.com/docker-library/docs/blob/master/postgres/README.md#pgdata).

In a second terminal:

```sh
cd server
go run ./cmd/api
```

In a third terminal:

```sh
cd app
bun install --frozen-lockfile
bun run dev
```

Open **http://localhost:5173**. Vite proxies `/api` to Go at `127.0.0.1:8080`, so the browser uses one origin. PostgreSQL listens at `127.0.0.1:54329`. Keep `localhost` in the browser URL: Go defaults `PUBLIC_ORIGIN` to `http://localhost:5173`, and Vite uses a strict port. If using another frontend port, set `PUBLIC_ORIGIN` to that exact origin before starting Go. `DATABASE_URL` overrides the disposable default database connection.

HTTPS origins set Secure cookies; local HTTP uses a persistent opaque HttpOnly, SameSite=Lax cookie. The credential never appears in JSON or localStorage. Clearing cookies establishes another guest. The application stores authoritative progress only in PostgreSQL; it does not offer account recovery or offline gameplay.

### Native PostgreSQL alternative (PowerShell)

If Docker is unavailable, start a separate native cluster. Adjust the binary directory for the installed version. These commands are for the first setup of a fresh ignored `output/postgres` directory and do not change another PostgreSQL service:

```powershell
$pgBin = 'C:\Program Files\PostgreSQL\18\bin'
New-Item -ItemType Directory -Force output | Out-Null
Set-Content -LiteralPath output/pg-password.txt -Value 'local_only'
& "$pgBin/initdb.exe" -D output/postgres -U case_platform --auth=scram-sha-256 --pwfile=output/pg-password.txt --encoding=UTF8
& "$pgBin/pg_ctl.exe" -D output/postgres -l output/postgres.log -o '-h 127.0.0.1 -p 54329' -w start
& "$pgBin/psql.exe" 'postgres://case_platform:local_only@127.0.0.1:54329/postgres' -v ON_ERROR_STOP=1 -c 'CREATE DATABASE case_platform'
foreach ($migration in @('0001_guests.sql','0002_playthroughs.sql','0003_action_events.sql','0004_action_receipts.sql')) {
  & "$pgBin/psql.exe" 'postgres://case_platform:local_only@127.0.0.1:54329/case_platform' -v ON_ERROR_STOP=1 -f "migrations/$migration"
  if ($LASTEXITCODE -ne 0) { throw 'Migration failed' }
}
```

Run `initdb` and database creation only once. Subsequent starts use the same `pg_ctl ... start`; stop this owned cluster with `& "$pgBin/pg_ctl.exe" -D output/postgres -m fast -w stop`. Use one PostgreSQL path at a time on port 54329. Native and Compose paths use the same migrations and API defaults. Local clusters, browser profiles, screenshots and private verification inputs belong under ignored `output/`.

## Playing the demos

Start a case from the launcher. In Phone, browse Messages, infer the gallery password from its public clues, enter an attempt, and open the server-unlocked gallery to complete. In Terminal, start with `help` and the public workstation notes; navigate the projected directories and follow the returned instructions. The permitted report read completes the case. These commands are entirely simulated.

**Continue** reads the server's current projection. **Restart** creates a new identity with fresh state and retains the old run. Reload and reopening the same browser profile restore committed progress and first completion. Case navigation and local flags grant no protected permission. A lost response offers an explicit original-request retry; a stale revision requires refresh. There is no automatic action retry or offline queue.

## Verification and ownership

Frontend, from `app/`: `bun run typecheck` and `bun run build`. Build includes the TypeScript platform/case/assembly import check. Backend, from `server/`: `go build -o bin/api ./cmd/api` and `go vet ./...`. They build independently and import no runtime code from the other language.

The grouped HTTP tests require the real migrated PostgreSQL database:

```powershell
cd server
$env:TEST_DATABASE_URL = 'postgres://case_platform:local_only@127.0.0.1:54329/case_platform?sslmode=disable'
go test ./... -count=1
```

On POSIX: `TEST_DATABASE_URL='postgres://case_platform:local_only@127.0.0.1:54329/case_platform?sslmode=disable' go test ./... -count=1`. Without the variable, database tests explicitly skip; a skip provides no live evidence. The tests cover both case journeys, ownership/protected delivery, revisions/concurrency/receipts, monotonic completion, and removed/restored version pins. [Real browser/system evidence](docs/verification/ticket-7-system.md) records full process restarts, case switching and genuine lost-response replay. [Contract evidence](docs/verification/ticket-7-contract.md) covers unchanged-state first completion and completed-pin restoration. Earlier frontend synthetic scenarios support independent development and are not final authority evidence.

The launcher compares the frontend assembly's opaque IDs/versions with the real catalogue at startup. An obvious mismatch blocks case execution with a clear error; this check does not require cross-side compilation or a manifest service. Historical unavailable run pins remain separate from current catalogue agreement.

Only each side's assembly imports/registers cases. `app/src/platform` contains no Phone/Terminal mechanics or case imports; `server/internal/kernel` contains no case imports, gameplay branches or parsers. Cases depend on neutral contracts/capabilities, not each other. Inspect backend dependencies with `go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/kernel/...` from `server/`.

```text
app/src/       platform/   cases/{phone,terminal}/   assembly/
server/        cmd/api/    internal/{kernel,cases/{phone,terminal},assembly}/
contracts/     neutral HTTP agreement and safe synthetic examples
migrations/    guest identity, playthroughs, relevant events, committed receipts
docs/          architecture, ADRs, acceptance, case agreements, verification
```

[Neutral wire contract](contracts/README.md), [Phone agreement](docs/cases/phone-wire.md) and [Terminal agreement](docs/cases/terminal-wire.md) describe the small transport shapes. Correct answers, private definitions and protected content remain server-owned. M0 adds no generic gameplay engine, real shell/filesystem, production deployment, or new infrastructure.

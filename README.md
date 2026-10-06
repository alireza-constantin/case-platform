# Case platform

Issue #2 bootstrap: a React/Vite launcher establishes an anonymous guest through Go and PostgreSQL, then loads an empty safe catalogue. Phone/Terminal gameplay, playthrough persistence/actions and protected delivery belong to later tickets. Their HTTP agreement is frozen in [contracts](contracts/README.md); safe case-owned examples are in [Phone](docs/cases/phone-wire.md) and [Terminal](docs/cases/terminal-wire.md).

## Local startup

Prerequisites: Bun 1.4+, Node.js 22.12+ (Vite tooling), Go 1.25+, and either Docker Compose with a running Docker engine or native PostgreSQL 17+. No cloud credentials. Bun's lockfile and Go's module/sum files pin dependencies. Local database credentials below are disposable development defaults.

From the repository root:

```sh
docker compose up -d --wait
docker compose exec -T db psql -U case_platform -d case_platform -v ON_ERROR_STOP=1 -f /migrations/0001_guests.sql
```

Apply the migration before starting the API. This first migration is transactional and safely repeatable; future numbered SQL migrations are applied explicitly in order. No migration runner or playthrough schema is built in this ticket. The named Docker volume retains guests across restarts. `docker compose down` stops the database without deleting its volume.

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

Open **http://localhost:5173**. Vite proxies `/api` to Go at `127.0.0.1:8080`; the browser uses one origin, with no CORS or TLS setup. PostgreSQL listens only on `127.0.0.1:54329`. Keep `localhost` in the browser URL: Origin validation defaults to `http://localhost:5173`, and strict Vite port selection prevents accidental changes. `DATABASE_URL` and `PUBLIC_ORIGIN` can override the Go defaults. HTTPS origins cause Secure cookies; local HTTP uses HttpOnly, SameSite=Lax, Path=/ and a one-year persistent lifetime. Only the stored random credential identifies the guest; it never appears in JSON or localStorage. Clearing cookies establishes a new guest. There is no recovery/cleanup system.

### Native PostgreSQL alternative (PowerShell)

If Docker is unavailable, use installed PostgreSQL binaries to start a separate local cluster. These commands do not modify another running PostgreSQL service. From the repository root, adjust the binary directory for your installed version:

```powershell
$pgBin = 'C:\Program Files\PostgreSQL\18\bin'
New-Item -ItemType Directory -Force output | Out-Null
Set-Content -LiteralPath output/pg-password.txt -Value 'local_only'
& "$pgBin/initdb.exe" -D output/postgres -U case_platform --auth=scram-sha-256 --pwfile=output/pg-password.txt --encoding=UTF8
& "$pgBin/pg_ctl.exe" -D output/postgres -l output/postgres.log -o '-h 127.0.0.1 -p 54329' -w start
& "$pgBin/psql.exe" 'postgres://case_platform:local_only@127.0.0.1:54329/postgres' -v ON_ERROR_STOP=1 -c 'CREATE DATABASE case_platform'
& "$pgBin/psql.exe" 'postgres://case_platform:local_only@127.0.0.1:54329/case_platform' -v ON_ERROR_STOP=1 -f migrations/0001_guests.sql
```

Run `initdb` and database creation only on first setup. Subsequent starts use `pg_ctl ... start`; stop with `& "$pgBin/pg_ctl.exe" -D output/postgres -m fast -w stop`. Do not start Docker's database and the native alternative on the same port simultaneously. Both paths use the same API defaults and migration. `output/` is ignored, including the local cluster and browser profiles.

## Verification and ownership

Frontend independently, from `app/`: `bun run typecheck`, `bun run build`. Backend independently, from `server/`: `go build -o bin/api ./cmd/api`, `go vet ./...`. The full HTTP integration suite requires the running migrated PostgreSQL database:

```powershell
cd server
$env:TEST_DATABASE_URL = 'postgres://case_platform:local_only@127.0.0.1:54329/case_platform?sslmode=disable'
go test ./... -count=1
```

On POSIX: `TEST_DATABASE_URL='postgres://case_platform:local_only@127.0.0.1:54329/case_platform?sslmode=disable' go test ./... -count=1`. Without that variable the integration test explicitly skips; a skip is not live acceptance evidence. The tests exercise real HTTP/SQL guest creation/reuse, API-instance restart, unknown-cookie replacement, safe empty catalogue, and invalid/cross-origin requests.

Browser verification: open the launcher, observe `POST /api/guest` → 204 and `GET /api/cases` → 200 with `{"cases":[]}`, reload, close/reopen the browser in the same persistent profile, and confirm the guest cookie stays the same. Stop Go, reload and confirm a visible failure rather than fake guest success; restart Go and use **Try again**. No automatic retries are made.

Dependency review is separate on each side: `app/src/platform` imports only React and its own neutral modules/styles; it must not import `app/src/cases` or assembly. Inspect frontend imports, and use `go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/kernel/...` from `server/` for the backend boundary. `internal/kernel` must not import `internal/cases` or assembly. Cases depend on neutral capabilities; only each side's assembly imports/registers implementations. In #2 both registries are intentionally empty; registration agreement and mount/runtime callbacks arrive when cases exist.

```text
app/src/       platform/   cases/   assembly/
server/        cmd/api/    internal/{kernel,cases,assembly}/
contracts/     neutral HTTP documentation and synthetic examples
migrations/    guest identity only
docs/          architecture, ADRs, acceptance and case-owned agreements
```

The two tracks can build separately with the contract/examples, without runtime imports across languages. Neither a shared package nor a universal gameplay schema is introduced.

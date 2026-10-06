# Issue #2 verification

Executed locally on 2026-10-06, Windows, using Bun 1.4.2, Node 24.19.0, Go 1.27.1 and native PostgreSQL 18. The Go module minimum is 1.25; frontend versions are pinned in `app/package.json` and `app/bun.lock`.

## Executed evidence

- Started an isolated PostgreSQL cluster at `127.0.0.1:54329` using the README native alternative. Applied `migrations/0001_guests.sql`, then applied it again successfully (existing table notice, transaction committed).
- `bun install --frozen-lockfile`, `bun run typecheck`, `bun run build` passed from `app/`; Vite produced the launcher bundle without a backend build dependency.
- `go build -o bin/api.exe ./cmd/api`, `go vet ./...`, and `TEST_DATABASE_URL=… go test ./... -count=1` passed from `server/`. HTTP tests used the real migrated PostgreSQL database, not mocks. The first test build failed before a kernel implementation existed. A real run also failed before database creation, then passed after native setup/migration. A missing test URL is explicitly skipped, not counted as proof.
- Playwright CLI opened the launcher in a persistent Microsoft Edge profile. Real requests through Vite were `POST http://localhost:5173/api/guest` → **204**, then `GET http://localhost:5173/api/cases` → **200**. UI displayed **Guest access ready** and **No cases available yet**.
- Cookie checks found persistent expiry, HttpOnly and SameSite=Lax. Reload and a full browser close/reopen with the same profile retained the credential (compared by SHA-256 fingerprint, never printed the credential). API-instance restart reuse is also covered by the HTTP test.
- PostgreSQL inspection confirmed stored guest rows after HTTP/browser bootstrap. No browser memory/localStorage identity implementation exists.
- Stopped Go and reloaded: the launcher displayed **Guest access could not be established. Check the connection and try again.** After Go restart, **Try again** restored the ready empty catalogue. No automatic bootstrap retry or fake success.
- Reviewed screenshots at 1280×720 and 390×844; no horizontal overflow at mobile width. A missing favicon request discovered in the initial browser pass was removed with an inline empty icon.
- Parsed `contracts/examples.json` with Node; `git diff --check` passed. Frontend import inspection found only React and local platform modules/styles. `go list` of kernel imports found standard library and pgxpool only, with no case/assembly dependencies. Both assemblies have intentionally empty registration and cases directories contain ownership guidance only.

## Environment limitation and scope

Docker Desktop 4.55.0 failed before its engine became available, reporting an inaccessible `dockerInference` socket while starting its own inference service. Consequently `docker compose up -d --wait` did not complete here. The Compose definition and mounted migration commands are provided for a working Docker engine; runtime evidence uses the documented native PostgreSQL alternative. No Docker settings or existing PostgreSQL service were changed.

The catalogue is intentionally empty. The five remaining playthrough/action/asset operations are frozen documentation/examples only and return 404; no gameplay, playthrough schema, action receipts, case runtime, protected fixtures, or deployment stack is built. Case/version assembly agreement will be checked when implementations register actual cases. Existing untracked `docs/planning/` files are excluded from this implementation.

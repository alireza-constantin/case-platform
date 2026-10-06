# Ticket #3 — Phone frontend verification

The default application uses the real same-origin API. Phone registers `phone-demo` / `m0-v1` through frontend assembly. The platform handles guest/run selection, `/play/{id}` routing, dedicated case mounting, and an opaque SDK. Messages, gallery navigation, password input, interpretation of outcomes, and completion presentation belong to Phone.

## Independent verification

These browser scenarios explicitly intercept HTTP with synthetic responses matching `contracts/README.md` and `docs/cases/phone-wire.md`. They prove frontend behavior only. They do **not** prove real password validation, protected authorization, guest ownership, transactional replay, database persistence, or browser-restart recovery. Real API/PostgreSQL integration remains required after the independent backend slice.

Run from the repository root using installed Bun, Node, and npx. Keep the Vite command running separately. The private port avoids the primary checkout's services.

```powershell
bun install --cwd app --frozen-lockfile
bun run --cwd app dev -- --port 5183
```

In another shell:

```powershell
bun run --cwd app typecheck
bun run --cwd app build
npx --yes --package @playwright/cli playwright-cli -s=ticket3 open http://localhost:5183 --persistent --profile=output/playwright/ticket3/profile
npx --yes --package @playwright/cli playwright-cli -s=ticket3 run-code --filename app/verification/phone-browser.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket3 run-code --filename app/verification/phone-recovery.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket3 run-code --filename app/verification/phone-conflicts.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket3 close
```

All three scenarios passed. No synthetic responses are imported into `app/src`, and no synthetic answer or image fixture appeared in the production bundle.

| Scenario | Observable evidence |
| --- | --- |
| Messages → gallery | Projected public clue displayed; locked navigation and rejected attempt issue no protected fetch; unlock uses the returned projection; open sends the next base revision; image bytes use a playthrough-scoped API URL; completion follows returned metadata. Request IDs differ across distinct actions. Leaving the case removes its image and revokes its temporary blob URL. |
| Uncertain action and lifecycle | Lost response leaves the gallery locked and makes no automatic retry. Explicit retry retains exactly the original request ID, base revision, tag, and payload. A newer refresh precedes an older replay; newer completion and message projection remain visible. Reload consumes the API projection. Restart requests a fresh identity and shows locked state. Returning to the launcher cleans the mounted root without React lifecycle errors. |
| Creation uncertainty and conflict | Lost creation response disables another creation until own-run reload discovers the synthetic created run. A definite revision conflict offers no uncertain retry; refreshing allows a new action with a new ID and current revision. |

TDD evidence: the first selection/message scenario failed for missing Start; the gallery extension failed for missing Gallery; lifecycle recovery exposed synchronous nested-root unmount errors; the creation extension exposed a blind creation retry. Each was corrected before moving to the next slice.

## Boundary and visual checks

`bun run --cwd app verify:boundary` checks TypeScript imports/re-exports/dynamic imports, rejects platform imports of cases or assembly, and rejects case imports of assembly or other cases. The production build runs it. A temporary forbidden platform → Phone re-export was rejected with exit 1; removing the probe restored a passing build. There are no configured source aliases or computed imports.

Typecheck and production build passed independently of the backend. Frontend platform source contains no Phone IDs, password/gallery semantics, or case imports. Authoritative snapshots are not stored in localStorage/sessionStorage. SDK disposal aborts transport and clears subscribers; Phone's image effects abort fetches and revoke object URLs. Each mount uses a separate host; nested React unmount runs after the outer commit.

Screenshots were inspected at 1200×900 and 390×844. The handset, messages, form labels, focus styles, protected-image presentation, and completion layout remain readable; long Phone content grows vertically. Local screenshots/profile/logs live under ignored `output/playwright/ticket3` and `.playwright-cli`.

No backend, shared neutral contract, real protected content, or actual answer changes are part of this ticket. Phone → Terminal → Phone and real transport authority/persistence checks remain part of later integration once Terminal and the backend are available.

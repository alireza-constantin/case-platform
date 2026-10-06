# Ticket #5 — Terminal frontend verification

Terminal registers `terminal-demo` / `m0-v1` in frontend assembly alongside Phone. It owns its rectangular workstation, command transcript, directory projection, file reader, command/output types, and presentation. It sends arbitrary entered strings as `workstation.command` with `{command:string}` and displays the server's `lines` / `ok` outcome. It contains no operating-system execution or real filesystem access.

The existing neutral SDK, API, mount contract, and platform are unchanged: **zero platform expansion**. Completion comes from snapshot `completed_at`; readiness, cwd, entries, and file availability come from the case projection. Reading a projected file submits the case's simulated `cat` command using its projected name. A successful outcome permits the local reader to request the projected asset through the existing playthrough-scoped SDK. No frontend maintenance sequence or protected file body is bundled.

## Independent browser verification

The following checks explicitly intercept HTTP with synthetic envelopes and synthetic case content. They prove frontend behavior, not actual authority, guest ownership, transactional replay, PostgreSQL persistence, or browser-restart recovery. Real API/system acceptance remains required after both backend and frontend slices are integrated. The test prerequisite names are deliberately synthetic and are unrelated to the server's maintenance sequence.

From the repository root, start the private Vite service in one shell:

```powershell
bun install --cwd app --frozen-lockfile
bun run --cwd app dev -- --port 5185
```

In another shell:

```powershell
bun run --cwd app typecheck
bun run --cwd app build
npx --yes --package @playwright/cli playwright-cli -s=ticket5 open http://localhost:5185 --persistent --profile=output/playwright/ticket5/profile
npx --yes --package @playwright/cli playwright-cli -s=ticket5 run-code --filename app/verification/terminal-browser.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket5 run-code --filename app/verification/terminal-recovery.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket5 run-code --filename app/verification/case-switching.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket5 run-code --filename app/verification/phone-browser.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket5 run-code --filename app/verification/phone-recovery.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket5 run-code --filename app/verification/phone-conflicts.cjs
npx --yes --package @playwright/cli playwright-cli -s=ticket5 close
```

The Phone scenarios now use the opened browser origin and a two-case synthetic catalogue; they still work with ticket #3's original port. All six scenarios passed, as did independent typecheck and production build with the existing frontend import-boundary enforcement.

| Grouped scenario | Observable evidence |
| --- | --- |
| Command → prerequisites → file | Command input reaches the API under the agreed action tag. Handled denial leaves the reader absent and performs no protected fetch. Synthetic prerequisite commands project readiness, followed by projected directory/file availability. Read submits a simulated command with a new request ID and latest revision; successful returned metadata shows completion; protected text arrives through the playthrough asset endpoint. The newest output stays visible inside the console viewport. |
| Uncertain command / conflict | Lost response leaves existing state visible without automatic retries or fake success. Explicit retry retains exactly the original command, base revision, and request ID. A newer refresh precedes an older replay; newer directory/completion survive and older output is not installed. A definite revision conflict requires refresh; a subsequent explicit command uses a new ID and current revision. |
| Phone → Terminal → Phone | Committed projected Phone completion/messages survive switching. Leaving Terminal while its protected file fetch is pending cancels that fetch. No reader/output leaks into Phone; remount starts with clean local navigation/transcript. Returning to Terminal and reloading consume committed projected directory/completion. No React lifecycle errors occur. |
| Phone regression | Existing message/gallery, uncertain retry/newer revision, Restart identity, creation uncertainty, conflict, and temporary image-URL cleanup checks remain green with both registrations present. |

TDD slices began with observable failures: missing Terminal Start, missing projected-file Read, and clipped latest command output. Each was implemented or corrected before the next slice. Recovery guarantees reuse the already-verified neutral SDK rather than a second retry implementation.

## Visual and boundary review

Desktop 1200×900 and mobile 390×844 screenshots were inspected under ignored `output/playwright/ticket5`. Terminal uses a steel-grey workbench, amber monospaced console, projected directory rail, and separate text reader; Phone remains a distinct rounded handset with messages/gallery. Forms have explicit labels, visible keyboard focus, and keyboard-accessible console scrolling. The transcript scrolls to new output and responds to console resizing. Its case-owned ResizeObserver disconnects on unmount; file-reader effects abort fetches and guard late callbacks. Transcript, reader, and entered command are cosmetic in-memory state only.

No platform, server, migrations, neutral contracts, shared case agreement, or main README changed. The production bundle excludes synthetic prerequisite inputs and file content. Backend authority, the actual discoverable maintenance sequence, real protected delivery, durable profile recovery, and full system acceptance remain integration evidence.

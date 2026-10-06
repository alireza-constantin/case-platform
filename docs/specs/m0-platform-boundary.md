# M0: Prove the platform/case boundary with Phone and Terminal demos

## Problem Statement

We need an investigation platform that can host genuinely different case applications. A fixed detective-game template would constrain every case to the same gameplay, state, navigation, or visual structure. Two different skins over one shared mechanic would provide insufficient evidence that the boundary is sound.

The first milestone must prove that the platform provides neutral capabilities while cases own gameplay. Phone Demo and Terminal Demo must coexist locally, persist their independent progress, and enforce protected mechanics on the server without introducing their gameplay concepts into the kernel. Implementation must stay deliberately small.

## Solution

Build a locally runnable, web-first application with exactly two first-party case slices. Phone Demo lets a player browse messages, infer a password, unlock a protected gallery through server validation, resume, and complete the slice by opening the gallery. Terminal Demo lets a player navigate a simulated workstation, perform a small simulated command sequence that changes server-owned workstation state, reveal a protected file, resume, and complete the slice by reading/reaching it. The terminal executes no operating-system commands.

Both use one neutral platform for anonymous guest access, case/version resolution, playthrough lifecycle, persistence, authoritative action dispatch, projected views, and protected asset delivery. They own unrelated frontend structures, action types, state schemas, transition logic, projections, authorization rules, and completion conditions. Application assembly registers them; platform/kernel code never imports their implementations.

Establish a small neutral app-to-server agreement first. Then frontend and backend implementation can proceed on separate branches, using the same wire agreement and small case-owned examples before converging on the real local system. This specification does not start implementation.

## User Stories

1. As a player, I want a launcher showing the two available demos, so that I can choose an investigation experience.
2. As a player, I want anonymous guest access without registration, so that I can begin immediately.
3. As a player, I want the same browser/profile to recognize my guest on a later visit, so that I can resume.
4. As a player, I want separate Phone and Terminal playthroughs, so that progress in one does not change the other.
5. As a player, I want Phone Demo to resemble a simulated phone, so that its interaction is appropriate to its case.
6. As a player, I want to browse Phone messages and public clues, so that I can infer the gallery password.
7. As a player, I want the Phone server to evaluate my password attempt, so that gallery access follows the case's actual rules.
8. As a player, I want a wrong password to leave protected gallery content unavailable, so that the puzzle retains integrity.
9. As a player, I want a correct attempt to unlock protected gallery material, so that I can advance the Phone investigation.
10. As a player, I want the Phone unlock to survive reload and browser restart, so that I do not repeat completed work.
11. As a player, I want opening the protected gallery to contribute to or complete Phone Demo under its own rules, so that its ending matches its gameplay.
12. As a player, I want Terminal Demo to present a simulated workstation and command interface, so that it feels distinct from Phone Demo.
13. As a player, I want to navigate a simulated filesystem, so that I can explore Terminal's case-owned environment.
14. As a player, I want a short simulated command sequence to change server-owned workstation state, so that commands have meaningful consequences.
15. As a player, I want premature protected-file access denied, so that skipping the interface cannot bypass Terminal's rules.
16. As a player, I want the correct workstation state to make its protected file available, so that I can progress.
17. As a player, I want Terminal progress to survive reload and browser restart, so that I can return to the investigation.
18. As a player, I want reading/reaching the protected file to contribute to or complete Terminal Demo under its own rules, so that the ending belongs to that case.
19. As a player, I want each case to own its navigation and visuals, so that different experiences are not forced into shared gameplay UI.
20. As a player, I want switching Phone to Terminal to Phone to clean up the previous case, so that its effects do not interfere with the next one.
21. As a player, I want Continue to resume an existing run, so that I can pick up committed progress.
22. As a player, I want Restart to create a fresh run without overwriting the previous run, so that restarting is nondestructive.
23. As a player, I want first completion to remain recorded even if the case allows further exploration, so that my accomplishment is preserved.
24. As a player, I want a missing development case version explained without silent reinterpretation of my state, so that an incompatible deployment cannot corrupt progress.
25. As a player, I want a restored compatible implementation to make my retained run usable again, so that temporary availability is not permanent invalidation.
26. As a player, I want unrevealed protected material absent from downloaded code and initial responses, so that inspecting the browser does not bypass progression.
27. As a player, I want my private playthrough inaccessible to another guest, so that knowing a URL does not grant access.
28. As a player, I want protected assets authorized for the requested run, so that unlocking another run does not unlock mine.
29. As a player, I want client-side manipulation unable to authorize protected progress, so that the server remains authoritative.
30. As a player, I want lost connectivity to produce clear failure or uncertainty rather than fake success, so that I can trust reported progress.
31. As a player, I want stale mutations rejected without overwriting newer state, so that accidental concurrent tabs cannot destroy progress.
32. As a player, I want an identical retry after a lost successful response to return the committed outcome, so that an action is not applied twice.
33. As a player, I want replayed older results to preserve newer known state, so that retry handling does not roll the interface backward.
34. As a case author, I want to define my own state, actions, projections, protected-content decisions, and completion rules, so that gameplay semantics remain inside the case.
35. As a case author, I want harmless cosmetic state to stay client-side, so that UI interactions do not require unnecessary server round trips.
36. As a case author, I want synchronous server computations over supplied private state and definitions, so that protected mechanics can be implemented without direct infrastructure access.
37. As a developer, I want separate platform, case, and assembly layers on both sides, so that dependency direction is explicit and verifiable.
38. As a developer, I want adding Terminal Demo to change case code and registration without teaching the kernel terminal mechanics, so that the architectural thesis is demonstrated.
39. As a frontend developer, I want a stable neutral wire agreement and representative responses, so that I can proceed without waiting for the entire backend.
40. As a backend developer, I want that same agreement and independent API verification, so that I can proceed without compiling or importing frontend case implementations.
41. As a developer, I want each demo's frontend/backend payload and projection agreement kept case-owned, so that parallel work does not create a universal gameplay schema.
42. As a developer, I want obvious frontend/backend case-version registration mismatches caught during assembly, so that incompatible implementations are not paired.
43. As a developer, I want reproducible local web/API/PostgreSQL startup and migrations, so that another developer can exercise the milestone without cloud infrastructure.
44. As a developer, I want grouped system checks of observable behavior, so that acceptance does not depend on implementation-mirroring tests or a speculative harness.
45. As a developer, I want only capabilities exercised by these demos, so that the milestone remains small enough to prove the boundary.

## Implementation Decisions

### Repository ownership and dependency direction

The following locations are explicit repository-structure decisions supplied by the user:

| Location | Responsibility |
| --- | --- |
| `app/` | Vite/React/TypeScript frontend |
| `app/src/platform` | Neutral frontend platform and thin SDK |
| `app/src/cases` | Case frontend implementations |
| `app/src/assembly` | Frontend implementation imports and registration |
| `server/` | Go backend |
| `server/internal/kernel` | Neutral backend kernel and infrastructure |
| `server/internal/cases` | Case backend implementations |
| `server/internal/assembly` | Backend implementation imports and registration |
| `contracts/` | Minimal neutral app-to-server agreement and representative examples |
| `migrations/` | PostgreSQL migrations |
| `docs/` | Existing glossary, architecture, acceptance criteria, ADRs, and specifications |

Enforce the three-layer boundary independently on each side. Frontend platform code cannot import frontend case implementations. Backend kernel code cannot import backend case implementations. Neither layer branches on case IDs or interprets gameplay. Each side's assembly imports and registers its cases. Cases depend on neutral contracts/capabilities, not each other. Registration lookups may contain opaque case IDs and versions as data.

The neutral contract location is not a shared runtime package. Do not create speculative utility packages, cross-language runtime dependencies, generated SDK infrastructure, or a shared union of Phone/Terminal actions and views. Both sides implement their language-specific envelope types against the same documented agreement. Bun is the working frontend tooling direction; precise versions and subordinate package layout are ordinary implementation choices.

### Case execution and frontend lifecycle

Use trusted first-party code in one Vite application/browser document and one Go backend process. Case frontends receive a dedicated root and own their whole gameplay surface, navigation, visuals, styles, and local state approach. The platform owns selection/loading, guest/playthrough context, SDK construction, and mounting/unmounting. Case-created listeners, timers, subscriptions, observers, temporary DOM changes, and other effects must be cleaned up. Prefer scoped styles where practical.

The platform owns the playthrough route prefix and the case owns its suffix. Cases may use URL routing or local navigation; bookmarkability of every internal view is not required. Routes never authorize secret content. The thin SDK transports projected views, case-owned actions/outcomes, and protected asset access; it imposes no gameplay components, routing library, or global gameplay store.

Backend case modules initialize and validate their state, synchronously handle actions, project player-visible data, authorize protected assets, and declare completion. These are neutral responsibilities; exact Go callback signatures are subordinate implementation details. Handlers compute over supplied state and server-only definitions without raw database access, arbitrary SQL, external HTTP calls, filesystem writes, background work, queues, or asynchronous side effects.

### Minimal neutral app-to-server agreement

Before parallel implementation begins, record the following small HTTP/JSON agreement and representative success/error/replay examples in the neutral contract documentation. The routes and field names below are concrete M0 defaults, not a new framework. This baseline is sufficient for independent frontend and backend work.

| Operation | Default HTTP surface | Agreement |
| --- | --- | --- |
| Establish/reuse guest | POST `/api/guest` | Empty JSON request; reuse a valid guest cookie or establish one; return 204. Never return the cookie credential in JSON. |
| List launchable cases | GET `/api/cases` | Return a `cases` array containing safe `case_id`, `case_version`, and `title` metadata. Public covers may be added only if used by the launcher. No private definitions or protected inventories. |
| List own runs | GET `/api/playthroughs` | Return a `playthroughs` array of current guest-owned summaries, newest created first, for Continue/Restart selection. No private case state. |
| Create/restart run | POST `/api/playthroughs` | Accept `case_id` and `case_version`; server initializes fresh case state. Return 201 with the new snapshot. Restart uses creation and does not overwrite the old run. |
| Read current view | GET `/api/playthroughs/{playthrough_id}` | Return a projected snapshot for an owned, available pinned implementation. |
| Submit case action | POST `/api/playthroughs/{playthrough_id}/actions` | Accept `request_id`, `base_revision`, `action_type`, and opaque JSON `payload`; return the safe action result or a neutral error. |
| Fetch protected asset | GET `/api/playthroughs/{playthrough_id}/assets/{asset_id}` | Verify guest ownership, pinned case/version, and case authorization; deliver bytes with their content type and no-store caching. Asset IDs are opaque identifiers, not raw storage paths. |

A playthrough summary contains `playthrough_id`, `case_id`, `case_version`, `revision`, nullable `completed_at`, `created_at`, `updated_at`, and `availability` (`available` or `unavailable`). IDs/versions are opaque strings; revision is a nonnegative integer; timestamps are serialized consistently as ISO-8601 timestamps. A snapshot contains `playthrough` summary metadata and opaque JSON `view`. An action result contains that safe snapshot plus `request_id` and opaque JSON `outcome`. Views/outcomes/payloads have no shared gameplay fields or cross-case type union. Stored private state and server-only definitions are not wire fields.

Use a small error envelope with `error.code` and a safe `error.message`. Baseline categories are invalid request (400), missing guest (401), unavailable/not-owned resource (404), stale revision or committed request-ID reuse conflict (409), unavailable pinned case version (409), case action validation rejection (422), and unexpected server failure (500). Neutral codes distinguish revision conflict, committed ID reuse, and unavailable version. A wrong puzzle answer can be a successfully handled case outcome rather than a transport error; the case decides its meaning. Errors never reveal private state or internal stack traces. Stale callers fetch the current view and submit a reconciled action under a new request ID; there is no automatic gameplay retry after a conflict.

The playthrough identifier in the route is authoritative; action bodies cannot select another case/version or assign private state. Bootstrap and mutations use JSON, non-GET methods, cookie authentication where needed, and simple Origin validation. Public assets may use normal frontend/static delivery; protected assets must not be served from a public location.

For retries, retain the original logical action's request ID and input while its result is uncertain. A returned revision describes the returned snapshot/outcome, not necessarily the latest server state if the response is a replay. Never install that snapshot over a newer known revision. The SDK and case UI can fetch current state to reconcile. Do not introduce an offline action queue or automatic retry framework. Guest/run creation does not need a new distributed idempotency system; avoid blind automatic retries and use the own-run list when reconciling uncertain creation.

Each demo's action tags, payloads, views, and outcomes need a small case-owned agreement or representative examples for its frontend/backend implementers. Phone and Terminal can use different shapes. Coordinate these examples as part of their slices; do not extend the neutral agreement with case fields or move case schemas into shared contract infrastructure.

### Persistence, concurrency, and completion

Use PostgreSQL with a platform-owned playthrough envelope and case-owned JSONB state. Store only the neutral metadata needed for guest ownership, case/version pin, revision, creation/update times, and first completion; avoid redundant generic gameplay status. Cases own initialization, state validation/transitions, and player-visible projection. Persistence, authority, and secrecy are independent properties, and cosmetic UI state need not be server-backed.

An ordinary database transaction coordinates a state-changing action, revision check, first completion if any, relevant case events, and committed outcome/input identity. Serialize or otherwise coordinate concurrent transitions so the same base revision cannot be overwritten twice. Check committed-request identity before rejecting an identical successful retry as stale. Concurrent duplicates must not commit twice. Exact SQL/locking choices remain small backend implementation details.

Advance revision once for a durable authoritative playthrough change, including first completion. Changing the case document and first completion together is one transition. No durable change or operational logging alone does not require advancement. A wrong password advances revision only if the case persists a counter or other durable change.

Persist request ID, input identity, committed revision, and safe outcome atomically with successful authoritative mutations. Identical retries return that recorded outcome without repeating state/events/completion. Different input under an already committed ID is rejected. Malformed, unauthorized, stale, and ordinary pre-commit failures do not require permanent receipt records. Automatic cleanup and external-side-effect deduplication are outside M0.

Completion is case-declared monotonic metadata. Record its first timestamp in the same commit; never clear it or impose a universal post-completion action lock. Relevant emitted case events may be stored with commits, while ordinary operational logs remain observational. Do not create event sourcing, a broker, subscriptions, or a product analytics service.

### Authority, assets, and anonymous access

Treat the browser as hostile and case authors as trusted. Protected progression submits attempts evaluated by case server logic; client-supplied unlock flags are not authoritative. Unrevealed correct-answer data, hidden material, private definitions, and protected content must be absent from frontend bundles, public assets, and initial views. Public clues may intentionally enable inference.

Use a persistent opaque guest-session cookie. In deployed HTTPS it has HttpOnly, Secure, and SameSite=Lax or an appropriately stricter setting; local HTTP configuration must remain usable. Guest ownership is checked for reads, mutations, and protected assets. Playthrough IDs/URLs are not bearer credentials. Keep simple same-origin mutation controls and reasonable request bounds; do not add rotating-token or anti-automation infrastructure.

Protected delivery checks guest ownership and the run's pinned case/version, then delegates asset authorization to the case. A copied URL does not grant another guest access; a locked second run cannot inherit an unlocked run's permission. A guest may still fetch its own legitimately unlocked run. Serve protected content through the application server with no-store and exclude it from service-worker caches. Do not log raw secret-bearing inputs/private state by default.

If connectivity is lost, protected actions fail or remain uncertain cleanly and never report fake success. Already-loaded non-sensitive UI may remain visible. A response lost after a commit is reconciled through replay/current-state retrieval. There is no offline validation, authoritative action queue, or synchronization.

### Compatibility and run lifecycle

Each playthrough's case ID/version pin stays fixed. The version indicates compatibility of state and authoritative interpretation. Incompatible state, action semantics, passwords/solutions, rules, or protected progression require a new version; compatible typo, CSS, wording, and image-quality changes do not. Developer discipline suffices for M0; do not hash all artifacts.

Check frontend/backend registration agreement at assembled build/startup or integration. Independent side builds must not require compiling the other's case runtime. A missing pinned implementation blocks execution without deleting state or permanently changing lifecycle. Restoring the implementation can make the run available again. Development may use explicit restart instead of retaining every version. The launcher can offer Continue and Restart without a save-history interface.

### Contract-first work sequence and parallel tracks

1. **Contract gate:** record the minimal neutral messages and representative examples; confirm both implementers can build their side independently; keep each demo's semantic examples case-owned. Adopt the existing system verification seam. No application code needs to be built before this agreement is settled.
2. **Frontend track:** establish the Vite/React application, platform launcher/SDK/mount lifecycle, frontend assembly, and two distinct case UIs. Use lightweight local development responses matching the agreed envelopes and case-owned examples while the backend is unfinished. These responses are development aids, not an authoritative gameplay implementation.
3. **Backend track:** establish the Go service, PostgreSQL migrations, guest/run access, persistence/actions/replay/assets, backend assembly, and two separate synchronous case modules. Verify API behavior independently of frontend runtime imports.
4. **Integration:** connect the real API, compare assembly registrations, exercise both complete slices, and run grouped system acceptance. Remove any reliance on mocked authority from final acceptance. Document reproducible one-origin local startup.

After the contract gate, the two tracks may branch and proceed in parallel. Changes to neutral messages are coordinated explicitly; case-owned gameplay evolution remains local to the relevant case. This is a work plan, not a requirement for new shared packages, multiple services, or immediate branch creation.

## Testing Decisions

Retain the existing approved high-level seam: browser behavior through the real Go API and PostgreSQL. Direct requests at the same API surface exercise tampering, ownership, stale revisions, and uncertain-delivery replay. Independently review frontend-platform and backend-kernel imports/registration boundaries. Contract examples support parallel development but do not substitute for the real system in acceptance.

Tests assert external behavior and durable outcomes rather than private functions, exact SQL, component structure, or implementation-mirroring snapshots. There is no existing application test suite for prior art; use the established M0 acceptance criteria as the behavioral baseline. Group related criteria into a small number of system scenarios rather than building a separate test per row or a specialized harness. Use focused case-transition tests only where actual rules warrant them.

The modules covered are frontend platform/SDK/assembly, both case frontends, backend kernel/assembly, both case modules, persistence and guest access, protected delivery, and neutral protocol behavior. The required acceptance evidence is:

1. A clean clone can run web, Go API, and PostgreSQL through one local browser origin without cloud credentials.
2. Both guest-owned case slices run with genuinely different UIs, navigation, actions, state, transitions, views, authorization, and completion conditions.
3. Neither frontend platform nor backend kernel imports case implementations or interprets Phone/Terminal semantics; assemblies register the cases.
4. Phone messages support password inference; case-server validation unlocks its protected gallery; opening it contributes to/completes the slice.
5. Terminal's small simulated command sequence changes server workstation state; premature file access is denied; reaching/reading the protected file contributes to/completes the slice without real shell execution.
6. Both runs resume committed progress/completion after reload and browser restart in the same profile. Restart creates a new identity and preserves the old run.
7. Phone-to-Terminal-to-Phone switching unmounts/cleans effects without losing legitimate committed state or retaining accidental transient interference.
8. Initial bundles/views/public assets withhold unrevealed secret material; raw client state manipulation cannot unlock protected progression.
9. Another guest cannot read, mutate, or fetch protected assets from a known run. Locked runs do not inherit another run's unlock, and copied URLs are not credentials.
10. Two distinct writes from one base revision cannot silently overwrite one another. Malformed, unauthorized, stale, and pre-commit rejected requests do not mutate progress.
11. A committed request retried with identical ID/input returns its stored outcome without another transition/event/revision. Concurrent identical submissions cannot commit twice; changed input under a committed ID fails.
12. No-op/log-only processing need not advance revision; an actual persisted attempt-counter change does. State plus first completion advances once; first completion alone is also a durable revisioned change.
13. An older replay never rolls newer known frontend state backward. Lost connectivity produces no fake success or offline queued authority.
14. Completion remains recorded while cases control post-completion gameplay. Obvious registration mismatch fails assembly, and temporarily absent versions retain state and can become available when restored.

Builds and tests may run in CI. Production deployment and new observability/testing systems are not acceptance requirements.

## Out of Scope

Third-party authors/uploads, runtime plugins, hot loading, module federation, WASM, process/document sandboxing, independently deployed case applications, speculative shared packages, generic phone/terminal/filesystem/detective frameworks, reusable gameplay engines, AI, realtime/multiplayer, payments/entitlements, accounts/cross-device recovery, offline gameplay/synchronization, product analytics, achievements/notifications/CMS/creator tools, custom case SQL, queues/external side effects, generic secret-management or workflow products, DRM/solution-sharing prevention, artifact hashing/content-addressed releases, and production hosting/domains/TLS automation/CDNs/deployment pipelines/autoscaling.

PWA installability remains independent of offline capability and is not required to prove this browser-based boundary. Long-term public-case retention/migration promises, save-history UI, and automated lifecycle cleanup are deferred.

## Further Notes

The existing M0 architecture, acceptance criteria, ADRs, glossary, and the user's explicit repository-layout decision are the source of truth. Repository paths above are included because that layout is an explicit user requirement; subordinate implementation paths and code snippets are intentionally unspecified.

There is no unresolved expensive architectural dilemma requiring another interview. Exact fictional content/password, simulated command vocabulary, case-local field names, tooling versions, styling details, callback signatures, and local ports should use the smallest choices that demonstrate the two slices. Keep unrevealed real demo answers out of public contract examples.

Before the first public case, revisit historical-version retention, migration, or expiration policies. M0's development restart concession is not a production compatibility promise. Hosted external testing is a separate later milestone.

Publishing this specification is the current deliverable. It does not start implementation, create feature branches, or commission frontend/backend work. The parallel tracks become actionable after the minimal contract agreement is established.

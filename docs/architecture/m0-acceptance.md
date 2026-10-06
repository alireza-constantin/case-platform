# M0 acceptance criteria

Status: review criteria, not executed test results. Implementation has not started. These criteria verify the decisions in [M0 architecture](m0-architecture.md) and the exact slices in [M0 scope](m0-scope.md).

## Primary verification boundary

Prefer browser-driven behavior through the real Go API and PostgreSQL. Use direct requests at that same API boundary for client tampering, access checks, stale revisions, and lost-response retries. Dependency review separately verifies the kernel/case import boundary. Add smaller case-transition tests only where they provide useful coverage of actual case rules; do not mirror implementation or invent testing subsystems.

There is no existing application/test suite to imitate. Tool selection and exact callback signatures belong to specification/implementation. Assertions should concern observable behavior and durable outcomes, not private helper names, SQL shapes, or component structure.

The rows below are review criteria, not a mandate for a separate automated test per row. Cover related guarantees with a small number of system scenarios and focused dependency review; introduce no specialized testing subsystem.

## Architectural proof

| ID | Scenario | Required evidence |
| --- | --- | --- |
| B1 | Inspect frontend platform and backend kernel dependencies/dispatch independently | Neither imports its case implementations, branches on Phone/Terminal IDs, or interprets gameplay. Each side's assembly imports/registers cases. Generic registry lookups are permitted. |
| B2 | Register and run Terminal Demo alongside Phone Demo | Terminal code and application assembly can supply it through the existing neutral contracts without adding terminal concepts to the kernel. |
| B3 | Compare the two case implementations | Their action types, state schemas, transitions, projections, protected-content decisions, completion rules, frontend structure, and navigation differ. Neither is implemented through a generic phone/terminal engine. |
| B4 | Exercise the SDK and persistence through both cases | Case-owned payloads/views/state pass through neutral infrastructure without a common gameplay schema or case-specific SDK operation. |
| B5 | Review the repository layout and integration contract | The agreed app/server/contracts/migrations/docs roots and frontend/backend layer locations are preserved. The minimal neutral contract permits separate side builds and coordinated parallel work without runtime imports across sides. |

## Local setup and lifecycle

| ID | Scenario | Expected behavior |
| --- | --- | --- |
| L1 | Clone and follow documented startup instructions | Web, Go API, and PostgreSQL start reproducibly; both demos run through one local browser origin. No production hosting or cloud credentials are required. |
| L2 | Start as a fresh browser guest | The application establishes an anonymous guest and can create separate Phone and Terminal playthroughs. |
| L3 | Reload and restart the browser in the same profile | Each run resumes its committed progress and completion metadata. Authoritative state is loaded from the server, not trusted localStorage. |
| L4 | Restart a demo | A fresh playthrough ID and initial state are created; the previous run remains stored. |
| L5 | Switch Phone → Terminal → Phone | Prior case trees unmount; listeners, timers, observers, subscriptions, and transient UI effects do not interfere with the next case. Returning restores committed state while accidental transient carryover is absent. |

## Phone Demo slice

| ID | Scenario | Expected behavior |
| --- | --- | --- |
| P1 | Browse the simulated phone/messages | The Phone case presents its own interaction surface and public clues without requiring phone concepts in the kernel. |
| P2 | Submit an incorrect password or forge an unlock flag | Phone server rules determine the outcome; client manipulation cannot authorize gallery access. Revision follows whether a durable attempt counter or other state actually changes. |
| P3 | Submit the correct password | The Phone server commits gallery access; authorized protected content becomes available. |
| P4 | Reload/restart, then open the gallery | The unlock persists; opening the gallery contributes to/completes the case under Phone-owned rules. Completion persists. |

## Terminal Demo slice

| ID | Scenario | Expected behavior |
| --- | --- | --- |
| T1 | Navigate the simulated workstation/filesystem | Terminal provides distinct command-based behavior; it executes no actual operating-system shell. |
| T2 | Attempt to reach protected content before prerequisites | Terminal server rules deny premature access, including direct API attempts that bypass the UI. |
| T3 | Perform the agreed small command sequence | Case-owned server transitions mutate workstation state and make the protected file available. Kernel code neither parses commands nor interprets directories/files/shell state. |
| T4 | Reload/restart, then read/reach the protected file | Workstation progress persists and Terminal-owned rules contribute to/complete the run. |

## Authority and disclosure

| ID | Scenario | Expected behavior |
| --- | --- | --- |
| S1 | Inspect initial frontend bundles, public assets, and API views for both cases | They contain no unrevealed protected content or server-only correct-answer/configuration data. Public clues may intentionally support inference. |
| S2 | Fetch a protected asset before/after unlock, copy its URL into a different guest's browser, and request the same asset through a locked second run's route | Only authorized guest/playthrough access succeeds. The different guest cannot use the copied URL, and the locked run does not inherit the first run's unlock. A guest may still access its own legitimately unlocked run. |
| S3 | Use a different guest to read, mutate, or fetch assets from another guest's run | Ownership is enforced by the server for all three operations. Knowing the playthrough ID grants no access. |
| S4 | Mutate local state, bypass UI controls, or navigate directly to a hidden view | Server progression and disclosure rules still hold. Raw private state is never the generic client response. |
| S5 | Lose connectivity during a protected interaction | The UI shows failure/uncertainty rather than fake success; authoritative actions are not queued for offline execution. Reconciliation handles a response lost after a commit. |
| S6 | Attempt a clearly cross-origin mutation | Simple Origin/method/JSON protections reject it without introducing elaborate CSRF infrastructure. |

## State and delivery guarantees

| ID | Scenario | Expected behavior |
| --- | --- | --- |
| C1 | Send two distinct mutations based on the same revision | At most one state-changing transition commits from that base; a stale request never silently overwrites the newer state. |
| C2 | Commit an action, discard its response, and retry with identical ID/input | The recorded outcome is returned without repeating the transition, durable events, or revision advancement. Concurrent identical submissions cannot commit twice. |
| C3 | Reuse a committed request ID with different input | The request is rejected without changing state. |
| C4 | Submit malformed, unauthorized, stale, or ordinary pre-commit invalid input | There is no durable playthrough mutation or revision advancement and no requirement to persist a failure receipt. |
| C5 | Process a no-op or write only an operational log | The revision need not advance. A persisted wrong-attempt counter is a real state change and does advance it. |
| C6 | Change case state and record first completion in one action | Case state, completed_at, outcome, relevant events, and revision commit together. The revision advances exactly once. |
| C7 | Commit first completion without changing the case document | The envelope's durable completion change advances the revision once and is protected by committed-mutation replay. This is a contract regression check, not an extra gameplay mechanic. |
| C8 | Replay an older committed response after newer state is known | The frontend keeps newer state and fetches the current view if needed; it does not roll back its visible snapshot. |
| C9 | Continue after completion | completed_at remains the first completion time. Case rules determine which further actions are allowed. |

## Compatibility and availability

| ID | Scenario | Expected behavior |
| --- | --- | --- |
| V1 | Assemble obviously mismatched frontend/backend case/version registrations | Build/startup fails clearly rather than pairing incompatible registrations. |
| V2 | Remove a pinned development implementation | Its existing run cannot silently execute against a different version; state and lifecycle are retained. Explicit restart creates a new run. |
| V3 | Restore the compatible pinned implementation | The retained run can become available again. |
| V4 | Review a cosmetic change and an incompatible gameplay change | Cosmetic edits may keep the compatibility version; incompatible authoritative interpretation/state requires a new version. No artifact-hashing system is required. |

## Review exit and next step

M0 succeeds when both slices and the observable guarantees pass, and dependency review confirms the ownership boundary. It does not require hosted infrastructure, product analytics, generic gameplay packages, offline play, multiplayer, or a sophisticated security/compatibility/workflow framework.

The next specification should describe these existing system seams and tests, then obtain review of their fit before publication. It should turn exact local commands, protocol shapes, and case-specific details into implementable work without expanding capability scope. This review document neither publishes a tracker issue nor authorizes implementation.

# M0 architecture closeout

Status: M0 complete and merged, recorded on 2026-10-07. [PR #8](https://github.com/alireza-constantin/case-platform/pull/8) delivered bootstrap and the neutral contract; [PR #9](https://github.com/alireza-constantin/case-platform/pull/9) delivered both cases and assembled verification. Issues [#1](https://github.com/alireza-constantin/case-platform/issues/1), [#2](https://github.com/alireza-constantin/case-platform/issues/2), [#3](https://github.com/alireza-constantin/case-platform/issues/3), [#4](https://github.com/alireza-constantin/case-platform/issues/4), [#5](https://github.com/alireza-constantin/case-platform/issues/5), [#6](https://github.com/alireza-constantin/case-platform/issues/6), and [#7](https://github.com/alireza-constantin/case-platform/issues/7) are closed.

This records the outcome of the [architecture](m0-architecture.md), [scope](m0-scope.md), [acceptance criteria](m0-acceptance.md), and [ADRs 0001–0009](../adr/). Those documents retain their historical design content. This closeout records M1 direction only; it publishes no M1 issue and begins no implementation.

## Implemented and verified

The local Vite/React/TypeScript application, Go API, and PostgreSQL implement all seven [neutral wire operations](../../contracts/README.md), persistent anonymous guest access, a two-case launcher, Continue, and nondestructive Restart.

- **Phone:** case-owned messages and navigation lead to a server-validated gallery password attempt, protected image delivery, and first completion when the gallery is opened.
- **Terminal:** a distinct simulated workstation owns command interpretation, directory/prerequisite state, protected report authorization, and first completion when the report is reached/read. It executes no real shell commands.
- **Lifecycle:** both committed runs survive reload, same-profile browser restart, and API restart. Switching Phone → Terminal → Phone cleans transient effects and pending delivery; Restart retains prior runs.
- **Authority and disclosure:** real API/browser checks cover withheld answers/content, forged client-state denial, cross-guest read/action/asset isolation, locked-run and copied-URL denial, malformed requests, and same-origin mutation controls.
- **Durable outcomes:** PostgreSQL checks cover stale-write rejection, identical and concurrent committed replay, changed-input conflicts, no-ops, atomic state/events/receipts/completion, first completion alone advancing once, monotonic completion, and removed/restored pinned implementations. Frontend checks preserve newer views over older replay snapshots and show uncertainty without fake success or offline queues.

Final acceptance used the real browser → Go → PostgreSQL path, with independent builds, frontend import enforcement, backend dependency review, bundle disclosure checks, and registration mismatch checks passing. The [system record](../verification/ticket-7-system.md) and [contract regressions](../verification/ticket-7-contract.md) contain commands and results. Earlier [bootstrap](../verification/issue-2-bootstrap.md), [Phone frontend](../verification/ticket-3.md), [Phone API](../verification/ticket-4.md), [Terminal frontend](../verification/ticket-5.md), and [Terminal API](../verification/ticket-6.md) records distinguish synthetic frontend checks from real authority/persistence evidence. These are recorded M0 results, not a new acceptance run for this documentation change.

## Architectural result and established defaults

**The thesis was validated within M0's two trusted, local, single-player slices: the platform provides capabilities; cases provide gameplay.** Phone and Terminal have different UIs, actions, private schemas, transitions, projections, permissions, and completion rules while sharing neutral transport and persistence.

Treat these seams as established defaults for M1:

| Seam | Default |
| --- | --- |
| Dependency and assembly | Platform/kernel imports no case implementations and interprets no gameplay. Each side's assembly imports/registers cases; case semantics remain case-owned. |
| Frontend runtime | Launcher/playthrough identity and mount lifecycle are platform-owned. Cases own the viewport, navigation, styles, cosmetic state, and effect cleanup. The thin SDK transports opaque actions/views/outcomes and authorized assets. |
| Backend execution | The existing `kernel.Case` interface supplies metadata, initialization, projection, synchronous actions, and asset permission/content. Callbacks receive supplied state, with no raw database or external-side-effect capability. |
| Persistence and authority | A neutral envelope plus case-owned JSONB, explicit safe projection, revisions, atomic committed receipts/events, nondestructive restart, and case-declared first completion. |
| Access and compatibility | Single-origin guest cookies; owned-playthrough asset authorization with no-store; fixed compatibility pins, reversible availability, and frontend/backend registration agreement. |

## Terminal reuse evidence

The [Terminal frontend commit](https://github.com/alireza-constantin/case-platform/commit/65c526f022c2f3f17bcecf4d5a51cf6b57db79bb) adds its case, assembly registration, and verification without changing Phone's frontend platform, SDK, or mount contract. The [Terminal backend commit](https://github.com/alireza-constantin/case-platform/commit/6a13d47a2e9ddb75aa358e2c12e505a0d72c000b) adds its module, registration, tests, and case-owned notes without changing kernel contracts, migrations, or the neutral wire protocol. Both reuse opaque envelopes and the same playthrough-scoped protected delivery; Phone regressions remained green.

Integration subsequently added neutral launcher registration validation and corrected the PostgreSQL 18 Compose volume path. The contract documentation's implementation-status paragraph changed; its protocol did not. These integration changes introduced no Terminal-specific platform capability.

## Deferred decisions and limits of the proof

Historical implementation/asset retention, case-authored migration, or expiration/restart guarantees remain deliberately unsettled and must be chosen before the first public case ([ADR-0006](../adr/0006-immutable-version-pins-with-development-restarts.md)). Accounts/cross-device recovery, offline play, save-history UI, automatic cleanup, independent deployment, untrusted authors/plugins, external side effects, and reusable gameplay systems remain deferred.

M0 did not prove a substantial investigation's depth, pacing, multi-surface coherence, or player engagement. It did not prove production scale/operations, hostile-code isolation, public save longevity, multiplayer, AI, payments, or creator workflows. Shared browser/process risks remain accepted for trusted code. Compose configuration validated, but container startup was not exercised because the host Docker engine failed; live acceptance used native PostgreSQL.

## Direction for M1

M1 should be the first substantial real investigation, combining multiple case-owned interaction surfaces inside one case. Prefer that case over another architecture demo. Do not extract generic phone, map, document, CCTV, or evidence systems merely because M1 uses them; new platform capabilities require a concrete M1 need. AI, multiplayer, payments, creator tooling, and production-scale infrastructure remain deferred unless separately chosen as the explicit next milestone. A detailed M1 specification remains future work.

# M0 boundary milestone

Prove the platform/case boundary with exactly two slices: Phone Demo and Terminal Demo. The two experiences must look and behave differently. Adding Terminal Demo may extend application composition and case code, but must not teach the platform kernel terminal-specific concepts. Neither demo is a reusable gameplay engine.

## Agreed gameplay slices

- **Phone Demo:** browse a simulated phone and messages, infer a password, submit a case-server-validated attempt, unlock a protected gallery, resume that progress after reload/browser restart, and open the gallery to contribute to or complete the demo.
- **Terminal Demo:** navigate a simulated workstation/filesystem, run a small simulated command sequence that changes server-owned workstation state, make a protected file available, resume that progress after reload/browser restart, and read/reach the file to contribute to or complete the demo. It executes no real system shell commands.

Each owns its action types, state schema, server transitions, player-visible projection, protected-content decisions, completion conditions, frontend structure, and navigation. They share neutral platform contracts.

## Accepted capability scope

- Anonymous guest identity.
- Explicit case/compatibility-version registration.
- Case frontend mounting and unmounting.
- Registered case-specific Go backend modules.
- Playthrough creation and same-browser resume.
- Case-owned persisted state and player-visible projection.
- Optimistic revisions and idempotent authoritative actions.
- Public/protected asset delivery exercised by the demos.
- Basic playthrough events and operational logging.
- Nondestructive restart and case-declared monotonic completion.

These are responsibilities, not a requirement for separate services. Persistence uses PostgreSQL JSONB. Runtime, state, dependency, lifecycle, and single-origin authentication guarantees are recorded in ADR-0001 through ADR-0009. The agreed model is consolidated in [M0 architecture](m0-architecture.md), with review criteria in [M0 acceptance](m0-acceptance.md). Implementation is not authorized.

## Required execution environment

A developer can clone the repository and reproducibly run the web application, Go API, and PostgreSQL, exercising both demos through one local browser origin. Exact ports and local proxy arrangements are implementation details. Hosting, domains, TLS automation, cloud infrastructure, deployment pipelines, CDN setup, production observability, and autoscaling are outside M0. CI may verify builds/tests; hosted external testing is a later milestone.

## Simplicity rule

Prefer the smallest implementation that proves the two demos and their ownership boundary. Defer mechanisms motivated primarily by hypothetical future cases. Strongly preserve case-neutral kernel dependencies, case-owned gameplay semantics and state, server authority where required, non-disclosure of unrevealed private material, and compatibility-version discipline.

Use ordinary transactional persistence rather than a workflow engine, queues, or event-sourced reconstruction. Operational logging should not record raw secret-bearing action inputs or private state by default. Protected-asset inventories and storage locations are not public manifests. These are minimal implementation assumptions, not additional subsystems.

## Deferred

Product analytics, AI gateways, realtime, multiplayer, payments, entitlements, achievements, notifications, creator tooling, CMS, a separate gameplay-secret product, independent case deployment, and marketplace/plugin infrastructure. Add further capabilities only when a concrete case exercises them.

## Gate before first public case

Resolve production version retention/migration/expiration guarantees (ADR-0006). Development restarts are an M0 concession, not a production save-compatibility promise.

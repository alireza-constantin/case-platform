# Compatibility version pins with explicit development restarts

Each playthrough's case ID and compatibility-version pin remain fixed from creation; registration and routing use that identity. Changes to state expectations, action semantics, passwords/solutions, authoritative server behavior, or protected progression rules require a new compatibility version. Compatible typo, wording, CSS, image-quality, and other cosmetic changes do not. A playthrough must never silently reinterpret its authoritative state under incompatible rules.

Frontend and backend registration declare the same case ID/version, and build/startup must reject obvious registration mismatches. Developer discipline enforces compatibility changes in M0; artifact hashing, content-addressed releases, automated byte-difference detection, and sophisticated compatibility declarations are deferred. An accidentally missed compatibility-version bump is a development defect.

M0 does not guarantee retention of every old development implementation. An unavailable development version blocks execution and may require an explicit restart. Availability is separate from lifecycle: removal of an implementation does not delete or permanently invalidate its playthrough data, and restoring it may make that playthrough runnable again. Restart behavior is specified in ADR-0008.

Before publishing the first real public case, choose and document historical implementation/asset retention, case-authored migration, or an expiration/restart policy. Paid/public cases are expected to need stronger retention, but this operational obligation is deferred beyond M0.

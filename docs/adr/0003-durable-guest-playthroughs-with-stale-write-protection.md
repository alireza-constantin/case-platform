# Durable guest playthroughs with stale-write protection

The first milestone uses anonymous guest identity and server-backed playthroughs. Authoritative progress must survive reloads, browser restarts, and returning later in the same browser/profile. Registration, recovery after clearing browser storage, and cross-device recovery are deferred.

Authoritative mutations use optimistic revisions: stale writes must be rejected or reconciled rather than silently overwrite newer progress. Revisions track committed durable playthrough state rather than every processed action. Request-level retry deduplication is specified in ADR-0005. Sophisticated collaborative reconciliation is deferred.

The two demos are online and single-player. On connection loss, protected actions pause or fail cleanly without reporting fake success; already-loaded non-sensitive UI may remain visible. Offline gameplay, queued authoritative actions, offline synchronization, realtime, and multiplayer are deferred. PWA installability is a separate concern, and the platform model should avoid unnecessary permanent restrictions on future participant counts.

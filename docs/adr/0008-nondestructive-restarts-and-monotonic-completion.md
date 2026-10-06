# Nondestructive restarts and monotonic completion

Restart creates a new playthrough ID and fresh case state; the old run remains. The data model permits multiple playthroughs per guest/case. The initial launcher may offer only Continue and Restart; save-slot and history interfaces are deferred. Case-version availability is separately resolved from current registration and must not permanently mutate lifecycle merely because an implementation is temporarily absent.

Completion is a durable, monotonic milestone declared by case server logic. The platform records the first completion time and may display it, but does not impose a generic lock on later actions. Cases decide whether post-completion exploration or other gameplay is allowed. M0 does not reverse completion or model multiple endings.

The revision covers durable authoritative playthrough state as a whole, including first completion. An action that changes both case state and first completion advances the revision once. Resulting case state, first completion time, revision, committed request outcome, and relevant case events are persisted atomically. Operational logging alone does not advance the revision, and an action with no durable playthrough change need not advance it. No separate completion subsystem is required.

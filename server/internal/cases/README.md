# Case backends

`phone/` and `terminal/` implement the two M0 cases through the existing neutral `kernel.Case` interface. Only `internal/assembly` imports/registers them; `internal/kernel` imports no case implementations and interprets no gameplay. Each module owns initialization, private state, projection, synchronous transitions, asset permission/content and completion decisions. Case callbacks have no database, OS shell, filesystem, external HTTP or background-work capability.

Case-owned handoff: [Phone](../../../docs/cases/phone-wire.md), [Terminal](../../../docs/cases/terminal-wire.md). These documents can be read without importing or compiling frontend code.

# Investigation Platform

Shared language for a platform that provides capabilities and cases that provide gameplay.

## Language

**Case**:
An independently designed interactive investigation experience with its own gameplay and presentation.
_Avoid_: Content pack, detective-game template

**Platform kernel**:
The shared foundation that provides capabilities usable across cases, while each case owns its gameplay concepts.
_Avoid_: Detective-game engine

**Guest**:
An anonymous participant recognized by the platform without a registered account.

**Playthrough**:
A particular run of a case, with its own progress and participant access.
_Avoid_: Case definition

**Case version**:
A compatibility identifier for a case's state expectations and authoritative gameplay interpretation. Compatible cosmetic or editorial changes may share a version.
_Avoid_: Artifact hash, deployment

**Protected progression**:
Progress that a case permits only after validating a required condition.

**Completion**:
A case-declared milestone indicating that a playthrough has first been completed. It remains a historical fact even when that case permits further play.

**Public asset**:
Case material safe to disclose without access to a particular playthrough.

**Protected asset**:
Case material whose disclosure depends on access to a particular playthrough and permission granted by that case.

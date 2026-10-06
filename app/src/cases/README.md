# Case frontends

`phone/` and `terminal/` implement the two launchable M0 cases. Only `src/assembly` imports/registers them. `src/platform` owns neutral selection, guest/playthrough context, SDK transport and dedicated mounting; it imports no case code and interprets no gameplay. The frontend build enforces the import boundary.

Case-owned handoff: [Phone](../../../docs/cases/phone-wire.md), [Terminal](../../../docs/cases/terminal-wire.md). Each case owns its viewport, scoped styles, navigation, view/outcome interpretation and effect cleanup. Protected content and actual solutions remain in the server; synthetic verification files are outside application source and excluded from the build.

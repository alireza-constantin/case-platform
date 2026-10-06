# Terminal-owned M0 examples

Gate vocabulary for the Terminal implementers; no gameplay is implemented in #2. Register `terminal-demo` / `m0-v1` together through each assembly when the slice exists. Case-owned JSON only; no common gameplay union. Coordinate incompatible changes with a version change.

Initial safe view:
```json
{ "cwd": "/", "entries": [{ "name": "readme.txt", "kind": "file" }], "workstation_ready": false }
```

Simulated input: `action_type: "workstation.command"`, payload `{"command":"help"}`. Safe outcome: `{"lines":["Simulated workstation help."],"ok":true}`. A read-only help command need not advance revision. Navigation and the eventual small state-changing sequence are interpreted only by Terminal. Never execute an OS shell. The actual unlocking sequence and server definitions are chosen privately in the later slice, not in these public examples.

Premature protected read may be a handled outcome `{"lines":["Access denied."],"ok":false}` with the unchanged locked view; a direct protected asset fetch returns neutral 404. Once legitimately permitted, a safe view can add `protected_file: {"asset_id":"synthetic-report","name":"report.txt"}`. Reaching/reading that file contributes to/completes Terminal. Its bytes remain behind the authorized asset endpoint; completion is in the neutral summary. No protected text or solutions appear in these examples.

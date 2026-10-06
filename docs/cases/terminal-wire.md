# Terminal-owned M0 examples

Terminal registers `terminal-demo` / `m0-v1` through each assembly. The backend slice is implemented by ticket #6. Case-owned JSON only; no common gameplay union. Coordinate incompatible changes with a version change. Cosmetic wording and presentation changes do not require an artifact hash.

Initial safe view:
```json
{ "cwd": "/", "entries": [{ "name": "readme.txt", "kind": "file" }, { "name": "maintenance", "kind": "directory" }], "workstation_ready": false }
```

Simulated input: `action_type: "workstation.command"`, payload `{"command":"help"}`. Safe outcome: `{"lines":["Simulated workstation help."],"ok":true}`. Every successful or handled-denied command returns `lines` (an array of strings) and `ok` (a boolean) inside the neutral action result. Unknown commands and premature commands are handled outcomes with `ok:false`; malformed case payloads use neutral `422 action_rejected`. Extra fields cannot assign private state.

The frontend renders server-produced entries, command output and instructions, and submits the entered command unchanged. It needs no hardcoded unlocking sequence, private prerequisites or protected text. `help` and the root `readme.txt` lead to public in-world instructions. Navigation and the short prerequisite sequence are interpreted solely by Terminal; no operating-system shell, real host, network call or asynchronous side effect is involved. Read-only help/location/listing/public text and repeated satisfied prerequisites do not advance revision. Navigation and newly satisfied prerequisites are durable case-state changes.

Premature protected read may be a handled outcome `{"lines":["Access denied."],"ok":false}` with the unchanged locked view; a direct protected asset fetch returns neutral 404. Once legitimately permitted, a safe view can add `protected_file: {"asset_id":"synthetic-report","name":"report.txt"}`. Reaching/reading that file contributes to/completes Terminal. Its bytes remain behind the authorized asset endpoint; completion is in the neutral summary. No protected text or solutions appear in these examples.

`workstation_ready:true` means Terminal's prerequisites permit the archive. Only then can projected entries expose the archive and `protected_file` identify its authorized report. Reaching the report through a permitted command records first completion in the same durable transition as Terminal's read milestone. Its output points to the protected file rather than embedding report bytes. The frontend uses `/api/playthroughs/{playthrough_id}/assets/{asset_id}` to fetch `text/plain; charset=utf-8` content with `Cache-Control: no-store`; a copied URL or a locked second run never inherits permission. Repeated reads preserve the first completion time without another revision.

The neutral revision/request agreement is unchanged: committed duplicates replay their safe recorded snapshot, changed input under a committed ID conflicts, stale writes require reconciliation, and a no-op needs no permanent receipt. Current views resume Terminal's directory and prerequisite state from PostgreSQL. The server package owns the actual simulated definitions and unrevealed report; this document remains safe handoff vocabulary.

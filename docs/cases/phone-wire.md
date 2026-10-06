# Phone-owned M0 examples

Phone registers `phone-demo` / `m0-v1` through both assemblies. Its frontend and authoritative backend are implemented in M0. These safe examples define case-owned JSON carried inside the neutral envelopes, not shared platform types. Incompatible changes require a coordinated version change; actual answers and protected bytes remain server-only.

Initial safe view:
```json
{ "messages": [{ "id": "public-example", "sender": "Demo contact", "text": "A public message." }], "gallery_unlocked": false }
```

Gallery attempt: `action_type: "gallery.attempt"`, payload `{"password":"synthetic-wrong-input"}`. Handled wrong answer outcome: `{"accepted":false}`; view remains locked. This input is deliberately wrong and never a real answer. M0 persists an attempt counter, so a new locked attempt advances revision. Already-unlocked attempts and repeated opening are no-ops that retain revision and first completion.

Opening after a legitimate unlock: `action_type: "gallery.open"`, payload `{}`; outcome `{"opened":true}`. Safe authorized view can add `gallery_assets: [{"asset_id":"synthetic-image","label":"Revealed image"}]`. Actual bytes only come from the playthrough-authorized asset endpoint. Opening contributes to/completes Phone; completion stays in the neutral summary. Never send correct answers, private definitions, protected inventories before unlock, or protected bytes in public fixtures.

Messages/navigation can be local cosmetic interactions. The case owns UI, server validation, projections and authorization; these examples establish no generic phone system.

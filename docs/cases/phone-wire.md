# Phone-owned M0 examples

Gate vocabulary for the Phone implementers; no gameplay is implemented in #2. Register `phone-demo` / `m0-v1` together through each assembly when the slice exists. These examples define case-owned JSON carried inside the neutral envelopes, not shared platform types. Incompatible changes require a coordinated version change.

Initial safe view:
```json
{ "messages": [{ "id": "public-example", "sender": "Demo contact", "text": "A public message." }], "gallery_unlocked": false }
```

Gallery attempt: `action_type: "gallery.attempt"`, payload `{"password":"synthetic-wrong-input"}`. Handled wrong answer outcome: `{"accepted":false}`; view remains locked. This input is deliberately wrong and never a real answer. The case decides whether a durable attempt counter exists and therefore whether revision advances.

Opening after a legitimate unlock: `action_type: "gallery.open"`, payload `{}`; outcome `{"opened":true}`. Safe authorized view can add `gallery_assets: [{"asset_id":"synthetic-image","label":"Revealed image"}]`. Actual bytes only come from the playthrough-authorized asset endpoint. Opening contributes to/completes Phone; completion stays in the neutral summary. Never send correct answers, private definitions, protected inventories before unlock, or protected bytes in public fixtures.

Messages/navigation can be local cosmetic interactions. The case owns UI, server validation, projections and authorization; these examples establish no generic phone system.

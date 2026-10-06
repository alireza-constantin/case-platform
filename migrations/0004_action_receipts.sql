BEGIN;
CREATE TABLE IF NOT EXISTS action_receipts (
 playthrough_id text NOT NULL REFERENCES playthroughs(playthrough_id),
 request_id text NOT NULL,
 input_hash bytea NOT NULL,
 result bytea NOT NULL,
 PRIMARY KEY (playthrough_id, request_id)
);
COMMIT;

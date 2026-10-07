BEGIN;
CREATE TABLE IF NOT EXISTS playthrough_events (
 event_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 playthrough_id text NOT NULL REFERENCES playthroughs(playthrough_id),
 revision bigint NOT NULL,
 event jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
COMMIT;

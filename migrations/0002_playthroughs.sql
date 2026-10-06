BEGIN;
CREATE TABLE IF NOT EXISTS playthroughs (
 playthrough_id text PRIMARY KEY,
 guest_token text NOT NULL REFERENCES guests(session_token),
 case_id text NOT NULL,
 case_version text NOT NULL,
 revision bigint NOT NULL DEFAULT 0 CHECK (revision >= 0),
 completed_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 state jsonb NOT NULL
);
CREATE INDEX IF NOT EXISTS playthroughs_guest_created ON playthroughs(guest_token, created_at DESC);
COMMIT;

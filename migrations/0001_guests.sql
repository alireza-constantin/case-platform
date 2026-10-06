BEGIN;
CREATE TABLE IF NOT EXISTS guests (
    session_token text PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now()
);
COMMIT;

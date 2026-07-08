-- account security and password recovery

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS recovery_email text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS password_updated_at timestamptz;

CREATE TABLE IF NOT EXISTS password_reset_codes (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  code_hash    text NOT NULL,
  expires_at   timestamptz NOT NULL,
  used_at      timestamptz,
  attempts     integer NOT NULL DEFAULT 0,
  requested_ip text NOT NULL DEFAULT '',
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS password_reset_codes_user_created_idx
  ON password_reset_codes (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS password_reset_codes_active_idx
  ON password_reset_codes (user_id, expires_at DESC)
  WHERE used_at IS NULL;
-- Denylist token untuk logout (invalidasi JWT sebelum kadaluarsa).
CREATE TABLE
    IF NOT EXISTS revoked_tokens (
        jti VARCHAR(64) PRIMARY KEY,
        expires_at TIMESTAMPTZ NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_revoked_tokens_expires_at ON revoked_tokens (expires_at);

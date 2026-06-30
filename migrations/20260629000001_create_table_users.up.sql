CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Satu username bisa punya banyak role non-admin (buyer/seller/driver).
CREATE TABLE
    IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        email VARCHAR(255) NOT NULL UNIQUE,
        password TEXT NOT NULL,
        is_admin BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

-- Role yang dimiliki user. Role aktif dipilih per sesi (lihat JWT claim).
CREATE TABLE
    IF NOT EXISTS user_roles (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        role VARCHAR(20) NOT NULL CHECK (role IN ('buyer', 'seller', 'driver')),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        UNIQUE (user_id, role)
    );

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id);

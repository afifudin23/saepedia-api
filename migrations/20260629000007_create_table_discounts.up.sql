-- Diskon: Voucher & Promo dibedakan lewat kolom `kind`.
-- Voucher punya usage_limit + used_count; Promo tidak (usage_limit NULL).
CREATE TABLE
    IF NOT EXISTS discounts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        code VARCHAR(50) NOT NULL UNIQUE,
        kind VARCHAR(10) NOT NULL CHECK (kind IN ('voucher', 'promo')),
        discount_type VARCHAR(10) NOT NULL CHECK (discount_type IN ('percent', 'fixed')),
        discount_value BIGINT NOT NULL CHECK (discount_value >= 0),
        max_discount BIGINT NOT NULL DEFAULT 0,
        min_spend BIGINT NOT NULL DEFAULT 0,
        expires_at TIMESTAMPTZ NOT NULL,
        usage_limit INTEGER,
        used_count INTEGER NOT NULL DEFAULT 0,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_discounts_kind ON discounts (kind);

-- Simpan kode diskon yang dipakai pada order (untuk laporan).
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount_code VARCHAR(50) NOT NULL DEFAULT '';

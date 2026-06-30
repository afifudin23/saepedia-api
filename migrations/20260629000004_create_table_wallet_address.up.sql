-- Dompet buyer. Saldo integer rupiah.
CREATE TABLE
    IF NOT EXISTS wallets (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
        balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

-- Riwayat transaksi dompet (topup, payment, refund).
CREATE TABLE
    IF NOT EXISTS wallet_transactions (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        wallet_id UUID NOT NULL REFERENCES wallets (id) ON DELETE CASCADE,
        type VARCHAR(20) NOT NULL CHECK (type IN ('topup', 'payment', 'refund')),
        amount BIGINT NOT NULL,
        balance_after BIGINT NOT NULL,
        description TEXT NOT NULL DEFAULT '',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_wallet_tx_wallet_id ON wallet_transactions (wallet_id, created_at DESC);

-- Alamat pengiriman buyer.
CREATE TABLE
    IF NOT EXISTS addresses (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        recipient_name VARCHAR(100) NOT NULL,
        phone VARCHAR(20) NOT NULL,
        full_address TEXT NOT NULL,
        is_primary BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses (user_id);

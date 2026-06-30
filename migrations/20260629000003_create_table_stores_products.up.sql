-- Toko milik seller. Nama toko WAJIB unik.
CREATE TABLE
    IF NOT EXISTS stores (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
        name VARCHAR(150) NOT NULL UNIQUE,
        description TEXT NOT NULL DEFAULT '',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

-- Produk milik toko. Harga & stok dalam satuan integer (rupiah / unit).
CREATE TABLE
    IF NOT EXISTS products (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        store_id UUID NOT NULL REFERENCES stores (id) ON DELETE CASCADE,
        name VARCHAR(200) NOT NULL,
        description TEXT NOT NULL DEFAULT '',
        price BIGINT NOT NULL CHECK (price >= 0),
        stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_products_store_id ON products (store_id);

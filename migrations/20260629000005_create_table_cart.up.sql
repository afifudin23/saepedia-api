-- Keranjang buyer. Single-store rule: satu cart hanya berisi produk dari SATU toko.
-- store_id di-set saat item pertama masuk, dan di-reset NULL saat cart kosong.
CREATE TABLE
    IF NOT EXISTS carts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        user_id UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
        store_id UUID REFERENCES stores (id) ON DELETE SET NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE TABLE
    IF NOT EXISTS cart_items (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        cart_id UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
        product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
        quantity INTEGER NOT NULL CHECK (quantity > 0),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        UNIQUE (cart_id, product_id)
    );

CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items (cart_id);

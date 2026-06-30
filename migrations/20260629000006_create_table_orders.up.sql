-- Order hasil checkout. Single-store: satu order = satu toko.
-- Semua nominal dalam integer rupiah. PPN 12% dihitung dari (subtotal - discount).
CREATE TABLE
    IF NOT EXISTS orders (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        buyer_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        store_id UUID NOT NULL REFERENCES stores (id) ON DELETE CASCADE,
        recipient_name VARCHAR(100) NOT NULL,
        phone VARCHAR(20) NOT NULL,
        full_address TEXT NOT NULL,
        delivery_method VARCHAR(20) NOT NULL CHECK (delivery_method IN ('instant', 'next_day', 'regular')),
        subtotal BIGINT NOT NULL CHECK (subtotal >= 0),
        discount BIGINT NOT NULL DEFAULT 0 CHECK (discount >= 0),
        delivery_fee BIGINT NOT NULL CHECK (delivery_fee >= 0),
        tax BIGINT NOT NULL CHECK (tax >= 0),
        total BIGINT NOT NULL CHECK (total >= 0),
        status VARCHAR(40) NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_orders_buyer_id ON orders (buyer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_store_id ON orders (store_id, created_at DESC);

-- Snapshot item saat order dibuat (nama & harga dikunci di waktu checkout).
CREATE TABLE
    IF NOT EXISTS order_items (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
        product_id UUID NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
        product_name VARCHAR(200) NOT NULL,
        price BIGINT NOT NULL CHECK (price >= 0),
        quantity INTEGER NOT NULL CHECK (quantity > 0),
        subtotal BIGINT NOT NULL CHECK (subtotal >= 0)
    );

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);

-- Riwayat perubahan status order, lengkap dengan timestamp.
CREATE TABLE
    IF NOT EXISTS order_status_histories (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
        status VARCHAR(40) NOT NULL,
        note TEXT NOT NULL DEFAULT '',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_order_status_order_id ON order_status_histories (order_id, created_at);

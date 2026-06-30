-- Delivery job dimodelkan pada order itu sendiri (lensa driver):
--   status 'Menunggu Pengirim' + driver_id NULL  = job tersedia
--   status 'Sedang Dikirim'                       = job diambil driver
--   status 'Pesanan Selesai'                      = job selesai
ALTER TABLE orders ADD COLUMN IF NOT EXISTS driver_id UUID REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS driver_earning BIGINT NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS taken_at TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;
-- Penanda agar refund overdue bersifat idempotent (cegah double refund).
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_orders_driver_id ON orders (driver_id);

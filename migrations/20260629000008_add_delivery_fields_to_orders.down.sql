DROP INDEX IF EXISTS idx_orders_driver_id;
DROP INDEX IF EXISTS idx_orders_status;
ALTER TABLE orders DROP COLUMN IF EXISTS refunded_at;
ALTER TABLE orders DROP COLUMN IF EXISTS completed_at;
ALTER TABLE orders DROP COLUMN IF EXISTS taken_at;
ALTER TABLE orders DROP COLUMN IF EXISTS driver_earning;
ALTER TABLE orders DROP COLUMN IF EXISTS driver_id;

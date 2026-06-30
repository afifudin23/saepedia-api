-- Review aplikasi/website (BUKAN review produk). Boleh diisi guest tanpa checkout.
CREATE TABLE
    IF NOT EXISTS app_reviews (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        reviewer_name VARCHAR(100) NOT NULL,
        rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
        comment TEXT NOT NULL,
        user_id UUID REFERENCES users (id) ON DELETE SET NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_app_reviews_created_at ON app_reviews (created_at DESC);

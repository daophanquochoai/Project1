\connect product_service;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- categories
CREATE TABLE IF NOT EXISTS categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- products
CREATE TABLE IF NOT EXISTS products (
                                        id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL,
    description    TEXT,
    price          NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    category_id    UUID NOT NULL REFERENCES categories(id),
    average_rating NUMERIC(3,2) NOT NULL DEFAULT 0,
    total_ratings  INTEGER NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
    );

-- ratings
CREATE TABLE IF NOT EXISTS ratings (
                                       id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    user_id    UUID NOT NULL, -- from User Service (no FK across DB)
    rating     INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
    );

-- product_related (many-to-many)
CREATE TABLE IF NOT EXISTS product_related (
    product_id     UUID NOT NULL REFERENCES products(id),
    related_id     UUID NOT NULL REFERENCES products(id),
    relation_type  TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (product_id, related_id, relation_type)
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_average_rating ON products(average_rating);
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);

CREATE INDEX IF NOT EXISTS idx_ratings_product_id ON ratings(product_id);
CREATE INDEX IF NOT EXISTS idx_ratings_user_id ON ratings(user_id);

-- UNIQUE: 1 user rate 1 product (soft-delete aware)
CREATE UNIQUE INDEX IF NOT EXISTS idx_ratings_user_product
    ON ratings(user_id, product_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_related_product_id ON product_related(product_id);
CREATE INDEX IF NOT EXISTS idx_product_related_related_id ON product_related(related_id);

-- updated_at auto
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_categories_updated_at ON categories;
CREATE TRIGGER trg_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_ratings_updated_at ON ratings;
CREATE TRIGGER trg_ratings_updated_at
    BEFORE UPDATE ON ratings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Recompute avg + count for a product (ignore soft-deleted ratings)
CREATE OR REPLACE FUNCTION recompute_product_rating(p_product_id UUID)
RETURNS VOID AS $$
BEGIN
UPDATE products p
SET
    average_rating = COALESCE((
                                  SELECT ROUND(AVG(r.rating)::numeric, 2)
                                  FROM ratings r
                                  WHERE r.product_id = p_product_id
                                    AND r.deleted_at IS NULL
                              ), 0),
    total_ratings = COALESCE((
                                 SELECT COUNT(*)
                                 FROM ratings r
                                 WHERE r.product_id = p_product_id
                                   AND r.deleted_at IS NULL
                             ), 0),
    updated_at = NOW()
WHERE p.id = p_product_id;
END;
$$ LANGUAGE plpgsql;

-- Trigger on ratings: insert/update/delete (soft delete)
CREATE OR REPLACE FUNCTION trg_ratings_recompute()
RETURNS TRIGGER AS $$
DECLARE
v_product_id UUID;
BEGIN
  v_product_id := COALESCE(NEW.product_id, OLD.product_id);
  PERFORM recompute_product_rating(v_product_id);
RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_ratings_after_change ON ratings;
CREATE TRIGGER trg_ratings_after_change
    AFTER INSERT OR UPDATE OR DELETE ON ratings
    FOR EACH ROW EXECUTE FUNCTION trg_ratings_recompute();

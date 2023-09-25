-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS "products" (
    "id" BIGINT PRIMARY KEY,
    "name" TEXT,
    "slug" TEXT,
    "price" DECIMAL(20,0),
    "description" TEXT,
    "quantity" INT,
    "created_at" TIMESTAMP NOT NULL DEFAULT 'NOW()',
    "updated_at" TIMESTAMP NOT NULL DEFAULT 'NOW()',
    "deleted_at" timestamp
);


ALTER TABLE IF EXISTS "products" ADD CONSTRAINT product_slug_unique UNIQUE (slug);
CREATE INDEX products_name_fts_index ON products USING gin(to_tsvector('english', name));

-- +migrate Down
DROP TABLE IF EXISTS "users";
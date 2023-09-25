-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS "transactions" (
    "id" BIGINT PRIMARY KEY,
    "total_price" DECIMAL(20,0) NOT NULL,
    "amount_paid" DECIMAL(20,0) NOT NULL,
    "change" DECIMAL(20,0) NOT NULL,
    "created_by" BIGINT NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT 'NOW()'
);

ALTER TABLE "transactions" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id");

CREATE TABLE IF NOT EXISTS "transaction_details" (
    "transaction_id" BIGINT NOT NULL,
    "product_id" BIGINT NOT NULL,
    "quantity" BIGINT NOT NULL,
    "subtotal" DECIMAL(20,0) NOT NULL
);

ALTER TABLE "transaction_details" ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id");
ALTER TABLE "transaction_details" ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id");

-- +migrate Down
DROP TABLE IF EXISTS "transactions";
DROP TABLE IF EXISTS "transaction_details";
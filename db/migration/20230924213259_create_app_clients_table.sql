-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS "app_clients" (
    "id" bigint PRIMARY KEY,
    "client_id" text NOT NULL,
    "client_secret" text NOT NULL,
    "updated_at" timestamp NOT NULL DEFAULT 'now()',
    "created_at" timestamp NOT NULL DEFAULT 'now()'
);
ALTER TABLE "app_clients" ADD CONSTRAINT client_unique UNIQUE("client_id");

-- +migrate Down
DROP TABLE IF EXISTS "app_clients";
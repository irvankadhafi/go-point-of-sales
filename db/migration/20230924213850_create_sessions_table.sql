-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS "sessions" (
    "id" bigint PRIMARY KEY,
    "user_id" bigint,
    "app_id" bigint NOT NULL,
    "access_token" text NOT NULL,
    "refresh_token" text NOT NULL,
    "access_token_expired_at" timestamp NOT NULL,
    "refresh_token_expired_at" timestamp NOT NULL,
    "user_agent" text NOT NULL,
    "ip_address" text NOT NULL,
    "updated_at" timestamp NOT NULL DEFAULT 'now()',
    "created_at" timestamp NOT NULL DEFAULT 'now()'
);
ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "sessions" ADD FOREIGN KEY ("app_id") REFERENCES "app_clients" ("id");

-- +migrate Down
DROP TABLE IF EXISTS "sessions";
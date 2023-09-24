-- +migrate Up notransaction
CREATE TYPE "user_status" AS ENUM (
    'PENDING',
    'ACTIVE',
    'INACTIVE'
    );

CREATE TABLE IF NOT EXISTS "users" (
    "id" bigint PRIMARY KEY,
    "name" text,
    "email" text,
    "password" text,
    "role" user_role,
    "status" user_status,
    "created_by" bigint,
    "updated_by" bigint,
    "created_at" timestamp NOT NULL DEFAULT 'now()',
    "updated_at" timestamp NOT NULL DEFAULT 'now()',
    "deleted_at" timestamp
);

ALTER TABLE "users" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id");
ALTER TABLE "users" ADD FOREIGN KEY ("updated_by") REFERENCES "users" ("id");
ALTER TABLE "users" ADD CONSTRAINT email_unique UNIQUE("email");

-- +migrate Down
DROP TABLE IF EXISTS "users";
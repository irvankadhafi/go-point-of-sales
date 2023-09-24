-- +migrate Up notransaction
CREATE TABLE IF NOT EXISTS "audits" (
    "user_id" BIGINT NOT NULL,
    "auditable_type" TEXT NOT NULL,
    "auditable_id" BIGINT NOT NULL,
    "action" TEXT NOT NULL,
    "audited_changes" JSON DEFAULT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX "audits_auditable_id_idx" ON "audits" ("auditable_id");
CREATE INDEX "audits_auditable_type_idx" ON "audits" ("auditable_type");
CREATE INDEX "audits_created_at_idx" ON "audits" ("created_at");

-- +migrate Down
DROP TABLE IF EXISTS "audits";
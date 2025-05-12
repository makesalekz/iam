-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "is_blocked" boolean NOT NULL DEFAULT false;

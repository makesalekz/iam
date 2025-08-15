-- Rename a column from "last_activity_at" to "last_seen"
ALTER TABLE "users" RENAME COLUMN "last_activity_at" TO "last_seen";

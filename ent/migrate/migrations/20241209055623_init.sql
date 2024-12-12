-- Modify "one_time_passwords" table
ALTER TABLE "one_time_passwords" ADD COLUMN "failed_attempts" bigint NOT NULL DEFAULT 0;

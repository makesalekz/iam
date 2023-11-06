-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "phone_verified" boolean NOT NULL DEFAULT false, ADD COLUMN "email_verified" boolean NOT NULL DEFAULT false;

-- Modify "user_credentials" table
ALTER TABLE "user_credentials" ADD COLUMN "external_user_id" bigint NULL, ADD COLUMN "phone" character varying NULL;

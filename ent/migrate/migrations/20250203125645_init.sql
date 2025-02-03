-- Drop index "user_credentials_mail_key" from table: "user_credentials"
DROP INDEX "user_credentials_mail_key";
-- Modify "user_credentials" table
ALTER TABLE "user_credentials" ALTER COLUMN "mail" DROP NOT NULL;

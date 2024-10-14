-- Modify "user_credentials" table
ALTER TABLE "user_credentials" ALTER COLUMN "mail" SET NOT NULL;
-- Create index "user_credentials_mail_key" to table: "user_credentials"
CREATE UNIQUE INDEX "user_credentials_mail_key" ON "user_credentials" ("mail");

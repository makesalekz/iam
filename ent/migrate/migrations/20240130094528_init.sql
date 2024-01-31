-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "username" character varying NULL;
-- Create index "user_username" to table: "users"
CREATE UNIQUE INDEX "user_username" ON "users" ("username") WHERE (username IS NOT NULL);
-- Create index "users_username_key" to table: "users"
CREATE UNIQUE INDEX "users_username_key" ON "users" ("username");

-- Update all null users.username columns to 'user[id]'
update users
set username = 'user' || id
where username is null;
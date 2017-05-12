
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "report_users" ("id" integer primary key autoincrement,"username" varchar(255) NOT NULL UNIQUE,"parent_username" varchar(255) NOT NULL,"hash" varchar(255));
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "report_users";

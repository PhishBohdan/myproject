-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "restricted_campaigns" ("id" integer primary key autoincrement,"restricted_user_id" integer NOT NULL,"campaign_id" integer NOT NULL);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE "restricted_campaigns";
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
-- SQLite3 doesnt support dropping columns

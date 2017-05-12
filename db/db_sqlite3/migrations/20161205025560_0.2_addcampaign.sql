
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "restricted_users" ADD COLUMN campaign_id integer;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
-- SQLite3 doesnt support dropping columns

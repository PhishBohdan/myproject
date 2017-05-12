
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "users" ADD COLUMN restricted boolean;
ALTER TABLE "users" ADD COLUMN restricted_parent_id integer;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
-- SQLite3 doesnt support dropping columns

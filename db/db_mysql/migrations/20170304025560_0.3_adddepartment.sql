-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "results" ADD COLUMN department varchar(255);
ALTER TABLE "targets" ADD COLUMN department varchar(255);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
-- SQLite3 doesnt support dropping columns

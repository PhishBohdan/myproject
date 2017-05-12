
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE "report_users" RENAME TO "restricted_users";
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE "restricted_users" RENAME TO "report_users";

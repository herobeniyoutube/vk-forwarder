-- +goose Up
-- +goose StatementBegin
ALTER TABLE idempotency_keys
	ADD COLUMN status text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE idempotency_keys
	DROP COLUMN status;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
ALTER TABLE refresh_tokens
ALTER COLUMN created_at TYPE TIMESTAMP
WITH
  TIME ZONE,
ALTER COLUMN expires_at TYPE TIMESTAMP
WITH
  TIME ZONE,
ALTER COLUMN revoked_at TYPE TIMESTAMP
WITH
  TIME ZONE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE refresh_tokens
ALTER COLUMN created_at TYPE TIMESTAMP,
ALTER COLUMN expires_at TYPE TIMESTAMP,
ALTER COLUMN revoked_at TYPE TIMESTAMP;

-- +goose StatementEnd

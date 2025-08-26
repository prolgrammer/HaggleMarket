-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_offers_created_at ON offers(created_at);
CREATE INDEX idx_offers_expires_at ON offers(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_offers_created_at;
DROP INDEX IF EXISTS idx_offers_expires_at;
-- +goose StatementEnd

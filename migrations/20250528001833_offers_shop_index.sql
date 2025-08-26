-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_offers_shop_id ON offers (shop_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_offers_shop_id;
-- +goose StatementEnd

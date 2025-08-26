-- +goose Up
-- +goose StatementBegin
CREATE TABLE offers (
    id SERIAL PRIMARY KEY,
    offer_price DECIMAL(10,2) NOT NULL,
    currency char(3) NOT NULL,
    status VARCHAR(50) NOT NULL default 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '7 days',
    shop_id INT NOT NULL,
    user_id INT NOT NULL,
    product_id INT NOT NULL,
    FOREIGN KEY (product_id, shop_id) REFERENCES shop_inventory(product_id, shop_id)
);

CREATE INDEX idx_offers_user_id ON offers(user_id);
CREATE INDEX idx_offers_product_id ON offers(product_id);
CREATE INDEX idx_offers_status_id ON offers(status) WHERE status = 'pending';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS offers;
-- +goose StatementEnd

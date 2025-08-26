-- +goose Up
-- +goose StatementBegin
CREATE TABLE seller_reviews (
    id SERIAL PRIMARY KEY,
    seller_id INT NOT NULL,
    user_id INT NOT NULL,
    rating INT NOT NULL,
    review TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (seller_id) REFERENCES users(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_seller_reviews_seller_id ON seller_reviews(seller_id);
CREATE INDEX idx_seller_reviews_user_id ON seller_reviews(user_id);
CREATE INDEX idx_seller_reviews_seller_user ON seller_reviews(seller_id, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS seller_reviews;
-- +goose StatementEnd

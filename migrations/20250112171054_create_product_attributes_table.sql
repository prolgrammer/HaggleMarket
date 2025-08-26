-- +goose Up
-- +goose StatementBegin
CREATE TABLE product_attributes (
    product_id INT PRIMARY KEY ,
    attributes JSONB NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_attributes;
-- +goose StatementEnd

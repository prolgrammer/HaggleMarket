-- +goose Up
-- Очистка всех таблиц и сброс последовательностей
-- +goose StatementBegin
TRUNCATE TABLE refresh_tokens CASCADE;

TRUNCATE TABLE seller_reviews CASCADE;

TRUNCATE TABLE product_reviews CASCADE;

TRUNCATE TABLE shop_point_inventory CASCADE;

TRUNCATE TABLE shop_points CASCADE;

TRUNCATE TABLE shop_inventory CASCADE;

TRUNCATE TABLE products CASCADE;

TRUNCATE TABLE categories CASCADE;

TRUNCATE TABLE shops CASCADE;

TRUNCATE TABLE users CASCADE;

-- Сброс последовательностей для всех таблиц
ALTER SEQUENCE users_id_seq
RESTART WITH 1;

ALTER SEQUENCE shops_id_seq
RESTART WITH 1;

ALTER SEQUENCE categories_id_seq
RESTART WITH 1;

ALTER SEQUENCE products_id_seq
RESTART WITH 1;

ALTER SEQUENCE shop_points_id_seq
RESTART WITH 1;

ALTER SEQUENCE product_reviews_id_seq
RESTART WITH 1;

ALTER SEQUENCE seller_reviews_id_seq
RESTART WITH 1;

-- +goose StatementEnd
-- Вставка пользователей (первая транзакция)
-- +goose StatementBegin
INSERT INTO
    users (
        name,
        phone_number,
        password_hash,
        email,
        is_store
    )
VALUES
    (
        'John Doe',
        '+1234567890',
        'hashed_password_1',
        'john.doe@example.com',
        FALSE
    ), -- Покупатель, id=1
    (
        'Jane Smith',
        '+1234567891',
        'hashed_password_2',
        'jane.smith@example.com',
        FALSE
    ), -- Покупатель, id=2
    (
        'Store Owner 1',
        '+1234567892',
        'hashed_password_3',
        'store1@example.com',
        TRUE
    ), -- Продавец, id=3
    (
        'Store Owner 2',
        '+1234567893',
        'hashed_password_4',
        'store2@example.com',
        TRUE
    );

-- Продавец, id=4
-- +goose StatementEnd
-- Вставка магазинов (вторая транзакция)
-- +goose StatementBegin
INSERT INTO
    shops (name, user_id)
VALUES
    ('Store 1', 3), -- Магазин продавца Store Owner 1, id=1
    ('Store 2', 4);

-- Магазин продавца Store Owner 2, id=2
-- +goose StatementEnd
-- Вставка категорий (третья транзакция)
-- +goose StatementBegin
INSERT INTO
    categories (name, lft, rgt, parent_id)
VALUES
    ('Electronics', 1, 6, NULL), -- id=1
    ('Phones', 2, 3, 1), -- id=2
    ('Laptops', 4, 5, 1), -- id=3
    ('Clothing', 7, 10, NULL), -- id=4
    ('Shirts', 8, 9, 4);

-- id=5
-- +goose StatementEnd
-- Вставка продуктов (четвёртая транзакция)
-- +goose StatementBegin
INSERT INTO
    products (name, category_id, description)
VALUES
    (
        'iPhone 14',
        2,
        'Latest iPhone model with advanced features'
    ), -- id=1, категория Phones
    (
        'MacBook Pro',
        3,
        'High-performance laptop for professionals'
    ), -- id=2, категория Laptops
    ('T-Shirt', 5, 'Comfortable cotton t-shirt');

-- id=3, категория Shirts
-- +goose StatementEnd
-- Вставка данных о наличии продуктов в магазинах (пятая транзакция)
-- +goose StatementBegin
INSERT INTO
    shop_inventory (product_id, shop_id, is_available)
VALUES
    (1, 1, TRUE), -- iPhone 14 доступен в Store 1
    (2, 1, TRUE), -- MacBook Pro доступен в Store 1
    (3, 2, TRUE);

-- T-Shirt доступен в Store 2
-- +goose StatementEnd
-- Вставка точек выдачи (шестая транзакция)
-- +goose StatementBegin
INSERT INTO
    shop_points (shop_id, address, phone)
VALUES
    (1, '123 Main St, City 1', '+1234567895'), -- id=1, точка выдачи для Store 1
    (2, '456 High St, City 2', '+1234567896');

-- id=2, точка выдачи для Store 2
-- +goose StatementEnd
-- Вставка инвентаря для точек выдачи (седьмая транзакция)
-- +goose StatementBegin
INSERT INTO
    shop_point_inventory (shop_point_id, product_id, price, quantity)
VALUES
    (1, 1, 999, 10), -- iPhone 14 в точке выдачи 1
    (1, 2, 1999, 5), -- MacBook Pro в точке выдачи 1
    (2, 3, 20, 50);

-- T-Shirt в точке выдачи 2
-- +goose StatementEnd
-- Вставка отзывов о продуктах (восьмая транзакция)
-- +goose StatementBegin
INSERT INTO
    product_reviews (product_id, user_id, rating, review, created_at)
VALUES
    (
        1,
        1,
        5,
        'Amazing phone, love the camera!',
        '2025-04-20 10:00:00'
    ), -- John Doe об iPhone 14
    (
        1,
        2,
        4,
        'Great phone, but battery life could be better.',
        '2025-04-20 11:00:00'
    ), -- Jane Smith об iPhone 14
    (
        2,
        1,
        5,
        'Best laptop I’ve ever used!',
        '2025-04-20 12:00:00'
    );

-- John Doe о MacBook Pro
-- +goose StatementEnd
-- Вставка отзывов о продавцах (девятая транзакция)
-- +goose StatementBegin
INSERT INTO
    seller_reviews (seller_id, user_id, rating, review, created_at)
VALUES
    (
        3,
        1,
        5,
        'Fast delivery, great service!',
        '2025-04-20 13:00:00'
    ), -- John Doe о Store Owner 1
    (
        3,
        2,
        4,
        'Good store, but communication could be improved.',
        '2025-04-20 14:00:00'
    ), -- Jane Smith о Store Owner 1
    (
        4,
        1,
        3,
        'Average experience, shipping was slow.',
        '2025-04-20 15:00:00'
    );

-- John Doe о Store Owner 2
-- +goose StatementEnd
-- Вставка refresh-токенов (десятая транзакция)
-- +goose StatementBegin
INSERT INTO
    refresh_tokens (
        uuid,
        created_at,
        expires_at,
        revoked_at,
        fingerprint,
        user_id
    )
VALUES
    (
        '550e8400-e29b-41d4-a716-446655440000',
        '2025-04-20 09:00:00',
        '2025-05-20 09:00:00',
        NULL,
        'fingerprint_1',
        1
    ), -- John Doe
    (
        '550e8400-e29b-41d4-a716-446655440001',
        '2025-04-20 09:00:00',
        '2025-05-20 09:00:00',
        NULL,
        'fingerprint_2',
        2
    ), -- Jane Smith
    (
        '550e8400-e29b-41d4-a716-446655440002',
        '2025-04-20 09:00:00',
        '2025-05-20 09:00:00',
        NULL,
        'fingerprint_3',
        3
    );

-- Store Owner 1
-- +goose StatementEnd
-- +goose Down
-- Удаление данных в обратном порядке из-за зависимостей
-- +goose StatementBegin
DELETE FROM refresh_tokens
WHERE
    uuid IN (
        '550e8400-e29b-41d4-a716-446655440000',
        '550e8400-e29b-41d4-a716-446655440001',
        '550e8400-e29b-41d4-a716-446655440002'
    );

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM seller_reviews
WHERE
    (seller_id, user_id) IN ((3, 1), (3, 2), (4, 1));

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM product_reviews
WHERE
    (product_id, user_id) IN ((1, 1), (1, 2), (2, 1));

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM shop_point_inventory
WHERE
    (shop_point_id, product_id) IN ((1, 1), (1, 2), (2, 3));

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM shop_points
WHERE
    address IN ('123 Main St, City 1', '456 High St, City 2');

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM shop_inventory
WHERE
    (product_id, shop_id) IN ((1, 1), (2, 1), (3, 2));

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM products
WHERE
    name IN ('iPhone 14', 'MacBook Pro', 'T-Shirt');

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM categories
WHERE
    name IN (
        'Electronics',
        'Phones',
        'Laptops',
        'Clothing',
        'Shirts'
    );

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM shops
WHERE
    name IN ('Store 1', 'Store 2');

-- +goose StatementEnd
-- +goose StatementBegin
DELETE FROM users
WHERE
    email IN (
        'john.doe@example.com',
        'jane.smith@example.com',
        'store1@example.com',
        'store2@example.com'
    );

-- +goose StatementEnd

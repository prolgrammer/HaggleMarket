-- Seed test data for all tables
-- This file populates the database with test data for development and testing purposes

-- Truncate all tables (in reverse dependency order to respect foreign keys)
TRUNCATE TABLE refresh_tokens CASCADE;
TRUNCATE TABLE image_keys CASCADE;
TRUNCATE TABLE offers CASCADE;
TRUNCATE TABLE shop_inventory CASCADE;
TRUNCATE TABLE product_attributes CASCADE;
TRUNCATE TABLE products CASCADE;
TRUNCATE TABLE categories CASCADE;
TRUNCATE TABLE shops CASCADE;
TRUNCATE TABLE notifications CASCADE;

-- Reset all serial sequences
ALTER SEQUENCE shops_id_seq RESTART WITH 1;
ALTER SEQUENCE notifications_id_seq RESTART WITH 1;
ALTER SEQUENCE categories_id_seq RESTART WITH 1;
ALTER SEQUENCE products_id_seq RESTART WITH 1;
ALTER SEQUENCE offers_id_seq RESTART WITH 1;


-- Insert test shops (owned by store owners)
INSERT INTO shops (name, user_id) VALUES
                                      ('shop1 name', 1),
                                      ('shop2 name', 2);

-- Insert test categories (using nested set model)
-- Seed Categories with a nested set model (lft, rgt)
-- We manually specify IDs to establish parent-child relationships.
INSERT INTO categories (id, name, lft, rgt, parent_id) VALUES
                                                           (1, 'Electronics', 1, 20, NULL),
                                                           (2, 'Computers & Accessories', 2, 11, 1),
                                                           (3, 'Laptops', 3, 4, 2),
                                                           (4, 'Desktops', 5, 6, 2),
                                                           (5, 'Monitors', 7, 8, 2),
                                                           (6, 'Keyboards', 9, 10, 2),
                                                           (7, 'Mobile Phones', 12, 19, 1),
                                                           (8, 'Smartphones', 13, 14, 7),
                                                           (9, 'Cases & Covers', 15, 16, 7),
                                                           (10, 'Chargers', 17, 18, 7),
                                                           (11, 'Home & Garden', 21, 28, NULL),
                                                           (12, 'Kitchen', 22, 25, 11),
                                                           (13, 'Cookware', 23, 24, 12),
                                                           (14, 'Furniture', 26, 27, 11),
                                                           (15, 'Books', 29, 30, NULL);

-- Update the sequence for the categories table's primary key.
SELECT setval(pg_get_serial_sequence('categories', 'id'), (SELECT MAX(id) FROM categories));

-- Seed Products into leaf categories
INSERT INTO products (name, category_id, description) VALUES
                                                          ('SuperFast Laptop', 3, 'A very fast laptop for all your needs.'),
                                                          ('Gamer Desktop PC', 4, 'High-end gaming desktop with RGB.'),
                                                          ('4K UltraWide Monitor', 5, 'Crisp and clear 34-inch ultrawide monitor.'),
                                                          ('Mechanical Keyboard', 6, 'Clicky and satisfying mechanical keyboard.'),
                                                          ('SmartyPhone X', 8, 'The latest and greatest smartphone.'),
                                                          ('Tough Phone Case', 9, 'A protective case for your SmartyPhone X.'),
                                                          ('Fast Wall Charger', 10, '100W USB-C fast charger.'),
                                                          ('Non-stick Pan Set', 13, 'A set of three non-stick pans.'),
                                                          ('Ergonomic Office Chair', 14, 'Comfortable chair for your home office.'),
                                                          ('The Art of Go', 15, 'A book about writing concurrent programs in Go.');

-- Insert product attributes (JSONB data)
-- Seed Product Attributes
INSERT INTO product_attributes (product_id, attributes) VALUES
                                                            (1, '{"ram": "16GB", "cpu": "Intel Core i7", "storage": "1TB SSD", "screen_size": "15.6 inches", "weight": "1.8kg"}'),
                                                            (2, '{"cpu": "AMD Ryzen 9", "gpu": "NVIDIA RTX 4080", "ram": "32GB", "storage": "2TB NVMe SSD", "case_form_factor": "Mid-Tower"}'),
                                                            (3, '{"screen_size": "34 inches", "resolution": "3440x1440", "refresh_rate": "144Hz", "panel_type": "IPS"}'),
                                                            (4, '{"switch_type": "Cherry MX Brown", "layout": "Tenkeyless", "backlight": "RGB", "connectivity": "USB-C, Bluetooth"}'),
                                                            (5, '{"screen_size": "6.7 inches", "storage": "256GB", "color": "Space Gray", "camera_resolution": "48MP"}'),
                                                            (6, '{"material": "TPU and Polycarbonate", "color": "Black", "compatibility": "SmartyPhone X", "drop_protection": "10 feet"}'),
                                                            (7, '{"wattage": "100W", "ports": 2, "connector_type": "USB-C"}'),
                                                            (8, '{"material": "Anodized Aluminum", "pieces": 3, "dishwasher_safe": true}'),
                                                            (9, '{"material": "Mesh", "color": "Black", "weight_capacity": "300 lbs", "adjustability": ["lumbar", "armrests", "height"]}'),
                                                            (10, '{"author": "Alan A. A. Donovan, Brian W. Kernighan", "pages": 416, "format": "Paperback", "isbn": "978-0134190440"}');

-- Insert shop inventory (products available in shops with prices)
INSERT INTO shop_inventory (product_id, shop_id, is_available, price, currency) VALUES
    -- Shop 1 (Electronics focus)
    (1, 1, TRUE, 1500.00, 'USD'), -- SuperFast Laptop
    (2, 1, TRUE, 2500.00, 'USD'), -- Gamer Desktop PC
    (3, 1, TRUE, 750.00, 'USD'),  -- 4K UltraWide Monitor
    (4, 1, TRUE, 150.00, 'USD'),  -- Mechanical Keyboard
    (5, 1, TRUE, 999.00, 'USD'),  -- SmartyPhone X
    (6, 1, FALSE, 40.00, 'USD'),   -- Tough Phone Case
    (7, 1, TRUE, 50.00, 'USD'),   -- Fast Wall Charger
    -- Shop 2 (General / Home goods)
    (5, 2, TRUE, 950.00, 'USD'),  -- SmartyPhone X
    (6, 2, TRUE, 35.00, 'USD'),   -- Tough Phone Case
    (7, 2, TRUE, 45.00, 'USD'),   -- Fast Wall Charger
    (8, 2, TRUE, 80.00, 'USD'),   -- Non-stick Pan Set
    (9, 2, TRUE, 250.00, 'USD'),  -- Ergonomic Office Chair
    (10, 2, TRUE, 25.00, 'USD');  -- The Art of Go

-- Insert test offers
INSERT INTO offers (offer_price, currency, status, created_at, updated_at, expires_at, shop_id, user_id, product_id) VALUES
    (1400.00, 'USD', 'pending', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW() + INTERVAL '6 days', 1, 3, 1),
    (700.00, 'USD', 'accepted', NOW() - INTERVAL '3 days', NOW() - INTERVAL '2 days', NOW() + INTERVAL '4 days', 1, 4, 3),
    (30.00, 'USD', 'rejected', NOW() - INTERVAL '4 days', NOW() - INTERVAL '3 days', NOW() + INTERVAL '3 days', 2, 3, 6),
    (20.00, 'USD', 'pending', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW() + INTERVAL '7 days', 2, 4, 10),
    (75.00, 'USD', 'pending', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours', NOW() + INTERVAL '6 days', 2, 3, 8);


-- Insert test notifications
INSERT INTO notifications (message, user_id) VALUES
    ('A new offer was placed on "SuperFast Laptop".', 1),      -- To owner of shop 1 (user 1)
    ('Your offer for "SuperFast Laptop" has been received.', 3), -- To user 3
    ('You accepted an offer for "4K UltraWide Monitor".', 1),   -- To owner of shop 1 (user 1)
    ('Your offer for "4K UltraWide Monitor" was accepted!', 4), -- To user 4
    ('You rejected an offer for "Tough Phone Case".', 2),       -- To owner of shop 2 (user 2)
    ('Your offer for "Tough Phone Case" was rejected.', 3),     -- To user 3
    ('A new offer was placed on "The Art of Go".', 2),          -- To owner of shop 2 (user 2)
    ('Your offer for "The Art of Go" has been received.', 4),    -- To user 4
    ('A new offer was placed on "Non-stick Pan Set".', 2),      -- To owner of shop 2 (user 2)
    ('Your offer for "Non-stick Pan Set" has been received.', 3); -- To user 3

-- Insert test refresh tokens
INSERT INTO refresh_tokens (uuid, created_at, expires_at, revoked_at, fingerprint, user_id) VALUES
                                                                                                ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NOW() - INTERVAL '1 day', NOW() + INTERVAL '29 days', NULL, 'Mozilla/5.0 Windows NT 10.0', 1),
                                                                                                ('b1ffdc99-9c0b-4ef8-bb6d-6bb9bd380a12', NOW() - INTERVAL '2 days', NOW() + INTERVAL '28 days', NULL, 'Mozilla/5.0 Macintosh Intel Mac OS X', 2),
                                                                                                ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', NOW() - INTERVAL '3 days', NOW() + INTERVAL '27 days', NULL, 'Mozilla/5.0 X11 Linux x86_64', 3),
                                                                                                ('d3ffbc99-9c0b-4ef8-bb6d-6bb9bd380a14', NOW() - INTERVAL '12 hours', NOW() + INTERVAL '30 days', NULL, 'Mozilla/5.0 iPhone OS 15_0', 4);

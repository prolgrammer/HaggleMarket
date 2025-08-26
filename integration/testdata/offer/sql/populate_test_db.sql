-- test data
insert into users (name, phone_number, password_hash, email, is_store)
values ('user1','user1phone', 'no','user1email', true);
insert into users (name, phone_number, password_hash, email, is_store)
values ('user2','user2phone', 'no','user2email', false);
insert into users (name, phone_number, password_hash, email, is_store)
values ('user3','user3phone', 'no','user3email', true);

insert into shops (name, user_id) values ('shop1', 1);
insert into shops (name, user_id) values ('shop2', 1);

insert into categories (name, lft, rgt, parent_id) values ('test_cat', 1, 1, 1);

insert into products (name, category_id, description) VALUES ('product1', 1, 'description1');
insert into products (name, category_id, description) VALUES ('product2', 1, 'description2');
insert into products (name, category_id, description) VALUES ('product3', 1, 'description3');
insert into products (name, category_id, description) VALUES ('product4', 1, 'description4');

insert into shop_inventory (product_id, shop_id, is_available, price, currency) VALUES (1, 1, true, 100.00, 'usd');
insert into shop_inventory (product_id, shop_id, is_available, price, currency) VALUES (2, 1, true, 120.00, 'usd');
insert into shop_inventory (product_id, shop_id, is_available, price, currency) VALUES (3, 1, true, 150.00, 'usd');
insert into shop_inventory (product_id, shop_id, is_available, price, currency) VALUES (4, 1, true, 180.00, 'usd');
insert into shop_inventory (product_id, shop_id, is_available, price, currency) VALUES (1, 2, true, 100.00, 'usd');
insert into shop_inventory (product_id, shop_id, is_available, price, currency) VALUES (2, 2, true, 120.00, 'usd');
insert into shop_inventory (product_id, shop_id, is_available, price, currency) VALUES (3, 2, true, 150.00, 'usd');
insert into shop_inventory (product_id, shop_id, is_available, price, currency) VALUES (4, 2, true, 180.00, 'usd');

insert into offers (offer_price, currency, status, created_at, updated_at, user_id, product_id, shop_id) VALUES (55, 'usd', default, default, default, 2, 1, 1);
insert into offers (offer_price, currency, status, created_at, updated_at, user_id, product_id, shop_id) VALUES (65, 'usd', default, default, default, 2, 2, 1);
insert into offers (offer_price, currency, status, created_at, updated_at, user_id, product_id, shop_id) VALUES (45, 'usd', default, default, default, 2, 3, 1);
insert into offers (offer_price, currency, status, created_at, updated_at, user_id, product_id, shop_id) VALUES (48, 'usd', default, default, default, 2, 4, 1);

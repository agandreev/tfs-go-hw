CREATE TABLE orders
(
    id      SERIAL PRIMARY KEY,
    name    TEXT,
    orderID TEXT,
    price   NUMERIC,
    amount  NUMERIC,
    side    TEXT
);
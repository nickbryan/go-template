CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
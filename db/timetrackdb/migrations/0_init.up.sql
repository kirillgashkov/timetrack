BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    passport_number text NOT NULL,
    surname text NOT NULL,
    name text NOT NULL,
    patronymic text,
    address text NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (passport_number)
);

COMMIT;

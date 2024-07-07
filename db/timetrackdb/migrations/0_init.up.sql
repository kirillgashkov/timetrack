BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id serial NOT NULL,
    passport_number text NOT NULL,
    surname text NOT NULL,
    name text NOT NULL,
    patronymic text,
    address text NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (passport_number)
);

COMMIT;

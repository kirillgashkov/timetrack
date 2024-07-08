BEGIN;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS users (
    passport_number text NOT NULL,
    surname text NOT NULL,
    name text NOT NULL,
    patronymic text,
    address text NOT NULL,
    PRIMARY KEY (passport_number)
);
CREATE INDEX IF NOT EXISTS users_passport_number_trgm_idx ON users USING gin (passport_number gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_surname_trgm_idx ON users USING gin (surname gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_name_trgm_idx ON users USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_patronymic_trgm_idx ON users USING gin (patronymic gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_address_trgm_idx ON users USING gin (address gin_trgm_ops);

COMMIT;

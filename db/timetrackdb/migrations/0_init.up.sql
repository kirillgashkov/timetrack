BEGIN;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

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
CREATE INDEX IF NOT EXISTS users_passport_number_trgm_idx ON users USING gin (passport_number gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_surname_trgm_idx ON users USING gin (surname gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_name_trgm_idx ON users USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_patronymic_trgm_idx ON users USING gin (patronymic gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_address_trgm_idx ON users USING gin (address gin_trgm_ops);

CREATE TABLE IF NOT EXISTS tasks (
    id serial NOT NULL,
    description text NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS times (
    id serial NOT NULL,
    task_id integer NOT NULL,
    user_id integer NOT NULL,
    started_at timestamp with time zone NOT NULL,
    ended_at timestamp with time zone,
    PRIMARY KEY (id),
    FOREIGN KEY (task_id) REFERENCES tasks (id),
    FOREIGN KEY (user_id) REFERENCES users (passport_number)
);

COMMIT;

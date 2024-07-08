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
CREATE INDEX IF NOT EXISTS tasks_description_trgm_idx ON tasks USING gin (description gin_trgm_ops);

CREATE TABLE IF NOT EXISTS tasks_users (
    task_id integer NOT NULL,
    user_id integer NOT NULL,
    status text NOT NULL,
    PRIMARY KEY (task_id, user_id),
    FOREIGN KEY (task_id) REFERENCES tasks (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    CHECK (status IN ('active', 'inactive'))
);
-- TODO: Add indexes.

CREATE TABLE IF NOT EXISTS tracks (
    id serial NOT NULL,
    type text NOT NULL,
    timestamp timestamp with time zone NOT NULL DEFAULT now(),
    task_id integer NOT NULL,
    user_id integer NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (task_id) REFERENCES tasks (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    CHECK (type IN ('start', 'stop'))
);
-- TODO: Add indexes.

COMMIT;

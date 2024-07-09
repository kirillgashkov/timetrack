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

CREATE TABLE IF NOT EXISTS works (
    id serial NOT NULL,
    started_at timestamp with time zone NOT NULL,
    stopped_at timestamp with time zone,
    task_id integer NOT NULL,
    user_id integer NOT NULL,
    status text NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CHECK (
        status = 'started' AND stopped_at IS NULL
        OR status = 'stopped' AND stopped_at IS NOT NULL
    )
);
CREATE UNIQUE INDEX IF NOT EXISTS works_task_id_user_id_status_idx
    ON works (task_id, user_id, status)
    WHERE status = 'started';

COMMIT;

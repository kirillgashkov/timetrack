BEGIN;

DROP INDEX IF EXISTS times_ended_at_idx;
DROP INDEX IF EXISTS times_started_at_idx;
DROP INDEX IF EXISTS times_user_id_idx;
DROP INDEX IF EXISTS times_task_id_idx;
DROP TABLE IF EXISTS times;

DROP INDEX IF EXISTS tasks_description_trgm_idx;
DROP TABLE IF EXISTS tasks;

DROP INDEX IF EXISTS users_address_trgm_idx;
DROP INDEX IF EXISTS users_patronymic_trgm_idx;
DROP INDEX IF EXISTS users_name_trgm_idx;
DROP INDEX IF EXISTS users_surname_trgm_idx;
DROP INDEX IF EXISTS users_passport_number_trgm_idx;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS pg_trgm;

COMMIT;

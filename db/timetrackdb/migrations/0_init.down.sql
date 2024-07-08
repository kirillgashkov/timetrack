BEGIN;

DROP INDEX IF EXISTS users_address_trgm_idx;
DROP INDEX IF EXISTS users_patronymic_trgm_idx;
DROP INDEX IF EXISTS users_name_trgm_idx;
DROP INDEX IF EXISTS users_surname_trgm_idx;
DROP INDEX IF EXISTS users_passport_number_trgm_idx;

DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS pg_trgm;

COMMIT;

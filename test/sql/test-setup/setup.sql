BEGIN;
    CREATE SCHEMA IF NOT EXISTS test_setup;
    CREATE TABLE IF NOT EXISTS test_setup.setuptable(
        val smallint NOT NULL
    );
    INSERT INTO test_setup.setuptable (val) VALUES (1);
COMMIT;

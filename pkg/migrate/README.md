# go-pg-migrate

Library for migrating the Postgres Database of PHDP services. It uses [golang-migrate V4](https://github.com/golang-migrate/migrate) for the migration.

## Setup Script

`go-pg-migrate` allows to run a setup script before the migration steps that will be handled by `golang-migrate`.
The script is optional and must be called `setup.sql` and be placed in the same folder as the other sql scripts.
The main use case for the setup script is creating an schema that will be used by `golang-migrate` for the migration table itself.
The setup script must be idempotent, as it will be run for every migration (unlike the migration steps that are skipped if the version is already present).

## Postgres foreign-data wrapper

`go-pg-migrate` allows to run additional scripts before and after the migration which are golang templated by a ForeignDatabase struct and the following fields:

- LocalUser string
- DBName    string
- Hostname  string
- Port      uint
- User      string
- Password  string

The scripts are optional and must be called `fdw.up.sql` and `fdw.down.sql` and be placed in the same folder as the other sql scripts. The placeholders can be used like this well-known notation within the scripts: `{{.LocalUser}}`.
The main use case for the scripts is to prepare the database for some foreign data migration like described in [Postgres FDW](https://www.postgresql.org/docs/12/postgres-fdw.html).

## Migration Table

`golang-migrate` needs a table that will contain the migration metadata (current version and the dirty status). This table will be created by the library with the given table name.
However, the schema where the table is created is not configurable for postgres as of version 4 of `golang-migrate`. Instead, the `golang-migrate` library will create the table with the unqualified name, which will have the effect of creating the table in the current schema. Therefore, if the table is intended to be created in a particular schema, that schema needs to be set as the current schema (first element in the search path).

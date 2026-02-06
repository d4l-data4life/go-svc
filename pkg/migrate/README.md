# go-pg-migrate

Library for migrating Postgres databases in services using `pkg/db`. It is built
on top of [golang-migrate V4](https://github.com/golang-migrate/migrate).

## Setup Script

`go-pg-migrate` allows running a setup script before the migration steps handled
by `golang-migrate`. The script is optional and must be called `setup.sql` and
placed in the same folder as the other SQL scripts.

The main use case is creating a schema that will be used by `golang-migrate` for
the migration table itself. The setup script must be idempotent, as it will be
run for every migration invocation (unlike migration steps which are skipped if
the version is already present).

## Postgres Foreign Data Wrapper (FDW)

`go-pg-migrate` can run additional scripts before and after the migration which
are templated by a `ForeignDatabase` struct:

- LocalUser string
- DBName    string
- Hostname  string
- Port      uint
- User      string
- Password  string

The scripts are optional and must be called `fdw.up.sql` and `fdw.down.sql`, and
placed in the same folder as the other SQL scripts. Placeholders can be used via
`{{.LocalUser}}` syntax. The main use case is preparing the database for foreign
data migration (see Postgres FDW docs).

## Migration Table

`golang-migrate` uses a table that contains migration metadata (current version
and dirty status). The table is created with the given name. For Postgres, the
schema is not configurable as of v4, so the table is created in the current
schema (first element in the search path). If the table must live in a specific
schema, that schema must be in the search path.

## Migration Flows

`pkg/db` supports two flows: legacy and versioned.

### Legacy Flow

- Single AutoMigrate (latest models).
- SQL migrations executed via `golang-migrate`:
  - `{version}_{name}.up.sql` / `{version}_{name}.down.sql`

### Versioned Flow

- Interleaves per migration version:
  1. Versioned before script (optional)
  2. AutoMigrate(version) (service implementation)
  3. Versioned after script (optional)
  4. Record version after the full sequence completes

Missing before/after scripts are skipped.

**Supported naming (versioned flow only):**
- Before: `{version}_{name}.before.sql` or `{version}_{name}.before.up.sql`
- After: `{version}_{name}.after.sql` or `{version}_{name}.after.up.sql`

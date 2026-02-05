# go-pg-migrate

Library for migrating Postgres databases in services using `pkg/db`. It is
built on top of `golang-migrate` and supports **two migration logics** used by
`go-svc`.

## Shared Scripts

These scripts are optional and run regardless of which migration logic is used.

- `setup.sql` (idempotent, runs once)
- `fdw.up.sql` and `fdw.down.sql` (optional, runs once around all versions)

## Legacy Logic (as on main)

- Single AutoMigrate (latest models).
- Target‑only before script:
  - `{version}_{name}.before.up.sql`
- SQL migrations executed via `golang-migrate`:
  - `{version}_{name}.up.sql` / `{version}_{name}.down.sql`

## Versioned Logic (new)

- Interleaves per migration version:
  1. `{version}_{name}.before.up.sql` (optional)
  2. AutoMigrate(version) (service implementation)
  3. `{version}_{name}.after.up.sql` (optional)

Missing before/after scripts are skipped.

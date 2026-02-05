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
- SQL migrations executed via `golang-migrate`:
  - `{version}_{name}.up.sql` / `{version}_{name}.down.sql`

## Versioned Logic (new)

- Interleaves per migration version:
  1. Versioned before script (optional)
  2. AutoMigrate(version) (service implementation)
  3. Versioned after script (optional)
  4. Record version after the full sequence completes

Missing before/after scripts are skipped.

**Supported naming (versioned path only):**
- Before: `{version}_{name}.before.sql` or `{version}_{name}.before.up.sql`
- After: `{version}_{name}.after.sql` or `{version}_{name}.after.up.sql`

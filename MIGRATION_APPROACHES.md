# Migration Approaches

This repository supports **two migration logics** for services using `pkg/db`.
The *legacy* logic mirrors what is on `main` today. The *versioned* logic is
new and enables interleaving per migration version.

## 1) Legacy Migration Logic (as on main)

Use when you want a single AutoMigrate call and minimal changes to existing
services.

**API**
- Provide `WithMigrationFunc(func(*gorm.DB) error)`
- Optionally set `WithMigrationVersion(version)`

**Behavior (main)**
- Executes a **target‑only** before script if present:
  - `{version}_{name}.before.up.sql`
- Runs **one** AutoMigrate (latest models).
- Runs `setup.sql` (idempotent), then `fdw.up.sql` (optional).
- Runs numbered SQL migrations via `golang-migrate`:
  - `{version}_{name}.up.sql` / `{version}_{name}.down.sql`
- Runs `fdw.down.sql` (optional).

## 2) Versioned Migration Logic (new)

Use when you need **interleaving** per migration version.

**API**
- Provide `WithVersionedMigrationFunc(func(*gorm.DB, uint) error)`
- Optionally set `WithMigrationVersion(version)`

**Behavior**
- Runs `setup.sql` once (idempotent).
- Runs `fdw.up.sql` / `fdw.down.sql` once (optional).
- For each version from current+1 to target:
  1. `{version}_{name}.before.up.sql` (optional)
  2. `AutoMigrate(version)` (service implementation)
  3. `{version}_{name}.after.up.sql` (optional)

**Notes**
- Missing before/after scripts are skipped.
- If no after script exists for a version, a no‑op migration is applied so the
  migration table advances.

## File Naming Summary

Legacy (main):
- Before: `{version}_{name}.before.up.sql`
- After: `{version}_{name}.up.sql`

Versioned (new):
- Before: `{version}_{name}.before.up.sql`
- After: `{version}_{name}.after.up.sql`

Shared:
- Setup: `setup.sql`
- FDW: `fdw.up.sql`, `fdw.down.sql`

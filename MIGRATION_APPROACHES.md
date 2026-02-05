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
  1. Versioned before script (optional)
  2. `AutoMigrate(version)` (service implementation)
  3. Versioned after script (optional)
  4. Record version **after** the full sequence completes

**Notes**
- Missing before/after scripts are skipped.
- The version bump represents completion of before + AutoMigrate + after.
- **Supported naming (versioned path only):**
  - Before: `{version}_{name}.before.sql` **or** `{version}_{name}.before.up.sql`
  - After: `{version}_{name}.after.sql` **or** `{version}_{name}.after.up.sql`

## File Naming Summary

Legacy (main):
- SQL migrations: `{version}_{name}.up.sql`

Versioned (new):
- Before: `{version}_{name}.before.sql` or `{version}_{name}.before.up.sql`
- After: `{version}_{name}.after.sql` or `{version}_{name}.after.up.sql`

Shared:
- Setup: `setup.sql`
- FDW: `fdw.up.sql`, `fdw.down.sql`

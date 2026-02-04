# Migration Approaches

This repository provides a migration flow that combines GORM AutoMigrate with
versioned SQL migrations powered by `golang-migrate`. This document describes
what is supported today, what is not supported, and the extension approach we
chose for "before AutoMigrate" use cases.

## Current Flow (Baseline)

For a service using `pkg/db`:

1. Optional `setup.sql` is executed by `pkg/migrate` when migrations are run.
2. GORM AutoMigrate runs once (migrates to latest model definitions in code).
3. `golang-migrate` runs numbered `*.up.sql` migrations to the target version.
4. Optional `fdw.up.sql` / `fdw.down.sql` scripts run around the numbered steps.

Key properties:
- AutoMigrate is **not version-aware** and always migrates to the latest code.
- SQL migrations are **versioned** and ordered by filename version.
- `setup.sql` must be **idempotent** because it runs on every migration call.

## Supported

- Versioned SQL migrations: `{version}_{name}.up.sql` / `{version}_{name}.down.sql`
- Optional `setup.sql` and FDW scripts
- Migrating to a specific target version (`MigrateDB(ctx, version, ...)`)

## Not Supported (by design)

- Per-version interleaving of AutoMigrate with SQL steps
- Version-specific AutoMigrate (GORM does not support this)
- Non-idempotent pre-AutoMigrate scripts without tracking

## Extension: Target-Version Before Script

Use case: GORM AutoMigrate cannot perform a change you need before the current
target version. To support this while keeping AutoMigrate, we allow a **single
target-version before script** that runs immediately before AutoMigrate.

Naming:
- `{version}_{name}.before.up.sql`

Rules:
- The before script runs **only** when `{version}` matches the target version.
- It is **not tracked**; it must be **idempotent**.
- All `.before.` files are excluded from the post-AutoMigrate `golang-migrate`
  run to prevent accidental execution as a normal migration.

This is intentionally limited to avoid interleaving issues with AutoMigrate.

## Examples (File Lists and Execution Sequences)

### Example A: Single-Version Upgrade (v6 -> v7)

Files:
```
006_previous.up.sql
007_add_column.up.sql
007_add_column.before.up.sql
```

Target version: `7`

Execution sequence:
1. `007_add_column.before.up.sql` (idempotent, target-only)
2. GORM AutoMigrate (to latest schema, i.e., v7)
3. `006_previous.up.sql`
4. `007_add_column.up.sql`

### Example B: Multi-Version Jump (v2 -> v5)

Files:
```
003_add_user.up.sql
004_add_order.up.sql
005_add_invoice.up.sql
005_add_invoice.before.up.sql
```

Target version: `5`

Execution sequence:
1. `005_add_invoice.before.up.sql` (target-only)
2. GORM AutoMigrate (to latest schema, i.e., v5)
3. `003_add_user.up.sql`
4. `004_add_order.up.sql`
5. `005_add_invoice.up.sql`

Note why we do **not** run `003.before.up.sql` or `004.before.up.sql`:
AutoMigrate cannot be constrained to v3 or v4, so interleaving would be
incorrect. The model is intentionally limited to the **target-only** before
script.

### Example C: No Before Script

Files:
```
001_init.up.sql
002_add_user.up.sql
```

Target version: `2`

Execution sequence:
1. GORM AutoMigrate
2. `001_init.up.sql`
3. `002_add_user.up.sql`

## Extension Options (Tradeoffs)

1. **Full SQL migrations** (no AutoMigrate)
   - Pros: Deterministic, version-aware ordering, supports complex changes.
   - Cons: More SQL to write and maintain; slower for rapid dev.

2. **Separate migration streams** (before/after tables)
   - Pros: Versioned before + after with tracking.
   - Cons: Still cannot interleave AutoMigrate by version.

3. **Version-aware migrations only**
   - Requires a migration system that can apply model changes incrementally,
     which GORM AutoMigrate does not provide.

The target-only before script is the pragmatic compromise: it preserves
backwards compatibility, maintains the existing flow, and supports critical
pre-AutoMigrate fixes without re-architecting the migration system.

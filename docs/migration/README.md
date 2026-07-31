# Migration Approaches

This repository supports **two migration flows** for services using `pkg/db`.
Choose the versioned flow if you need per‑version interleaving, or the legacy
flow for a single AutoMigrate pass.

## 1) Versioned Migration Flow

Use when you need **interleaving** per migration version.

**API**
- `WithVersionedMigrationFunc(func(*gorm.DB, uint) error)`
- Optional: `WithMigrationVersion(version)`

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
- **Recording a version** means updating the `migrations` table used by
  `golang-migrate` with the new version number.
- **Supported naming (versioned flow only):**
  - Before: `{version}_{name}.before.sql` **or** `{version}_{name}.before.up.sql`
  - After: `{version}_{name}.after.sql` **or** `{version}_{name}.after.up.sql`

**Diagram**
```mermaid
flowchart TB
  %% Versioned Migration Logic

  B0["Start migrations"] --> B1["setup.sql (optional, idempotent)"]
  B1 --> B2["fdw.up.sql (optional)"]
  B2 --> V1["For each version v = current+1 .. target"]
  V1 --> V2["{version}_{name}.before.sql or .before.up.sql (optional)"]
  V2 --> V3["AutoMigrate(version)"]
  V3 --> V4["{version}_{name}.after.sql or .after.up.sql (optional)"]
  V4 --> V5["Record version v (migrations table)"]
  V5 -.-> V1
  V5 --> B3["fdw.down.sql (optional)"]
```

## 2) Legacy Migration Flow

Use when you want a single AutoMigrate call and minimal changes to existing
services.

**API**
- `WithMigrationFunc(func(*gorm.DB) error)`
- Optional: `WithMigrationVersion(version)`

**Behavior**
- Runs **one** AutoMigrate (latest models).
- Runs `setup.sql` (idempotent), then `fdw.up.sql` (optional).
- Runs numbered SQL migrations via `golang-migrate`:
  - `{version}_{name}.up.sql` / `{version}_{name}.down.sql`
- Runs `fdw.down.sql` (optional).

**Diagram**
```mermaid
flowchart TB
  %% Legacy Migration Logic

  A0["Start migrations"] --> A1["AutoMigrate (once, latest models)"]
  A1 --> A2["setup.sql (optional, idempotent)"]
  A2 --> A3["fdw.up.sql (optional)"]
  A3 --> L1["For each version v = current+1 .. target"]
  L1 --> L2["{version}_{name}.up.sql"]
  L2 -.-> L1
  L2 --> A4["fdw.down.sql (optional)"]
```

## File Naming Summary

Versioned flow:
- Before: `{version}_{name}.before.sql` or `{version}_{name}.before.up.sql`
- After: `{version}_{name}.after.sql` or `{version}_{name}.after.up.sql`

Legacy flow:
- SQL migrations: `{version}_{name}.up.sql`

Shared:
- Setup: `setup.sql`
- FDW: `fdw.up.sql`, `fdw.down.sql`

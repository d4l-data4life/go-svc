# Migration flows

Choose exactly one callback:

- `WithMigrationFunc(func(*gorm.DB) error)` retains legacy behavior: AutoMigrate once, then golang-migrate numbered `.up.sql`/`.down.sql` scripts.
- `WithVersionedMigrationFunc(func(*gorm.DB, uint) error)` opts into the serialized, forward-only flow below.

`WithMigrationVersion(0)` calls the selected callback once (with version 0 for the versioned callback), without SQL or version tracking. Services should interpret 0 as AutoMigrate current models. This is useful for isolated model tests, not production migration verification.

## Versioned execution

1. Pin a database connection and acquire the PostgreSQL advisory lock keyed identically to golang-migrate (database/schema/migration table). Wait up to 30 seconds, or until the startup context is cancelled.
2. Read the migration record under the lock. Reject dirty state and versions newer than the configured target. `WithMinimumMigrationVersion(N)` optionally rejects existing starting versions below N.
3. If no version exists, require an otherwise empty current schema. The metadata table alone is allowed. With the default `MigrationStartFromZero=false`, run current models once and record the target, without historical hooks. With `true`, replay from version 1; this is incompatible with a positive minimum starting version.
4. For an existing supported database, run versions `current+1 … target`: mark N dirty, optional before(N), AutoMigrate(N), optional after(N), record N clean. The final version becomes clean only after shared FDW cleanup succeeds. Setup and FDW scripts run once for a pending migration under the same lock; they do not run for an already-current database.
5. Release the migration connection and lock. A second startup rereads the version after acquiring the lock. Versioned failures always stop initialization, regardless of the legacy halt-on-error option.

Hooks, callback and metadata writes use the pinned connection. Callbacks must use the supplied GORM handle synchronously; do not open independent connections or launch background migration work. Lock cleanup releases the migration session only, not the shared application pool. There is no temporary SQL source directory and no nested Force lock.

The current schema must exist before initialization. Like the legacy postgres version store, metadata lives in PostgreSQL's current schema. GORM table prefixes alone do not change PostgreSQL search_path; configure these consistently. Schemas other than public require their own deployment validation.

## SQL hooks

Place hooks in `sql/`:

- `N_name.before.sql` or `N_name.before.up.sql`
- `N_name.after.sql` or `N_name.after.up.sql`

At most one file per version/phase is allowed; duplicate or competing naming variants fail. Missing phases are optional. Discovery does not recurse into archived legacy directories. Keep the active SQL directory in the built application image and verify packaging when adding hooks.

`setup.sql`, `fdw.up.sql` and `fdw.down.sql` remain optional. Versioned startup has no ForeignDatabase template configuration; callers needing that facility can still use the separate legacy Migration API.

## Failure and recovery

A whole migration step is not a single transaction. Before/AutoMigrate/after may each commit work. A crash or failure leaves the attempted version dirty and stops subsequent startups before any replay. This includes errors during fresh initialization, the version write, and final FDW cleanup. Earlier completed versions can remain clean.

Hooks should still be idempotent where possible for deliberate recovery, but individual idempotency does not guarantee whole-step replay: an after-hook may remove a column a before-hook needs. Never automatically clear dirty state. Inspect the database and logs, then restore a known backup or perform a reviewed repair. Record a clean version only after verifying its complete schema/data postconditions, while holding the migration lock and controlling other writers. A hook's own BEGIN/COMMIT does not include the callback or metadata write.

A lock serializes participating migration runners, not normal application traffic. Deployment owners must ensure old replicas remain compatible or arrange a maintenance rollout. Forward-only runtime migration requires a backup/restore rollback procedure.

## Verification

Run `scripts/test-versioned-migrations.sh` (also called by CI). It creates its own PostgreSQL 15 container on an ephemeral loopback port and runs the concurrency/recovery suite with the race detector. It does not reuse an existing local database. Coverage includes fresh initialization/restart, phase order, competing startups, cancellation while waiting, dirty rejection, before/auto/after/record/cleanup failures, and unsupported or populated unversioned states.

The general test suite's existing local DB helpers accept `GO_SVC_TEST_PORT` to avoid assuming port 5432 is free. The `test` package uses its existing PG_* settings. Use only disposable databases for these tests.

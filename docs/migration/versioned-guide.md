# Versioned migrations: adopting, changing the schema, recovering

This guide is for services that use, or want to use, the versioned flow described in [README.md](README.md). It covers three situations:

- [Part 1](#part-1-adopting-the-versioned-flow): switching an existing service from `WithMigrationFunc` (a new service starts at step 1.3);
- [Part 2](#part-2-changing-the-schema): changing the schema later — when to bump the migration version and what to do;
- [Part 3](#part-3-recovering-from-a-failed-migration): getting a service running again after a migration failed.

## Part 1: Adopting the versioned flow

### 1.1 Check the prerequisites

- **Driver:** the service opens its database with go-svc's default driver (`DefaultPostgresDriver`, pgx). The versioned flow refuses other drivers, because it inspects pgx's transaction status after every step.
- **Schema:** migration metadata lives in PostgreSQL's current schema. Keep the GORM table prefix and the connection's `search_path` consistent.
- **Packaging:** the service image contains a `sql/` directory, even an empty one. Every upgrade looks for hooks there and fails if the directory is missing. A fresh database skips that lookup, so tests that only start from an empty database don't catch it.
- **Startup time:** under `standard.Main`, a service has 120 seconds from start until its database must be ready. Connecting, waiting up to 30 seconds for the migration lock, and the migration itself all count towards it. A migration still running after that is cancelled and left dirty. See [Startup time limit and dirty state](README.md#startup-time-limit-and-dirty-state).

### 1.2 Choose the first versioned target: N+1, never N

Look up the version your databases record today (`SELECT version, dirty FROM migrations`) and the `MigrationVersion` your service configures. Call it N.

Set the first versioned target to **N+1** and the minimum starting version to **N**. Don't keep N as the target.

Why: the legacy flow runs `AutoMigrate` with the current models on every start, regardless of the recorded version, so model changes have accumulated without a version bump. The versioned flow does nothing when the recorded version equals the target. With target N, databases at N would never receive those accumulated changes. Step N+1 applies the current models once, explicitly.

Databases below N are refused. Upgrade them with the last legacy release first, or plan a separate route.

### 1.3 Write the versioned callback

The callback receives the version being migrated. Apply the models that are valid **at that version**, and reject versions you don't support:

```go
func MigrationFunc(db *gorm.DB, version uint) error {
	switch version {
	case 0, CurrentVersion: // 0: untracked mode used by model tests
		return db.AutoMigrate(models.All()...)
	default:
		return fmt.Errorf("unsupported migration version %d", version)
	}
}
```

Keep the model list in one function (`models.All()` here) shared by migration and tests, so a new model can't be forgotten in one of them.

### 1.4 Wire it up

```go
db.NewConnection(
	// ... connection options as before ...
	db.WithVersionedMigrationFunc(models.MigrationFunc),
	db.WithMigrationVersion(CurrentVersion),        // N+1 for the first switch
	db.WithMinimumMigrationVersion(MinimumVersion), // N
)
```

Remove `WithMigrationFunc`; configuring both callbacks is an error. The versioned flow always stops startup on errors, whatever `WithMigrationHaltOnError` says. Leave `MigrationStartFromZero` at its default (false), so an empty database is created directly at the target.

Move the legacy numbered `*.up.sql`/`*.down.sql` scripts into a subdirectory such as `sql/legacy/`. The versioned flow doesn't execute them, and keeping them out of `sql/` avoids confusion with hooks.

Add a schema test that compares the live models with a frozen copy of the current version's models (see [2.2](#22-what-to-do-when-bumping)). From now on it fails whenever someone changes the schema without bumping the version.

### 1.5 Test

- Existing model tests that use `TXDBPostgresDriver` without a migration version keep working: they run the callback with version 0.
- Tests of real migration behaviour need their own disposable database. Running the versioned flow with a target version through `TXDBPostgresDriver` is refused: inside one test-wide transaction, neither the lock nor the dirty marker means anything.
- Test the upgrade from every starting state your environments actually have: an empty database, and each released version recorded in an environment. Databases created by different releases can share a version number but differ in schema.

### 1.6 Roll out

- Before deploying, check each environment's recorded version and dirty flag, and take a backup.
- Migrations are forward-only. Rolling back means restoring the backup and deploying the previous release.
- The lock coordinates migrating replicas only. Old replicas keep serving traffic during a rolling update, so they must cope with the new schema.

## Part 2: Changing the schema

### 2.1 When to bump the migration version

**Bump** whenever the database must change:

- a persisted model is added or removed, or its table name changes;
- a persisted field is added, removed or renamed;
- a field's type, size, precision, nullability or default changes, including through `gorm` tags such as `type:uuid` or `size:1024`;
- an index, unique constraint, foreign key or check constraint is added, removed or changed;
- existing data must be moved, filled, converted or cleaned up, even without any model change.

**Don't bump** for changes the database never sees:

- fields excluded with `gorm:"-"`, JSON tags, methods, validation, queries.

Rule of thumb: if `AutoMigrate` against the previous schema would issue any DDL, or if any row must change, bump. The schema test from 1.4 fails when a model change slips through without a bump.

**One version per change set.** Several changes in the same release can share one new version, as long as that version hasn't been deployed anywhere yet. Once a version has run in **any** environment (including dev or sandbox), its meaning is fixed: further changes need the next version.

### 2.2 What to do when bumping

1. **Bump** `CurrentVersion` from N to N+1.
2. **Freeze the new shape.** Add a frozen copy of the models as they are at N+1, for example `models/snapshots/vN1`, and point the schema test at it. Never edit an existing frozen copy.
3. **Route the callback.** `case 0, N+1:` uses the live models. If databases can still start below N, an intermediate step N uses the frozen vN models: live models will keep changing, and step N must keep producing the schema of version N.
4. **Add SQL hooks only if data needs them:**
   - `N+1_name.before.sql` runs **before** the model change, against the schema of version N, for example to prepare or copy data;
   - `N+1_name.after.sql` runs **after** it, for example to fill a new column or clean up.
   - Version hooks don't run on a fresh database. It is created directly from the models at the target; only `setup.sql` and the FDW scripts run there. Anything a fresh database needs must come from the models or from `setup.sql`, not from a version hook.
   - A hook may use its own `BEGIN`/`COMMIT`, but must not leave a transaction open; the step fails otherwise.
5. **Destructive changes go in two releases.** Old replicas keep running during a rolling update. To drop or rename a column: first release stops using it (and adds the replacement); a later release removes it.
6. **Raise the minimum** (`WithMinimumMigrationVersion`) when you stop supporting older starting versions. Frozen copies below the minimum can then be deleted.
7. **Test:**
   - the schema test passes against the new frozen copy;
   - an upgrade from version N with representative data preserves it;
   - an upgrade from every older version still recorded in some environment;
   - a fresh database initialises at N+1;
   - a second start at N+1 does nothing.
8. **Note it in the release:** the new migration version, whether it is destructive, and how long it took on realistic data (it must fit into the startup time).

## Part 3: Recovering from a failed migration

### 3.1 What you see

Every start fails with one of:

- `database migration is dirty at version N; inspect and recover before restarting` — a migration step failed or was interrupted;
- `... left a non-idle transaction ...; migration remains dirty` — a hook or the callback left a transaction open (the step is also dirty);
- `database version N is newer than supported target M` — a newer release has already migrated this database;
- `unsupported database migration version N; lowest supported starting version is M`, or `cannot baseline a populated database without migration metadata` — the database is in a state this release doesn't support. The application's tables and data were not touched; the library may only have created an empty `migrations` table.

Only "dirty" needs the recovery below. For the others, deploy a release that supports the database's version, or plan a migration route for it.

### 3.2 Why "dirty" needs a person

A step is not one transaction: the before-hook, each `AutoMigrate` statement and the after-hook may each have committed. "Dirty at version N" means: step N started, and it is unknown how far it got. It can even mean that step N finished and only cleanup or recording failed. Replaying it automatically could break the database, so every start refuses until someone has decided.

### 3.3 Get back to "green"

1. **Stop the rollout.** New replicas fail at startup anyway; old replicas (if any) keep serving. Scale the new release down to 0 while you work, so no start interferes.
2. **Find out what failed.** Read the failing start's error and the log lines before it. Errors from before- and after-hooks name the hook file; for the callback, setup, cleanup or version recording, the error text and the preceding log lines are your best clues. Check `SELECT version, dirty FROM migrations;`.
3. **Find out where the database is.** Compare the actual schema (and, for hook steps, the data) with the expected shapes at N-1 and N: the frozen models and the hook SQL for step N.
4. **Repair, usually by hand:**
   - **Finish step N** (the usual case). Apply the rest of step N with reviewed SQL, until the database matches version N.
   - **Undo step N**, if finishing isn't possible, for example because the hook or model change itself was wrong. Revert step N's partial changes with reviewed SQL, until the database matches version N-1.
   - **Restore the backup** taken before the deployment, only if the state can't be understood. This loses writes made since the backup.
5. **Verify before clearing the flag.** The service doesn't check the schema itself; once the flag is clean, it trusts the database. So compare the repaired schema with a reference:
   1. Create an empty database on the same PostgreSQL major version, with the same extensions available. Start the release whose target is the version you're repairing to (N, or N-1 after an undo) against it, with the same configuration. That release must use the versioned flow. It then creates its target schema from its models, plus whatever `setup.sql` and the FDW scripts create.
   2. Dump both schemas: `pg_dump --schema-only --no-owner --no-privileges`.
   3. Compare the dumps, for example with `diff`. The reference shows what the models require; it doesn't contain objects that only version hooks or older scripts created, so the repaired database may legitimately have more. Every difference must be explained. A missing or different table, column, type, constraint or index means the repair isn't finished.

   For data changes made by hooks, also check the data the step was meant to produce, for example that every row received its new value.
6. **Mark the version clean**, only after step 5, and only while no instance of the service is running.
   - First check `SELECT version, dirty FROM migrations;`. It must return exactly one row, with version N and `dirty = true`. If it doesn't, stop and find out why before changing anything.
   - Then update it with the concrete numbers, and check that PostgreSQL reports `UPDATE 1`:

   ```sql
   -- after finishing step N (example: N = 9):
   UPDATE migrations SET dirty = false WHERE version = 9 AND dirty;
   -- after undoing step N (example: back to 8):
   UPDATE migrations SET version = 8, dirty = false WHERE version = 9 AND dirty;
   ```

   - After a restore, don't update anything. Check instead that the restored row shows the expected version with `dirty = false`.

7. **Fix the cause** in the code, hook or data, if step N will run again (after an undo or a restore).
8. **Start the service again.** With a clean version below the target, it runs the remaining steps automatically, as long as that version isn't below the configured minimum. With a clean version equal to the target, it starts without migrating.

Never set `dirty = false` just to get the service started: the service never checks the schema again, and every later step builds on whatever is there. Write down what you did and why; that record is the only history of a manual repair.

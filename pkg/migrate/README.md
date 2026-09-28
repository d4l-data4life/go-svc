# PostgreSQL migrations

See [migration flows and recovery](../../docs/migration/README.md) for the service-facing configuration and operational contract.

The legacy `Migration.MigrateDB` API executes setup/FDW SQL and delegates numbered up/down files to golang-migrate. `MigrateToVersion` remains part of that live legacy path.

The versioned runner uses `OpenVersionStore` to own one PostgreSQL connection and a compatible advisory lock. `Version` and `SetVersion` run under that lock; `NewVersionedMigration` executes hooks on the same connection. `Close` rolls back any unfinished explicit hook transaction, releases the lock and returns the connection. Cleanup failure discards the connection instead of returning a possibly locked session to the pool. It never closes the application's shared pool.

Versioned hooks use before/after names and record dirty intent before mutation. A failed step is **not automatically retried**. Keep scripts idempotent where practical, but require inspected recovery after any dirty migration. Setup runs only when work is pending, and final cleanup must succeed before the target is clean. See the linked recovery instructions before changing metadata manually.

With a configured target of 0, service callbacks run AutoMigrate current models without version tracking; no version store or SQL hook runner is created. This is distinct from replaying an empty database from version 1.

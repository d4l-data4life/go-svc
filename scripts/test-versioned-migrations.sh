#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
name="go-svc-migration-test-$$"
trap 'docker rm -f "$name" >/dev/null 2>&1 || true' EXIT
# The dedicated container is independent of the repository's legacy test service.
docker run --rm -d --name "$name" -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=migration_test \
  -p 127.0.0.1::5432 postgres:15 >/dev/null
ready=false
for _ in {1..60}; do
  if docker exec "$name" pg_isready -U postgres -d migration_test >/dev/null 2>&1; then ready=true; break; fi
  sleep 1
done
[ "$ready" = true ] || { echo 'Local PostgreSQL did not become ready' >&2; exit 1; }
export GO_SVC_MIGRATION_TEST_PORT
a=$(docker port "$name" 5432/tcp)
GO_SVC_MIGRATION_TEST_PORT="${a##*:}"
go test -v -race ./pkg/db ./pkg/migrate -run 'Test(MigrationConfiguration|MigrationMinimum|LegacyTestHelper|VersionedMigration|WithMigration|Migration_parse|Find)' -count=1 -timeout=120s

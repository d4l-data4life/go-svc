GO_VERSION := 1.22
CILINT_VERSION := v1.58

DB_CONTAINER_NAME=go-svc-postgres
DB_PORT=5432

define LOCAL_VARIABLES
PG_HOST=localhost \
PG_PORT=$(DB_PORT) \
PG_NAME=test \
PG_USER=user \
PG_PASSWORD=test \
PG_USE_SSL="false"
endef

.PHONY: test-gh-action
test-gh-action: ## Run tests natively in verbose mode and storing the results in out file
	$(LOCAL_VARIABLES) \
	go test -timeout 300s -cover -covermode=atomic -v ./... 2>&1 | tee test-result.out

.PHONY: test
test: lint unit-test-postgres

.PHONY: lint
lint:
	@golangci-lint --version
	golangci-lint run ./...

.PHONY: unit-test-postgres
unit-test-postgres: docker-database local-test clean

.PHONY: local-test lt
local-test lt:      ## Run tests natively
	$(LOCAL_VARIABLES) \
	go test -timeout 30s -cover -covermode=atomic ./...

.PHONY: docker-database
docker-database ddb: clean ## Run database in Docker
	docker run --name $(DB_CONTAINER_NAME) -d \
		-e POSTGRES_DB=test \
		-e POSTGRES_USER=user \
		-e POSTGRES_PASSWORD=test \
		-p $(DB_PORT):5432 postgres
	@until docker container exec -t $(DB_CONTAINER_NAME) pg_isready; do \
		>&2 echo "Postgres is unavailable - waiting for it... 😴"; \
		sleep 1; \
	done

.PHONY: clean
clean:
	-docker rm -f $(DB_CONTAINER_NAME)

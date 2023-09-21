GO_VERSION := 1.20
CILINT_VERSION := v1.53

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

test-gh-action: ## Run tests natively in verbose mode
	$(LOCAL_VARIABLES) \
	go test -timeout 300s -cover -covermode=atomic -v ./... 2>&1 | tee test-result.out

test: lint test-gh-action

lint:
	DOCKER_BUILDKIT=1 \
		docker build \
			--build-arg CILINT_VERSION=${CILINT_VERSION} \
			--build-arg GITHUB_USER_TOKEN \
			-t "go-svc:lint" \
			-f build/lint.Dockerfile \
			.
		docker run --rm "go-svc:lint"

.PHONY: docker-database
docker-database: clean ## Run database in Docker
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

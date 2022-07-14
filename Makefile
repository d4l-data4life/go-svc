GO_VERSION := 1.18
CILINT_VERSION := v1.46

PKG="./..."

define LOCAL_VARIABLES
endef

parallel-test:               ## Specifies space-separated list of targets that can be run in parallel
	@echo "lint unit-test integration-test"

test: lint unit-test integration-test

lint:
	DOCKER_BUILDKIT=1 \
		docker build \
			--build-arg CILINT_VERSION=${CILINT_VERSION} \
			--build-arg GITHUB_USER_TOKEN \
			-t "go-svc:lint" \
			-f build/lint.Dockerfile \
			.
		docker run --rm "go-svc:lint"

unit-test:          ## Run unit tests inside the Docker image
	docker build \
		--build-arg GO_VERSION=${GO_VERSION} \
		--build-arg GITHUB_USER_TOKEN \
		-t go-svc:test \
		-f build/test.Dockerfile \
		.
	docker run --rm go-svc:test

local-test:      ## Run tests natively
	go vet ./... && \
	go test -v -cover -covermode=atomic ./pkg/...

.PHONY: integration-test
integration-test: clean integration-test-dirty clean

.PHONY: integration-test-dirty
integration-test-dirty:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 \
	TARGET=integration-test \
		docker-compose \
		--file build/docker-compose.yml \
		build \
		--build-arg GO_VERSION="$(GO_VERSION)"

	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 \
	TARGET=integration-test \
	docker-compose \
		--file build/docker-compose.yml \
		up \
		--attach-dependencies \
		--exit-code-from migrate-test \
		migrate-test

scan-docker-images:
	@echo ""

docker-push:
	true

.PHONY: clean
clean:
	docker-compose \
		-f build/docker-compose.yml \
		down --volumes

docker-build:
	true

build:
	true

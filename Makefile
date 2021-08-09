GO_VERSION := 1.16
CILINT_VERSION := v1.41

PKG="./..."

define LOCAL_VARIABLES
endef

parallel-test:               ## Specifies space-separated list of targets that can be run in parallel
	@echo "lint unit-test"

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

integration-test:
	TARGET=code docker-compose \
		-f build/docker-compose.yml \
		build \
		--build-arg GO_VERSION="$(GO_VERSION)" migrate-test
	VERBOSE=false docker-compose \
		-f build/docker-compose.yml \
		run --rm migrate-test \
		go test -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./test

scan-docker-images:
	@echo ""

docker-push:
	true

clean:
	docker-compose \
		-f build/docker-compose.yml \
		down --volumes

docker-build:
	true

build:
	true

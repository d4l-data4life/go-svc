GO_VERSION := 1.15
CILINT_VERSION := v1.30

PKG="./..."

define LOCAL_VARIABLES
endef

parallel-test:               ## Specifies space-separated list of targets that can be run in parallel
	@echo "lint unit-test"

test: lint unit-test

lint:
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
	go test -v -cover -covermode=atomic ./...

scan-docker-images:
	@echo ""

docker-push:
	true

clean:
	true

docker-build:
	true

build:
	true

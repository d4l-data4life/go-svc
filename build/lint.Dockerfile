# syntax = docker/dockerfile:1-experimental
ARG CILINT_VERSION

# LINTER stage: get the linter executable
FROM golangci/golangci-lint:${CILINT_VERSION} AS lint-base

FROM lint-base AS lint
COPY --from=lint-base /usr/bin/golangci-lint /usr/bin/golangci-lint

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Set the environment variables for the `go` command.
ENV GOOS=linux GOARCH=amd64

# Import the code from the context.
COPY ./ ./

CMD ["golangci-lint", "run", "-v", "--timeout=3m"]
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    golangci-lint run -v --timeout 10m0s

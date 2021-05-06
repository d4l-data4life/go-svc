# syntax = docker/dockerfile:1-experimental
ARG CILINT_VERSION

# LINTER stage: get the linter executable
FROM golangci/golangci-lint:${CILINT_VERSION} AS lint-base

FROM lint-base AS lint
COPY --from=lint-base /usr/bin/golangci-lint /usr/bin/golangci-lint

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Use github token auth to be able to access internal repositories
ARG GITHUB_USER_TOKEN
RUN git config --global url."https://${GITHUB_USER_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

# Set the environment variables for the `go` command.
ENV GOPRIVATE="github.com/gesundheitscloud/*"
ENV GOOS=linux GOARCH=amd64

# Import the code from the context.
COPY ./ ./

CMD ["golangci-lint", "run", "-v", "--timeout=3m"]
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    golangci-lint run -v --timeout 10m0s

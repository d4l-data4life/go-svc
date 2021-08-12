# syntax = docker/dockerfile:1-experimental

ARG IMG_BASE=phdp-snapshots.hpsgc.de/
ARG GO_VERSION=1.16

FROM ${IMG_BASE}golang:${GO_VERSION}-alpine AS code

RUN apk add --no-cache ca-certificates git gcc musl-dev make tzdata

WORKDIR /src
COPY go.* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Add docker-compose-wait tool -------------------
# This is used only for integration-test and is not included in the final image
ENV WAIT_VERSION 2.8.0
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/$WAIT_VERSION/wait /wait
RUN chmod +x /wait
RUN echo "da75829985a9d2cccb072a77168356708901ca9ec767cb5605eb79b4758d8d00  /wait" > /shasums.txt
RUN sha256sum -c /shasums.txt

COPY ./ ./

FROM code AS integration-test
WORKDIR /src
ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod vendor

CMD ["sh", "-c", "/wait && go test -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... ./test"]

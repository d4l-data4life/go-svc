# Accept the Go version for the image to be set as a build argument.
ARG GO_VERSION=1.16

# First stage: code and flags.
# This stage is used for running the tests with a valid Go environment, with
# `/vendor` directory support enabled.
FROM golang:${GO_VERSION} AS code

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Import the code from the context.
COPY ./ ./

# Second stage: build the executable
FROM golang:${GO_VERSION}-alpine AS builder

# Create the user and group files that will be used in the running container to
# run the process an unprivileged user.
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

# Import the Certificate-Authority certificates for the app to be able to send
# requests to HTTPS endpoints.
RUN apk add --no-cache ca-certificates

# Accept the version of the app that will be injected into the compiled
# executable.
ARG APP_VERSION=undefined

# Set the environment variables for the build command.
ENV CGO_ENABLED=0

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /bin

# Import the code from the first stage.
COPY --from=code /src ./

# Build the executable to `/app`. Mark the build as statically linked and
# inject the version as a global variable.
RUN go build \
    -installsuffix 'static' \
    -ldflags "-X main.Version=${APP_VERSION}" \
    -o /app \
    ./migrate

# FINAL stage: the running container
# `gcr.io/distroless/base:latest` is like `scratch` but Anchore can recognize it (Anchore does not recognize scratch as distro and issues a warning)
# What it contains: https://github.com/GoogleContainerTools/distroless/blob/master/base/README.md
FROM gcr.io/distroless/base:latest AS final

ARG GO_VERSION=undefined
ARG APP_VERSION=undefined
ARG GIT_COMMIT=undefined
ARG BUILD_DATE=undefined

LABEL org.label-schema.build-date="$BUILD_DATE"
LABEL org.label-schema.name="mail"
LABEL org.label-schema.description="Sends emails via mailjet"
LABEL org.label-schema.vcs-url="https://github.com/gesundheitscloud/phdp-mail"
LABEL org.label-schema.vcs-ref="$GIT_COMMIT"
LABEL org.label-schema.vendor="data4life gGmbH"
LABEL org.label-schema.version="$APP_VERSION"
LABEL org.label-schema.schema-version="1.0"
LABEL go-version="$GO_VERSION"

# Declare the port on which the application will be run.
EXPOSE 8080

# Import the user and group files.
COPY --from=builder /user/group /user/passwd /etc/

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the compiled executable frmo the second stage.
COPY --from=builder /app /app

# Run the container as an unprivileged user.
USER nobody:nobody

ENTRYPOINT ["/app"]

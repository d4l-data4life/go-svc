# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

- [logging] Format of logs and omit almost all empty fields

### Deprecated

### Removed

- [logging] Remove generic event types

### Fixed

### Security

## [v1.80.1] - 2024-01-08

### Removed

- [db2] SQlite support

## [v1.80.0] - 2023-12-04

### Added

- [ticket] package to facilitate authorization for downloads from data-receiver

## [v1.79.0] - 2023-10-18

### Changed

- Upgrade dependencies

### Removed

- [db] package for gorm v1

## [v1.78.0] - 2023-06-14

### Added

- [standard] GRPC default functions for ListenAndServe and Gateway

## [v1.77.0] - 2023-05-12

### Added

- [db2] Option to set schema name for gorm

### Fixed

- [db2] Normal connection not using `DefaultPostgresDriver`

## [v1.76.0] - 2023-03-09

### Changed

- [clients] User preferences single setting get does not error for setting not found

## [v1.75.0] - 2023-02-15

### Added

- [clients] flowmailer support for raw emails and attachments

## [v1.74.0] - 2023-02-10

### Added

- [clients] more detailed consent service mock

## [v1.73.0] - 2023-02-07

### Added

- New jwt based authentication middleware for azure AD

## [v1.72.0] - 2023-02-01

### Added

- [clients] Flowmailer

## [v1.71.1] - 2023-01-20

### Fixed

- missing `TxSync` in Listmonk client interface

## [v1.71.0] - 2023-01-19

### Added

- [clients] Listmonk client support for synchronous Transactional Message

## [v1.70.0] - 2023-01-18

### Added

- [clients] fromName to Listmonk Transactional Message

### Changed

- [clients] cleanup cds-notification client

### Removed

- [clients] very old cds-notification clients

## [v1.69.1] - 2022-11-16

### Fixed

- [bi] Log and HTTP out have different event-id

## [v1.69.0] - 2022-10-28

### Changed

- [bi] Stop repeating sending after 5 Attempts with exponential backoff

## [v1.68.0] - 2022-10-18

### Added

- Listmonk client and mock

## [v1.67.0] - 2022-10-11

### Added

- [bi] Add async http bi-event sender

### Fixed

- [standard] Wait for serviceMain done on runCtx cancellation

## [v1.66.0] - 2022-09-29

### Added

- `hasAppID` to tut functions

## [v1.65.0] - 2022-09-19

### Changed

- [bievents]: move timestamp into event to allow specifying it

## [v1.64.0] - 2022-09-19

### Added

- [bievents pkg]: event `login-complete` includes client version

## [v1.63.0] - 2022-08-26

### Changed

- [db2] Decouple log level from general logging level

## [v1.62.0] - 2022-08-16

### Added

- [tut pkg] `ValueEqualsAnyOrder` value check

## [v1.61.2] - 2022-08-10

### Changed

- Upgraded github.com/onsi/gomega to v1.19.0 to get rid of go mod tidy issues
- Updated header obfuscation configuration

### Fixed

- [db2] Using gorm default logger in test DB connections

## [v1.61.1] - 2022-07-14

### Fixed

- Revert not_before and not_after in jwt from time.Time to string
- JWT test files format to match actual format from vault

## [v1.61.0] - 2022-07-14

### Added

- [client] FeatureFlagging Get with user authorization

### Changed

- Update go 1.17 -> 1.18
- Update all dependencies

## [v1.60.0] - 2022-07-12

### Added

- [errorV2 pkg] New package for ErrorV2 from openapi
- [tut pkg] Some test utils functions for ErrorV2

### Fixed

- upgrade gopkg.in/yaml.v3 v3.0.1

## [v1.59.0] - 2022-06-30

### Added

- [tut pkg] Added some more test util functions
- [tut pkg] `CookieCheckFunc` type for cookie checks

### Changed

- Upgraded github.com/onsi/gomega to v1.19.0 to get rid of go mod tidy issues
- [tut pkg] `RespHasSetCookie` now takes `CookieCheckFunc` as checks

### Removed

- [tut pkg] Removed `RespBodyEqualsText` and `RespBodyTextMatchesRegex`. `RespHasTextBody` should be used instead

### Security

- Uograded containerd too fix security issuue.

## [v1.58.0] - 2022-06-07

### Fixed

- bug in notification client error handling

## [v1.57.0] - 2022-06-07

### Added

- [tut pkg] `tut` (a.k.a `test utils`) package with reusable test tools

## [v1.56.0] - 2022-05-24

### Fixed

- http-out-response header obfuscation attempt in case of nil response

## [v1.55.0] - 2022-05-11

### Added

- [log pkg]: Logging HTTP headers for in/out requests and responses with options to ignore or obfuscate certain header keys. Obfuscation by replacing the value with its length.
- [bievents pkg]: events `login-email` includes more details about the cause in case of failure
- [bievents pkg]: events `login-email`, `login-sms` and `login-eid`
- `WithMigrationHaltOnError` to halt on errors during DB migration
- bievents: added `LogCtx` which extracts `sessionID` and `userID` from Context
- bievents: added default logging of event-source
- [client pkg] transport middleware for forwarding JWT access token from the context
- [jwt pkg] extract JWT raw token and store in context

## [v1.54.0] - 2022-05-02

### Added

- [Transport pkg]: Retry transport mechanism depending on http status codes

### Changed

- [JWT pkg] keymgmt scopes adjusted to follow CRUD style (from former read, write, append)

## [v1.53.1] - 2022-04-29

### Fixed

- add missing checks in UserPreferences client

## [v1.53.0] - 2022-04-28

### Changed

- upgrade UserPreferences client to APIv2

## [v1.52.0] - 2022-04-06

### Added

- [bievents pkg]: events `login-email` includes more details about the cause in case of failure
- [bievents pkg]: events `login-email`, `login-sms` and `login-eid`
- `WithMigrationHaltOnError` to halt on errors during DB migration

### Deprecated

- [bievents pkg]: events `login-start` and `phone-verify`

### Fixed

- `waitForDB` didn't halt on connection errors
- `waitForDB` didn't wait for all connectors, but only the first

## [v1.51.0] - 2022-03-30

### Changed

- publish channels package

## [v1.50.0] - 2022-03-22

### Added

- [JWT pkg]: New scope: `PushNotifDevice`

### Changed

- [JWT pkg]: Replace deprecated `StandardClaims` with `RegisteredClaims`

## [v1.49.0] - 2022-03-16

### Added

- added helper functions LogAuditRead() and LogAuditBulkRead()

## [v1.48.1] - 2022-03-14

### Fixed

- fix missing scopes in KnownTokens

## [v1.48.0] - 2022-03-10

### Added

- bievents: account-delete activity type and additional properties (for hades)

## [v1.47.0] - 2022-03-10

### Added

- Mock for vega internal client

## [v1.46.0] - 2022-03-09

### Security

- Fixes containerd security vulnerability

## [v1.45.0] - 2022-03-01

### Added

- client: add http-out logging and monitoring using transport pkg for all clients

## [v1.44.0] - 2022-02-28

### Added

- feature-flagging: simulate cancelled requests and failing service

## [v1.43.0] - 2022-02-23

### Added

- granular JWT scopes for services: cov-survey, program-management, user-preferences, consent-management, cds-userdata
- IP address obfuscator

## [v1.42.0] - 2022-02-15

### Security

- Fixes docker/distribution security vulnerability

## [v1.41.0] - 2022-02-14

### Added

- Service Secret Auth middleware (moved from Vega)

### Changed

- Go version 1.16 -> 1.17

### Security

- Fixes containerd security vulnerability
- Fixes image-spec security vulnerability

## [v1.40.0] - 2022-02-11

### Changed

- [breaking] [d4lcontext]: `GetUserID`, `GetUserIDFromCtx` and `WithUserID` work with UUIDs instead of strings

## [v1.39.0] - 2022-02-10

### Added

- Introduced unique BI event-id to be E2E duplicate-proof

## [v1.38.0] - 2022-02-07

### Removed

- [JWT pkg]: `auth.VerifyAny` (use `auth.Verify` instead)

### Security

- [JWT pkg]: auth middleware: access cookies are considered only if the CSRF factors are valid (using nosurf's double submit cookie) or if the method is CSRF-safe.

## [v1.37.0] - 2022-02-02

### Added

- [db2] Add `SkipDefaultTransaction` option

## [v1.36.0] - 2022-01-19

### Added

- [JWT pkg]: New scope: `TokenLoginConsent`

## [v1.35.1] - 2022-01-19

### Fixed

- Repair linter pipeline by using newest jenkins pipeline version.
- Standard: Fix database connection timeout duration

## [v1.35.0] - 2022-01-11

### Added

- [db2] Add `TXDBPostgresDriverWithoutSavepoint` for concurrency testing

## [v1.34.0] - 2022-01-05

### Added

- [client] for Feature Flagging Service

## [v1.33.0] - 2021-12-22

### Added

- Mobile phone number obfuscator

### Changed

- [JWT pkg]: auth middleware functions support (optionally) the access token being sent in a cookie (with the `phdp-access-token` name)

### Removed

- Offboarding Claas and Piotr

## [v1.32.0] - 2021-12-10

### Added

- [JWT pkg]: New scopes: `TokenPasswordUpdate`, `TokenLoginSecondFactor`

### Changed

- [JWT pkg]: Use CRUD operations for device scopes (`TokenDeviceCreate` instead of `TokenDeviceAppend`, `TokenDeviceUpdate` and `TokenDeviceDelete` instead of `TokenDeviceWrite`).

## [v1.31.0] - 2021-12-08

### Added

- client: vega v2 client

### Changed

- [JWT pkg]: unknown scopes are ignored by `Scope`'s `Scan` and `UnmarshalJSON` methods (but not by `NewScope` method).
- [JWT pkg]: `WithAllScopes` and `WithAnyScope` don't fail on unknown scopes; instead the unknown scopes are ignored.
- [JWT pkg]: a warning is logged for every unknown scope encountered (and ignored) when deserializing scopes.

### Deprecated

- [JWT pkg]: `NewScope` method (use `Parse` instead)

### Removed

- [JWT pkg]: `WithScopes` auth checker function. Use `WithAllScopes` instead
- [JWT pkg]: `WithGorillaMux` middleware. Use `WithGorillaOwner` instead.
- [JWT pkg]: `New` function. Use `NewAuthenticator` instead.
- [JWT pkg]: `middleware.Auth`. Use `NewAuthenticator` instead.

## [v1.30.0] - 2021-11-22

### Added

- middleware to handle context cancellation caused by the client connection (return status 499 instead of 500)
- Obfuscator for HTTP Out Request and Response logging

### Changed

- [JWT pkg]: Introduce deprecated scopes mechanism (deprecated scopes will not be issued)

### Deprecated

- [JWT pkg]: `terradb` scope

## [v1.29.0] - 2021-11-02

### Changed

- SQL & HTTP logging now partially obfuscates emails

## [v1.28.0] - 2021-10-27

### Added

- JWT pkg - token for appending a recovery password

### Removed

- JWT pkg - `AllScopes` variable. Use `KnownScopes` instead.

## [v1.27.0] - 2021-10-26

### Added

- A pgx Logger adapter

## [v1.26.0] - 2021-10-07

### Fixed

- Issue with hot-reloading of JWT public keys

## [v1.25.0] - 2021-10-07

### Added

- bievents - event type `password-reset`

## [v1.24.0] - 2021-10-05

### Added

- log - New sql logging function

## [v1.23.0] - 2021-09-29

### Added

- dynamic - Metrics for monitoring bootstrapping and hot-reloading of JWT keys
- dynamic - `ViperConfig.WithServiceName` to provide information about service name for more informative metrics
- Add commonpb package

### Changed

- dynamic - Renamed `ViperConfig.Merge` to `ViperConfig.MergeAndDisableHotReload` to increase awareness about its side-effects

### Deprecated

- dynamic - `ViperConfig.Merge / MergeAndDisableHotReload` shall not be used anymore to prevent occurrences of hard-to-debug side-effects (it disables hot-reload)

## [v1.22.0] - 2021-09-14

### Added

- DB2 - New Logger for gorm
- client - Azure Oauth2 client

## [v1.21.0] - 2021-09-07

### Added

- http client CSRF transport

## [v1.20.0] - 2021-09-06

### Added

- Support for *easier secrets rotation* for JWT keys by adding new constructor `middlewares.NewAuthentication(...)`

### Deprecated

- Constructor `middlewares.NewAuth(...)` in favor of `middlewares.NewAuthentication(...)`
    For 1:1 backwards compatibility use: `NewAuthentication(token, AuthWithRSAPublicKey(key), handlerFactory, options...)` instead of `NewAuth(token, key, handlerFactory, options...)`.
    To support *easier secrets rotation* use: `NewAuthentication(token, AuthWithPublicKeyProvider(viperConfig), handlerFactory, options...)`

## [v1.19.0] - 2021-08-31

### Added

- bievents - Event add event-source

## [v1.18.0] - 2021-08-26

### Added

- bievents - property `authn-type` to event data for type `login-start`

### Changed

- bievents - event type `eid-login-start` remodelled to `login-start` with `authn-type=eID`

### Removed

- bievents - event type `eid-login-start`

## [v1.17.0] - 2021-08-23

### Changed

- Inconsistent BI events names: changed `token-refreshed` to `token-refresh`, `sharing-revoked` to `sharing-revoke`

## [v1.16.0] - 2021-08-13

### Added

- bievents - use cuc ID instead of cuc name for onboarding BI data

### Changed

- bievents - round timestamp of events to seconds
- bievents - remove `source` from UserRegisterData
- bievents - rename events: `login` to `login-start` and `eid-login` to `eid-login-start`

### Fixed

- migrate - fox broken integration-tests

## [v1.15.0] - 2021-08-11

### Added

- bievents - add blob and record event types
- bievents - activity types are moved to this library (were in services code)
- bievents - more events data types

### Changed

- Ops: Pull Docker images from Nexus instead of DockerHub
- dynamic - refactor interfaces and add support for JWT private keys

### Security

- Replace vulnerable `dgrijalva/jwt-go` with `golang-jwt/jwt`

## [v1.14.1] - 2021-08-05

### Fixed

- DB2 - Set txdb postgres driver to pgx

## [v1.14.0] - 2021-08-03

### Added

- client - instrumented HTTP client implementation (logging, monitoring, traceID)

### Fixed

- bievents - fix missing state property on logged events

## [v1.13.0] - 2021-08-02

### Added

- DB2 pkg - use migrate pkg after the execution of the gorm migrations

## [v1.12.0] - 2021-08-02

### Added

- bievents - add optional state to events
- bievents - `WithTenantID` option for emitter
- bievents - more events data types
- d4lcontext - `GetUserIDFromCtx`, `GetClientIDFromCtx`, `GetTenantIDFromCtx` to extract information from a context object directly

## [v1.11.0] - 2021-07-19

### Added

- Add package `db2` which upgrades `github.com/jinzhu/gorm` to `gorm.io/gorm`
- channels - added FanIn

## [v1.10.0] - 2021-07-13

### Added

- Add source field to OnBoardingData in bievents

## [v1.9.0] - 2021-07-08

### Added

- Add scope for user keys migrate [ku:m]

## [v1.8.0] - 2021-06-09

### Added

- bievents - provide method GetEmailTypeNoError for simpler error handling in client code

## [v1.7.0] - 2021-06-03

### Added

- Client pkg - support for `SendRaw` in the notification client

## [v1.6.1] - 2021-06-01

### Changed

- Use semantic versions for minVersion of consents

## [v1.6.0] - 2021-05-27

### Added

- JWT pkg - `Extract` middleware that extracts the JWT information without doing any access control
- JWT pkg - function for creating a signed access token
- JWT pkg - token for mail verification
- d4lcontext pkg - methods allowing to add values to a request's context using the d4lcontext keys
- DB pkg - use migrate pkg after the execution of the gorm migrations

### Changed

- [breaking] Move `ParseRequesterID` function from `d4lcontext` to `d4lhandler` package
- Log pkg - support the `d4lcontext` keys as a fallback for user ID, client ID and tenant ID
- d4lcontext pkg - GetTenantID doesn't fall back to 'd4l' if the tenant ID is missing in the context

## [v1.5.0] - 2021-05-21

### Added

- JWT pkg - tokens for user keys, app keys and devices scopes

## [v1.4.1] - 2021-05-06

### Fixed

- Boostrapping error was not nil even if `dynamic.NewViperConfig` succeeded

## [v1.4.0] - 2021-05-06

### Added

- JWT pkg - ability to run `Verify` against multiple JWT public keys
- Package `dynamic` to dynamically load and update config/secrets using `viper` without restarting a service

### Deprecated

- Constructor for auth middleware `jwt.New()` in favor of `jwt.NewAuthenticator()`

## [v1.3.0] - 2021-05-06

### Added

- UrlValidator middleware to early detect and reject malformed queries

## [v1.2.0] - 2021-04-27

### Added

- Optional redis client used for caching
- Extended BI Events by consent document key

### Changed

- The main wrapper supports now opening connections to Postgres and Redis using functional options.

## [v1.1.1] - 2021-04-21

### Fixed

- Notification client encoding of request body as JSON

## [v1.1.0] - 2021-04-20

### Changed

- Update client for `cds-notification` service to support `arbitraryEmailAddress` parameter used in `>= v0.13.0`
- Client `NotificationService` implementing the previous version of the interface (`NotificationV4`) is renamed to `NotificationServiceLegacyV4`.

### Deprecated

- Deprecate client interface `NotificationV4`.
    Migrate from `NotificationV4` to `NotificationV5` by setting both `Arbitrary...` fields to empty strings or switch to `NotificationServiceLegacyV4`.

### Fixed

- Notification client no longer considers reply code 200 as error

## [v1.0.0] - 2021-04-08

### Added

- go-log, go-jwt, go-pg-migrate, go-cors, go-monitoring merged into this library

### Changed

- Go version 1.15 -> 1.16
- Linter version 1.30 -> 1.38

## [v0.16.1] - 2021-03-05

### Fixed

- Make SSLRootCertPath optional

## [v0.16.0] - 2021-03-03

### Changed

- Encrypt connection to the Database and verify the server certificate (BSI)
- Default value for DB connection parameter `SSLMode` is now `verify-full` (was `disable`)

## [v0.15.0] - 2021-03-02

### Deprecated

- Deprecate XSRF middleware, it no longer denies access and should be removed from the services
- Deprecate XSRF handler, it hands out a constant token and should be removed after all clients have been updated

## [v0.14.1] - 2021-02-22

### Changed

- DB instrumenter produces no unformatted log messages anymore

## [v0.14.0] - 2021-02-18

### Added

- Add LogAuditSuccess(), LogAuditFailure() to audit log successful and failed accesses

### Removed

- Remove deprecated LogAudit()

### Fixed

- Add missing line breaks in error handlers

## [v0.13.0] - 2021-01-13

### Added

- Add middleware to extract tenant id from X-Tenant-ID header (for non-JWT endpoints)

## [v0.12.0] - 2021-01-13

### Changed

- Deprecated UUID package from `satori` to `gofrs`

## [v0.11.0] - 2020-12-17

### Changed

- Remove one level of indirection for GetUserID(), GetClientID(), GetTenantID()

## [v0.10.1] - 2020-12-16

### Added

- Tenant id is taken from JWT token (default is 'd4l')

## [v0.10.0] - 2020-11-24

### Added

- Trace transport client
- Trace middleware

### Changed

- Http clients use trace transport

## [v0.9.2] - 2020-11-19

### Fixed

- Bug in auth middleware that returned wrong error message if auth header was empty

## [v0.9.1] - 2020-11-02

### Fixed

- Bug in counting notified users in notification mock

## [v0.9.0] - 2020-10-30

### Added

- Client for `consent-management` service `>= v0.7.0` (operation: batch-fetching user consents)
- Notification client (interface `NotificationV4`) supports now
- NotificationMock client implementing `NotificationV3` and `NotificationV4`

## [v0.8.0] - 2020-10-28

### Added

- Auth middlewares for JWT and service secret based authentication
- XSRF middleware
- XSRF handler

## [v0.7.0] - 2020-10-22

### Added

- Client for `user-preferences` service
- Logging functions accepting context e.g., `LogInfofCtx`, `LogErrorfCtx`

### Changed

- Update client for `cds-notification` service to support `consentGuardKey` and `minConsentVersion` parameters used in `>= v0.6.x`

### Deprecated

- Deprecate client interfaces `NotificationV3`.
    Please migrate to compatible `NotificationV4` client.
    When using this version of `go-svc`, all `NotificationClient`s that rely on `cds-notification` `< v0.6.0` should be changed to `NotificationClientLegacy`,

## [v0.6.1] - 2020-10-01

### Fixed

- Add mutex to map accesses in gorm instrumenter

## [v0.6.0] - 2020-10-01

### Added

- Add instrumentation for gorm

## [v0.5.0] - 2020-09-25

### Added

- Functions for audit logging
- Audit logging functions require context
- Add functions to set environment, hostname and pod name for audit logs

## [v0.4.0] - 2020-08-19

### Added

- Interface client.NotificationV2 and its implementations to enable tests in survey-svc

## [v0.3.0] - 2020-08-17

### Added

- Support for Postgres test-DB with custom TXDB driver

### Changed

- Go version upgraded to 1.15
- Clint version upgraded to 1.30

## [v0.2.0] - 2020-08-14

### Added

- Client library for `cds-notification`

## [v0.1.0] - 2020-07-31

### Added

- Initial state: standards for Main, HTTP Server, DB access (gorm), Logging, Instrumented-Handler, and K8s Probe

[Unreleased]: https://github.com/gesundheitscloud/go-svc/compare/v1.80.1...HEAD
[v1.80.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.80.0...v1.80.1
[v1.80.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.79.0...v1.80.0
[v1.79.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.78.0...v1.79.0
[v1.78.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.77.0...v1.78.0
[v1.77.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.76.0...v1.77.0
[v1.76.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.75.0...v1.76.0
[v1.75.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.74.0...v1.75.0
[v1.74.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.73.0...v1.74.0
[v1.73.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.72.0...v1.73.0
[v1.72.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.71.1...v1.72.0
[v1.71.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.71.0...v1.71.1
[v1.71.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.70.0...v1.71.0
[v1.70.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.69.1...v1.70.0
[v1.69.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.69.0...v1.69.1
[v1.69.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.68.0...v1.69.0
[v1.68.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.67.0...v1.68.0
[v1.67.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.66.0...v1.67.0
[v1.66.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.65.0...v1.66.0
[v1.65.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.64.0...v1.65.0
[v1.64.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.63.0...v1.64.0
[v1.63.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.62.0...v1.63.0
[v1.62.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.61.2...v1.62.0
[v1.61.2]: https://github.com/gesundheitscloud/go-svc/compare/v1.61.1...v1.61.2
[v1.61.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.61.0...v1.61.1
[v1.61.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.60.0...v1.61.0
[v1.60.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.59.0...v1.60.0
[v1.59.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.58.0...v1.59.0
[v1.58.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.57.0...v1.58.0
[v1.57.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.56.0...v1.57.0
[v1.56.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.55.0...v1.56.0
[v1.55.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.54.0...v1.55.0
[v1.54.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.53.1...v1.54.0
[v1.53.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.53.0...v1.53.1
[v1.53.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.52.0...v1.53.0
[v1.52.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.51.0...v1.52.0
[v1.51.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.50.0...v1.51.0
[v1.50.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.49.0...v1.50.0
[v1.49.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.48.1...v1.49.0
[v1.48.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.48.0...v1.48.1
[v1.48.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.47.0...v1.48.0
[v1.47.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.46.0...v1.47.0
[v1.46.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.45.0...v1.46.0
[v1.45.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.44.0...v1.45.0
[v1.44.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.43.0...v1.44.0
[v1.43.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.42.0...v1.43.0
[v1.42.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.41.0...v1.42.0
[v1.41.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.40.0...v1.41.0
[v1.40.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.39.0...v1.40.0
[v1.39.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.38.0...v1.39.0
[v1.38.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.37.0...v1.38.0
[v1.37.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.36.0...v1.37.0
[v1.36.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.35.1...v1.36.0
[v1.35.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.35.0...v1.35.1
[v1.35.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.34.0...v1.35.0
[v1.34.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.33.0...v1.34.0
[v1.33.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.32.0...v1.33.0
[v1.32.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.31.0...v1.32.0
[v1.31.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.30.0...v1.31.0
[v1.30.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.29.0...v1.30.0
[v1.29.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.28.0...v1.29.0
[v1.28.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.27.0...v1.28.0
[v1.27.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.26.0...v1.27.0
[v1.26.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.25.0...v1.26.0
[v1.25.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.24.0...v1.25.0
[v1.24.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.23.0...v1.24.0
[v1.23.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.22.0...v1.23.0
[v1.22.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.21.0...v1.22.0
[v1.21.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.20.0...v1.21.0
[v1.20.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.19.0...v1.20.0
[v1.19.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.18.0...v1.19.0
[v1.18.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.17.0...v1.18.0
[v1.17.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.16.0...v1.17.0
[v1.16.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.15.0...v1.16.0
[v1.15.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.14.1...v1.15.0
[v1.14.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.14.0...v1.14.1
[v1.14.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.13.0...v1.14.0
[v1.13.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.12.0...v1.13.0
[v1.12.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.11.0...v1.12.0
[v1.11.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.10.0...v1.11.0
[v1.10.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.9.0...v1.10.0
[v1.9.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.8.0...v1.9.0
[v1.8.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.7.0...v1.8.0
[v1.7.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.6.1...v1.7.0
[v1.6.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.6.0...v1.6.1
[v1.6.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.5.0...v1.6.0
[v1.5.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.4.1...v1.5.0
[v1.4.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.4.0...v1.4.1
[v1.4.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.1.1...v1.2.0
[v1.1.1]: https://github.com/gesundheitscloud/go-svc/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/gesundheitscloud/go-svc/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.16.1...v1.0.0
[v0.16.1]: https://github.com/gesundheitscloud/go-svc/compare/v0.16.0...v0.16.1
[v0.16.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.15.0...v0.16.0
[v0.15.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.14.1...v0.15.0
[v0.14.1]: https://github.com/gesundheitscloud/go-svc/compare/v0.14.0...v0.14.1
[v0.14.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.13.0...v0.14.0
[v0.13.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.12.0...v0.13.0
[v0.12.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.11.0...v0.12.0
[v0.11.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.10.1...v0.11.0
[v0.10.1]: https://github.com/gesundheitscloud/go-svc/compare/v0.10.0...v0.10.1
[v0.10.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.9.2...v0.10.0
[v0.9.2]: https://github.com/gesundheitscloud/go-svc/compare/v0.9.1...v0.9.2
[v0.9.1]: https://github.com/gesundheitscloud/go-svc/compare/v0.9.0...v0.9.1
[v0.9.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.6.1...v0.7.0
[v0.6.1]: https://github.com/gesundheitscloud/go-svc/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/gesundheitscloud/go-svc/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/gesundheitscloud/go-svc/releases/tag/v0.1.0

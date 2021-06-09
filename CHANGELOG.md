# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

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

[Unreleased]: https://github.com/gesundheitscloud/go-svc/compare/v1.8.0...HEAD
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

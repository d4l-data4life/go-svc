# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add LogAuditSuccess(), LogAuditFailure() to audit log successful and failed accesses

### Changed

### Deprecated

### Removed

- Remove deprecated LogAudit()

### Fixed

- Add missing line breaks in error handlers

### Security

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

[Unreleased]: https://github.com/gesundheitscloud/go-svc/compare/v0.13.0...HEAD
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

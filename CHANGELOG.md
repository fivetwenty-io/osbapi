# Changelog

## v2.0.5

### Security

- The basic-auth middleware now compares the username and password in constant time. It hashes both the supplied and the configured values with SHA-256 and compares the digests with `crypto/subtle.ConstantTimeCompare`, so the time a request takes no longer depends on how many leading characters were right or on the length of either value. It always checks both fields, so a 401 never reveals which one was wrong.

### Added

- The repository now carries a `LICENSE` file with the full Apache License 2.0 text, which the README already named as the license.

## v2.0.4

### Added

- `ErrInstanceAlreadyProvisioned` is a new sentinel a broker returns from `Provision` when the instance already exists, is fully provisioned, and the request is identical. The provision handler answers 200 with the returned `ProvisionResponse`, even when it is empty and encodes as `{}`, which a broker could not send before because `ErrInstanceAlreadyExists` answers 409 and a nil error answers 201 or 202.

- A wrapped `ErrInstanceAlreadyProvisioned` is recognized with `errors.Is`. If it reaches the shared error mapping from any other operation it answers 409, the same as `ErrInstanceAlreadyExists`, and `IsConflict` does not report it because it signals success.

- The example broker now returns `ErrInstanceAlreadyProvisioned` when an identical provision repeats.

## v2.0.3

### Fixed

- `ErrUnauthorized` now answers 401, `ErrPlanQuotaExceeded` answers 422, and `ErrInvalidParameters` answers 400, each with the error's message as the `description`. They previously fell through to a 500, which contradicted the sentinel table in the README.

- `ErrBindingAlreadyExists` now answers 409 when the broker returns it with an empty `BindResponse`, and the bind handler still answers 200 with the response when the broker returns a non-empty one. The same error from any other operation also answers 409 instead of 500.

## v2.0.2

### Changed

- A broker error that the server has no mapping for now produces a 500 whose `description` is the error's message instead of the bare "internal server error", so the platform shows the operator the cause.

- Every error that maps to a 5xx is logged at error level with the operation, method, path, instance and binding IDs, status, and error message.

- A handler built without `WithLogger` now logs through `slog.Default()` instead of staying silent, and this also applies to the malformed originating-identity warning.

## v2.0.1

### Added

- Server middleware now decodes the `X-Broker-API-Originating-Identity`
  header (OSB v2.17 §5.1) and stashes the resulting `OriginatingIdentity`
  on the request `context.Context`.

- New public helpers `osbapi.ContextWithOriginatingIdentity` and
  `osbapi.OriginatingIdentityFromContext` let broker handlers retrieve the
  identity, and let tests inject one without going through HTTP.

### Behavior

- The header is treated as optional and lenient: an absent header leaves
  the context untouched; a malformed header is logged via the configured
  `Logger` (when set) and dropped, rather than rejected with 400, so a
  misbehaving platform cannot block otherwise-valid traffic. Brokers that
  require the identity must check `OriginatingIdentityFromContext` and
  treat absence as untrusted.

## v2.0.0

Initial release implementing OSB API v2.17.

### Features

- Complete server library with ServiceBroker interface

- Complete HTTP client with all 10 OSB API operations

- Async operation support with polling helpers

- HTTP Basic Authentication

- API version header validation

- Comprehensive error handling with OSBError types

- Integration tests for sync and async workflows

- Example broker and platform client

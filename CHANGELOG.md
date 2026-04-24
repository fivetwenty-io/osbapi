# Changelog

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

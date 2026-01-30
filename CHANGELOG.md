# Changelog

All notable changes to the Pincho Go Client Library will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Pincho Go Client Library
- Core client implementation with `NewClient()`
- `Send()` method with full options support
- `SendSimple()` convenience method
- Context support for cancellation and timeouts
- Functional options pattern for client configuration
- Custom error types: `Error`, `AuthError`, `ValidationError`, `RateLimitError`
- Comprehensive test suite with >95% coverage
- Complete documentation and examples
- Support for notification types, tags, images, and action URLs
- Zero dependencies (standard library only)

## [1.0.0] - TBD

Initial stable release.

### Features
- Send push notifications via Pincho API
- Context support for timeouts and cancellation
- Customizable HTTP client
- Type-safe Go API
- Comprehensive error handling
- Full test coverage
- Production-ready

---

## Version History

- **1.0.0** - Initial stable release (TBD)

## Upgrading

### To 1.0.0

Initial release - no migration needed.

## Development

### Version Format

We use [Semantic Versioning](https://semver.org/):
- **MAJOR** - Incompatible API changes
- **MINOR** - New functionality (backward compatible)
- **PATCH** - Bug fixes (backward compatible)

### Release Process

1. Update CHANGELOG.md with release notes
2. Update version in documentation
3. Create Git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tag: `git push origin v1.0.0`
5. GitLab CI/CD will handle the release

## Support

For questions about specific versions:
- Check the [documentation](https://pkg.go.dev/gitlab.com/pincho/pincho-go)
- Open an [issue](https://gitlab.com/pincho/pincho-go/-/issues)
- Email support@pincho.com

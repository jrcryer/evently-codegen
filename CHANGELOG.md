# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive build and release configuration
- Docker support with multi-stage builds
- GitHub Actions CI/CD pipeline
- Cross-platform build support
- Automated release process
- Installation scripts for easy setup
- Enhanced documentation and examples

### Changed
- Improved Makefile with comprehensive build targets
- Enhanced version information and build metadata
- Updated project structure for better organization

### Fixed
- Build configuration and dependency management
- Release artifact generation and packaging

## [1.0.0] - TBD

### Added
- Initial release of AsyncAPI Go Code Generator
- Support for AsyncAPI 2.x and 3.x specifications
- JSON and YAML input format support
- Go struct generation with proper naming conventions
- CLI interface with comprehensive options
- Library API for programmatic usage
- Schema reference resolution ($ref support)
- Comprehensive test coverage
- Integration and performance tests
- Docker containerization
- Cross-platform binary releases

### Features
- Parse AsyncAPI specifications from JSON/YAML files
- Generate strongly-typed Go structs from message schemas
- Handle nested objects and array types
- Support for optional and required fields
- Proper Go naming conventions (PascalCase)
- JSON struct tags for serialization
- Schema comments as Go field comments
- External reference resolution
- Error handling and validation
- CLI with help and usage information
- Library for programmatic integration

### Supported Platforms
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

---

## Release Notes Template

When creating a new release, use this template:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features and capabilities

### Changed
- Changes to existing functionality

### Deprecated
- Features that will be removed in future versions

### Removed
- Features that have been removed

### Fixed
- Bug fixes and corrections

### Security
- Security-related changes and fixes
```

## Version Numbering

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version when you make incompatible API changes
- **MINOR** version when you add functionality in a backwards compatible manner
- **PATCH** version when you make backwards compatible bug fixes

### Pre-release Versions

- **alpha**: Early development versions with potential breaking changes
- **beta**: Feature-complete versions undergoing testing
- **rc**: Release candidates that are potentially stable

Examples:
- `1.0.0-alpha.1` - First alpha release
- `1.0.0-beta.1` - First beta release
- `1.0.0-rc.1` - First release candidate
- `1.0.0` - Stable release
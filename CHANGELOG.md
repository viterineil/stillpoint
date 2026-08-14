# Changelog

All notable changes will be documented here. This project follows
[Semantic Versioning](https://semver.org/) after the first tagged release.

## Unreleased

### Added

- Initial public project scaffold.
- Strict configuration v1 model and JSON Schema.
- `doctor`, `config validate`, `plan`, and `version` commands.
- Canonical path-containment checks.
- Backup transaction state machine.
- Failure-safe agent barrier coordinator with cleanup tests.
- Architecture, threat model, agent protocol, Docker/data, restore, and roadmap
  documentation.

### Safety gate

- `backup` is intentionally disabled until Btrfs, Restic, durable run recovery,
  and restore integration tests are implemented.

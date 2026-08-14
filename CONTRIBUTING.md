# Contributing

Stillpoint welcomes issues, design review, documentation improvements, tests,
and code. Backup software has an unusually high cost for false confidence, so
correct failure behavior matters more than feature velocity.

## Start here

1. Read the [architecture](docs/architecture.md) and
   [threat model](docs/threat-model.md).
2. Search existing issues before opening a new one.
3. Discuss large protocol, storage, or restore changes before implementation.
4. Keep changes focused and add tests at the layer where behavior is promised.

## Development

Requires Go 1.23 or newer.

```bash
git clone https://github.com/viterineil/stillpoint.git
cd stillpoint
make check
make build
./bin/stillpoint help
```

Run `make fmt` before committing. Do not add secrets, real repository
credentials, production paths, or unredacted diagnostic fixtures.

## Safety review checklist

Changes that affect capture, retention, or restore should answer:

- What happens if this operation fails before changing state?
- What happens if it changes state and then reports failure?
- What happens on cancellation, timeout, SIGTERM, or host restart?
- Can a path, glob, symlink, or shell metacharacter expand the operation's
  scope?
- Is any secret observable in argv, environment dumps, logs, manifests, or
  temporary files?
- Does a successful test restore the result and validate its contents?
- Can cleanup delete the only complete copy?

New adapter execution must use explicit argument arrays, timeouts, structured
output, and redaction. It must not interpolate user input into shell commands.

## Pull requests

Describe the user-visible outcome, safety assumptions, tests, and documentation
changes. CI must pass. Maintainers may ask for destructive failure injection or
a restore fixture before accepting code that participates in a backup
transaction.

By contributing, you agree that your contribution is licensed under Apache-2.0.

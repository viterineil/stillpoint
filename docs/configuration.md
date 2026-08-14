# Configuration

Stillpoint uses strict, versioned YAML. Unknown keys are rejected so a typo
cannot silently disable a safety control. The current schema version is `1`.

Validate without performing a backup:

```bash
stillpoint config validate --config stillpoint.yaml
stillpoint plan --config stillpoint.yaml --set development
```

The public machine-readable definition is
[`schemas/stillpoint.schema.json`](../schemas/stillpoint.schema.json).

## Sources and snapshots

Every source path must be absolute and names a filesystem tree to protect.
Restic includes dotfiles, `.git`, ignored files, untracked files, symlinks,
permissions, and supported extended attributes beneath that path.

The MVP snapshot provider is Btrfs. `live_subvolume` identifies the writable
subvolume and `snapshots_directory` holds short-lived read-only snapshots. A
repository destination inside any source is rejected after canonical paths and
symlinks are resolved.

## Exclusions

The `node-dependencies` preset contains only:

```text
**/node_modules/**
**/.npm/**
**/.pnpm-store/**
**/.cache/node-gyp/**
```

Stillpoint intentionally does not exclude `packages`, `dist`, `build`,
`.yarn/cache`, Python environments, Rust `target`, or language `vendor`
directories automatically. Some contain irreplaceable or tracked source.

Before execution is enabled, exclusion planning will estimate saved space and
check for Git-tracked files. With `refuse_to_exclude_git_tracked_files: true`,
an exclusion that matches tracked data cannot proceed silently.

## Secret policy

`secrets.mode` is one of `include-encrypted`, `exclude`, `select`, or `fail`.
Interactive setup defaults to selection; unattended operation should use an
explicit noninteractive policy. Included files receive Restic's normal
whole-repository encryption.

Detection covers `.env`, `.env.*` (excluding documented templates), private-key
filenames, package-manager credentials, credential JSON, Compose `env_file`,
and file-backed Compose secrets/configs. Additional filename patterns can be
added without placing secret values in configuration.

## Agent barrier

`agents.mode: hybrid` combines command-based cooperative acknowledgements with
a `systemd-slice` fence. The configured arrays are executable plus arguments,
not shell strings. They are never passed to a shell.

`agents.mode: none` explicitly requests a crash-consistent run and will always
produce a visible warning.

See [Agent protocol](agent-protocol.md) for adapter behavior.

## Repository and credentials

The primary repository should be local, outside the source WSL VHD, and
preferably on another physical disk. `password_command` is an argument array
that writes only the Restic password to stdout. It must not accept the password
as an argument.

Replica `repository` values are Restic repository locators. Examples:

```yaml
repository: s3:s3.us-east-1.amazonaws.com/private-bucket/stillpoint/development
repository: rclone:gdrive:Backups/Stillpoint/development
```

Replication copies completed snapshots into separately initialized Restic
repositories. It never moves the only local copy.

## Docker and data

Compose discovery is opt-in per backup set. Every external bind mount must be
added as a source or explicitly omitted. Named volumes use application-aware
database adapters when possible and a generic filesystem archive only with a
clear consistency policy.

See [Docker and databases](docker-and-databases.md).

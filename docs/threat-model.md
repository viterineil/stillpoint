# Threat model

Stillpoint coordinates tools that already have broad access to development
data. Its safety claims need a clear boundary.

## Protected scenarios

The completed design should help recover from:

- loss or corruption of a WSL virtual disk;
- accidental deletion or destructive agent edits;
- untracked, ignored, hidden, and secret-bearing files missing from Git;
- files captured across different moments during a live directory scan;
- loss or corruption of a local backup repository when a replica survives;
- cloud storage disclosure, because Restic encrypts and authenticates content
  before upload.

## Outside the protection boundary

Stillpoint alone cannot protect against:

- a compromised source machine reading plaintext or credentials before
  encryption;
- an attacker who can delete every local and remote copy, unless independent
  immutable retention is configured;
- application state that remains only in memory and has no quiesce or dump
  mechanism;
- writers that are neither discoverable nor contained by a fence;
- remote Docker volumes or storage plugins without a compatible adapter;
- loss of the only Restic password or password-command recovery method;
- a backup that was never restored and tested.

## Trust assumptions

- The Linux kernel, snapshot implementation, Restic binary, and credential
  provider are trusted.
- Agent adapters truthfully enumerate writers and acknowledgements.
- The backup controller is outside the cgroup it freezes.
- A read-only snapshot provider never returns a writable live path.
- Repository credentials grant only the access required for their destination.

## Secrets

Restic provides whole-repository authenticated encryption. Stillpoint does not
add file-by-file encryption because doing so would complicate deduplication,
metadata restoration, rotation, and recovery.

Users choose one of four policies:

- `include-encrypted`: include detected secret files in encrypted backups;
- `exclude`: omit them and record the omission in the encrypted manifest;
- `select`: decide for each detected path interactively;
- `fail`: stop unattended runs when a secret path has no explicit policy.

Detection is advisory and cannot prove that every secret has been found.
Secret values are never valid log or manifest fields. Repository passwords must
come from `RESTIC_PASSWORD_COMMAND`, an OS keyring adapter, or an explicitly
protected `0600` file—not YAML or process arguments.

## Remote resilience

S3 credentials should be scoped to a single bucket prefix with public access
blocked. Bucket versioning is recommended. An independent Object Lock replica
can resist deletion, but its retention window must be designed around Restic
prune behavior.

Google Drive replication uses a completed encrypted Restic repository through
rclone. Raw source files and live-mutating repository directories are never
mirrored to Drive.

## Security reporting

Do not open a public issue for a suspected vulnerability. Follow
[SECURITY.md](../SECURITY.md).

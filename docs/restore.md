# Restore design

A backup is useful only if restoration is understandable and repeatable.

The planned command pair is:

```bash
stillpoint restore plan <backup-set-id>
stillpoint restore run <backup-set-id> --target <empty-directory>
```

## Plan

The plan will show:

- required repository and password-command method;
- filesystem roots and destination space estimate;
- exclusions and intentionally omitted secrets;
- Compose files and external bind mounts;
- named-volume archive formats;
- database adapters and required client tools;
- conflicts with existing files, containers, databases, or volumes;
- completeness and verification status of every component.

An incomplete logical set is not restorable by default.

## Execute

Restore defaults to a new empty directory. It restores secret-bearing files
with restrictive permissions, creates new Docker volumes, and targets new or
explicitly selected databases. Existing live paths are never overwritten
without a separate explicit confirmation flag.

Services are not started unless `--start` is supplied. A dry Compose start plan
is emitted first so network bindings, credentials, migrations, and external
dependencies can be reviewed.

## Drills

Recommended verification cadence:

- after every backup: validate component completion and Restic exit status;
- weekly: run a Restic structural check;
- monthly: read a sample of pack data;
- quarterly: restore into a temporary empty location;
- optionally: run a Compose smoke test in isolated volumes with no published
  ports.

The restore catalog must be rebuildable from encrypted manifests in the
repository. A lost local catalog must not make a valid repository unusable.

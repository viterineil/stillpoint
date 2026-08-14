# Architecture

## Design objective

Stillpoint must produce a backup set whose components can be understood as one
recoverable point in time while keeping the writer pause short. Long-running
copying and uploading happens only after an immutable filesystem snapshot
exists.

## Transaction

Every run receives a unique `backup_set_id`. Its monotonic state is:

```text
PLANNED
  -> DRAINING
  -> QUIESCENT
  -> FENCED
  -> FS_SNAPSHOTTED
  -> RESUMED
  -> COMPONENTS_BACKED_UP
  -> LOCAL_COMPLETE
  -> REPLICATED
  -> VERIFIED
```

`ABORTED` is allowed before `LOCAL_COMPLETE`. Replication failure does not
invalidate or delete `LOCAL_COMPLETE`; it remains retryable.

The consistent capture sequence is:

1. Create the run record and close the agent manager's dispatch gate.
2. Ask every discovered writer to checkpoint and acknowledge the barrier ID.
3. Abort if the acknowledgement deadline expires.
4. Freeze the dedicated agent cgroup to close the acknowledgement-to-snapshot
   race.
5. Flush pending filesystem writes and create a read-only Btrfs snapshot.
6. Unfreeze and resume agents, then reopen dispatch.
7. Back up files, database dumps, and volume archives to local Restic snapshots
   tagged with the same backup-set ID.
8. Write an encrypted completeness manifest and mark `LOCAL_COMPLETE`.
9. Remove temporary filesystem snapshots only after local completion.
10. Copy the completed Restic snapshots to configured replicas.

The controller runs outside the fenced cgroup. Cleanup actions are idempotent
and are registered before each state-changing call, so a partial adapter error
still triggers thaw/resume/open operations.

## Components

```text
cmd/stillpoint/       CLI entry point
internal/config/      Strict versioned YAML model and validation
internal/cli/         User commands and stable exit behavior
internal/doctor/      Environment capability detection
internal/safety/      Canonical paths and containment rules
internal/runstate/    Monotonic transaction state machine
internal/barrier/     Agent checkpoint/fence coordinator
internal/snapshot/    Snapshot-provider interface
internal/excludes/    Conservative built-in exclusions
schemas/              Public configuration and manifest schemas
docs/                 Protocol, security, restore, and operator guidance
```

Planned packages will add Restic execution, Btrfs, Docker discovery, database
adapters, manifests, replication, and restore execution.

## Adapter boundaries

Interfaces are narrow and external extensions will use a versioned JSON Lines
protocol. Go runtime plugins are intentionally avoided because they tightly
couple builds and complicate distribution.

Subprocesses must be invoked with `exec.CommandContext` and explicit argument
arrays. User paths must never be interpolated into `sh -c`. Tools should emit
JSON where supported; output passes through structured redaction before logs.

## Filesystem providers

The MVP assumes each live source is a Btrfs subvolume with a read-only sibling
snapshot directory:

```text
/mnt/stillpoint-dev/live
/mnt/stillpoint-dev/snapshots/<backup_set_id>
```

LVM and ZFS fit behind the same provider interface. A staged-copy fallback may
be added, but it must keep agents paused for the entire copy, estimate that
pause in advance, and never call itself instantaneous.

## Backup-set manifest

A logical backup may span multiple Restic snapshots. The encrypted final
manifest records source snapshots, exclusions, safe acknowledgement metadata,
Compose projects, database and volume components, Restic snapshot IDs, tool
versions, and completeness. The local catalog is only a cache: it must be
rebuildable from repository manifests.

## Repository topology

```text
Live WSL Btrfs subvolume
  -> read-only Btrfs snapshot
  -> encrypted local Restic repository on another disk
  -> encrypted Restic replica in S3
  -> optional separate Restic repository via rclone/Google Drive
```

The local repository should be outside the WSL source VHD and preferably on a
different physical disk. Copying is separate from capture so a network outage
cannot extend the agent pause.

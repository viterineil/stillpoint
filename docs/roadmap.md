# Roadmap

The roadmap is ordered by restoration trust, not feature count.

## Phase 0 — bootstrap (current)

- public repository, license, security and contribution policy;
- strict v1 YAML model and JSON Schema;
- environment doctor and dry safety plan;
- path containment tests;
- monotonic run-state model;
- coordinator cleanup tests for partial agent/fence/snapshot failures;
- CI build, vet, race tests, and formatting checks.

## Phase 1 — trustworthy local file backups

- durable run journal and crash recovery;
- command-hook agent adapter and acknowledgement-set validation;
- systemd-slice launch/freeze/thaw adapter;
- Btrfs read-only snapshot provider;
- safe Restic JSON wrapper and local repository initialization;
- secret-path detection and redacted structured logs;
- encrypted completeness manifests;
- listing, checking, file restore, and destructive failure injection;
- privileged Btrfs loopback integration suite.

The `backup` command remains disabled until this phase's end-to-end restore and
always-thaw tests pass.

## Phase 2 — Compose and stateful data

- Compose canonical-model discovery;
- external bind-mount coverage checks;
- file-backed configs, secrets, Dockerfiles, and build contexts;
- generic named-volume archive/restore;
- PostgreSQL and SQLite adapters, followed by MySQL/MariaDB;
- complete restore planner and isolated Compose smoke tests.

## Phase 3 — replicas and scheduling

- Restic snapshot copy to S3;
- Google Drive replication through a separate rclone Restic repository;
- retry queue, bandwidth controls, retention, and repository checks;
- systemd timer and Windows Task Scheduler launcher for WSL;
- optional immutable S3 replica guidance.

## Phase 4 — managed storage and ecosystem

- assisted Btrfs VHD creation, mount, migration, and verified rollback for WSL;
- LVM and ZFS providers;
- additional database adapters;
- stable versioned external plugin SDK;
- recovery-kit export and optional terminal UI.

## Release gates

An alpha release needs tested local capture and restore. A beta needs Docker
data, remote copy, interrupted-run recovery, and documented WSL drills. A 1.0
release needs schema/protocol compatibility commitments and restore testing from
a clean machine using only the repository, credentials, and public docs.

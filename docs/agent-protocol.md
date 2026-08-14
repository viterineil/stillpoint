# Agent coordination protocol

An agent can acknowledge a logical safe point, while a kernel cgroup fence can
prevent it from writing during the final snapshot race. Stillpoint uses both in
`hybrid` mode.

## Required lifecycle

1. `close`: stop dispatching new work.
2. `discover`: enumerate every active writer in scope.
3. `checkpoint`: ask each writer to finish or journal its current atomic unit.
4. `acknowledgements`: wait until the exact discovered set acknowledges the
   barrier ID, or fail at the deadline.
5. `fence`: freeze `agents.slice` recursively.
6. Create the read-only filesystem snapshot.
7. `unfence`, `resume`, and `open`, in that order.

Cleanup operations must be idempotent. Stillpoint may call them after a command
partially succeeds and then reports an error.

## Command adapter contract

Configured commands are argument arrays and are invoked directly—never through
a shell. Protocol messages use one JSON object on stdin and stdout.

Request envelope:

```json
{
  "protocol_version": 1,
  "operation": "checkpoint",
  "barrier_id": "01J...",
  "deadline": "2026-08-14T18:30:00Z",
  "agents": [{"id": "worker-1"}]
}
```

Successful response envelope:

```json
{
  "protocol_version": 1,
  "ok": true,
  "barrier_id": "01J...",
  "agents": [
    {
      "id": "worker-1",
      "worktree": "/home/neil/dev/project-a",
      "phase": "idle",
      "acknowledged_at": "2026-08-14T18:29:52Z",
      "child_writers_found": true
    }
  ]
}
```

Errors use `ok: false`, a stable `code`, and a redacted human-readable
`message`. Prompts, source contents, environment values, tokens, and credentials
must never be returned.

The process receives a minimal environment plus
`STILLPOINT_PROTOCOL_VERSION=1` and `STILLPOINT_BARRIER_ID`. Secret material is
not sent in JSON.

## Acknowledgement rules

- An acknowledgement applies only to its exact barrier ID.
- The acknowledged agent set must equal the last discovered writer set.
- A newly discovered writer closes the barrier and restarts acknowledgement.
- Timeout aborts a consistent backup; it never silently degrades consistency.
- Metadata may include ID, adapter, worktree, timestamp, high-level phase, and
  whether child writers exist—never task contents.

## systemd fence

Agents launched under Stillpoint belong to a dedicated `agents.slice`. The
controller belongs to another unit. `systemctl freeze agents.slice` recursively
stops ordinary descendants, and `thaw` is always attempted during cleanup.

Containers, remote jobs, privileged processes that escape the cgroup, and
independent editors require separate discovery/fencing or an explicit
crash-consistent policy.

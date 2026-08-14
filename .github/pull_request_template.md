## Outcome

Describe the user-visible result and the recovery problem it solves.

## Safety analysis

- [ ] Partial failure and cancellation behavior is documented.
- [ ] Cleanup is idempotent and cannot delete the only complete copy.
- [ ] Paths and commands do not expand through a shell.
- [ ] Secrets do not appear in config, argv, logs, manifests, or fixtures.

## Verification

- [ ] Unit or integration tests cover the change.
- [ ] Failure injection is included where state changes.
- [ ] Restore behavior is tested where backup contents change.
- [ ] Documentation and schema are updated.

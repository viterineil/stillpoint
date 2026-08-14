# Docker Compose, volumes, and databases

## Discovery

Stillpoint will resolve Compose projects from their canonical model and Docker
Engine metadata. Discovery offers the following source artifacts for backup:

- Compose files, overrides, includes, and project `.env`;
- referenced `env_file` paths and file-backed secrets/configs;
- Dockerfiles, `.dockerignore`, and build contexts;
- bind mounts and their host paths;
- named volumes, service ownership, and image references/digests.

Discovery must avoid logging interpolated environment values. Image layers are
not backed up by default because they are normally reproducible or pullable; an
optional `docker save` component can protect a locally built irreplaceable
image.

## Bind mounts

A host path inside a selected source is already covered by its filesystem
snapshot. A host path outside all sources must be added or explicitly omitted.
The encrypted manifest records host path, container destination, read-only
flag, and associated project/service.

## Named volumes

Preferred order:

1. Use an application-aware hot-backup adapter.
2. If none exists, pause/stop the owning service and archive the mounted volume
   read-only through a short-lived helper container.
3. Label the result crash-consistent only when the application supports that
   recovery model.

A generic tar archive provides a filesystem view; it does not make a live
database logically consistent.

## Initial database adapters

- PostgreSQL logical dump, with optional roles/globals;
- SQLite online backup operation;
- MySQL/MariaDB transactional logical dump.

Image or service-name detection only suggests an adapter. Users confirm it.
Credentials come from the process environment, native credential files, or a
password command—not YAML or argv. Dumps should stream directly into Restic so
plaintext temporary files are not required.

All filesystem, database, and volume components share a `backup_set_id`. The
final completeness manifest is written only after every required component
succeeds. A failed database dump therefore cannot be mistaken for a complete
restore point.

## Restore

Named volumes are restored into newly created volumes. Databases are restored
into user-selected targets. Stillpoint renders a final Compose start plan but
does not start services or publish ports unless the user explicitly asks.

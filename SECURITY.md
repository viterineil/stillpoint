# Security policy

## Supported versions

Stillpoint has not published a supported release. The default branch is
pre-alpha and must not yet be relied upon as the only backup mechanism.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting flow on the repository's
**Security** tab. Please do not open a public issue containing an exploit,
credential, private path listing, repository password, or unredacted log.

Include the affected commit or version, operating environment, impact, minimal
reproduction, and whether the issue can expose data, leave writers frozen,
delete data, or produce a falsely complete backup.

Maintainers will acknowledge a report as soon as practical, coordinate a fix
and disclosure window, and credit reporters who want attribution.

## Operational warning

Until a release explicitly removes the safety gate, `stillpoint backup` is not
implemented. Keep independent tested backups and do not replace an existing
recovery system with development builds.

# Security policy

## Reporting a vulnerability

If you discover a security vulnerability in Tolvi, please **do not** open a public GitHub issue.

Instead, report it privately by [opening a security advisory](https://github.com/tolvi-labs/tolvi/security/advisories/new) in this repository. We will acknowledge within 48 hours.

Once a mailbox is live, security reports may also be sent to `security@tolvilabs.com`.

## Scope

The following are in scope:

- Vulnerabilities in code published in this repository
- Vulnerabilities in the published `tolvi` CLI binary, the `tolvi-server` Docker image, or the `@tolvi-labs/sdk` npm package
- Vulnerabilities that allow unauthorized read or write access to vault content via the `tolvi` server API

The following are out of scope (please file as regular issues):

- Theoretical vulnerabilities without proof-of-concept
- Issues in third-party dependencies (please report upstream)
- Issues in the user's own deployment configuration

## Supported versions

| Version | Supported |
|---|---|
| pre-1.0 | Best-effort; security fixes will land on `main` |

A formal supported-versions policy will be published when 1.0 ships.

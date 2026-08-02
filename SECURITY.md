# Security Policy

Ababil SSO is an identity provider — a compromise here compromises every
relying party. Please report vulnerabilities privately.

## Reporting a vulnerability

Email **developer.mozadded@gmail.com** with:

- A description of the issue and its impact.
- Steps to reproduce (a minimal PoC helps a lot).
- Affected version/commit.

Do not open a public GitHub issue for security reports. You will get an
acknowledgment within 72 hours. We ask for 90 days to ship a fix before
public disclosure, and will credit you in the release notes unless you
prefer otherwise.

## Scope

In scope: `server/` (Go OAuth2.1/OIDC implementation), `src/` (Next.js
UI), the Docker Compose deployment as shipped.

Out of scope: vulnerabilities requiring a compromised database,
Redis instance, or `KEY_ENCRYPTION_KEY` — those are trusted-infrastructure
assumptions, not application bugs. Self-hosted deployments where the
operator disabled TLS or exposed Postgres/Redis publicly.

## Supported versions

Only the latest `main` is supported until the project reaches a tagged
1.0.

## Design-level security notes

For the threat model behind specific decisions (pairwise subject IDs,
PKCE-only authorization code flow, refresh-token rotation with reuse
detection, Argon2id admission control, etc.), see `docs/architecture.md`.

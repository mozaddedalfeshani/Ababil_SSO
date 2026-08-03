# Ababil SSO

Open-source, privacy-first OAuth 2.1 / OpenID Connect identity provider.
Go owns every protocol and security decision; Next.js owns every pixel.

- Pairwise subject identifiers by default — relying parties can't
  correlate the same user with each other.
- No telemetry, no third-party runtime calls, hashed IPs, bounded audit
  retention.
- PKCE-only authorization code flow, opaque rotated refresh tokens with
  reuse detection, ES256 JWKS.

See `docs/architecture.md` for the design rationale and threat model,
`docs/self-hosting.md` for deployment + key-backup guidance,
`docs/integrating.md` for connecting a relying party, and
`SECURITY.md` to report a vulnerability.

Want to see the whole protocol flow run end to end? `examples/rp-demo/run.sh`
drives a full authorize → consent → token → refresh-rotation →
reuse-detection walkthrough against a running instance.

## Quickstart (Docker Compose)

```bash
cp .env.example .env
# fill in DATABASE_URL / REDIS_URL if not using the bundled containers,
# and generate a key encryption key:
openssl rand -base64 32   # paste into KEY_ENCRYPTION_KEY

docker compose up --build
```

Then open http://localhost:5680.

## Local development

### Option A — db + redis + backend in Docker, frontend on host (recommended)

Backend hot-reloads on file save (via [air](https://github.com/air-verse/air)); no rebuild needed while iterating.

```bash
cp .env.dev.example .env.dev
# fill in KEY_ENCRYPTION_KEY: openssl rand -base64 32

docker compose -f docker-compose.dev.yml up --build
```

This starts 3 containers: `postgres` (:5433 on host), `redis` (:6380 on
host), `go` (:7897, hot-reloading), plus a one-shot `migrate` job that
runs first. In another shell:

```bash
pnpm install
pnpm dev               # listens on :5680, proxies /oauth,/api,/.well-known to :7897
```

### Option B — everything on host, no Docker

Backend (`server/`, Go 1.25+):

```bash
cd server
go run ./cmd/migrate   # apply schema (safe to re-run)
go run ./cmd/serve     # listens on :7897
```

Frontend (`src/`, requires pnpm):

```bash
pnpm install
pnpm dev               # listens on :5680, proxies /oauth,/api,/.well-known to :7897
```

Both read configuration from `.env` at the repo root — see
`.env.example` for every variable and what it controls. Option A uses
`.env.dev` instead (see `.env.dev.example`) since its Postgres/Redis
ports differ from the full Compose stack.

## Repository layout

```
server/     Go OAuth2.1/OIDC provider (Gin, pgx, Redis)
src/        Next.js UI (App Router, Tailwind, shadcn/ui)
examples/   Example relying party proving the OAuth loop end to end
docs/       Architecture, threat model, self-host guide
```

## License

Apache 2.0 — see `LICENSE`.

# Self-Hosting Guide

## Requirements

- Docker + Docker Compose (or Postgres 16+ / Redis 7+ / Go 1.25+ / Node 22+ if running services directly)
- A domain pointed at your server, if deploying for real use (Caddy auto-provisions TLS)

## Quickstart

```bash
git clone <this-repo>
cd Ababil_SSO
cp .env.example .env
```

Fill in `.env`:

- `SSO_ISSUER` — the public URL this instance will be reachable at (e.g. `https://auth.example.com`). This is baked into every issued token as `iss`; changing it later invalidates every outstanding token and breaks every relying party's discovery cache.
- `KEY_ENCRYPTION_KEY` — generate with `openssl rand -base64 32`. **Back this up before going live** — see "Key backup" below.
- `NEXT_PUBLIC_APP_URL` — same as `SSO_ISSUER` in the standard same-origin topology.
- `SMTP_*` — required in production (`APP_ENV=production`); the server refuses to boot without SMTP configured in that mode, since the console mailer would otherwise silently log password-reset links only to your own server logs.
- `TRUSTED_PROXY_CIDRS` — the IP(s)/CIDR(s) of whatever sits in front of Go. In the bundled Compose stack that's Caddy on the Docker network; leave the default (`127.0.0.1/32,::1/128`) if Caddy and Go run in the same network namespace via Docker Compose's internal DNS, or set it explicitly to your Docker bridge subnet if health checks show requests being misattributed.

Then, for a real deployment, edit `Caddyfile`: replace `:80` with your domain and remove the `auto_https off` block — Caddy will provision a Let's Encrypt certificate automatically.

```bash
docker compose up --build -d
docker compose run --rm migrate   # first boot only; safe to re-run
```

Open your domain (or `http://localhost:5680` for local testing).

## Architecture recap

Caddy terminates TLS and routes by path: `/oauth/*`, `/.well-known/*`, `/api/*` go to the Go container; everything else goes to Next.js. This is the same routing `src/proxy.ts` does in local `next dev` — Caddy exists only because `next dev`'s proxy isn't meant for production TLS termination. See `docs/architecture.md` for why this topology (same-origin, first-party cookies) was chosen.

## Key backup (`KEY_ENCRYPTION_KEY`)

This key seals, at rest:

- Every TOTP secret (`users.totp_secret_enc`)
- Every OAuth signing key's private half (`signing_keys.private_key_enc`)

**Losing it is unrecoverable** — not "hard to recover," actually unrecoverable. Every user with 2FA enabled would need to re-enroll (locking them out until they do, since the stored secret becomes permanently undecryptable), and the OIDC signing key would need full replacement (every previously-issued JWT — access tokens up to their TTL, and any long-lived assumptions relying parties made about your JWKS — breaks until they refresh).

**Before your first production boot:**

1. Store `KEY_ENCRYPTION_KEY` in your organization's secret manager (1Password, Vault, AWS Secrets Manager, whatever you already use) — not just in `.env` on the server's disk.
2. Keep at least one offline copy (printed, or on air-gapped storage) if this instance protects anything you can't afford to lose access to.
3. If you ever need to rotate it: there is currently no built-in re-wrap tool. Rotating requires decrypting every sealed value under the old key and re-sealing under the new one in a single maintenance transaction — file an issue if you need this; it's a reasonable feature request, just not built yet because no deployment has needed it.

## Signing key rotation

Unlike `KEY_ENCRYPTION_KEY`, OIDC signing keys (ES256) rotate routinely and safely — see `services/keys.Manager.Rotate` in the Go server. Rotation:

1. Generates a new key, published in JWKS as `status: "next"` immediately (so relying parties cache-warm before it's used).
2. Promotes it to `active`.
3. Retires the previous key but keeps it in JWKS for a grace period (long enough for any token signed just before rotation to still validate).

There's no scheduled automatic rotation yet — call `Rotate` from an admin tool or a cron job if you want periodic rotation; the default posture (one long-lived active key) is fine for most self-hosted deployments.

## Scheduled maintenance

Run this on a cron / systemd timer / Kubernetes CronJob — it is **not** started automatically by `docker compose up`:

```bash
docker compose run --rm retention
```

Purges `audit_logs` rows older than `AUDIT_RETENTION_DAYS` (default 90). See `docs/architecture.md` "Privacy" for why this is bounded by default.

## Backups

Back up the `pg_data` volume (or your external Postgres, if you're not using the bundled one) like any other production database — this is where everything except `KEY_ENCRYPTION_KEY` lives. Redis holds only ephemeral state (in-flight logins, rate-limit counters); losing it costs nothing but a few active login attempts.

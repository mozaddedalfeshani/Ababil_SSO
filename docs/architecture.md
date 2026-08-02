# Architecture & Threat Model

Ababil SSO is a self-hostable OAuth 2.1 / OpenID Connect provider. **Go
owns every protocol and security decision; Next.js owns every pixel.**
Nothing in the frontend is trusted as an authorization boundary — every
protected action is re-validated by Go against the session cookie.

## Locked design decisions

| Area | Decision | Why |
|---|---|---|
| First factor | Password (Argon2id) + TOTP 2FA | No third-party dependency (passkeys/social) for the core flow; keeps self-hosting simple |
| Tenancy | Single issuer, global users, orgs own OAuth clients | Simplest model that still lets teams share client ownership |
| Access tokens | JWT, ES256, `typ: at+jwt` (RFC 9068) | Offline verification via JWKS; explicit `typ` prevents confusion with ID tokens |
| Refresh tokens | Opaque, hashed, rotated, reuse-detected | Refresh tokens are long-lived — must be revocable and replay-safe |
| Subject IDs | **Pairwise per client** | Two relying parties cannot correlate the same user by comparing `sub` — the core privacy guarantee of this project |
| Token/session binding | Tokens without `offline_access` die with the session; `offline_access` survives logout | Matches OIDC scope semantics and the "authorized apps" mental model |
| Redis outage | Auth endpoints fail closed; read-only traffic fails open | An attacker who can take down Redis must not thereby unlock unlimited credential stuffing |
| Email verification | Required to approve consent or own a client; not required to log in | Blocks account-squatting from reaching a relying party without blocking signup |

## Why authorization codes live in Postgres, not Redis

A single-use code needs two properties: exactly-once redemption, and a
surviving record to detect replay. `GETDEL` in Redis gives the first and
destroys the second. Instead:

```sql
UPDATE authorization_codes
SET consumed_at = now()
WHERE code_hash = $1 AND consumed_at IS NULL
RETURNING refresh_family_id;
```

One atomic statement gives single-use redemption. The row survives
consumption, so if the same code is presented again, the handler finds
`consumed_at IS NOT NULL`, revokes the entire `refresh_family_id`, and
logs `code_replay` — treating it as evidence of a leaked code, not a
client bug.

## Authorization error handling (RFC 6749 §4.1.2.1)

- Invalid `client_id` or unregistered `redirect_uri` → **rendered error
  page**. Redirecting here would make the IdP an open redirector.
- Every other error (`invalid_scope`, `unsupported_response_type`,
  `access_denied`, `login_required`) → **redirect back** to the
  relying party with `error`, `error_description`, `state` (echoed
  verbatim), and `iss` (RFC 9207, prevents IdP mix-up attacks against
  RPs configured with multiple providers).

## Session / MFA state machine

Password success does **not** mint the real session cookie. It writes
`mfa_pending:<id>` to Redis (5 minutes, single purpose) and only
`/api/auth/login/totp` may consume it. This closes the gap where a
client that ignores the "go verify TOTP" response would otherwise be
sitting on a fully authenticated session with password alone.

The session token itself rotates on every privilege change — after
login, after the TOTP step, after password change — to prevent session
fixation via a cookie planted before authentication.

## Password hashing under load

Argon2id at 64 MB/hash is deliberately memory-hard, which means
concurrent login attempts exhaust RAM long before they exhaust CPU. A
counting semaphore bounds in-flight hash operations so a login flood
degrades to queuing/timeouts instead of OOM-killing the process.
Client secrets, by contrast, are high-entropy random values — they're
verified with HMAC-SHA256, not Argon2id, because hashing a secret that's
already unguessable just adds a self-inflicted DoS surface.

## Key management

ES256 signing keys are generated on first boot under the same
`pg_advisory_lock` used for migrations, so two replicas booting
simultaneously can't both decide they're the one to mint the "active"
key. Private halves are sealed with AES-256-GCM under
`KEY_ENCRYPTION_KEY`. **Losing that key is unrecoverable** — every
session, TOTP secret, and signing key depends on it; back it up before
going to production. Retired keys stay published in JWKS for at least
the maximum access-token TTL plus clock skew, so in-flight tokens don't
fail validation mid-rotation.

## Data model

See `server/db/migrations/0001_init.sql` for the authoritative schema.
Notable constraints:

- `users_email_lower_idx` — case-insensitive uniqueness without a
  `citext` extension dependency; the app layer NFKC-normalizes email.
- `consents_active_idx`, `refresh_tokens_session_idx` — partial unique
  indexes scoped to the *active* row, so revoked history doesn't block
  a new grant.
- `signing_keys_one_active_idx` — at most one `active` signing key at
  any time, enforced at the database level as a backstop to the
  advisory-lock discipline above.

## Out of scope for v1

Device authorization grant, dynamic client registration, OIDC
back-channel logout. All deferred to keep the first conformance target
(OpenID Foundation basic OP profile) achievable.

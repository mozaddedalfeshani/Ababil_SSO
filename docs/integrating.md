# Integrating with Ababil SSO

This guide is for **relying parties** (your apps) that sign users in through a
self-hosted Ababil SSO instance. For deploying the IdP itself, see
[`self-hosting.md`](./self-hosting.md). For why the protocol choices look the
way they do, see [`architecture.md`](./architecture.md).

## 1. Register an application

1. Sign in to the Ababil SSO console.
2. Create an **organization** (if you do not have one).
3. Open the org → **New application**.
4. Choose:
   - **Confidential** — server-side apps (web backends). You get a client
     secret **once** at create/rotate time.
   - **Public** — SPAs / native. No secret; **PKCE is mandatory**.
5. Add every exact `redirect_uri` you will use (no wildcards).

Copy the **Client ID**. Store any **client secret** in your secrets manager —
the console never shows it again.

## 2. Discovery

Prefer OpenID Connect discovery over hard-coded paths:

```
GET {issuer}/.well-known/openid-configuration
```

`issuer` is the public URL of your SSO instance (same value as `SSO_ISSUER` /
`NEXT_PUBLIC_APP_URL`).

Useful fields: `authorization_endpoint`, `token_endpoint`,
`userinfo_endpoint`, `jwks_uri`, `end_session_endpoint`,
`code_challenge_methods_supported` (must include `S256`).

## 3. Authorization Code + PKCE

Ababil SSO expects the OAuth 2.1 authorization code flow with PKCE.

1. Create a high-entropy `code_verifier` and S256 `code_challenge`.
2. Redirect the browser to:

```
{authorization_endpoint}
  ?client_id=…
  &redirect_uri=…          # must match registration exactly
  &response_type=code
  &scope=openid profile email
  &state=…                 # CSRF
  &code_challenge=…
  &code_challenge_method=S256
```

3. After consent (and login / email verification as required), the user returns
   with `?code=…&state=…&iss=…` (RFC 9207 `iss`).
4. Exchange at the token endpoint:

```http
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=…
&redirect_uri=…
&client_id=…
&code_verifier=…
```

Confidential clients also authenticate with their client secret
(`client_secret_post` / method advertised at registration).

## 4. Tokens

| Token | Form | Notes |
|---|---|---|
| Access token | JWT (ES256), `typ: at+jwt` | Verify with JWKS; check `iss`, `aud`, `exp` |
| ID token | JWT (when `openid` granted) | Standard OIDC claims; `email` only with `email` scope |
| Refresh token | Opaque | Rotated on use; reuse of an old token revokes the family |

**Subjects are pairwise per client.** The same human has a different `sub` at
each client — do not use `sub` to join users across products.

Access tokens without `offline_access` die with the browser session.
`offline_access` refresh tokens survive logout.

## 5. UserInfo & logout

```http
GET /oauth/userinfo
Authorization: Bearer {access_token}
```

RP-initiated logout:

```
GET /oauth/logout?id_token_hint=…&post_logout_redirect_uri=…
```

`post_logout_redirect_uri` must be registered on the client.

## 6. Email verification

Users can log into the SSO console before verifying email, but **cannot
approve OAuth consent** (or own clients) until email is verified via the
6-digit OTP flow.

## 7. End-to-end demo

```bash
# with a running local instance
./examples/rp-demo/run.sh
```

That script walks authorize → consent → token → refresh rotation → reuse
detection against your instance.

## Checklist

- [ ] Redirect URIs exact-match registered
- [ ] PKCE S256 on every authorize
- [ ] Confidential secret stored out of git
- [ ] JWKS-based JWT validation (ES256)
- [ ] Treat `sub` as per-client, not global
- [ ] Handle refresh rotation / reuse detection

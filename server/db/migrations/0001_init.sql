-- Ababil SSO initial schema.
-- Design notes are in /docs/architecture (see repo plan): pairwise subject
-- IDs, opaque refresh tokens, Postgres-backed single-use auth codes.

CREATE TABLE users (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email              text NOT NULL,
    email_verified_at  timestamptz,
    password_hash      text NOT NULL,
    totp_secret_enc    bytea,
    totp_last_step     bigint,
    totp_enabled_at    timestamptz,
    status             text NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'disabled')),
    created_at         timestamptz NOT NULL DEFAULT now(),
    last_login_at      timestamptz
);
-- Case-insensitive uniqueness without the citext extension dependency.
-- Application layer NFKC-normalizes email before every read/write.
CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE user_recovery_codes (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash    text NOT NULL,
    consumed_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX user_recovery_codes_user_idx ON user_recovery_codes (user_id);

CREATE TABLE sessions (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash          text NOT NULL,
    ip_hash             text,
    user_agent          text,
    amr                 text[] NOT NULL DEFAULT '{}',
    auth_time           timestamptz NOT NULL,
    idle_expires_at     timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at          timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX sessions_token_hash_idx ON sessions (token_hash);
CREATE INDEX sessions_user_idx ON sessions (user_id) WHERE revoked_at IS NULL;

CREATE TABLE email_tokens (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose      text NOT NULL CHECK (purpose IN ('verify', 'reset')),
    token_hash   text NOT NULL,
    expires_at   timestamptz NOT NULL,
    consumed_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX email_tokens_hash_idx ON email_tokens (token_hash);
CREATE INDEX email_tokens_user_idx ON email_tokens (user_id, purpose);

CREATE TABLE organizations (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name           text NOT NULL,
    slug           text NOT NULL,
    owner_user_id  uuid NOT NULL REFERENCES users(id),
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX organizations_slug_idx ON organizations (slug);

CREATE TABLE organization_members (
    org_id      uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        text NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, user_id)
);

CREATE TABLE oauth_clients (
    id                          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                      uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id                   text NOT NULL,
    client_secret_hmac         text,
    name                        text NOT NULL,
    logo_url                    text,
    client_type                 text NOT NULL CHECK (client_type IN ('public', 'confidential')),
    token_endpoint_auth_method text NOT NULL
                                 CHECK (token_endpoint_auth_method IN
                                        ('client_secret_basic', 'client_secret_post', 'none')),
    redirect_uris                text[] NOT NULL DEFAULT '{}',
    post_logout_redirect_uris    text[] NOT NULL DEFAULT '{}',
    grant_types                  text[] NOT NULL DEFAULT '{}',
    allowed_scopes               text[] NOT NULL DEFAULT '{}',
    subject_type                 text NOT NULL DEFAULT 'pairwise'
                                  CHECK (subject_type IN ('pairwise', 'public')),
    sector_identifier             text,
    require_consent               boolean NOT NULL DEFAULT true,
    created_by                    uuid NOT NULL REFERENCES users(id),
    disabled_at                   timestamptz,
    created_at                    timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX oauth_clients_client_id_idx ON oauth_clients (client_id);
CREATE INDEX oauth_clients_org_idx ON oauth_clients (org_id);

-- Stable pairwise subject identifier per (user, client): the load-bearing
-- privacy control that keeps two relying parties from correlating a user.
CREATE TABLE client_subjects (
    client_id      uuid NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pairwise_sub   text NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (client_id, user_id)
);
CREATE UNIQUE INDEX client_subjects_sub_idx ON client_subjects (client_id, pairwise_sub);

CREATE TABLE oauth_scopes (
    name         text PRIMARY KEY,
    description  text NOT NULL,
    is_default   boolean NOT NULL DEFAULT false
);
INSERT INTO oauth_scopes (name, description, is_default) VALUES
    ('openid', 'Authenticate via OpenID Connect', true),
    ('profile', 'Read basic profile information', false),
    ('email', 'Read email address and verification status', false),
    ('offline_access', 'Retain access when you are not present (refresh tokens survive logout)', false);

CREATE TABLE consents (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id    uuid NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
    scopes       text[] NOT NULL DEFAULT '{}',
    granted_at   timestamptz NOT NULL DEFAULT now(),
    revoked_at   timestamptz
);
-- Only one *active* consent per (user, client); history is kept, not overwritten.
CREATE UNIQUE INDEX consents_active_idx ON consents (user_id, client_id) WHERE revoked_at IS NULL;

-- Authorization codes live in Postgres, not Redis: the row must survive
-- consumption so a replay attempt can be traced to its refresh family.
CREATE TABLE authorization_codes (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash          text NOT NULL,
    client_id          uuid NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
    user_id            uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id         uuid NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    redirect_uri       text NOT NULL,
    scope              text NOT NULL,
    nonce              text,
    code_challenge     text NOT NULL,
    auth_time          timestamptz NOT NULL,
    refresh_family_id  uuid NOT NULL DEFAULT gen_random_uuid(),
    expires_at         timestamptz NOT NULL,
    consumed_at        timestamptz,
    created_at         timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX authorization_codes_hash_idx ON authorization_codes (code_hash);

CREATE TABLE refresh_tokens (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash     text NOT NULL,
    family_id      uuid NOT NULL,
    client_id      uuid NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id     uuid REFERENCES sessions(id) ON DELETE SET NULL,
    scope          text NOT NULL,
    -- Tokens without offline_access die with the session; offline_access
    -- grants (session_bound = false) survive logout by design.
    session_bound  boolean NOT NULL DEFAULT true,
    expires_at     timestamptz NOT NULL,
    consumed_at    timestamptz,
    rotated_to     uuid REFERENCES refresh_tokens(id),
    revoked_at     timestamptz,
    revoke_reason  text,
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX refresh_tokens_hash_idx ON refresh_tokens (token_hash);
CREATE INDEX refresh_tokens_family_idx ON refresh_tokens (family_id);
CREATE INDEX refresh_tokens_session_idx ON refresh_tokens (session_id) WHERE session_bound;

CREATE TABLE signing_keys (
    kid               text PRIMARY KEY,
    alg               text NOT NULL,
    private_key_enc   bytea NOT NULL,
    public_jwk        jsonb NOT NULL,
    status            text NOT NULL CHECK (status IN ('active', 'next', 'retired')),
    activates_at      timestamptz NOT NULL,
    retires_at        timestamptz,
    created_at        timestamptz NOT NULL DEFAULT now()
);
-- Only one active signing key at a time; enforced with the migration's
-- advisory lock at generation time, checked again here as a backstop.
CREATE UNIQUE INDEX signing_keys_one_active_idx ON signing_keys ((true)) WHERE status = 'active';

CREATE TABLE audit_logs (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id  uuid REFERENCES users(id) ON DELETE SET NULL,
    org_id         uuid REFERENCES organizations(id) ON DELETE SET NULL,
    client_id      uuid REFERENCES oauth_clients(id) ON DELETE SET NULL,
    event          text NOT NULL,
    ip_hash        text,
    user_agent     text,
    meta           jsonb NOT NULL DEFAULT '{}',
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_logs_actor_idx ON audit_logs (actor_user_id, created_at DESC);
CREATE INDEX audit_logs_org_idx ON audit_logs (org_id, created_at DESC);

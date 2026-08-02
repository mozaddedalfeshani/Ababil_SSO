#!/usr/bin/env bash
# CLI walkthrough of the full authorization_code + PKCE flow — see
# README.md. Fails fast (set -e) so a broken step stops the script
# with the actual curl response still visible above.
set -euo pipefail

BASE_URL="${SSO_BASE_URL:-http://localhost:5680}"
EMAIL="rp-demo-$(date +%s)@example.com"
PASSWORD="rp-demo-password-$(date +%s)"
REDIRECT_URI="http://localhost:9999/callback"
COOKIES="$(mktemp)"
trap 'rm -f "$COOKIES"' EXIT

step() { echo; echo "=== $1 ==="; }

json_get() {
  # $1 = key path passed to python (e.g. "['user']['id']"), reads JSON from stdin
  python3 -c "import json,sys; d=json.load(sys.stdin); print(d$1)"
}

pkce_pair() {
  # python3 -c, not a heredoc: a heredoc here reads from the *script's*
  # stdin machinery, which misbehaves when this function runs inside a
  # process substitution and the script's own stdin has already hit EOF
  # (as it does after step 1's `read` consumes it).
  python3 -c "
import base64, hashlib, secrets
verifier = base64.urlsafe_b64encode(secrets.token_bytes(32)).rstrip(b'=').decode()
challenge = base64.urlsafe_b64encode(hashlib.sha256(verifier.encode()).digest()).rstrip(b'=').decode()
print(verifier)
print(challenge)
"
}

step "1. Register + verify test user ($EMAIL)"
curl -sf -X POST "$BASE_URL/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" >/dev/null
echo "Registered. This demo relies on the console mailer — check your"
echo "server logs for the verify-email link, or run against a fresh dev"
echo "instance where SMTP is unset (the default)."
echo
echo "Paste the verification token from the server log (the part after"
echo "?token=), or press enter to skip verification (consent will fail):"
read -r VERIFY_TOKEN
if [ -n "$VERIFY_TOKEN" ]; then
  curl -sf -X POST "$BASE_URL/api/auth/verify-email" \
    -H 'Content-Type: application/json' \
    -d "{\"token\":\"$VERIFY_TOKEN\"}" >/dev/null
  echo "Verified."
fi

step "2. Log in"
curl -sf -c "$COOKIES" -X POST "$BASE_URL/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" >/dev/null
echo "Logged in."

csrf() { grep -i 'sso_csrf' "$COOKIES" | awk '{print $7}'; }

step "3. Create an organization"
ORG=$(curl -sf -b "$COOKIES" -X POST "$BASE_URL/api/orgs" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $(csrf)" \
  -d '{"name":"RP Demo Org"}')
ORG_ID=$(echo "$ORG" | json_get "['id']")
echo "Org: $ORG_ID"

step "4. Register a confidential OAuth client"
CLIENT=$(curl -sf -b "$COOKIES" -X POST "$BASE_URL/api/orgs/$ORG_ID/clients" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $(csrf)" \
  -d "{\"name\":\"RP Demo App\",\"client_type\":\"confidential\",\"redirect_uris\":[\"$REDIRECT_URI\"],\"allowed_scopes\":[\"openid\",\"email\",\"offline_access\"]}")
CLIENT_ID=$(echo "$CLIENT" | json_get "['client']['client_id']")
CLIENT_SECRET=$(echo "$CLIENT" | json_get "['client_secret']")
echo "Client ID: $CLIENT_ID"

step "5. /oauth/authorize with PKCE (S256)"
# Plain command substitution + sed, not mapfile — macOS ships bash 3.2
# by default, which has neither mapfile nor associative arrays.
PKCE_OUTPUT=$(pkce_pair)
VERIFIER=$(echo "$PKCE_OUTPUT" | sed -n '1p')
CHALLENGE=$(echo "$PKCE_OUTPUT" | sed -n '2p')
LOCATION=$(curl -sf -b "$COOKIES" -D - -o /dev/null \
  "$BASE_URL/oauth/authorize?client_id=$CLIENT_ID&redirect_uri=$REDIRECT_URI&response_type=code&scope=openid%20email%20offline_access&state=demo-state&code_challenge=$CHALLENGE&code_challenge_method=S256" \
  | grep -i '^location:' | sed 's/^[Ll]ocation: //' | tr -d '\r\n')
REQ_ID=$(echo "$LOCATION" | sed -n 's/.*req=\([^&]*\).*/\1/p')
echo "Parked authorization request: $REQ_ID"

step "6. Approve consent"
CONSENT=$(curl -sf -b "$COOKIES" -X POST "$BASE_URL/api/authorize/consent" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $(csrf)" \
  -d "{\"req_id\":\"$REQ_ID\",\"approved\":true}")
REDIRECT_TO=$(echo "$CONSENT" | json_get "['redirect_to']")
CODE=$(echo "$REDIRECT_TO" | sed -n 's/.*code=\([^&]*\).*/\1/p' | python3 -c "import sys,urllib.parse;print(urllib.parse.unquote(sys.stdin.read().strip()))")
echo "Authorization code issued."

step "7. Exchange code for tokens"
TOKENS=$(curl -sf -u "$CLIENT_ID:$CLIENT_SECRET" -X POST "$BASE_URL/oauth/token" \
  -d "grant_type=authorization_code&code=$CODE&redirect_uri=$REDIRECT_URI&code_verifier=$VERIFIER")
ACCESS_TOKEN=$(echo "$TOKENS" | json_get "['access_token']")
REFRESH_TOKEN=$(echo "$TOKENS" | json_get "['refresh_token']")
echo "Got access_token, id_token, refresh_token."

step "8. Call /oauth/userinfo"
curl -sf -H "Authorization: Bearer $ACCESS_TOKEN" "$BASE_URL/oauth/userinfo"
echo

step "9. Rotate the refresh token"
ROTATED=$(curl -sf -u "$CLIENT_ID:$CLIENT_SECRET" -X POST "$BASE_URL/oauth/token" \
  -d "grant_type=refresh_token&refresh_token=$REFRESH_TOKEN")
NEW_REFRESH_TOKEN=$(echo "$ROTATED" | json_get "['refresh_token']")
echo "Rotated OK — old token is now consumed."

step "10. Replay the OLD refresh token (expect rejection + family revoked)"
set +e
REPLAY=$(curl -s -u "$CLIENT_ID:$CLIENT_SECRET" -w '\n%{http_code}' -X POST "$BASE_URL/oauth/token" \
  -d "grant_type=refresh_token&refresh_token=$REFRESH_TOKEN")
set -e
echo "$REPLAY"

step "11. Confirm the family is dead — even the legitimately-rotated token fails now"
set +e
curl -s -u "$CLIENT_ID:$CLIENT_SECRET" -w '\n%{http_code}\n' -X POST "$BASE_URL/oauth/token" \
  -d "grant_type=refresh_token&refresh_token=$NEW_REFRESH_TOKEN"
set -e

echo
echo "Done. Steps 10-11 should both show invalid_grant / 400 — that's"
echo "refresh-token reuse detection working as designed (see"
echo "docs/architecture.md)."

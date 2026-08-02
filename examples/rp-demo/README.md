# RP Demo

A relying party normally has a browser in the loop for the
authorize/consent step. This demo is a CLI walkthrough instead — it
drives the exact same HTTP calls a browser-based RP would make (using
cookies to stand in for the browser session), so it doubles as a
regression check for the whole authorization_code + PKCE flow without
needing a UI.

## Requirements

`curl`, `python3` (used only for JSON parsing and PKCE generation —
no dependencies beyond the standard library).

## Usage

Start the stack first (`docker compose up` or `go run ./cmd/serve` +
`pnpm dev`), then:

```bash
./run.sh
```

It will:

1. Register and verify a throwaway user (via the console mailer log)
2. Create a test organization and a confidential OAuth client
3. Run the full `/oauth/authorize` → consent → `/oauth/token` exchange with PKCE (S256)
4. Call `/oauth/userinfo` with the resulting access token
5. Rotate the refresh token once, then replay the old one to prove reuse detection revokes the family

Each step prints what it's doing and the response it got — read it top
to bottom to see the protocol in action.

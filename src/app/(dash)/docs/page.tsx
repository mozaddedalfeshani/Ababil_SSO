import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { PageHeader } from "@/components/dash/page-header";
import { CopyableField } from "@/components/dash/copyable-field";

function issuerBase() {
  return (process.env.NEXT_PUBLIC_APP_URL || process.env.SSO_ISSUER || "http://localhost:5680").replace(/\/$/, "");
}

export default function DocsPage() {
  const issuer = issuerBase();
  const discovery = `${issuer}/.well-known/openid-configuration`;
  const authorize = `${issuer}/oauth/authorize`;
  const token = `${issuer}/oauth/token`;
  const userinfo = `${issuer}/oauth/userinfo`;
  const jwks = `${issuer}/oauth/jwks.json`;
  const logout = `${issuer}/oauth/logout`;

  return (
    <div className="flex flex-col gap-8">
      <PageHeader
        title="Documentation"
        description="Wire a relying party to this Ababil SSO instance. Discovery is the single source of truth for endpoints."
      />

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Issuer & discovery</CardTitle>
          <CardDescription>
            Point your OIDC library at the issuer. Prefer discovery over hard-coding paths.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <CopyableField label="Issuer" value={issuer} />
          <CopyableField label="OpenID discovery" value={discovery} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Protocol endpoints</CardTitle>
          <CardDescription>Same-origin with the console when served behind Caddy / Next proxy.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4">
          <CopyableField label="Authorize" value={authorize} />
          <CopyableField label="Token" value={token} />
          <CopyableField label="UserInfo" value={userinfo} />
          <CopyableField label="JWKS" value={jwks} />
          <CopyableField label="Logout" value={logout} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Quick integrate</CardTitle>
          <CardDescription>Authorization Code + PKCE (required). Confidential clients also use a client secret at the token endpoint.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-5 text-sm leading-relaxed text-muted-foreground">
          <ol className="list-decimal space-y-3 pl-5 text-foreground">
            <li>
              Create an <strong>organization</strong>, then a <strong>New application</strong> with your redirect URI(s).
            </li>
            <li>
              Copy the <strong>Client ID</strong> (and secret once, for confidential apps).
            </li>
            <li>
              Send the user to <code className="rounded bg-muted px-1 font-mono text-xs">/oauth/authorize</code> with{" "}
              <code className="rounded bg-muted px-1 font-mono text-xs">response_type=code</code>,{" "}
              <code className="rounded bg-muted px-1 font-mono text-xs">code_challenge</code> (S256),{" "}
              <code className="rounded bg-muted px-1 font-mono text-xs">scope</code>, and{" "}
              <code className="rounded bg-muted px-1 font-mono text-xs">redirect_uri</code>.
            </li>
            <li>
              Exchange the code at <code className="rounded bg-muted px-1 font-mono text-xs">/oauth/token</code> with{" "}
              <code className="rounded bg-muted px-1 font-mono text-xs">code_verifier</code>.
            </li>
            <li>
              Validate access tokens via JWKS (<code className="rounded bg-muted px-1 font-mono text-xs">ES256</code>,{" "}
              <code className="rounded bg-muted px-1 font-mono text-xs">typ: at+jwt</code>). Subjects are{" "}
              <strong>pairwise per client</strong>.
            </li>
          </ol>
          <Separator />
          <div>
            <p className="mb-2 font-medium text-foreground">Example authorize URL</p>
            <pre className="overflow-x-auto rounded-xl border border-border/60 bg-muted/40 p-3 font-mono text-[11px] leading-5 text-foreground">
{`${authorize}?
  client_id=YOUR_CLIENT_ID
  &redirect_uri=https://app.example.com/callback
  &response_type=code
  &scope=openid%20profile%20email
  &state=…
  &code_challenge=…
  &code_challenge_method=S256`}
            </pre>
          </div>
          <p>
            Full walkthrough (including refresh rotation): see repo{" "}
            <code className="rounded bg-muted px-1 font-mono text-xs">docs/integrating.md</code> and{" "}
            <code className="rounded bg-muted px-1 font-mono text-xs">examples/rp-demo/</code>.
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Scopes</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <p>
            <code className="rounded bg-muted px-1 font-mono text-xs">openid</code> — required for OIDC; issues an ID token.
          </p>
          <p>
            <code className="rounded bg-muted px-1 font-mono text-xs">profile</code> — basic profile claims when granted.
          </p>
          <p>
            <code className="rounded bg-muted px-1 font-mono text-xs">email</code> — email + email_verified.
          </p>
          <p>
            <code className="rounded bg-muted px-1 font-mono text-xs">offline_access</code> — refresh token that survives logout of the browser session.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

import Link from "next/link";
import { notFound } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { PageHeader } from "@/components/dash/page-header";
import { CopyableField } from "@/components/dash/copyable-field";
import { ClientActions } from "./client-actions";

type ClientDetail = {
  id: string;
  org_id: string;
  name: string;
  client_id: string;
  client_type: "public" | "confidential";
  token_endpoint_auth_method: string;
  redirect_uris: string[];
  post_logout_redirect_uris: string[];
  grant_types: string[];
  allowed_scopes: string[];
  subject_type: string;
  require_consent: boolean;
  disabled: boolean;
  created_at: string;
};

function MetaRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1.5 sm:flex-row sm:items-start sm:justify-between sm:gap-6">
      <span className="shrink-0 text-xs font-medium text-muted-foreground sm:w-44 sm:pt-0.5">{label}</span>
      <div className="min-w-0 flex-1 text-sm">{children}</div>
    </div>
  );
}

export default async function ClientDetailPage({ params }: { params: Promise<{ clientId: string }> }) {
  const { clientId } = await params;
  const { data: client, status } = await serverFetch<ClientDetail>(`/api/clients/${clientId}`);

  if (status === 404 || !client) notFound();

  return (
    <div className="flex flex-col gap-8">
      <div className="space-y-4">
        <Link href={`/orgs/${client.org_id}`} className="text-xs font-medium text-muted-foreground hover:text-foreground">
          ← Back to organization
        </Link>
        <PageHeader
          title={client.name}
          description={
            <span className="flex flex-wrap items-center gap-2">
              {client.disabled ? <Badge variant="destructive">Disabled</Badge> : <Badge variant="secondary">Active</Badge>}
              <Badge variant="outline" className="font-mono text-[10px] uppercase">
                {client.client_type}
              </Badge>
              <span>Created {new Date(client.created_at).toLocaleDateString()}</span>
            </span>
          }
        />
        <ClientActions clientId={client.id} clientType={client.client_type} disabled={client.disabled} />
      </div>

      <Card className="overflow-hidden">
        <CardHeader className="border-b border-border/60 bg-muted/20">
          <CardTitle className="text-base">Credentials</CardTitle>
          <CardDescription>
            Use these at the token endpoint. Secrets are never shown again after create/rotate.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-5 pt-6 sm:grid-cols-2">
          <CopyableField label="Client ID" value={client.client_id} />
          <div className="flex flex-col gap-1.5">
            <span className="text-xs font-medium text-muted-foreground">Client secret</span>
            <div className="rounded-xl border border-dashed border-border/80 bg-muted/20 px-3 py-2 text-sm text-muted-foreground">
              {client.client_type === "public"
                ? "Public clients have no secret (PKCE only)."
                : "Hidden after creation. Rotate to mint a new one."}
            </div>
          </div>
          <MetaRow label="Token auth method">
            <code className="font-mono text-xs">{client.token_endpoint_auth_method}</code>
          </MetaRow>
          <MetaRow label="Subject type">
            <code className="font-mono text-xs">{client.subject_type}</code>
          </MetaRow>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Application settings</CardTitle>
          <CardDescription>Redirects, grants, and scopes allowed for this client.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          <MetaRow label="Redirect URIs">
            <ul className="flex flex-col gap-1.5 font-mono text-xs">
              {client.redirect_uris.map((u) => (
                <li key={u} className="rounded-lg bg-muted/40 px-2 py-1 break-all">
                  {u}
                </li>
              ))}
            </ul>
          </MetaRow>
          <Separator />
          <MetaRow label="Grant types">
            <div className="flex flex-wrap gap-1.5">
              {client.grant_types.map((g) => (
                <Badge key={g} variant="outline" className="font-mono text-xs">
                  {g}
                </Badge>
              ))}
            </div>
          </MetaRow>
          <Separator />
          <MetaRow label="Allowed scopes">
            <div className="flex flex-wrap gap-1.5">
              {client.allowed_scopes.map((s) => (
                <Badge key={s} variant="outline" className="font-mono text-xs">
                  {s}
                </Badge>
              ))}
            </div>
          </MetaRow>
          <Separator />
          <MetaRow label="Consent">{client.require_consent ? "Required" : "Skipped (first-party)"}</MetaRow>
        </CardContent>
      </Card>
    </div>
  );
}

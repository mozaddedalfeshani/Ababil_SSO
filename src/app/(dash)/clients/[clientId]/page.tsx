import { notFound } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-xs font-medium text-muted-foreground">{label}</span>
      <div className="text-sm">{children}</div>
    </div>
  );
}

export default async function ClientDetailPage({ params }: { params: Promise<{ clientId: string }> }) {
  const { clientId } = await params;
  const { data: client, status } = await serverFetch<ClientDetail>(`/api/clients/${clientId}`);

  if (status === 404 || !client) notFound();

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight">{client.name}</h1>
            {client.disabled && <Badge variant="destructive">Disabled</Badge>}
          </div>
          <p className="text-sm text-muted-foreground">
            Created {new Date(client.created_at).toLocaleDateString()}
          </p>
        </div>
        <ClientActions clientId={client.id} clientType={client.client_type} disabled={client.disabled} />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Credentials</CardTitle>
          <CardDescription>Used to authenticate at the token endpoint.</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <Field label="Client ID">
            <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{client.client_id}</code>
          </Field>
          <Field label="Type">
            <Badge variant="secondary">{client.client_type}</Badge>
          </Field>
          <Field label="Token endpoint auth method">
            <code className="font-mono text-xs">{client.token_endpoint_auth_method}</code>
          </Field>
          <Field label="Subject type">
            <code className="font-mono text-xs">{client.subject_type}</code>
          </Field>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Configuration</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <Field label="Redirect URIs">
            <ul className="flex flex-col gap-1 font-mono text-xs">
              {client.redirect_uris.map((u) => (
                <li key={u}>{u}</li>
              ))}
            </ul>
          </Field>
          <Field label="Grant types">
            <div className="flex flex-wrap gap-1">
              {client.grant_types.map((g) => (
                <Badge key={g} variant="outline" className="font-mono text-xs">
                  {g}
                </Badge>
              ))}
            </div>
          </Field>
          <Field label="Allowed scopes">
            <div className="flex flex-wrap gap-1">
              {client.allowed_scopes.map((s) => (
                <Badge key={s} variant="outline" className="font-mono text-xs">
                  {s}
                </Badge>
              ))}
            </div>
          </Field>
          <Field label="Requires consent">{client.require_consent ? "Yes" : "No (first-party)"}</Field>
        </CardContent>
      </Card>
    </div>
  );
}

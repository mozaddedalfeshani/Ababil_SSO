"use client";

import { use, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { api, ApiError } from "@/lib/api";

const SCOPES = ["openid", "profile", "email", "offline_access"];

type CreateClientResult = { client: { id: string }; client_secret?: string };

export default function NewAppPage({ params }: { params: Promise<{ id: string }> }) {
  const { id: orgId } = use(params);
  const router = useRouter();

  const [name, setName] = useState("");
  const [clientType, setClientType] = useState<"confidential" | "public">("confidential");
  const [redirectURIs, setRedirectURIs] = useState("");
  const [scopes, setScopes] = useState<string[]>(["openid"]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<CreateClientResult | null>(null);

  function toggleScope(scope: string) {
    setScopes((prev) => (prev.includes(scope) ? prev.filter((s) => s !== scope) : [...prev, scope]));
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const uris = redirectURIs
        .split("\n")
        .map((s) => s.trim())
        .filter(Boolean);
      const res = await api.post<CreateClientResult>(`/api/orgs/${orgId}/clients`, {
        name,
        client_type: clientType,
        redirect_uris: uris,
        allowed_scopes: scopes,
      });
      setResult(res);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  if (result) {
    return (
      <Card className="mx-auto max-w-lg">
        <CardHeader>
          <CardTitle>Application created</CardTitle>
          <CardDescription>
            {result.client_secret
              ? "Save this client secret now — it will never be shown again."
              : "This is a public client; no secret is issued."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {result.client_secret && (
            <Alert>
              <AlertDescription className="break-all font-mono text-xs">{result.client_secret}</AlertDescription>
            </Alert>
          )}
        </CardContent>
        <CardFooter>
          <Button onClick={() => router.push(`/clients/${result.client.id}`)}>View application</Button>
        </CardFooter>
      </Card>
    );
  }

  return (
    <Card className="mx-auto max-w-lg">
      <CardHeader>
        <CardTitle>New application</CardTitle>
        <CardDescription>Register an OAuth 2.1 / OIDC client.</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="flex flex-col gap-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" required value={name} onChange={(e) => setName(e.target.value)} />
          </div>

          <div className="flex flex-col gap-2">
            <Label>Client type</Label>
            <Select value={clientType} onValueChange={(v) => setClientType(v as typeof clientType)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="confidential">Confidential (server-side app)</SelectItem>
                <SelectItem value="public">Public (SPA / native / CLI)</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="redirects">Redirect URIs (one per line)</Label>
            <textarea
              id="redirects"
              required
              rows={3}
              className="rounded-md border border-input bg-transparent px-3 py-2 font-mono text-sm shadow-xs outline-none focus-visible:ring-2 focus-visible:ring-ring"
              value={redirectURIs}
              onChange={(e) => setRedirectURIs(e.target.value)}
              placeholder="https://app.example.com/callback"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label>Allowed scopes</Label>
            <div className="flex flex-col gap-2">
              {SCOPES.map((scope) => (
                <label key={scope} className="flex items-center gap-2 text-sm">
                  <Checkbox checked={scopes.includes(scope)} onCheckedChange={() => toggleScope(scope)} />
                  <span className="font-mono">{scope}</span>
                </label>
              ))}
            </div>
          </div>

          <Button type="submit" disabled={loading || !name || scopes.length === 0}>
            {loading ? "Creating…" : "Create application"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

import Link from "next/link";
import { notFound } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

type Organization = { id: string; name: string; slug: string };
type ClientSummary = {
  id: string;
  name: string;
  client_id: string;
  client_type: "public" | "confidential";
  disabled: boolean;
};

export default async function OrgDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  const [{ data: org, status }, { data: clientsData }] = await Promise.all([
    serverFetch<Organization>(`/api/orgs/${id}`),
    serverFetch<{ clients: ClientSummary[] | null }>(`/api/orgs/${id}/clients`),
  ]);

  if (status === 404 || !org) notFound();
  const clients = clientsData?.clients ?? [];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">{org.name}</h1>
          <p className="font-mono text-xs text-muted-foreground">{org.slug}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" render={<Link href={`/orgs/${id}/members`} />}>
            Members
          </Button>
          <Button render={<Link href={`/orgs/${id}/apps/new`} />}>New app</Button>
        </div>
      </div>

      {clients.length === 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>No applications yet</CardTitle>
            <CardDescription>Register an OAuth client to start integrating.</CardDescription>
          </CardHeader>
        </Card>
      ) : (
        <div className="flex flex-col gap-3">
          {clients.map((c) => (
            <Link key={c.id} href={`/clients/${c.id}`}>
              <Card className="transition-colors hover:border-foreground/30">
                <CardContent className="flex items-center justify-between py-4">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{c.name}</span>
                      {c.disabled && <Badge variant="destructive">Disabled</Badge>}
                    </div>
                    <p className="font-mono text-xs text-muted-foreground">{c.client_id}</p>
                  </div>
                  <Badge variant="secondary">{c.client_type}</Badge>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

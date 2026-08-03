import Link from "next/link";
import { notFound } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/dash/page-header";

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
    <div className="flex flex-col gap-8">
      <PageHeader
        title={org.name}
        description={<span className="font-mono text-xs">{org.slug}</span>}
        actions={
          <>
            <Button variant="outline" render={<Link href={`/orgs/${id}/members`} />}>
              Members
            </Button>
            <Button render={<Link href={`/orgs/${id}/apps/new`} />}>New application</Button>
          </>
        }
      />

      <section className="space-y-3">
        <div className="flex items-end justify-between gap-2">
          <div>
            <h2 className="text-sm font-semibold tracking-tight">Applications</h2>
            <p className="text-xs text-muted-foreground">OAuth clients registered under this organization.</p>
          </div>
        </div>

        {clients.length === 0 ? (
          <Card className="border-dashed">
            <CardHeader>
              <CardTitle className="text-base">No applications yet</CardTitle>
              <CardDescription>Register an OAuth client to start integrating a relying party.</CardDescription>
            </CardHeader>
          </Card>
        ) : (
          <div className="flex flex-col gap-2">
            {clients.map((c) => (
              <Link key={c.id} href={`/clients/${c.id}`} className="group">
                <Card className="transition-colors group-hover:border-foreground/25 group-hover:bg-muted/20">
                  <CardContent className="flex items-center justify-between gap-4 py-4">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-medium">{c.name}</span>
                        {c.disabled ? <Badge variant="destructive">Disabled</Badge> : null}
                      </div>
                      <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">{c.client_id}</p>
                    </div>
                    <Badge variant="secondary" className="shrink-0 capitalize">
                      {c.client_type}
                    </Badge>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

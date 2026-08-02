import Link from "next/link";
import { serverFetch } from "@/lib/server-fetch";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { NewOrgDialog } from "./new-org-dialog";

type Organization = { id: string; name: string; slug: string; created_at: string };

export default async function DashboardPage() {
  const { data } = await serverFetch<{ organizations: Organization[] | null }>("/api/orgs");
  const orgs = data?.organizations ?? [];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Organizations</h1>
          <p className="text-sm text-muted-foreground">Organizations own the OAuth clients your apps use.</p>
        </div>
        <NewOrgDialog />
      </div>

      {orgs.length === 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>No organizations yet</CardTitle>
            <CardDescription>Create one to register your first OAuth client.</CardDescription>
          </CardHeader>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {orgs.map((org) => (
            <Link key={org.id} href={`/orgs/${org.id}`}>
              <Card className="transition-colors hover:border-foreground/30">
                <CardHeader>
                  <CardTitle className="text-base">{org.name}</CardTitle>
                  <CardDescription className="font-mono text-xs">{org.slug}</CardDescription>
                </CardHeader>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

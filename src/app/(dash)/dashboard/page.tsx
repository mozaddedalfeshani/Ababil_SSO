import Link from "next/link";
import { serverFetch } from "@/lib/server-fetch";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/dash/page-header";
import { NewOrgDialog } from "./new-org-dialog";

type Organization = { id: string; name: string; slug: string; created_at: string };

export default async function DashboardPage() {
  const { data } = await serverFetch<{ organizations: Organization[] | null }>("/api/orgs");
  const orgs = data?.organizations ?? [];

  return (
    <div className="flex flex-col gap-8">
      <PageHeader
        title="Organizations"
        description="Organizations own OAuth applications your products use to sign users in."
        actions={<NewOrgDialog />}
      />

      {orgs.length === 0 ? (
        <Card className="border-dashed">
          <CardHeader className="py-10 text-center">
            <CardTitle>Create your first organization</CardTitle>
            <CardDescription className="mx-auto max-w-sm">
              Then register an application to get a client ID (and secret for confidential apps).
            </CardDescription>
          </CardHeader>
        </Card>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {orgs.map((org) => (
            <Link key={org.id} href={`/orgs/${org.id}`} className="group">
              <Card className="h-full transition-colors group-hover:border-foreground/25 group-hover:bg-muted/20">
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

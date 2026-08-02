import { serverFetch } from "@/lib/server-fetch";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { AddMemberForm } from "./add-member-form";

type Member = { user_id: string; role: "owner" | "admin" | "member"; created_at: string };

export default async function OrgMembersPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { data } = await serverFetch<{ members: Member[] | null }>(`/api/orgs/${id}/members`);
  const members = data?.members ?? [];

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold tracking-tight">Members</h1>

      <Card>
        <CardHeader>
          <CardTitle>Add a member</CardTitle>
        </CardHeader>
        <CardContent>
          <AddMemberForm orgId={id} />
        </CardContent>
      </Card>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>User ID</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Joined</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {members.map((m) => (
            <TableRow key={m.user_id}>
              <TableCell className="font-mono text-xs">{m.user_id}</TableCell>
              <TableCell>
                <Badge variant={m.role === "owner" ? "default" : "secondary"}>{m.role}</Badge>
              </TableCell>
              <TableCell className="text-muted-foreground">{new Date(m.created_at).toLocaleDateString()}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

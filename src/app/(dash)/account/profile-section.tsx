import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { SessionUser } from "@/lib/session";
import { ResendVerificationButton } from "./resend-verification-button";

export function ProfileSection({ session }: { session: SessionUser }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Profile</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-1">
          <span className="text-xs font-medium text-muted-foreground">Email</span>
          <div className="flex items-center gap-2 text-sm">
            {session.email}
            {session.email_verified ? (
              <Badge variant="secondary">Verified</Badge>
            ) : (
              <Badge variant="destructive">Unverified</Badge>
            )}
          </div>
          {!session.email_verified && <ResendVerificationButton />}
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs font-medium text-muted-foreground">User ID</span>
          <code className="w-fit rounded bg-muted px-1.5 py-0.5 font-mono text-xs">{session.id}</code>
          <p className="text-xs text-muted-foreground">Share this with an org admin to be added as a member.</p>
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-xs font-medium text-muted-foreground">Member since</span>
          <span className="text-sm">{new Date(session.created_at).toLocaleDateString()}</span>
        </div>
      </CardContent>
    </Card>
  );
}

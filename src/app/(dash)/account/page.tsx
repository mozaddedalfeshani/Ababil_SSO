import { getSession } from "@/lib/session";
import { serverFetch } from "@/lib/server-fetch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageHeader } from "@/components/dash/page-header";
import { ProfileSection } from "./profile-section";
import { PasswordSection } from "./password-section";
import { TOTPSection } from "./totp-section";
import { SessionsSection } from "./sessions-section";
import { DangerSection } from "./danger-section";

export type SessionRow = {
  id: string;
  user_agent: string | null;
  amr: string[];
  auth_time: string;
  created_at: string;
  is_current: boolean;
};

export default async function AccountPage() {
  const session = await getSession();
  const [{ data: sessionsData }, { data: recoveryData }] = await Promise.all([
    serverFetch<{ sessions: SessionRow[] | null }>("/api/me/sessions"),
    serverFetch<{ remaining: number }>("/api/me/recovery-codes"),
  ]);

  if (!session) return null; // layout already guards this

  return (
    <div className="flex flex-col gap-8">
      <PageHeader title="Account" description="Profile, security, sessions, and data controls." />

      <Tabs defaultValue="profile">
        <TabsList>
          <TabsTrigger value="profile">Profile</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
          <TabsTrigger value="sessions">Sessions</TabsTrigger>
          <TabsTrigger value="danger">Data &amp; deletion</TabsTrigger>
        </TabsList>

        <TabsContent value="profile" className="mt-6">
          <ProfileSection session={session} />
        </TabsContent>

        <TabsContent value="security" className="mt-6 flex flex-col gap-6">
          <PasswordSection />
          <TOTPSection totpEnabled={session.totp_enabled} recoveryCodesRemaining={recoveryData?.remaining ?? 0} />
        </TabsContent>

        <TabsContent value="sessions" className="mt-6">
          <SessionsSection initialSessions={sessionsData?.sessions ?? []} />
        </TabsContent>

        <TabsContent value="danger" className="mt-6">
          <DangerSection />
        </TabsContent>
      </Tabs>
    </div>
  );
}

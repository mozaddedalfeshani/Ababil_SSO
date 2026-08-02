import { redirect } from "next/navigation";
import { serverFetch } from "@/lib/server-fetch";
import { ConsentCard } from "./consent-card";

type AuthRequestInfo = {
  client: { name: string; logo_url: string | null };
  scopes: string[];
  requires_login: boolean;
  requires_consent: boolean;
  email_unverified: boolean;
};

const SCOPE_DESCRIPTIONS: Record<string, string> = {
  openid: "Confirm your identity",
  profile: "Read your basic profile information",
  email: "Read your email address and verification status",
  offline_access: "Retain access when you're not present",
};

export default async function AuthorizePage({
  searchParams,
}: {
  searchParams: Promise<{ req?: string }>;
}) {
  const { req } = await searchParams;
  if (!req) {
    redirect("/login");
  }

  const { data, status } = await serverFetch<AuthRequestInfo>(`/api/auth-request/${req}`);

  if (status === 404) {
    return (
      <div className="rounded-lg border border-destructive/30 bg-destructive/10 p-6 text-center text-sm">
        This authorization request has expired. Return to the application and try again.
      </div>
    );
  }
  if (!data) {
    redirect("/login");
  }

  if (data.requires_login) {
    redirect(`/login?next=${encodeURIComponent(`/authorize?req=${req}`)}`);
  }

  return (
    <ConsentCard
      reqId={req}
      clientName={data.client.name}
      scopes={data.scopes.map((s) => ({ scope: s, description: SCOPE_DESCRIPTIONS[s] ?? s }))}
      requiresConsent={data.requires_consent}
      emailUnverified={data.email_unverified}
    />
  );
}

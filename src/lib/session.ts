import "server-only";
import { cookies } from "next/headers";

const BACKEND_URL = process.env.SSO_BACKEND_URL ?? "http://localhost:7897";

export type SessionUser = {
  id: string;
  email: string;
  email_verified: boolean;
  totp_enabled: boolean;
  created_at: string;
  last_login_at: string | null;
};

/**
 * The single place a server component reads the session. Every call
 * forwards the session cookie to Go and re-validates there — nothing
 * about auth state is trusted client-side or cached. `cache: "no-store"`
 * is non-negotiable here: a cached authed response served to a
 * different visitor is a full account disclosure (see
 * docs/architecture.md).
 */
export async function getSession(): Promise<SessionUser | null> {
  const cookieStore = await cookies();
  const cookieHeader = cookieStore.toString();
  if (!cookieHeader) return null;

  const res = await fetch(`${BACKEND_URL}/api/me`, {
    headers: { cookie: cookieHeader },
    cache: "no-store",
  });

  if (!res.ok) return null;
  return (await res.json()) as SessionUser;
}

export async function requireSession(): Promise<SessionUser> {
  const session = await getSession();
  if (!session) {
    throw new Error("unauthenticated");
  }
  return session;
}

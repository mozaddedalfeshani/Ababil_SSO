import "server-only";
import { cookies } from "next/headers";

const BACKEND_URL = process.env.SSO_BACKEND_URL ?? "http://localhost:7897";

/**
 * Server-component fetch to the Go backend, forwarding the session
 * cookie. Always `no-store` — see lib/session.ts for why that's
 * non-negotiable. Returns null on any non-OK response; callers decide
 * how to handle "not found" vs. "unauthenticated" from status if they
 * need to.
 */
export async function serverFetch<T>(path: string): Promise<{ data: T | null; status: number }> {
  const cookieStore = await cookies();
  const res = await fetch(`${BACKEND_URL}${path}`, {
    headers: { cookie: cookieStore.toString() },
    cache: "no-store",
  });

  if (!res.ok) return { data: null, status: res.status };
  return { data: (await res.json()) as T, status: res.status };
}

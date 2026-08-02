import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const BACKEND_URL = process.env.SSO_BACKEND_URL ?? "http://localhost:7897";

/**
 * Rewrites protocol/API paths to the Go backend so the browser only
 * ever talks to one origin (first-party cookies, no CORS — see
 * docs/architecture.md "Topology").
 *
 * This is NOT the authorization boundary. Proxy runs on Vercel's Edge
 * runtime and can only do optimistic, cheap checks — it must never be
 * the thing that decides whether a request is allowed. Every actual
 * protected read/write re-verifies the session against Go itself (see
 * lib/session.ts). All this does is avoid flashing a dashboard page
 * before redirecting an obviously-unauthenticated visitor to /login.
 */
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (
    pathname.startsWith("/oauth/") ||
    pathname.startsWith("/.well-known/") ||
    pathname.startsWith("/api/")
  ) {
    const target = new URL(pathname + request.nextUrl.search, BACKEND_URL);
    return NextResponse.rewrite(target);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/oauth/:path*", "/.well-known/:path*", "/api/:path*"],
};

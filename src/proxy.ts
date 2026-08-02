import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const BACKEND_URL = process.env.SSO_BACKEND_URL ?? "http://localhost:7897";

/**
 * Two unrelated jobs share this file because Next 16 allows exactly
 * one proxy.ts:
 *
 * 1. Rewrites protocol/API paths to the Go backend so the browser only
 *    ever talks to one origin (first-party cookies, no CORS — see
 *    docs/architecture.md "Topology"). NOT the authorization boundary:
 *    every protected read/write still re-verifies the session against
 *    Go itself (see lib/session.ts) — this only avoids flashing a
 *    dashboard page before redirecting an unauthenticated visitor.
 *
 * 2. Issues a per-request CSP nonce (script-src 'nonce-<value>', no
 *    unsafe-inline) so an injected <script> from a successful XSS
 *    still can't execute — this matters more than usual here since
 *    every auth page on this origin handles credentials or tokens.
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

  const nonce = Buffer.from(crypto.randomUUID()).toString("base64");
  const isDev = process.env.NODE_ENV !== "production";

  const csp = [
    "default-src 'self'",
    // 'unsafe-eval' only in dev — Turbopack's HMR runtime needs it;
    // production never receives it.
    `script-src 'self' 'nonce-${nonce}' 'strict-dynamic'${isDev ? " 'unsafe-eval'" : ""}`,
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data:",
    "font-src 'self'",
    "connect-src 'self'",
    "frame-ancestors 'none'",
    "base-uri 'self'",
    "form-action 'self'",
  ].join("; ");

  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-nonce", nonce);

  const response = NextResponse.next({ request: { headers: requestHeaders } });
  response.headers.set("Content-Security-Policy", csp);
  return response;
}

export const config = {
  matcher: [
    "/oauth/:path*",
    "/.well-known/:path*",
    "/api/:path*",
    // Everything else, excluding static assets — see Next's CSP guide.
    "/((?!_next/static|_next/image|favicon.ico).*)",
  ],
};

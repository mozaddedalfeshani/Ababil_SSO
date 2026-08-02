/** Reads the (non-HttpOnly, by design) CSRF cookie so client
 * components can echo it back in the X-CSRF-Token header — the
 * double-submit pattern the Go backend enforces on every
 * state-changing /api/* request. */
export function readCSRFCookie(): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(/(?:^|;\s*)(?:__Host-)?sso_csrf=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

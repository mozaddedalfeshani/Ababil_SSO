"use client";

import { Suspense, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { api, ApiError } from "@/lib/api";

/**
 * RP-initiated logout requires interactive confirmation — a bare GET
 * must never be able to end a session (see docs/architecture.md
 * "Logout hardening"). This page is that confirmation step; the
 * actual revocation happens on POST /api/oauth/logout/confirm.
 */
export default function LogoutPage() {
  return (
    <Suspense>
      <LogoutConfirm />
    </Suspense>
  );
}

function LogoutConfirm() {
  const params = useSearchParams();
  const idTokenHint = params.get("id_token_hint") ?? "";
  const postLogoutRedirectURI = params.get("post_logout_redirect_uri") ?? "";
  const state = params.get("state") ?? "";

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function confirm() {
    setLoading(true);
    setError(null);
    try {
      const res = await api.post<{ redirect_to: string }>("/api/oauth/logout/confirm", {
        id_token_hint: idTokenHint,
        post_logout_redirect_uri: postLogoutRedirectURI,
        state,
      });
      window.location.href = res.redirect_to || "/login";
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");
      setLoading(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Sign out?</CardTitle>
        <CardDescription>This will end your Ababil SSO session.</CardDescription>
      </CardHeader>
      <CardContent>
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </CardContent>
      <CardFooter className="flex gap-3">
        <Button variant="outline" className="flex-1" onClick={() => window.history.back()}>
          Cancel
        </Button>
        <Button className="flex-1" disabled={loading} onClick={confirm}>
          {loading ? "Signing out…" : "Sign out"}
        </Button>
      </CardFooter>
    </Card>
  );
}

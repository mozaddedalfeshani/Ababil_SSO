"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { api, ApiError } from "@/lib/api";

type ScopeItem = { scope: string; description: string };

type Props = {
  reqId: string;
  clientName: string;
  scopes: ScopeItem[];
  requiresConsent: boolean;
  emailUnverified: boolean;
};

export function ConsentCard({ reqId, clientName, scopes, requiresConsent, emailUnverified }: Props) {
  const [loading, setLoading] = useState(false);
  const [autoApproving, setAutoApproving] = useState(!requiresConsent && !emailUnverified);
  const [error, setError] = useState<string | null>(null);

  async function decide(approved: boolean) {
    setLoading(true);
    setError(null);
    try {
      const res = await api.post<{ redirect_to: string }>("/api/authorize/consent", {
        req_id: reqId,
        approved,
      });
      window.location.href = res.redirect_to;
    } catch (err) {
      setAutoApproving(false);
      setError(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (autoApproving) {
      decide(true);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (autoApproving) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Redirecting…</CardTitle>
          <CardDescription>Signing you in to {clientName}.</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Authorize {clientName}</CardTitle>
        <CardDescription>This application is requesting access to your account.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        {emailUnverified ? (
          <Alert variant="destructive">
            <AlertDescription>
              Verify your email before authorizing an application.{" "}
              <Link href="/account" className="underline">
                Go to account settings
              </Link>
            </AlertDescription>
          </Alert>
        ) : (
          <ul className="flex flex-col gap-2 text-sm">
            {scopes.map((s) => (
              <li key={s.scope} className="flex items-start gap-2">
                <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" />
                <span>{s.description}</span>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
      <CardFooter className="flex gap-3">
        <Button variant="outline" className="flex-1" disabled={loading} onClick={() => decide(false)}>
          Deny
        </Button>
        <Button className="flex-1" disabled={loading || emailUnverified} onClick={() => decide(true)}>
          {loading ? "Authorizing…" : "Authorize"}
        </Button>
      </CardFooter>
    </Card>
  );
}

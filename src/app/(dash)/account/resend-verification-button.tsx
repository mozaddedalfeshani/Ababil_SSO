"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { api, ApiError } from "@/lib/api";

export function ResendVerificationButton() {
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function resend() {
    setLoading(true);
    setError(null);
    try {
      await api.post("/api/auth/resend-verification");
      setSent(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  if (sent) return <p className="text-xs text-muted-foreground">Verification email sent.</p>;

  return (
    <div className="flex items-center gap-2">
      <Button size="sm" variant="outline" disabled={loading} onClick={resend}>
        Resend verification email
      </Button>
      {error && <span className="text-xs text-destructive">{error}</span>}
    </div>
  );
}

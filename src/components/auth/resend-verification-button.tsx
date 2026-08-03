"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { api, ApiError } from "@/lib/api";

export function ResendVerificationButton({ email }: { email: string }) {
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function resend() {
    setLoading(true);
    setError(null);
    try {
      await api.post("/api/auth/resend-verification", { email });
      setSent(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  if (sent) {
    return <p className="text-xs text-muted-foreground">Verification code sent.</p>;
  }

  return (
    <div className="flex flex-col gap-1">
      <Button size="sm" variant="outline" disabled={loading || !email} onClick={resend}>
        {loading ? "Sending…" : "Resend code"}
      </Button>
      {error && <span className="text-xs text-destructive">{error}</span>}
    </div>
  );
}

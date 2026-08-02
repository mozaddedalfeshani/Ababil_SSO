"use client";

import { Suspense, useState, type FormEvent } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp";
import { api, ApiError } from "@/lib/api";

export default function LoginTOTPPage() {
  return (
    <Suspense>
      <LoginTOTPForm />
    </Suspense>
  );
}

function LoginTOTPForm() {
  const router = useRouter();
  const params = useSearchParams();
  const mfaPendingId = params.get("mfa") ?? "";
  const next = params.get("next") ?? "/dashboard";

  const [code, setCode] = useState("");
  const [useRecovery, setUseRecovery] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await api.post("/api/auth/login/totp", {
        mfa_pending_id: mfaPendingId,
        code,
        recovery_code: useRecovery,
      });
      router.push(next);
      router.refresh();
    } catch (err) {
      if (err instanceof ApiError && err.code === "mfa_expired") {
        setError("Verification session expired. Please log in again.");
      } else {
        setError("Invalid code. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Two-factor verification</CardTitle>
        <CardDescription>
          {useRecovery
            ? "Enter one of your recovery codes."
            : "Enter the 6-digit code from your authenticator app."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {useRecovery ? (
            <div className="flex flex-col gap-2">
              <Label htmlFor="recovery">Recovery code</Label>
              <Input
                id="recovery"
                placeholder="xxxx-xxxx"
                autoComplete="off"
                required
                value={code}
                onChange={(e) => setCode(e.target.value)}
              />
            </div>
          ) : (
            <div className="flex flex-col items-center gap-2">
              <Label>Verification code</Label>
              <InputOTP maxLength={6} value={code} onChange={setCode}>
                <InputOTPGroup>
                  <InputOTPSlot index={0} />
                  <InputOTPSlot index={1} />
                  <InputOTPSlot index={2} />
                  <InputOTPSlot index={3} />
                  <InputOTPSlot index={4} />
                  <InputOTPSlot index={5} />
                </InputOTPGroup>
              </InputOTP>
            </div>
          )}

          <Button type="submit" disabled={loading || !mfaPendingId} className="mt-2">
            {loading ? "Verifying…" : "Verify"}
          </Button>

          <button
            type="button"
            className="text-center text-sm text-muted-foreground hover:text-foreground"
            onClick={() => {
              setUseRecovery((v) => !v);
              setCode("");
              setError(null);
            }}
          >
            {useRecovery ? "Use authenticator app instead" : "Use a recovery code instead"}
          </button>
        </form>
      </CardContent>
    </Card>
  );
}

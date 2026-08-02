"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp";
import { api, ApiError } from "@/lib/api";

type Props = { totpEnabled: boolean; recoveryCodesRemaining: number };

type EnrollResponse = { secret: string; otpauth_url: string };

export function TOTPSection({ totpEnabled, recoveryCodesRemaining }: Props) {
  const router = useRouter();
  const [step, setStep] = useState<"idle" | "enrolling" | "confirming">("idle");
  const [enrollment, setEnrollment] = useState<EnrollResponse | null>(null);
  const [code, setCode] = useState("");
  const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function startEnroll() {
    setError(null);
    setLoading(true);
    try {
      const res = await api.post<EnrollResponse>("/api/me/totp/enroll");
      setEnrollment(res);
      setStep("confirming");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  async function confirmEnroll(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await api.post<{ recovery_codes: string[] }>("/api/me/totp/verify", { code });
      setRecoveryCodes(res.recovery_codes);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Invalid code.");
    } finally {
      setLoading(false);
    }
  }

  async function disable(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await api.post("/api/me/totp/disable", { current_password: password });
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  if (recoveryCodes) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Save your recovery codes</CardTitle>
          <CardDescription>
            Each code can be used once if you lose access to your authenticator. They will not be shown again.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="grid grid-cols-2 gap-2 rounded-md border bg-muted/40 p-4 font-mono text-sm">
            {recoveryCodes.map((c) => (
              <span key={c}>{c}</span>
            ))}
          </div>
          <Button
            className="w-fit"
            onClick={() => {
              setRecoveryCodes(null);
              setStep("idle");
              router.refresh();
            }}
          >
            Done
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (step === "confirming" && enrollment) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Scan the QR code</CardTitle>
          <CardDescription>Or enter this key manually in your authenticator app.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <code className="w-fit rounded bg-muted px-2 py-1 font-mono text-sm">{enrollment.secret}</code>
          <form onSubmit={confirmEnroll} className="flex flex-col items-start gap-3">
            <Label>Enter the 6-digit code to confirm</Label>
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
            <div className="flex gap-2">
              <Button type="submit" disabled={loading || code.length !== 6}>
                {loading ? "Confirming…" : "Confirm"}
              </Button>
              <Button type="button" variant="ghost" onClick={() => setStep("idle")}>
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Two-factor authentication</CardTitle>
        <CardDescription>
          {totpEnabled
            ? `Enabled. ${recoveryCodesRemaining} recovery code${recoveryCodesRemaining === 1 ? "" : "s"} remaining.`
            : "Add an authenticator app for a second sign-in factor."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        {totpEnabled ? (
          <form onSubmit={disable} className="flex max-w-sm flex-col gap-3">
            <Label htmlFor="disable-password">Current password (required to disable)</Label>
            <Input
              id="disable-password"
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <Button type="submit" variant="destructive" disabled={loading} className="w-fit">
              {loading ? "Disabling…" : "Disable two-factor authentication"}
            </Button>
          </form>
        ) : (
          <Button onClick={startEnroll} disabled={loading}>
            {loading ? "Starting…" : "Enable two-factor authentication"}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

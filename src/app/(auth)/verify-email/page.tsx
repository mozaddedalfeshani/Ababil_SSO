"use client";

import { Suspense, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp";
import { ResendVerificationButton } from "@/components/auth/resend-verification-button";
import { api, ApiError } from "@/lib/api";

export default function VerifyEmailPage() {
  return (
    <Suspense>
      <VerifyEmailForm />
    </Suspense>
  );
}

function VerifyEmailForm() {
  const router = useRouter();
  const params = useSearchParams();
  const email = (params.get("email") ?? "").trim();

  const [otp, setOtp] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    if (!email) {
      setError("Missing email. Register again or open the link from your signup screen.");
      return;
    }
    if (otp.length !== 6) {
      setError("Enter the 6-digit code from your email.");
      return;
    }
    setLoading(true);
    try {
      await api.post("/api/auth/verify-email", { email, otp });
      setDone(true);
    } catch (err) {
      if (err instanceof ApiError && err.code === "invalid_token") {
        setError("That code is invalid or expired. Request a new one.");
      } else if (err instanceof ApiError && err.code === "rate_limited") {
        setError(err.message);
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  }

  if (done) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Email verified</CardTitle>
          <CardDescription>Your email address has been confirmed.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button className="w-full" onClick={() => router.push("/login")}>
            Continue to sign in
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Verify your email</CardTitle>
        <CardDescription>
          {email
            ? <>Enter the 6-digit code sent to <strong>{email}</strong>.</>
            : "Enter the 6-digit code from your email. Open this page from the signup screen so we know which address to verify."}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="flex flex-col items-center gap-2">
            <Label>Verification code</Label>
            <InputOTP maxLength={6} value={otp} onChange={setOtp} autoFocus>
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
          <Button type="submit" disabled={loading || !email || otp.length !== 6}>
            {loading ? "Verifying…" : "Verify email"}
          </Button>
        </form>
        {email ? <ResendVerificationButton email={email} /> : null}
        <Link href="/login" className="text-sm font-medium hover:underline">
          Back to sign in
        </Link>
      </CardContent>
    </Card>
  );
}

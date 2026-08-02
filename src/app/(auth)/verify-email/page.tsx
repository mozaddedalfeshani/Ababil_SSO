"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";

type Status = "verifying" | "success" | "error";

export default function VerifyEmailPage() {
  return (
    <Suspense>
      <VerifyEmailStatus />
    </Suspense>
  );
}

function VerifyEmailStatus() {
  const params = useSearchParams();
  const token = params.get("token");
  // No-token is derivable directly from the render input — only the
  // actual network call needs an effect.
  const [status, setStatus] = useState<Status>(token ? "verifying" : "error");

  useEffect(() => {
    if (!token) return;
    api
      .post("/api/auth/verify-email", { token })
      .then(() => setStatus("success"))
      .catch(() => setStatus("error"));
  }, [token]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {status === "verifying" && "Verifying your email…"}
          {status === "success" && "Email verified"}
          {status === "error" && "Verification failed"}
        </CardTitle>
        <CardDescription>
          {status === "success" && "Your email address has been confirmed."}
          {status === "error" && "This link is invalid or has expired. Request a new one from your account."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Link href="/login" className="text-sm font-medium hover:underline">
          Back to sign in
        </Link>
      </CardContent>
    </Card>
  );
}

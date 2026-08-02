"use client";

import { useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { api, ApiError } from "@/lib/api";

export function DangerSection() {
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  function downloadExport() {
    // Uses the browser's own cookie-authenticated navigation rather
    // than fetch+blob — simplest way to trigger a same-origin
    // download without re-implementing auth for it.
    window.open("/api/me/export", "_blank");
  }

  async function deleteAccount(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await api.delete("/api/me", { current_password: password });
      window.location.href = "/login";
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong.");
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Export your data</CardTitle>
          <CardDescription>Download everything Ababil SSO holds about your account.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" onClick={downloadExport}>
            Download export
          </Button>
        </CardContent>
      </Card>

      <Card className="border-destructive/40">
        <CardHeader>
          <CardTitle>Delete account</CardTitle>
          <CardDescription>
            Permanently deletes your account, sessions, and authorized apps. This cannot be undone.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <AlertDialog>
            <AlertDialogTrigger render={<Button variant="destructive" />}>Delete my account</AlertDialogTrigger>
            <AlertDialogContent>
              <form onSubmit={deleteAccount}>
                <AlertDialogHeader>
                  <AlertDialogTitle>Confirm account deletion</AlertDialogTitle>
                  <AlertDialogDescription>
                    Enter your password to permanently delete your account.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <div className="py-4">
                  <Label htmlFor="delete-password" className="sr-only">
                    Password
                  </Label>
                  <Input
                    id="delete-password"
                    type="password"
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
                <AlertDialogFooter>
                  <AlertDialogCancel type="button">Cancel</AlertDialogCancel>
                  <AlertDialogAction type="submit" variant="destructive" disabled={loading}>
                    {loading ? "Deleting…" : "Delete account"}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </form>
            </AlertDialogContent>
          </AlertDialog>
        </CardContent>
      </Card>
    </div>
  );
}

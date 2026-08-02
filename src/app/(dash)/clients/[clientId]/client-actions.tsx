"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
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
import { Alert, AlertDescription } from "@/components/ui/alert";
import { api } from "@/lib/api";

type Props = { clientId: string; clientType: "public" | "confidential"; disabled: boolean };

export function ClientActions({ clientId, clientType, disabled }: Props) {
  const router = useRouter();
  const [newSecret, setNewSecret] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function rotateSecret() {
    setBusy(true);
    try {
      const res = await api.post<{ client_secret: string }>(`/api/clients/${clientId}/rotate-secret`);
      setNewSecret(res.client_secret);
    } finally {
      setBusy(false);
    }
  }

  async function disableClient() {
    setBusy(true);
    try {
      await api.delete(`/api/clients/${clientId}`);
      router.refresh();
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col items-end gap-3">
      <div className="flex gap-2">
        {clientType === "confidential" && (
          <AlertDialog>
            <AlertDialogTrigger render={<Button variant="outline" />}>Rotate secret</AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Rotate client secret?</AlertDialogTitle>
                <AlertDialogDescription>
                  The current secret stops working immediately — every deployed instance of this app must be
                  updated with the new one.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction disabled={busy} onClick={rotateSecret}>
                  Rotate
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}
        {!disabled && (
          <AlertDialog>
            <AlertDialogTrigger render={<Button variant="destructive" />}>Disable</AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Disable this application?</AlertDialogTitle>
                <AlertDialogDescription>
                  No new authorizations or token exchanges will succeed for this client.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction disabled={busy} onClick={disableClient}>
                  Disable
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}
      </div>
      {newSecret && (
        <Alert className="max-w-md">
          <AlertDescription className="break-all font-mono text-xs">
            New secret (shown once): {newSecret}
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}

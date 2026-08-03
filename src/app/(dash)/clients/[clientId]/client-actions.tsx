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
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { CopyButton } from "@/components/dash/copy-button";
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
    <>
      <div className="flex flex-wrap justify-end gap-2">
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

      {newSecret ? (
        <Alert className="border-amber-500/40 bg-amber-500/10 text-foreground">
          <AlertTitle className="flex items-center justify-between gap-2">
            <span>New client secret</span>
            <CopyButton value={newSecret} label="Copy secret" />
          </AlertTitle>
          <AlertDescription className="mt-2 space-y-2">
            <p className="text-sm text-muted-foreground">
              Shown once. Store it in your secrets manager — you can&apos;t retrieve it later.
            </p>
            <code className="block rounded-xl border border-border/60 bg-background/80 px-3 py-2 font-mono text-xs break-all">
              {newSecret}
            </code>
          </AlertDescription>
        </Alert>
      ) : null}
    </>
  );
}

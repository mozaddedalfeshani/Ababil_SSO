"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";
import type { SessionRow } from "./page";

export function SessionsSection({ initialSessions }: { initialSessions: SessionRow[] }) {
  const [sessions, setSessions] = useState(initialSessions);
  const [busyId, setBusyId] = useState<string | null>(null);

  async function revoke(id: string) {
    setBusyId(id);
    try {
      await api.delete(`/api/me/sessions/${id}`);
      setSessions((prev) => prev.filter((s) => s.id !== id));
    } finally {
      setBusyId(null);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Active sessions</CardTitle>
        <CardDescription>Devices and browsers currently signed in to your account.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {sessions.length === 0 && <p className="text-sm text-muted-foreground">No active sessions.</p>}
        {sessions.map((s) => (
          <div key={s.id} className="flex items-center justify-between rounded-md border p-3">
            <div>
              <div className="flex items-center gap-2 text-sm">
                {s.user_agent ?? "Unknown device"}
                {s.is_current && <Badge variant="secondary">This device</Badge>}
              </div>
              <p className="text-xs text-muted-foreground">
                Signed in {new Date(s.auth_time).toLocaleString()} · {s.amr.join(", ")}
              </p>
            </div>
            <Button
              size="sm"
              variant="outline"
              disabled={busyId === s.id}
              onClick={() => revoke(s.id)}
            >
              Revoke
            </Button>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

"use client";

import { useState, type ReactNode } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { HugeiconsIcon } from "@hugeicons/react";
import {
  BookOpen01Icon,
  Building03Icon,
  Menu01Icon,
  UserCircleIcon,
  Cancel01Icon,
} from "@hugeicons/core-free-icons";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { api, ApiError } from "@/lib/api";
import { readCSRFCookie } from "@/lib/csrf";
import { AbabilLogoMark } from "@/components/brand/ababil-mark";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/dashboard", label: "Organizations", icon: Building03Icon, match: (p: string) => p === "/dashboard" || p.startsWith("/orgs") || p.startsWith("/clients") },
  { href: "/docs", label: "Docs", icon: BookOpen01Icon, match: (p: string) => p.startsWith("/docs") },
  { href: "/account", label: "Account", icon: UserCircleIcon, match: (p: string) => p.startsWith("/account") },
];

export function DashShell({ email, children }: { email: string; children: ReactNode }) {
  const pathname = usePathname();
  const [mobileOpen, setMobileOpen] = useState(false);

  async function signOut() {
    if (!readCSRFCookie()) return;
    try {
      await api.post("/api/auth/logout");
    } catch (err) {
      if (!(err instanceof ApiError)) throw err;
    }
    window.location.href = "/login";
  }

  const sidebar = (
    <div className="flex h-full flex-col">
      <div className="flex h-14 items-center gap-2.5 border-b border-sidebar-border px-4">
        <Link href="/dashboard" className="flex items-center gap-2.5 font-semibold tracking-tight" onClick={() => setMobileOpen(false)}>
          <AbabilLogoMark className="bg-sidebar-primary text-sidebar-primary-foreground" />
          <span className="text-sm">Ababil SSO</span>
        </Link>
      </div>

      <nav className="flex flex-1 flex-col gap-1 p-3">
        <p className="px-2 pb-1 text-[11px] font-medium tracking-wide text-muted-foreground uppercase">Console</p>
        {nav.map((item) => {
          const active = item.match(pathname);
          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => setMobileOpen(false)}
              className={cn(
                "flex items-center gap-2.5 rounded-xl px-2.5 py-2 text-sm transition-colors",
                active
                  ? "bg-sidebar-accent font-medium text-sidebar-accent-foreground"
                  : "text-sidebar-foreground/70 hover:bg-sidebar-accent/70 hover:text-sidebar-foreground",
              )}
            >
              <HugeiconsIcon icon={item.icon} strokeWidth={2} className="size-4 shrink-0 opacity-80" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="border-t border-sidebar-border p-3">
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <button className="flex w-full items-center gap-2.5 rounded-xl px-2 py-2 text-left transition-colors hover:bg-sidebar-accent" />
            }
          >
            <Avatar className="size-8">
              <AvatarFallback className="text-[10px]">{email.slice(0, 2).toUpperCase()}</AvatarFallback>
            </Avatar>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{email}</p>
              <p className="text-[11px] text-muted-foreground">Signed in</p>
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="w-56">
            <div className="px-2 py-1.5 text-sm text-muted-foreground">{email}</div>
            <DropdownMenuSeparator />
            <DropdownMenuItem render={<Link href="/account" />}>Account settings</DropdownMenuItem>
            <DropdownMenuItem onClick={signOut}>Sign out</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );

  return (
    <div className="flex min-h-screen bg-background">
      <aside className="sticky top-0 hidden h-screen w-60 shrink-0 border-r border-sidebar-border bg-sidebar text-sidebar-foreground md:block">
        {sidebar}
      </aside>

      {mobileOpen ? (
        <div className="fixed inset-0 z-50 md:hidden">
          <button type="button" className="absolute inset-0 bg-black/40" aria-label="Close menu" onClick={() => setMobileOpen(false)} />
          <aside className="relative h-full w-72 max-w-[85vw] bg-sidebar text-sidebar-foreground shadow-xl">
            <button
              type="button"
              className="absolute top-3 right-3 rounded-lg p-2 text-muted-foreground hover:bg-sidebar-accent"
              aria-label="Close"
              onClick={() => setMobileOpen(false)}
            >
              <HugeiconsIcon icon={Cancel01Icon} strokeWidth={2} className="size-4" />
            </button>
            {sidebar}
          </aside>
        </div>
      ) : null}

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-40 flex h-14 items-center gap-3 border-b border-border bg-background/80 px-4 backdrop-blur md:hidden">
          <Button type="button" variant="ghost" size="icon-sm" onClick={() => setMobileOpen(true)} aria-label="Open menu">
            <HugeiconsIcon icon={Menu01Icon} strokeWidth={2} className="size-4" />
          </Button>
          <AbabilLogoMark className="size-6 rounded-md" markClassName="size-3.5" />
          <span className="text-sm font-semibold">Ababil SSO</span>
        </header>
        <main className="mx-auto w-full max-w-5xl flex-1 px-4 py-8 sm:px-6 sm:py-10">{children}</main>
      </div>
    </div>
  );
}

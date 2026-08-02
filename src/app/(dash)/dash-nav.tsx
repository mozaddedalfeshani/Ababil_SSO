"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { api, ApiError } from "@/lib/api";
import { readCSRFCookie } from "@/lib/csrf";

const links = [
  { href: "/dashboard", label: "Organizations" },
  { href: "/account", label: "Account" },
];

export function DashNav({ email }: { email: string }) {
  const pathname = usePathname();

  async function signOut() {
    // Logout requires CSRF but not a JSON body — api.post already
    // attaches the header via readCSRFCookie().
    if (!readCSRFCookie()) return;
    try {
      await api.post("/api/auth/logout");
    } catch (err) {
      if (!(err instanceof ApiError)) throw err;
    }
    window.location.href = "/login";
  }

  return (
    <header className="border-b border-border">
      <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
        <div className="flex items-center gap-8">
          <Link href="/dashboard" className="flex items-center gap-2 text-sm font-semibold">
            <span className="flex h-6 w-6 items-center justify-center rounded bg-primary text-xs text-primary-foreground">
              A
            </span>
            Ababil SSO
          </Link>
          <nav className="flex items-center gap-4">
            {links.map((l) => (
              <Link
                key={l.href}
                href={l.href}
                className={
                  "text-sm " +
                  (pathname.startsWith(l.href) ? "font-medium text-foreground" : "text-muted-foreground hover:text-foreground")
                }
              >
                {l.label}
              </Link>
            ))}
          </nav>
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger render={<button className="flex items-center gap-2 rounded-full" />}>
            <Avatar className="h-7 w-7">
              <AvatarFallback className="text-xs">{email.slice(0, 2).toUpperCase()}</AvatarFallback>
            </Avatar>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <div className="px-2 py-1.5 text-sm text-muted-foreground">{email}</div>
            <DropdownMenuSeparator />
            <DropdownMenuItem render={<Link href="/account" />}>Account settings</DropdownMenuItem>
            <DropdownMenuItem onClick={signOut}>Sign out</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}

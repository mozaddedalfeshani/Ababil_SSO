import { redirect } from "next/navigation";
import { getSession } from "@/lib/session";
import { DashShell } from "@/components/dash/dash-shell";

export default async function DashLayout({ children }: { children: React.ReactNode }) {
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  return <DashShell email={session.email}>{children}</DashShell>;
}

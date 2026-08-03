import Link from "next/link";
import { AbabilLogoMark } from "@/components/brand/ababil-mark";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-8 bg-background px-4 py-12">
      <Link href="/" className="flex items-center gap-2.5 text-lg font-semibold tracking-tight">
        <AbabilLogoMark className="size-8 rounded-md" markClassName="size-[1.15rem]" />
        Ababil SSO
      </Link>
      <div className="w-full max-w-sm">{children}</div>
    </div>
  );
}

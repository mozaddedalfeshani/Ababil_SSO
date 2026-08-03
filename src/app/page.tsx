import Link from "next/link";
import { Button } from "@/components/ui/button";
import { AbabilLogoMark } from "@/components/brand/ababil-mark";

const FEATURES = [
  {
    title: "Pairwise identities",
    body: "Every relying party gets a different subject identifier for the same user — apps can't correlate your identity across each other.",
  },
  {
    title: "PKCE-only, always",
    body: "Authorization code + mandatory S256 PKCE, opaque refresh tokens with rotation and reuse detection.",
  },
  {
    title: "No telemetry",
    body: "No analytics, no third-party runtime calls, hashed IPs, bounded audit retention.",
  },
];

export default function Home() {
  return (
    <div className="flex flex-1 flex-col">
      <header className="mx-auto flex w-full max-w-5xl items-center justify-between px-6 py-6">
        <div className="flex items-center gap-2.5 text-sm font-semibold">
          <AbabilLogoMark />
          Ababil SSO
        </div>
        <div className="flex items-center gap-3">
          <Button variant="ghost" render={<Link href="/login" />}>
            Sign in
          </Button>
          <Button render={<Link href="/register" />}>Get started</Button>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col items-center justify-center gap-8 px-6 py-24 text-center">
        <h1 className="text-4xl font-semibold tracking-tight text-balance sm:text-5xl">
          Open-source, privacy-first identity
        </h1>
        <p className="max-w-xl text-lg text-muted-foreground text-balance">
          A self-hostable OAuth 2.1 / OpenID Connect provider. Go owns every protocol and security decision; you own
          the data.
        </p>
        <div className="flex gap-3">
          <Button size="lg" render={<Link href="/register" />}>
            Create an account
          </Button>
          <Button size="lg" variant="outline" render={<a href="https://github.com" target="_blank" rel="noreferrer" />}>
            View on GitHub
          </Button>
        </div>
      </main>

      <section className="mx-auto grid w-full max-w-5xl gap-6 px-6 pb-24 sm:grid-cols-3">
        {FEATURES.map((f) => (
          <div key={f.title} className="rounded-lg border border-border p-6">
            <h3 className="mb-2 text-sm font-semibold">{f.title}</h3>
            <p className="text-sm text-muted-foreground">{f.body}</p>
          </div>
        ))}
      </section>
    </div>
  );
}

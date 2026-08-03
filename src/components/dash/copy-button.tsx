"use client";

import { useState } from "react";
import { HugeiconsIcon } from "@hugeicons/react";
import { Copy01Icon, Tick02Icon } from "@hugeicons/core-free-icons";
import { cn } from "@/lib/utils";

type Props = {
  value: string;
  className?: string;
  label?: string;
};

export function CopyButton({ value, className, label = "Copy" }: Props) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      /* clipboard may be denied — ignore */
    }
  }

  return (
    <button
      type="button"
      onClick={copy}
      aria-label={copied ? "Copied" : label}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-lg px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground",
        className,
      )}
    >
      <HugeiconsIcon icon={copied ? Tick02Icon : Copy01Icon} strokeWidth={2} className="size-3.5" />
      {copied ? "Copied" : label}
    </button>
  );
}

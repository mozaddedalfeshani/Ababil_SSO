import type { ReactNode } from "react";
import { CopyButton } from "./copy-button";
import { cn } from "@/lib/utils";

type Props = {
  label: string;
  value: string;
  mono?: boolean;
  className?: string;
  children?: ReactNode;
};

export function CopyableField({ label, value, mono = true, className, children }: Props) {
  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-medium text-muted-foreground">{label}</span>
        <CopyButton value={value} />
      </div>
      <div
        className={cn(
          "rounded-xl border border-border/80 bg-muted/40 px-3 py-2 text-sm break-all",
          mono && "font-mono text-xs",
        )}
      >
        {children ?? value}
      </div>
    </div>
  );
}

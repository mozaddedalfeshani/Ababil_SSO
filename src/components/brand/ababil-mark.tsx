import { cn } from "@/lib/utils";

type MarkProps = {
  className?: string;
  title?: string;
};

/**
 * Ababil mark — shield + keyhole.
 * Shield = trust boundary / IdP. Keyhole = gated access to identity.
 * Uses currentColor so tiles can theme via text-* utilities.
 */
export function AbabilMark({ className, title }: MarkProps) {
  return (
    <svg
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={cn("size-full", className)}
      aria-hidden={title ? undefined : true}
      role={title ? "img" : undefined}
    >
      {title ? <title>{title}</title> : null}
      <path
        d="M16 2.8 27 7.2v8.6c0 6.6-4.6 12.2-11 14-6.4-1.8-11-7.4-11-14V7.2L16 2.8Z"
        fill="currentColor"
        fillOpacity="0.2"
      />
      <path
        d="M16 2.8 27 7.2v8.6c0 6.6-4.6 12.2-11 14-6.4-1.8-11-7.4-11-14V7.2L16 2.8Z"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinejoin="round"
      />
      {/* Keyhole: circle + stem (union as paths for crisp small sizes) */}
      <circle cx="16" cy="13.2" r="3.4" fill="currentColor" />
      <path
        d="M13.35 15.6c.55 1.15 1.5 2.55 2.65 4.05 1.15-1.5 2.1-2.9 2.65-4.05H13.35Z"
        fill="currentColor"
      />
      <path d="M14.15 19.2h3.7v3.35c0 .55-.45 1-1 1h-1.7c-.55 0-1-.45-1-1V19.2Z" fill="currentColor" />
    </svg>
  );
}

type LogoMarkProps = {
  className?: string;
  markClassName?: string;
};

/** Rounded tile used in nav / auth headers. Parent sets bg + text color. */
export function AbabilLogoMark({ className, markClassName }: LogoMarkProps) {
  return (
    <span
      className={cn(
        "inline-flex size-7 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground",
        className,
      )}
    >
      <AbabilMark className={cn("size-[1.1rem]", markClassName)} />
    </span>
  );
}

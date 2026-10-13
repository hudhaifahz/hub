import ExternalLink from "src/components/ExternalLink";
import { cn } from "src/lib/utils";
import {
  GRYNVAULT_PRODUCT_NAME,
  GRYNVAULT_SOURCE_URL,
  UPSTREAM_ALBY_VERSION,
} from "src/lib/grynvault";

export function GrynvaultForkNotice({ className }: { className?: string }) {
  return (
    <aside
      aria-label="Independent fork notice"
      className={cn(
        "border-b bg-amber-50 px-4 py-2 text-center text-xs text-amber-950 dark:bg-amber-950 dark:text-amber-50",
        className
      )}
    >
      <span className="font-semibold">{GRYNVAULT_PRODUCT_NAME}</span> is an
      experimental, independently maintained fork built from Alby Hub{" "}
      {UPSTREAM_ALBY_VERSION}. It is not an official Alby release.{" "}
      <ExternalLink className="underline" to={GRYNVAULT_SOURCE_URL}>
        Review source
      </ExternalLink>
      .
    </aside>
  );
}

import React from "react";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";

export function useBanner() {
  const { data: info } = useInfo();
  const { data: albyMe } = useAlbyMe();
  const [showBanner, setShowBanner] = React.useState(false);
  const isDismissedRef = React.useRef(false);

  React.useEffect(() => {
    if (!info || info.hideUpdateBanner || isDismissedRef.current) {
      return;
    }

    // vss migration (alby cloud only)
    // TODO: remove after 2026-08-01
    const vssMigrationRequired =
      info.oauthRedirect &&
      !!albyMe?.subscription.plan_code.includes("buzz") &&
      info.vssSupported &&
      !info.ldkVssEnabled;

    // Grynvault is a tested fork, not a skin over the current upstream binary.
    // Do not turn an upstream Alby release into an actionable update prompt.
    // Compatible Grynvault releases are announced only after their own release
    // gates pass.
    setShowBanner(vssMigrationRequired);
  }, [info, albyMe?.subscription.plan_code]);

  const dismissBanner = () => {
    isDismissedRef.current = true;
    setShowBanner(false);
  };

  return { showBanner, dismissBanner };
}

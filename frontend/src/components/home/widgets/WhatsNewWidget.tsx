import { compare } from "compare-versions";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { useAlbyInfo } from "src/hooks/useAlbyInfo";
import { useInfo } from "src/hooks/useInfo";
import { GRYNVAULT_UPDATE_URL } from "src/lib/grynvault";

export function WhatsNewWidget() {
  const { data: info } = useInfo();
  const { data: albyInfo } = useAlbyInfo();

  if (
    !info ||
    !albyInfo ||
    !albyInfo.hub.latestReleaseNotes ||
    info.hideUpdateBanner
  ) {
    return null;
  }

  const upToDate =
    info.version &&
    info.version.startsWith("v") &&
    compare(info.version.substring(1), albyInfo.hub.latestVersion, ">=");

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {upToDate
            ? "Upstream Alby Hub release notes"
            : "Upstream Alby Hub update available"}
        </CardTitle>
        <CardDescription>{albyInfo.hub.latestReleaseNotes}</CardDescription>
      </CardHeader>
      {!upToDate && (
        <CardContent className="text-right">
          <ExternalLinkButton to={GRYNVAULT_UPDATE_URL} size="sm">
            Review Compatibility
          </ExternalLinkButton>
        </CardContent>
      )}
    </Card>
  );
}

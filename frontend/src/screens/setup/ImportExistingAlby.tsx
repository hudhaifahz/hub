import { ShieldCheckIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router";
import Container from "src/components/Container";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { handleRequestError } from "src/utils/handleRequestError";
import { isHttpMode } from "src/utils/isHttpMode";
import { request } from "src/utils/request";

type ImportResult = {
  cancelled: boolean;
  restartRequired: boolean;
  backupLocation?: string;
};

export function ImportExistingAlby() {
  const [loading, setLoading] = React.useState(false);
  const [result, setResult] = React.useState<ImportResult>();
  const desktop = !isHttpMode();

  const importExistingAlby = async () => {
    try {
      setLoading(true);
      const response = await request<ImportResult>(
        "/api/import-existing-alby",
        { method: "POST" }
      );
      if (response && !response.cancelled) {
        setResult(response);
      }
    } catch (error) {
      handleRequestError("Could not import Alby Hub", error);
    } finally {
      setLoading(false);
    }
  };

  if (result?.restartRequired) {
    return (
      <Container>
        <title>Import verified · Grynvault Rebalance</title>
        <div className="grid gap-5 text-center">
          <ShieldCheckIcon className="mx-auto size-12 text-primary" />
          <div className="grid gap-2">
            <h1 className="font-headline text-2xl font-semibold">
              Your wallet copy is verified
            </h1>
            <p className="text-muted-foreground">
              Grynvault Rebalance will close to finish the import. Reopen it,
              unlock with your existing password, and verify your node and
              channels. Your original Alby Hub data remains untouched.
            </p>
          </div>
          {result.backupLocation && (
            <div className="rounded-lg border bg-muted/30 p-3 text-left text-sm">
              <div className="font-medium">Recovery copy</div>
              <div className="mt-1 break-all font-mono text-xs text-muted-foreground">
                {result.backupLocation}
              </div>
            </div>
          )}
        </div>
      </Container>
    );
  }

  return (
    <Container>
      <title>Import Alby Hub · Grynvault Rebalance</title>
      <div className="grid gap-5 text-center">
        <div className="grid gap-2">
          <h1 className="font-headline text-2xl font-semibold">
            Bring your existing Alby Hub LDK wallet
          </h1>
          <p className="text-muted-foreground">
            Your channels are not missing. Grynvault uses a separate data folder
            so it cannot change your Alby Hub without permission. This direct
            importer is for Alby Hub&apos;s embedded LDK wallet.
          </p>
        </div>

        <div className="grid gap-3 rounded-lg border p-4 text-left text-sm">
          <div className="flex gap-3">
            <ShieldCheckIcon className="mt-0.5 size-5 shrink-0 text-primary" />
            <div className="grid gap-2">
              <p className="font-medium">Safe, non-destructive import</p>
              <ul className="list-disc space-y-1 pl-4 text-muted-foreground">
                <li>Alby Hub must be closed before import.</li>
                <li>A hash-verified recovery copy is created automatically.</li>
                <li>The original Alby Hub data is never deleted or moved.</li>
                <li>Grynvault restarts before using the copied wallet data.</li>
              </ul>
            </div>
          </div>
        </div>

        {desktop ? (
          <LoadingButton
            className="w-full"
            loading={loading}
            onClick={importExistingAlby}
          >
            Choose Alby Hub data
          </LoadingButton>
        ) : (
          <p className="text-sm text-muted-foreground">
            Direct import is available in the Grynvault desktop application.
          </p>
        )}

        <Button asChild variant="secondary" className="w-full">
          <Link to="/setup/node-restore">Use an encrypted migration file</Link>
        </Button>
        <Button asChild variant="ghost" className="w-full">
          <Link to="/welcome">Set up a new wallet instead</Link>
        </Button>
      </div>
    </Container>
  );
}

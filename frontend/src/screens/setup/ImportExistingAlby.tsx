import {
  CheckCircle2Icon,
  CircleAlertIcon,
  FolderOpenIcon,
  ShieldCheckIcon,
  WifiOffIcon,
} from "lucide-react";
import React from "react";
import { Link } from "react-router";
import Container from "src/components/Container";
import PasswordInput from "src/components/password/PasswordInput";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { handleRequestError } from "src/utils/handleRequestError";
import { isHttpMode } from "src/utils/isHttpMode";
import { onMigrationProgress } from "src/utils/migrationProgress";
import { request } from "src/utils/request";

type Counts = {
  apps: number;
  permissions: number;
  transactions: number;
  userConfigs: number;
  migrations: number;
  rebalanceRecords: number;
};

type SourceSummary = {
  sourcePath: string;
  backendType?: string;
  walletFingerprint?: string;
  walletFingerprintHint?: string;
  nwcIntegrity?: string;
  ldkIntegrity?: string;
  counts?: Counts;
  durableFileCount?: number;
  latestDurableChange?: string;
  defaultLocation: boolean;
  ready: boolean;
  blockingReason?: string;
  conflictingPath?: string;
};

type ChannelSummary = {
  channelId: string;
  userChannelId: string;
  counterpartyNodeId: string;
  capacitySats: number;
  outboundSats: number;
  inboundSats: number;
  ready: boolean;
  usable: boolean;
};

type VerificationSummary = {
  nodeId: string;
  network: string;
  channels: ChannelSummary[];
  channelCount: number;
  storedPaymentCount: number;
  totalOnchainSats: number;
  spendableOnchainSats: number;
  totalLightningSats: number;
  pendingClosureSats: number;
  apps: number;
  permissions: number;
  transactions: number;
  rebalanceRecords: number;
  staticChannelBackupFiles: number;
  latestStaticChannelBackup?: string;
  latestStaticBackupSha256?: string;
  releaseVersion?: string;
  sourceCommit?: string;
  ldkLibrarySha256?: string;
  sourceBytesStillMatch: boolean;
  destinationBytesStillMatch: boolean;
};

type MigrationStatus = {
  active: boolean;
  phase?: string;
  sourcePath?: string;
  targetPath?: string;
  backupLocation?: string;
  fingerprintHint?: string;
  activationConfirmation?: string;
  blockingReason?: string;
  verification?: VerificationSummary;
  lightningOffline: boolean;
};

type SelectionResponse = {
  cancelled: boolean;
  summary?: SourceSummary;
};

type StageResult = {
  cancelled: boolean;
  restartRequired: boolean;
  backupLocation?: string;
  summary: SourceSummary;
};

type MigrationFixture = {
  source?: SourceSummary;
  status?: MigrationStatus;
  verified?: VerificationSummary;
  activated?: boolean;
};

const syntheticNodeId =
  "027cf7c81dc777e46572ff964f17650a0f6f783319fcc140fb9a69425e2fbdc93c";

const syntheticVerification: VerificationSummary = {
  nodeId: syntheticNodeId,
  network: "regtest",
  channelCount: 2,
  storedPaymentCount: 2,
  totalOnchainSats: 1_986_412,
  spendableOnchainSats: 1_964_870,
  totalLightningSats: 793_000,
  pendingClosureSats: 0,
  apps: 3,
  permissions: 8,
  transactions: 2,
  rebalanceRecords: 0,
  staticChannelBackupFiles: 2,
  latestStaticChannelBackup: "channel-backup-20261016T235900Z.json",
  latestStaticBackupSha256:
    "41cd8d451fd2f142977daf69bd014c49b1e69fdb32dfa6e41d88d8a626cc30f1",
  releaseVersion: "grynvault-rebalance-v0.1.0-rc.next",
  sourceCommit: "0000000000000000000000000000000000000000",
  ldkLibrarySha256:
    "c4a7e68619a9a6b3d7a73d72b25f77205871df5a5c921f61b7c660b9c4318f43",
  sourceBytesStillMatch: true,
  destinationBytesStillMatch: true,
  channels: [
    {
      channelId:
        "804dc9fa56c8254dd56bc4c5943ea74ea2658af608d74239a4e27dbce7254fd1",
      userChannelId: "145263719338420795779428504468414423215",
      counterpartyNodeId:
        "03a34b99f22c790c8fbce8b7e6c52b65755a294acbb03a15b5e8a4f4f82bc4ea56",
      capacitySats: 500_000,
      outboundSats: 389_000,
      inboundSats: 107_000,
      ready: true,
      usable: false,
    },
    {
      channelId:
        "3d22c780f797f5feca81d72c430e04d596b08f91f0edb27f21eb1771f21bc20f",
      userChannelId: "238761907298847108441007374629302849134",
      counterpartyNodeId:
        "028d7500dd4c12685d1f568b4c2b5048e02c62f1f0dc9dd76b4e820b5732230ff1",
      capacitySats: 500_000,
      outboundSats: 404_000,
      inboundSats: 92_000,
      ready: true,
      usable: false,
    },
  ],
};

function syntheticSource(sourcePath: string, defaultLocation: boolean) {
  return {
    sourcePath,
    backendType: "LDK",
    walletFingerprintHint: "a73c421f29b8",
    nwcIntegrity: "ok",
    ldkIntegrity: "ok",
    counts: {
      apps: 3,
      permissions: 8,
      transactions: 2,
      userConfigs: 19,
      migrations: 42,
      rebalanceRecords: 0,
    },
    durableFileCount: 9,
    latestDurableChange: "2026-10-16T23:59:00Z",
    defaultLocation,
    ready: true,
  } satisfies SourceSummary;
}

function getSyntheticFixture(): MigrationFixture | undefined {
  if (!import.meta.env.DEV) {
    return undefined;
  }
  const name = new URLSearchParams(window.location.search).get(
    "migration-fixture"
  );
  const standardPath =
    "/Users/synthetic/Library/Containers/com.getalby.Alby-Hub/Data/Library/Application Support/albyhub";
  const customPath = "/Volumes/Wallets/Alby 雷/primary node";
  const commonStatus = {
    active: true,
    sourcePath: standardPath,
    targetPath:
      "/Users/synthetic/Library/Containers/com.grynvault.rebalance/Data/Library/Application Support/grynvault-rebalance",
    backupLocation:
      "/Users/synthetic/Library/Containers/com.grynvault.rebalance/Data/Library/Application Support/grynvault-rebalance-migration-backups/20261016T235900Z-a1b2c3d4",
    fingerprintHint: "a73c421f29b8",
    lightningOffline: true,
  } satisfies MigrationStatus;
  switch (name) {
    case "default":
      return { source: syntheticSource(standardPath, true) };
    case "custom":
      return { source: syntheticSource(customPath, false) };
    case "stale":
      return {
        status: {
          ...commonStatus,
          phase: "blocked",
          blockingReason:
            "Another matching wallet copy was found at /Users/synthetic/Library/Application Support/albyhub. Grynvault cannot prove which copy is current, so no wallet data was changed.",
        },
      };
    case "verified":
      return {
        status: {
          ...commonStatus,
          phase: "verified",
          activationConfirmation: "CONNECT REGTEST a73c421f29b8",
          verification: syntheticVerification,
        },
        verified: syntheticVerification,
      };
    case "activation-failed":
      return {
        status: {
          ...commonStatus,
          phase: "activation_failed",
          blockingReason:
            "NodeError: ConnectionFailed: Network connection closed.",
          verification: syntheticVerification,
        },
        verified: syntheticVerification,
      };
    case "complete":
      return {
        status: {
          ...commonStatus,
          active: false,
          phase: "activated",
          lightningOffline: false,
          verification: syntheticVerification,
        },
        verified: syntheticVerification,
        activated: true,
      };
    default:
      return undefined;
  }
}

function formatSats(value: number) {
  return `${new Intl.NumberFormat().format(value)} sats`;
}

function EvidenceRow({
  label,
  value,
}: {
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="grid gap-1 sm:grid-cols-[minmax(0,12rem)_minmax(0,1fr)] sm:gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="min-w-0 break-words font-medium">{value}</dd>
    </div>
  );
}

export function ImportExistingAlby() {
  const fixture = React.useMemo(() => getSyntheticFixture(), []);
  const desktop = fixture !== undefined || !isHttpMode();
  const manualSourceSelectionStarted = React.useRef(false);
  const [loading, setLoading] = React.useState(false);
  const [status, setStatus] = React.useState<MigrationStatus | undefined>(
    fixture?.status
  );
  const [source, setSource] = React.useState<SourceSummary | undefined>(
    fixture?.source
  );
  const [stageResult, setStageResult] = React.useState<StageResult>();
  const [verified, setVerified] = React.useState<
    VerificationSummary | undefined
  >(fixture?.verified);
  const [activated, setActivated] = React.useState(fixture?.activated ?? false);
  const [unlockPassword, setUnlockPassword] = React.useState("");
  const [nodeConfirmation, setNodeConfirmation] = React.useState("");
  const [activationConfirmation, setActivationConfirmation] =
    React.useState("");
  const [activationAcknowledged, setActivationAcknowledged] =
    React.useState(false);
  const [activationReviewOpen, setActivationReviewOpen] = React.useState(false);
  const [sourceConfirmed, setSourceConfirmed] = React.useState(false);
  const [migrationProgress, setMigrationProgress] = React.useState("");

  const loadInitialState = React.useCallback(async () => {
    if (!desktop || fixture) {
      return;
    }
    try {
      const migrationStatus = await request<MigrationStatus>(
        "/api/import-existing-alby/status"
      );
      if (migrationStatus?.active) {
        setStatus(migrationStatus);
        setVerified(migrationStatus.verification);
        return;
      }
      const discovered = await request<SourceSummary>(
        "/api/import-existing-alby/discover",
        { method: "POST" }
      );
      if (discovered && !manualSourceSelectionStarted.current) {
        setSource(discovered);
      }
    } catch (error) {
      handleRequestError("Could not inspect Alby Hub", error);
    }
  }, [desktop, fixture]);

  React.useEffect(() => {
    void loadInitialState();
  }, [loadInitialState]);

  React.useEffect(
    () =>
      onMigrationProgress((progress) => setMigrationProgress(progress.detail)),
    []
  );

  async function chooseSource() {
    // A slow automatic discovery response must never replace a folder the
    // owner has started choosing explicitly.
    manualSourceSelectionStarted.current = true;
    try {
      setLoading(true);
      setMigrationProgress("");
      const response = await request<SelectionResponse>(
        "/api/import-existing-alby/select",
        { method: "POST" }
      );
      if (response && !response.cancelled && response.summary) {
        setSource(response.summary);
        setSourceConfirmed(false);
      }
    } catch (error) {
      handleRequestError("Could not inspect the selected folder", error);
    } finally {
      setLoading(false);
    }
  }

  async function stageImport() {
    if (!source?.ready || !sourceConfirmed) {
      return;
    }
    try {
      setLoading(true);
      setMigrationProgress("Starting the read-only wallet safety checks…");
      const response = await request<StageResult>(
        "/api/import-existing-alby/stage",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ sourcePath: source.sourcePath }),
        }
      );
      if (response) {
        setStageResult(response);
      }
    } catch (error) {
      handleRequestError("Could not stage the wallet migration", error);
    } finally {
      setLoading(false);
    }
  }

  async function verifyImport(event: React.FormEvent) {
    event.preventDefault();
    try {
      setLoading(true);
      const response = await request<VerificationSummary>(
        "/api/import-existing-alby/verify",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ unlockPassword }),
        }
      );
      if (response) {
        setVerified(response);
        let migrationStatus: MigrationStatus | undefined;
        try {
          migrationStatus = await request<MigrationStatus>(
            "/api/import-existing-alby/status"
          );
        } catch {
          // Verification is already durable. A reload recovers the
          // server-derived activation phrase without repeating the check.
        }
        setStatus((current) =>
          migrationStatus?.active
            ? migrationStatus
            : current
              ? { ...current, phase: "verified", verification: response }
              : current
        );
        setUnlockPassword("");
        setNodeConfirmation("");
        setActivationConfirmation("");
        setActivationAcknowledged(false);
        setActivationReviewOpen(false);
      }
    } catch (error) {
      setUnlockPassword("");
      handleRequestError("Migration verification failed", error);
    } finally {
      setLoading(false);
    }
  }

  async function activateImport(event: React.FormEvent) {
    event.preventDefault();
    if (
      !window.confirm(
        "Start Lightning networking for this copied wallet now? This leaves offline verification and connects the verified node to its Bitcoin network."
      )
    ) {
      return;
    }
    try {
      setLoading(true);
      await request("/api/import-existing-alby/activate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          unlockPassword,
          nodeIdConfirmation: nodeConfirmation.trim(),
          activationConfirmation: activationConfirmation.trim(),
        }),
      });
      setUnlockPassword("");
      setNodeConfirmation("");
      setActivationConfirmation("");
      setActivationAcknowledged(false);
      setActivationReviewOpen(false);
      setActivated(true);
    } catch (error) {
      setUnlockPassword("");
      setNodeConfirmation("");
      setActivationConfirmation("");
      setActivationAcknowledged(false);
      setActivationReviewOpen(false);
      try {
        const migrationStatus = await request<MigrationStatus>(
          "/api/import-existing-alby/status"
        );
        if (migrationStatus?.active) {
          setStatus(migrationStatus);
          setVerified(migrationStatus.verification);
        }
      } catch {
        // Keep the original activation error as the actionable message.
      }
      handleRequestError("Could not activate the imported wallet", error);
    } finally {
      setLoading(false);
    }
  }

  async function openEvidenceFolder() {
    try {
      setLoading(true);
      await request("/api/import-existing-alby/open-evidence", {
        method: "POST",
      });
    } catch (error) {
      handleRequestError("Could not open the recovery folder", error);
    } finally {
      setLoading(false);
    }
  }

  async function returnToSource() {
    try {
      setLoading(true);
      await request("/api/import-existing-alby/return-to-source", {
        method: "POST",
      });
    } catch (error) {
      handleRequestError("Could not quit migration mode", error);
      setLoading(false);
    }
  }

  async function abandonImport() {
    if (
      !window.confirm(
        "Archive this failed import and restart Grynvault? The source wallet, staged copy, and recovery evidence will be preserved locally."
      )
    ) {
      return;
    }
    try {
      setLoading(true);
      await request("/api/import-existing-alby/abandon", { method: "POST" });
    } catch (error) {
      handleRequestError("Could not archive the failed import", error);
      setLoading(false);
    }
  }

  if (!desktop) {
    return (
      <Container>
        <title>Import Alby Hub · Grynvault Rebalance</title>
        <div className="mx-auto grid max-w-2xl gap-5 text-center">
          <h1 className="font-headline text-2xl font-semibold">
            Desktop migration required
          </h1>
          <p className="text-muted-foreground">
            Direct migration is available in the Grynvault desktop application.
            An encrypted Alby migration file can still be restored here.
          </p>
          <Button asChild>
            <Link to="/setup/node-restore">
              Use an encrypted migration file
            </Link>
          </Button>
        </div>
      </Container>
    );
  }

  if (activated && verified) {
    return (
      <Container>
        <title>Migration complete · Grynvault Rebalance</title>
        <div className="mx-auto grid max-w-3xl gap-5">
          {fixture && (
            <p className="text-center text-xs font-medium text-muted-foreground">
              Synthetic documentation example — no wallet is connected
            </p>
          )}
          <div className="grid gap-3 text-center">
            <CheckCircle2Icon className="mx-auto size-12 text-primary" />
            <h1 className="font-headline text-2xl font-semibold">
              Migration complete
            </h1>
            <p className="text-muted-foreground">
              Grynvault activated the verified wallet. Lightning is connected
              with the same node identity and stored channels.
            </p>
          </div>
          <dl className="grid gap-3 rounded-lg border p-4 text-sm">
            <EvidenceRow
              label="Node ID"
              value={
                <span className="break-all font-mono text-xs">
                  {verified.nodeId}
                </span>
              }
            />
            <EvidenceRow label="Network" value={verified.network} />
            <EvidenceRow
              label="Stored channels"
              value={verified.channelCount}
            />
            <EvidenceRow
              label="Stored payments"
              value={verified.storedPaymentCount}
            />
            <EvidenceRow
              label="Recovery copy"
              value={
                <span className="break-all font-mono text-xs">
                  {status?.backupLocation}
                </span>
              }
            />
          </dl>
          <div className="rounded-lg border border-primary/30 bg-primary/5 p-4 text-sm">
            Keep the original Alby Hub stopped. Do not run both applications
            against copies of this node.
          </div>
          <Button type="button" onClick={() => window.location.assign("/")}>
            Open Grynvault Rebalance
          </Button>
        </div>
      </Container>
    );
  }

  if (stageResult?.restartRequired) {
    return (
      <Container>
        <title>Wallet copy verified · Grynvault Rebalance</title>
        <div className="mx-auto grid max-w-2xl gap-5 text-center">
          <ShieldCheckIcon className="mx-auto size-12 text-primary" />
          <div className="grid gap-2">
            <h1 className="font-headline text-2xl font-semibold">
              The recovery copy is verified
            </h1>
            <p className="text-muted-foreground">
              Grynvault Rebalance will close and reopen automatically to
              continue the guided verification. Lightning will remain offline
              until you unlock, review the same node and channels, and
              explicitly activate it.
            </p>
          </div>
          <dl className="grid gap-3 rounded-lg border p-4 text-left text-sm">
            <EvidenceRow
              label="Wallet fingerprint"
              value={stageResult.summary.walletFingerprintHint}
            />
            <EvidenceRow
              label="Recovery copy"
              value={
                <span className="break-all font-mono text-xs">
                  {stageResult.backupLocation}
                </span>
              }
            />
          </dl>
        </div>
      </Container>
    );
  }

  if (status?.active) {
    const blocked = status.phase === "blocked";
    const activationFailed = status.phase === "activation_failed";
    const requiresOfflineVerification =
      !verified || status.phase === "verification_needed" || activationFailed;
    const activationConfirmationText = status.activationConfirmation || "";
    const activationNetworkLabel =
      verified?.network === "bitcoin"
        ? "Bitcoin mainnet"
        : verified?.network || "the selected Bitcoin network";
    return (
      <Container>
        <title>Verify migration · Grynvault Rebalance</title>
        <div className="mx-auto grid max-w-3xl gap-5">
          {fixture && (
            <p className="text-center text-xs font-medium text-muted-foreground">
              Synthetic documentation example — no wallet is connected
            </p>
          )}
          <div className="grid gap-3 text-center">
            {blocked ? (
              <CircleAlertIcon className="mx-auto size-12 text-destructive" />
            ) : (
              <WifiOffIcon className="mx-auto size-12 text-primary" />
            )}
            <h1 className="font-headline text-2xl font-semibold">
              {blocked
                ? "Migration stopped safely"
                : "Migration verification — Lightning is offline"}
            </h1>
            <p className="text-muted-foreground">
              {blocked
                ? "Grynvault did not activate the staged wallet. Your source wallet and recovery evidence were preserved."
                : "No peer, invoice, payment, rebalance, swap, or channel service is running while you review this copy."}
            </p>
          </div>

          <dl className="grid gap-3 rounded-lg border p-4 text-sm">
            <EvidenceRow
              label="Source folder"
              value={
                <span className="break-all font-mono text-xs">
                  {status.sourcePath}
                </span>
              }
            />
            <EvidenceRow
              label="Wallet fingerprint"
              value={status.fingerprintHint}
            />
            <EvidenceRow
              label="Recovery copy"
              value={
                <span className="break-all font-mono text-xs">
                  {status.backupLocation}
                </span>
              }
            />
            <EvidenceRow label="Lightning" value="Offline and locked" />
          </dl>

          {blocked && (
            <div className="grid gap-3 rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm">
              <div>
                <p className="font-medium">Why Grynvault stopped</p>
                <p className="mt-1 break-words text-muted-foreground">
                  {status.blockingReason}
                </p>
              </div>
              <LoadingButton
                type="button"
                variant="secondary"
                loading={loading}
                onClick={abandonImport}
              >
                Archive failed attempt and restart
              </LoadingButton>
              <p className="text-xs text-muted-foreground">
                This preserves the source, failed staged copy, and evidence in
                the recovery folder before allowing a fresh guided import.
              </p>
            </div>
          )}

          {activationFailed && (
            <div className="grid gap-2 rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm">
              <p className="font-medium">Lightning did not connect</p>
              <p className="break-words text-muted-foreground">
                {status.blockingReason ||
                  "The activation attempt stopped before Lightning connected."}
              </p>
              <p className="text-muted-foreground">
                The copied wallet remains offline and its successful
                verification evidence is preserved. Restore network access, then
                verify again below. A failed activation cannot be retried from
                the old confirmation.
              </p>
            </div>
          )}

          <div className="grid gap-2 sm:grid-cols-2">
            <LoadingButton
              type="button"
              variant="secondary"
              loading={loading}
              onClick={openEvidenceFolder}
            >
              Open recovery evidence
            </LoadingButton>
            <LoadingButton
              type="button"
              variant="ghost"
              loading={loading}
              onClick={returnToSource}
            >
              Quit and return to source wallet
            </LoadingButton>
          </div>

          <p className="text-center text-xs text-muted-foreground">
            Returning quits Grynvault without starting Lightning. Reopen the
            original Alby Hub yourself only after Grynvault has fully closed.
          </p>

          {!blocked && requiresOfflineVerification && (
            <form className="grid gap-4" onSubmit={verifyImport}>
              <div className="grid gap-1.5">
                <Label htmlFor="migration-password">
                  Existing unlock password
                </Label>
                <PasswordInput
                  id="migration-password"
                  value={unlockPassword}
                  onChange={setUnlockPassword}
                  autoFocus
                />
                <p className="text-xs text-muted-foreground">
                  Your password stays in this local application and is not saved
                  in the migration evidence.
                </p>
              </div>
              <LoadingButton
                type="submit"
                loading={loading}
                disabled={!unlockPassword}
              >
                {activationFailed
                  ? "Re-verify wallet offline"
                  : "Verify wallet offline"}
              </LoadingButton>
            </form>
          )}

          {!blocked && verified && (
            <>
              <div className="grid gap-4 rounded-lg border p-4">
                <div className="flex items-start gap-3">
                  <CheckCircle2Icon className="mt-0.5 size-5 shrink-0 text-primary" />
                  <div>
                    <h2 className="font-semibold">Continuity checks passed</h2>
                    <p className="text-sm text-muted-foreground">
                      The copied wallet was inspected without starting LDK
                      networking.
                    </p>
                  </div>
                </div>
                <dl className="grid gap-3 text-sm">
                  <EvidenceRow
                    label="Node ID"
                    value={
                      <span className="break-all font-mono text-xs">
                        {verified.nodeId}
                      </span>
                    }
                  />
                  <EvidenceRow label="Network" value={verified.network} />
                  <EvidenceRow label="Channels" value={verified.channelCount} />
                  <EvidenceRow
                    label="Lightning balance"
                    value={formatSats(verified.totalLightningSats)}
                  />
                  <EvidenceRow
                    label="Spendable on-chain"
                    value={formatSats(verified.spendableOnchainSats)}
                  />
                  <EvidenceRow
                    label="Pending channel closures"
                    value={formatSats(verified.pendingClosureSats)}
                  />
                  <EvidenceRow label="Apps" value={verified.apps} />
                  <EvidenceRow
                    label="Permissions"
                    value={verified.permissions}
                  />
                  <EvidenceRow
                    label="Transactions"
                    value={verified.transactions}
                  />
                  <EvidenceRow
                    label="Static channel backups"
                    value={
                      verified.staticChannelBackupFiles > 0
                        ? `${verified.staticChannelBackupFiles} file(s); latest ${verified.latestStaticChannelBackup}`
                        : "No saved static backup files"
                    }
                  />
                  {verified.latestStaticBackupSha256 && (
                    <EvidenceRow
                      label="Latest backup SHA-256"
                      value={
                        <span className="break-all font-mono text-xs">
                          {verified.latestStaticBackupSha256}
                        </span>
                      }
                    />
                  )}
                  {verified.releaseVersion && (
                    <EvidenceRow
                      label="Grynvault release"
                      value={verified.releaseVersion}
                    />
                  )}
                  {verified.sourceCommit && (
                    <EvidenceRow
                      label="Embedded source commit"
                      value={
                        <span className="break-all font-mono text-xs">
                          {verified.sourceCommit}
                        </span>
                      }
                    />
                  )}
                  {verified.ldkLibrarySha256 && (
                    <EvidenceRow
                      label="Embedded LDK SHA-256"
                      value={
                        <span className="break-all font-mono text-xs">
                          {verified.ldkLibrarySha256}
                        </span>
                      }
                    />
                  )}
                </dl>
              </div>

              {verified.channels.length > 0 && (
                <div className="grid gap-3">
                  <h2 className="font-semibold">Stored channels</h2>
                  {verified.channels.map((channel) => (
                    <div
                      key={channel.userChannelId}
                      className="grid gap-2 rounded-lg border p-4 text-sm"
                    >
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <span className="font-medium">
                          {formatSats(channel.capacitySats)} capacity
                        </span>
                        <span className="text-muted-foreground">
                          {channel.ready ? "Ready" : "Stored"}
                        </span>
                      </div>
                      <div className="break-all font-mono text-xs text-muted-foreground">
                        Peer: {channel.counterpartyNodeId}
                      </div>
                      <div className="break-all font-mono text-xs text-muted-foreground">
                        Channel: {channel.userChannelId}
                      </div>
                      <div className="grid gap-1 text-xs text-muted-foreground sm:grid-cols-2">
                        <span>
                          Outbound: {formatSats(channel.outboundSats)}
                        </span>
                        <span>Inbound: {formatSats(channel.inboundSats)}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {status.phase === "verified" && !activationReviewOpen && (
                <div className="grid gap-3 rounded-lg border p-4">
                  <div className="flex items-start gap-3">
                    <ShieldCheckIcon className="mt-0.5 size-5 shrink-0 text-primary" />
                    <div>
                      <h2 className="font-semibold">
                        Offline verification is complete
                      </h2>
                      <p className="text-sm text-muted-foreground">
                        The copied wallet remains offline. Activation is a
                        separate step and is not required to preserve this
                        verified result.
                      </p>
                    </div>
                  </div>
                  <Button
                    type="button"
                    variant="secondary"
                    disabled={!activationConfirmationText}
                    onClick={() => {
                      setUnlockPassword("");
                      setNodeConfirmation("");
                      setActivationConfirmation("");
                      setActivationAcknowledged(false);
                      setActivationReviewOpen(true);
                    }}
                  >
                    Review network activation
                  </Button>
                  {!activationConfirmationText && (
                    <p className="text-xs text-muted-foreground">
                      Reload Grynvault to recover the server-validated
                      activation phrase.
                    </p>
                  )}
                </div>
              )}

              {status.phase === "verified" && activationReviewOpen && (
                <form
                  className="grid gap-4 rounded-lg border border-destructive/40 bg-destructive/5 p-4"
                  onSubmit={activateImport}
                >
                  <div className="grid gap-1">
                    <h2 className="font-semibold text-destructive">
                      Connect this copied wallet to {activationNetworkLabel}
                    </h2>
                    <p className="text-sm text-muted-foreground">
                      This starts Lightning networking for the verified node.
                      Keep the source Alby Hub fully stopped. Verification
                      expires after five minutes, and any failed attempt
                      requires another offline verification.
                    </p>
                  </div>
                  <div className="grid gap-1.5">
                    <Label htmlFor="activation-password">
                      Re-enter the existing unlock password
                    </Label>
                    <PasswordInput
                      id="activation-password"
                      value={unlockPassword}
                      onChange={setUnlockPassword}
                      autoComplete="off"
                    />
                  </div>
                  <div className="grid gap-1.5">
                    <Label htmlFor="node-confirmation">
                      Confirm the complete node ID
                    </Label>
                    <Input
                      id="node-confirmation"
                      value={nodeConfirmation}
                      onChange={(event) =>
                        setNodeConfirmation(event.target.value)
                      }
                      className="font-mono text-xs"
                      autoComplete="off"
                      spellCheck={false}
                    />
                    <p className="break-all text-xs text-muted-foreground">
                      Type or locally copy: {verified.nodeId}
                    </p>
                  </div>
                  <div className="grid gap-1.5">
                    <Label htmlFor="activation-confirmation">
                      Type the network activation phrase
                    </Label>
                    <Input
                      id="activation-confirmation"
                      value={activationConfirmation}
                      onChange={(event) =>
                        setActivationConfirmation(event.target.value)
                      }
                      className="font-mono text-xs"
                      autoComplete="off"
                      spellCheck={false}
                    />
                    <p className="break-all text-xs text-muted-foreground">
                      Type exactly: {activationConfirmationText}
                    </p>
                  </div>
                  <div className="flex items-start gap-3">
                    <Checkbox
                      id="activation-acknowledgement"
                      checked={activationAcknowledged}
                      onCheckedChange={(checked) =>
                        setActivationAcknowledged(checked === true)
                      }
                    />
                    <Label
                      htmlFor="activation-acknowledgement"
                      className="text-sm font-normal leading-5"
                    >
                      I understand this leaves offline verification and starts
                      the copied production wallet on {activationNetworkLabel}.
                    </Label>
                  </div>
                  <div className="grid gap-2 sm:grid-cols-2">
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        setUnlockPassword("");
                        setNodeConfirmation("");
                        setActivationConfirmation("");
                        setActivationAcknowledged(false);
                        setActivationReviewOpen(false);
                      }}
                    >
                      Keep wallet offline
                    </Button>
                    <LoadingButton
                      type="submit"
                      variant="destructive"
                      loading={loading}
                      disabled={
                        !activationAcknowledged ||
                        !unlockPassword ||
                        nodeConfirmation.trim() !== verified.nodeId ||
                        activationConfirmation.trim() !==
                          activationConfirmationText
                      }
                    >
                      Connect verified wallet now
                    </LoadingButton>
                  </div>
                </form>
              )}
            </>
          )}
        </div>
      </Container>
    );
  }

  return (
    <Container>
      <title>Import Alby Hub · Grynvault Rebalance</title>
      <div className="mx-auto grid max-w-3xl gap-5">
        {fixture && (
          <p className="text-center text-xs font-medium text-muted-foreground">
            Synthetic documentation example — no wallet is connected
          </p>
        )}
        <div className="grid gap-2 text-center">
          <h1 className="font-headline text-2xl font-semibold">
            Bring your existing Alby Hub wallet
          </h1>
          <p className="text-muted-foreground">
            Grynvault checks the source, creates a verified recovery copy, and
            keeps Lightning offline until you confirm the same node and
            channels.
          </p>
        </div>

        <div className="grid gap-3 rounded-lg border p-4">
          <div className="flex items-start gap-3">
            <FolderOpenIcon className="mt-0.5 size-5 shrink-0 text-primary" />
            <div className="min-w-0 flex-1">
              <p className="font-medium">
                {source?.ready
                  ? source.defaultLocation
                    ? "Alby Hub wallet found"
                    : "Selected Alby Hub wallet"
                  : "Choose your Alby Hub data folder"}
              </p>
              <p className="mt-1 break-all font-mono text-xs text-muted-foreground">
                {source?.sourcePath || "No folder selected"}
              </p>
              {source?.blockingReason && (
                <p className="mt-2 text-sm text-muted-foreground">
                  {source.blockingReason}
                </p>
              )}
            </div>
          </div>
          <LoadingButton
            type="button"
            variant="secondary"
            loading={loading}
            onClick={chooseSource}
          >
            {source?.ready
              ? "Choose another Alby data folder"
              : "Choose folder"}
          </LoadingButton>
        </div>

        {source?.ready && source.counts && (
          <>
            <div className="grid gap-4 rounded-lg border p-4">
              <div className="flex items-start gap-3">
                <ShieldCheckIcon className="mt-0.5 size-5 shrink-0 text-primary" />
                <div>
                  <h2 className="font-semibold">Ready for the safety check</h2>
                  <p className="text-sm text-muted-foreground">
                    This is an embedded-LDK wallet with readable databases. A
                    final quiescence and competing-copy check runs before any
                    data is staged.
                  </p>
                </div>
              </div>
              <dl className="grid gap-3 text-sm">
                <EvidenceRow
                  label="Wallet fingerprint"
                  value={source.walletFingerprintHint}
                />
                <EvidenceRow label="Backend" value={source.backendType} />
                <EvidenceRow
                  label="Database integrity"
                  value={`${source.nwcIntegrity} / ${source.ldkIntegrity}`}
                />
                <EvidenceRow label="Apps" value={source.counts.apps} />
                <EvidenceRow
                  label="Permissions"
                  value={source.counts.permissions}
                />
                <EvidenceRow
                  label="Transactions"
                  value={source.counts.transactions}
                />
                <EvidenceRow
                  label="Durable files"
                  value={source.durableFileCount}
                />
                <EvidenceRow
                  label="Latest durable change"
                  value={source.latestDurableChange || "Unavailable"}
                />
              </dl>
            </div>

            <label className="flex items-start gap-3 rounded-lg border p-4 text-sm">
              <Checkbox
                checked={sourceConfirmed}
                onCheckedChange={(checked) =>
                  setSourceConfirmed(checked === true)
                }
              />
              <span>
                I have fully quit Alby Hub and confirm this is its current live
                data folder, not a backup or historical copy.
              </span>
            </label>

            <LoadingButton
              type="button"
              loading={loading}
              disabled={!sourceConfirmed}
              onClick={stageImport}
            >
              Create recovery copy and migrate
            </LoadingButton>
            {loading && migrationProgress && (
              <div
                aria-live="polite"
                className="rounded-lg border bg-muted/40 p-3 text-sm text-muted-foreground"
              >
                <p className="font-medium text-foreground">
                  Preparing your verified migration
                </p>
                <p className="mt-1">{migrationProgress}</p>
              </div>
            )}
          </>
        )}

        <div className="grid gap-2 sm:grid-cols-2">
          <Button asChild variant="secondary">
            <Link to="/setup/node-restore">
              Use an encrypted migration file
            </Link>
          </Button>
          <Button asChild variant="ghost">
            <Link to="/welcome">Set up a new wallet instead</Link>
          </Button>
        </div>
      </div>
    </Container>
  );
}

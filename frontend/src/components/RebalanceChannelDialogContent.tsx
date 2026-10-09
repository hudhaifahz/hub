import { AlertTriangleIcon, CheckCircle2Icon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import Loading from "src/components/Loading";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useChannels } from "src/hooks/useChannels";
import { Channel } from "src/types";
import { request } from "src/utils/request";
import {
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./ui/alert-dialog";

type Props = {
  incomingChannel: Channel;
  closeDialog(): void;
};

const localRebalanceRecoveryKey = "alby-local-rebalance-recovery-v1";

type LocalRouteHop = {
  nodePubkey: string;
  shortChannelId: string;
  feeMsat: number;
  cltvDelta: number;
};

type LocalRoutePath = {
  hops: LocalRouteHop[];
  amountMsat: number;
  feeMsat: number;
};

type RebalanceQuote = {
  quoteId: string;
  routeFingerprint: string;
  amountMsat: number;
  totalRoutingFeeMsat: number;
  maxRoutingFeeMsat: number;
  maxTotalDebitMsat: number;
  outgoingChannelId: string;
  outgoingNodePubkey: string;
  outgoingShortChannelId: string;
  incomingChannelId: string;
  incomingNodePubkey: string;
  incomingShortChannelId: string;
  outgoingSpendableSnapshotMsat: number;
  incomingReceivableSnapshotMsat: number;
  outgoingLocalSnapshotMsat: number;
  outgoingRemoteSnapshotMsat: number;
  outgoingLocalReserveMsat: number;
  outgoingRemoteReserveMsat: number;
  incomingLocalSnapshotMsat: number;
  incomingRemoteSnapshotMsat: number;
  incomingLocalReserveMsat: number;
  incomingRemoteReserveMsat: number;
  paths: LocalRoutePath[];
  quotedAt: string;
  expiresAt: string;
  executionEnabled: boolean;
  blockedReason: string;
  executionConfirmation: string;
};

type LocalRebalanceOperation = {
  quoteId: string;
  routeFingerprint: string;
  state: "executing" | "succeeded" | "failed";
  phase: "acquired" | "prepared" | "submitted" | "succeeded" | "failed";
  operationId: string;
  paymentHash: string;
  outboundPaymentId: string;
  amountMsat: number;
  quotedRoutingFeeMsat: number;
  maxRoutingFeeMsat: number;
  actualRoutingFeeMsat?: number;
  outgoingChannelId: string;
  outgoingNodePubkey: string;
  incomingChannelId: string;
  incomingNodePubkey: string;
  outgoingLocalSnapshotMsat: number;
  outgoingRemoteSnapshotMsat: number;
  outgoingLocalReserveMsat: number;
  outgoingRemoteReserveMsat: number;
  incomingLocalSnapshotMsat: number;
  incomingRemoteSnapshotMsat: number;
  incomingLocalReserveMsat: number;
  incomingRemoteReserveMsat: number;
  preparedAt?: string;
  submittedAt?: string;
  lightningTerminalAt?: string;
  reconciledAt?: string;
  terminalEvidenceHash?: string;
  failureReason?: string;
  requiresExactRetry: boolean;
  reconciliationPending: boolean;
};

function exactBtc(amountMsat: number) {
  return (amountMsat / 1000 / 100_000_000).toFixed(8);
}

function sats(amountMsat: number) {
  return Math.floor(amountMsat / 1000).toLocaleString();
}

function channelIdentity(channel: Channel) {
  return `${channel.remotePubkey.slice(0, 12)}… · ${channel.id}`;
}

export function RebalanceChannelDialogContent({
  incomingChannel,
  closeDialog,
}: Props) {
  const { data: channels } = useChannels();
  const outgoingChannels = (channels || []).filter(
    (channel) => channel.id !== incomingChannel.id
  );
  const [outgoingChannelId, setOutgoingChannelId] = React.useState("");
  const [amountSat, setAmountSat] = React.useState("");
  const [maxRoutingFeeSat, setMaxRoutingFeeSat] = React.useState("1000");
  const [isQuoting, setQuoting] = React.useState(false);
  const [isExecuting, setExecuting] = React.useState(false);
  const [quote, setQuote] = React.useState<RebalanceQuote>();
  const [confirmation, setConfirmation] = React.useState("");
  const [operation, setOperation] = React.useState<LocalRebalanceOperation>();
  const [resumeStoredOperation, setResumeStoredOperation] =
    React.useState(false);

  const refreshOperation = React.useCallback(async () => {
    if (!quote) {
      return;
    }
    const response = await request<LocalRebalanceOperation>(
      "/api/channels/rebalance/local-status",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          quoteId: quote.quoteId,
          routeFingerprint: quote.routeFingerprint,
        }),
      }
    );
    if (response) {
      setOperation(response);
    }
  }, [quote]);

  React.useEffect(() => {
    if (!operation?.reconciliationPending) {
      return;
    }
    const timer = window.setTimeout(() => {
      refreshOperation().catch((error) => console.error(error));
    }, 2000);
    return () => window.clearTimeout(timer);
  }, [operation, refreshOperation]);

  React.useEffect(() => {
    if (operation?.state === "succeeded" || operation?.state === "failed") {
      window.localStorage.removeItem(localRebalanceRecoveryKey);
    }
  }, [operation]);

  React.useEffect(() => {
    try {
      const stored = window.localStorage.getItem(localRebalanceRecoveryKey);
      if (!stored) {
        return;
      }
      const storedQuote = JSON.parse(stored) as RebalanceQuote;
      if (storedQuote.incomingChannelId !== incomingChannel.id) {
        return;
      }
      setQuote(storedQuote);
      setOutgoingChannelId(storedQuote.outgoingChannelId);
      setAmountSat(String(storedQuote.amountMsat / 1000));
      setMaxRoutingFeeSat(String(storedQuote.maxRoutingFeeMsat / 1000));
      setResumeStoredOperation(true);
    } catch (error) {
      console.error(error);
      window.localStorage.removeItem(localRebalanceRecoveryKey);
    }
  }, [incomingChannel.id]);

  React.useEffect(() => {
    if (!resumeStoredOperation || !quote) {
      return;
    }
    refreshOperation()
      .catch((error) => console.error(error))
      .finally(() => setResumeStoredOperation(false));
  }, [quote, refreshOperation, resumeStoredOperation]);

  if (!channels) {
    return <Loading />;
  }

  const outgoingChannel = outgoingChannels.find(
    (channel) => channel.id === outgoingChannelId
  );

  async function createQuote(event: React.FormEvent) {
    event.preventDefault();
    if (!outgoingChannel) {
      toast.error("Choose the exact channel the sats must leave through.");
      return;
    }
    setQuoting(true);
    setQuote(undefined);
    setConfirmation("");
    setOperation(undefined);
    window.localStorage.removeItem(localRebalanceRecoveryKey);
    try {
      const response = await request<RebalanceQuote>(
        "/api/channels/rebalance/local-quote",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            outgoingChannelId: outgoingChannel.id,
            outgoingNodePubkey: outgoingChannel.remotePubkey,
            incomingChannelId: incomingChannel.id,
            incomingNodePubkey: incomingChannel.remotePubkey,
            amountMsat: Number(amountSat) * 1000,
            maxRoutingFeeMsat: Number(maxRoutingFeeSat) * 1000,
          }),
        }
      );
      if (!response) {
        throw new Error("No quote response received");
      }
      setQuote(response);
    } catch (error) {
      console.error(error);
      toast.error(String(error));
    } finally {
      setQuoting(false);
    }
  }

  async function executeQuote() {
    if (!quote || confirmation !== quote.executionConfirmation) {
      return;
    }
    setExecuting(true);
    window.localStorage.setItem(
      localRebalanceRecoveryKey,
      JSON.stringify(quote)
    );
    try {
      const response = await request<LocalRebalanceOperation>(
        "/api/channels/rebalance/local-execute",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            quoteId: quote.quoteId,
            routeFingerprint: quote.routeFingerprint,
            confirmation,
          }),
        }
      );
      if (!response) {
        throw new Error("No execution response received");
      }
      setOperation(response);
      if (response.state === "succeeded") {
        toast.success("Exact-channel rebalance succeeded and was reconciled.");
      } else if (response.state === "failed") {
        toast.error(
          "Exact-channel rebalance failed; both payment legs were reconciled."
        );
      } else {
        toast.info(
          "Exact route submitted. Waiting for durable two-leg reconciliation."
        );
      }
    } catch (error) {
      console.error(error);
      toast.error(String(error));
      try {
        await refreshOperation();
      } catch (statusError) {
        console.error(statusError);
      }
    } finally {
      setExecuting(false);
    }
  }

  return (
    <AlertDialogContent className="max-h-[calc(100dvh-2rem)] min-w-0 max-w-2xl overflow-hidden p-0">
      <form
        className="flex max-h-[calc(100dvh-2rem)] min-h-0 min-w-0 flex-col"
        onSubmit={createQuote}
      >
        <AlertDialogHeader className="min-h-0 min-w-0 flex-1 overflow-x-hidden overflow-y-auto p-6 pb-4">
          <AlertDialogTitle>Exact-channel rebalance</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="min-w-0 space-y-4 text-left">
              <p>
                Choose both exact channels. This searches the local routing
                graph and saves an expiring review record; it creates no
                invoice, probe, HTLC, or payment.
              </p>

              <div className="space-y-2">
                <Label>Leave through this channel</Label>
                <Select
                  disabled={Boolean(operation)}
                  value={outgoingChannelId}
                  onValueChange={(value) => {
                    setOutgoingChannelId(value);
                    setQuote(undefined);
                  }}
                >
                  <SelectTrigger className="min-w-0 [&>span]:truncate">
                    <SelectValue placeholder="Select exact outgoing channel" />
                  </SelectTrigger>
                  <SelectContent>
                    {outgoingChannels.map((channel) => (
                      <SelectItem
                        key={`${channel.remotePubkey}:${channel.id}`}
                        value={channel.id}
                        disabled={channel.status !== "online"}
                      >
                        {channelIdentity(channel)} ·{" "}
                        {Math.floor(
                          channel.localSpendableBalanceSat
                        ).toLocaleString()}{" "}
                        sats spendable
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {outgoingChannel && (
                  <p className="break-all text-xs text-muted-foreground">
                    Peer: {outgoingChannel.remotePubkey}
                    <br />
                    Channel: {outgoingChannel.id}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label>Return through this channel</Label>
                <div className="min-w-0 rounded-md border p-3 text-sm">
                  <div className="truncate">
                    {channelIdentity(incomingChannel)}
                  </div>
                  <div className="mt-1 break-all text-xs text-muted-foreground">
                    Peer: {incomingChannel.remotePubkey}
                    <br />
                    Channel: {incomingChannel.id}
                    <br />
                    Receiving capacity:{" "}
                    {Math.floor(
                      incomingChannel.remoteBalanceSat
                    ).toLocaleString()}{" "}
                    sats
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div className="min-w-0">
                  <Label htmlFor="rebalance-amount">Principal (sats)</Label>
                  <Input
                    disabled={Boolean(operation)}
                    id="rebalance-amount"
                    type="number"
                    required
                    min={10000}
                    step={1}
                    value={amountSat}
                    onChange={(event) => {
                      setAmountSat(event.target.value);
                      setQuote(undefined);
                    }}
                  />
                </div>
                <div className="min-w-0">
                  <Label htmlFor="routing-fee-cap">Routing fee cap</Label>
                  <Input
                    disabled={Boolean(operation)}
                    id="routing-fee-cap"
                    type="number"
                    required
                    min={0}
                    step={1}
                    value={maxRoutingFeeSat}
                    onChange={(event) => {
                      setMaxRoutingFeeSat(event.target.value);
                      setQuote(undefined);
                    }}
                  />
                </div>
              </div>

              {quote && (
                <div className="space-y-3 rounded-md border p-4">
                  <h3 className="font-medium">Quote review</h3>
                  <dl className="grid min-w-0 grid-cols-1 gap-2 text-sm [&>dd]:min-w-0 [&>dd]:break-words [&>dd]:text-left sm:grid-cols-2 sm:[&>dd]:text-right">
                    <dt>Principal</dt>
                    <dd className="text-right">
                      <FormattedBitcoinAmount amountMsat={quote.amountMsat} />
                    </dd>
                    <dt>Estimated routing fee</dt>
                    <dd className="text-right">
                      <FormattedBitcoinAmount
                        amountMsat={quote.totalRoutingFeeMsat}
                      />
                    </dd>
                    <dt>Routing fee cap</dt>
                    <dd className="text-right">
                      <FormattedBitcoinAmount
                        amountMsat={quote.maxRoutingFeeMsat}
                      />
                    </dd>
                    <dt className="font-medium">Maximum source debit</dt>
                    <dd className="text-right font-medium">
                      <FormattedBitcoinAmount
                        amountMsat={quote.maxTotalDebitMsat}
                      />
                    </dd>
                    <dt>Principal in BTC</dt>
                    <dd className="text-right font-mono">
                      {exactBtc(quote.amountMsat)} BTC
                    </dd>
                    <dt>Outgoing spendable at quote</dt>
                    <dd className="text-right">
                      {sats(quote.outgoingSpendableSnapshotMsat)} sats
                    </dd>
                    <dt>Projected outgoing spendable</dt>
                    <dd className="text-right">
                      {sats(
                        quote.outgoingSpendableSnapshotMsat -
                          quote.maxTotalDebitMsat
                      )}{" "}
                      sats
                    </dd>
                  </dl>
                  <div className="space-y-2 rounded-md border p-3 text-xs">
                    <div className="font-medium">
                      Fingerprint-bound channel balance evidence
                    </div>
                    <dl className="grid min-w-0 grid-cols-1 gap-1 [&>dd]:min-w-0 [&>dd]:break-words [&>dd]:text-left sm:grid-cols-2 sm:[&>dd]:text-right">
                      <dt>Outgoing local → projected</dt>
                      <dd className="text-right">
                        {sats(quote.outgoingLocalSnapshotMsat)} →{" "}
                        {sats(
                          quote.outgoingLocalSnapshotMsat -
                            quote.maxTotalDebitMsat
                        )}{" "}
                        sats
                      </dd>
                      <dt>Outgoing remote → projected</dt>
                      <dd className="text-right">
                        {sats(quote.outgoingRemoteSnapshotMsat)} →{" "}
                        {sats(
                          quote.outgoingRemoteSnapshotMsat +
                            quote.maxTotalDebitMsat
                        )}{" "}
                        sats
                      </dd>
                      <dt>Outgoing reserves (local / remote)</dt>
                      <dd className="text-right">
                        {sats(quote.outgoingLocalReserveMsat)} /{" "}
                        {sats(quote.outgoingRemoteReserveMsat)} sats
                      </dd>
                      <dt>Incoming local → projected</dt>
                      <dd className="text-right">
                        {sats(quote.incomingLocalSnapshotMsat)} →{" "}
                        {sats(
                          quote.incomingLocalSnapshotMsat + quote.amountMsat
                        )}{" "}
                        sats
                      </dd>
                      <dt>Incoming remote → projected</dt>
                      <dd className="text-right">
                        {sats(quote.incomingRemoteSnapshotMsat)} →{" "}
                        {sats(
                          quote.incomingRemoteSnapshotMsat - quote.amountMsat
                        )}{" "}
                        sats
                      </dd>
                      <dt>Incoming reserves (local / remote)</dt>
                      <dd className="text-right">
                        {sats(quote.incomingLocalReserveMsat)} /{" "}
                        {sats(quote.incomingRemoteReserveMsat)} sats
                      </dd>
                      <dt>Expected on-chain change</dt>
                      <dd className="text-right">0 sats</dd>
                    </dl>
                    <p className="text-muted-foreground">
                      Projections use the quoted fixed-route debit. While the
                      payment is in flight, the amount is temporarily committed
                      in HTLCs; final balances are accepted only after both
                      durable payment records reconcile.
                    </p>
                  </div>
                  <div className="rounded-md bg-muted p-3 text-xs">
                    <div className="break-all">Quote ID: {quote.quoteId}</div>
                    <div className="break-all">
                      Route fingerprint: {quote.routeFingerprint}
                    </div>
                    <div>
                      Quoted: {new Date(quote.quotedAt).toLocaleString()}
                    </div>
                    <div>
                      Expires: {new Date(quote.expiresAt).toLocaleString()}
                    </div>
                    <div className="break-all">
                      First hop SCID: {quote.outgoingShortChannelId}
                    </div>
                    <div className="break-all">
                      Final hop SCID/alias: {quote.incomingShortChannelId}
                    </div>
                  </div>
                  {quote.paths.map((path, pathIndex) => (
                    <div
                      className="space-y-1 rounded-md border p-3 text-xs"
                      key={`${pathIndex}:${path.feeMsat}`}
                    >
                      <div className="font-medium">
                        Path {pathIndex + 1} · {path.hops.length} hops ·{" "}
                        {path.feeMsat.toLocaleString()} msat fee
                      </div>
                      {path.hops.map((hop, hopIndex) => (
                        <div
                          className="break-all text-muted-foreground"
                          key={`${hopIndex}:${hop.shortChannelId}`}
                        >
                          {hopIndex + 1}. {hop.nodePubkey} via SCID{" "}
                          {hop.shortChannelId}
                        </div>
                      ))}
                    </div>
                  ))}
                  {(!operation || operation.requiresExactRetry) && (
                    <div className="space-y-2 rounded-md border border-destructive p-3">
                      <Label htmlFor="rebalance-confirmation">
                        Type this exact confirmation
                      </Label>
                      <code className="block break-all text-xs">
                        {quote.executionConfirmation}
                      </code>
                      <Input
                        autoComplete="off"
                        id="rebalance-confirmation"
                        value={confirmation}
                        onChange={(event) =>
                          setConfirmation(event.target.value)
                        }
                      />
                    </div>
                  )}
                </div>
              )}

              {operation && (
                <Alert
                  variant={
                    operation.state === "succeeded" ? "default" : "destructive"
                  }
                >
                  {operation.state === "succeeded" ? (
                    <CheckCircle2Icon className="h-4 w-4" />
                  ) : (
                    <AlertTriangleIcon className="h-4 w-4" />
                  )}
                  <AlertTitle>
                    {operation.state === "succeeded"
                      ? "Rebalance reconciled"
                      : operation.state === "failed"
                        ? "Rebalance failed"
                        : operation.requiresExactRetry
                          ? "Exact-operation recovery required"
                          : "Payment submitted; reconciliation pending"}
                  </AlertTitle>
                  <AlertDescription>
                    <div className="space-y-1 break-all text-xs">
                      <div>
                        State/phase: {operation.state} / {operation.phase}
                      </div>
                      <div>Operation ID: {operation.operationId}</div>
                      {operation.paymentHash && (
                        <div>Payment hash: {operation.paymentHash}</div>
                      )}
                      {operation.outboundPaymentId && (
                        <div>
                          Outbound payment ID: {operation.outboundPaymentId}
                        </div>
                      )}
                      {operation.actualRoutingFeeMsat !== undefined && (
                        <div>
                          Actual routing fee:{" "}
                          {operation.actualRoutingFeeMsat.toLocaleString()} msat
                        </div>
                      )}
                      {operation.terminalEvidenceHash && (
                        <div>
                          Evidence hash: {operation.terminalEvidenceHash}
                        </div>
                      )}
                      {operation.failureReason && (
                        <div>{operation.failureReason}</div>
                      )}
                      {operation.reconciliationPending && (
                        <div>
                          Status is read from both durable payment records every
                          two seconds.
                        </div>
                      )}
                    </div>
                  </AlertDescription>
                </Alert>
              )}

              {!operation && quote && (
                <Alert variant="destructive">
                  <AlertTriangleIcon className="h-4 w-4" />
                  <AlertTitle>This action moves funds</AlertTitle>
                  <AlertDescription>
                    One fixed-route payment will debit at most{" "}
                    {Math.ceil(quote.maxTotalDebitMsat / 1000).toLocaleString()}{" "}
                    sats from channel {quote.outgoingChannelId}, return{" "}
                    {Math.floor(quote.amountMsat / 1000).toLocaleString()} sats
                    through channel {quote.incomingChannelId}, use no automatic
                    fallback route, and remain locked until both durable payment
                    records agree.
                  </AlertDescription>
                </Alert>
              )}

              {!quote && (
                <Alert>
                  <AlertTriangleIcon className="h-4 w-4" />
                  <AlertTitle>Quote first</AlertTitle>
                  <AlertDescription>
                    Finding a route is non-paying. Execution requires a fresh
                    quote and a separate exact confirmation.
                  </AlertDescription>
                </Alert>
              )}
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="mt-0 shrink-0 border-t p-4 [&>*]:w-full sm:flex-wrap sm:p-6 sm:pt-4 sm:[&>*]:w-auto">
          <AlertDialogCancel onClick={closeDialog}>Close</AlertDialogCancel>
          <LoadingButton
            disabled={Boolean(operation)}
            loading={isQuoting}
            type="submit"
          >
            Find local route
          </LoadingButton>
          <LoadingButton
            disabled={
              !quote?.executionEnabled ||
              confirmation !== quote.executionConfirmation ||
              Boolean(operation && !operation.requiresExactRetry)
            }
            loading={isExecuting}
            onClick={executeQuote}
            type="button"
            variant="destructive"
          >
            {operation?.requiresExactRetry
              ? "Retry exact operation"
              : "Execute exact route"}
          </LoadingButton>
        </AlertDialogFooter>
      </form>
    </AlertDialogContent>
  );
}

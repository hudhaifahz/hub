import { AlertTriangleIcon, LockIcon } from "lucide-react";
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
  paths: LocalRoutePath[];
  expiresAt: string;
  executionEnabled: boolean;
  blockedReason: string;
};

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
  const [quote, setQuote] = React.useState<RebalanceQuote>();

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

  return (
    <AlertDialogContent className="max-w-2xl">
      <form onSubmit={createQuote}>
        <AlertDialogHeader>
          <AlertDialogTitle>Exact-channel rebalance</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-4 text-left">
              <p>
                Choose both exact channels. This searches the local routing
                graph and saves an expiring review record; it creates no
                invoice, probe, HTLC, or payment.
              </p>

              <div className="space-y-2">
                <Label>Leave through this channel</Label>
                <Select
                  value={outgoingChannelId}
                  onValueChange={(value) => {
                    setOutgoingChannelId(value);
                    setQuote(undefined);
                  }}
                >
                  <SelectTrigger>
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
                <div className="rounded-md border p-3 text-sm">
                  <div>{channelIdentity(incomingChannel)}</div>
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
                <div>
                  <Label htmlFor="rebalance-amount">Principal (sats)</Label>
                  <Input
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
                <div>
                  <Label htmlFor="routing-fee-cap">Routing fee cap</Label>
                  <Input
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
                  <dl className="grid grid-cols-2 gap-2 text-sm">
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
                  </dl>
                  <div className="rounded-md bg-muted p-3 text-xs">
                    <div className="break-all">Quote ID: {quote.quoteId}</div>
                    <div className="break-all">
                      Route fingerprint: {quote.routeFingerprint}
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
                </div>
              )}

              <Alert variant="destructive">
                <AlertTriangleIcon className="h-4 w-4" />
                <AlertTitle>Payment execution remains locked</AlertTitle>
                <AlertDescription>
                  {quote?.blockedReason ||
                    "Local circular-route execution is not implemented or authorized."}{" "}
                  This route result cannot spend funds.
                </AlertDescription>
              </Alert>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="mt-4">
          <AlertDialogCancel onClick={closeDialog}>Close</AlertDialogCancel>
          <LoadingButton loading={isQuoting} type="submit">
            Find local route
          </LoadingButton>
          <LoadingButton disabled type="button" variant="destructive">
            <LockIcon /> Execute locked
          </LoadingButton>
        </AlertDialogFooter>
      </form>
    </AlertDialogContent>
  );
}

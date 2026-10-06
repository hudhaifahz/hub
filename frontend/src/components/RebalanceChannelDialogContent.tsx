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

type RebalanceQuote = {
  quoteId: string;
  amountMsat: number;
  providerFeeMsat: number;
  maxProviderFeeMsat: number;
  maxRoutingFeeMsat: number;
  maxTotalDebitMsat: number;
  outgoingChannelId: string;
  outgoingNodePubkey: string;
  incomingChannelId: string;
  incomingNodePubkey: string;
  outgoingSpendableSnapshotMsat: number;
  incomingReceivableSnapshotMsat: number;
  expiresAt: string;
  executionEnabled: boolean;
  blockedReason?: string;
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
  const [maxProviderFeeSat, setMaxProviderFeeSat] = React.useState("2500");
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
        "/api/channels/rebalance/quote",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            outgoingChannelId: outgoingChannel.id,
            outgoingNodePubkey: outgoingChannel.remotePubkey,
            incomingChannelId: incomingChannel.id,
            incomingNodePubkey: incomingChannel.remotePubkey,
            amountMsat: Number(amountSat) * 1000,
            maxProviderFeeMsat: Number(maxProviderFeeSat) * 1000,
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
                Choose both exact channels and separate fee limits. Creating a
                quote does not pay it.
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

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
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
                  <Label htmlFor="provider-fee-cap">Provider fee cap</Label>
                  <Input
                    id="provider-fee-cap"
                    type="number"
                    required
                    min={0}
                    step={1}
                    value={maxProviderFeeSat}
                    onChange={(event) => {
                      setMaxProviderFeeSat(event.target.value);
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
                    <dt>Provider fee</dt>
                    <dd className="text-right">
                      <FormattedBitcoinAmount
                        amountMsat={quote.providerFeeMsat}
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
                  <p className="text-xs text-muted-foreground">
                    Quote {quote.quoteId} · expires{" "}
                    {new Date(quote.expiresAt).toLocaleString()}
                  </p>
                </div>
              )}

              <Alert variant="destructive">
                <AlertTriangleIcon className="h-4 w-4" />
                <AlertTitle>Payment execution remains locked</AlertTitle>
                <AlertDescription>
                  {quote?.blockedReason ||
                    "The incoming-channel claim must be proven atomic before this build is allowed to move sats."}{" "}
                  A quote cannot spend funds.
                </AlertDescription>
              </Alert>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="mt-4">
          <AlertDialogCancel onClick={closeDialog}>Close</AlertDialogCancel>
          <LoadingButton loading={isQuoting} type="submit">
            Create non-paying quote
          </LoadingButton>
          <LoadingButton disabled type="button" variant="destructive">
            <LockIcon /> Execute locked
          </LoadingButton>
        </AlertDialogFooter>
      </form>
    </AlertDialogContent>
  );
}

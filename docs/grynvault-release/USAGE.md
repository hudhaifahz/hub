# Using exact-channel rebalance

Grynvault Rebalance moves Lightning liquidity between two channels on the same
node. It sends one circular payment that must leave through the selected
outgoing channel and return through the selected incoming channel. It does not
move funds on-chain.

This provides a local circular-payment alternative when the former
swap-backed rebalance flow is unavailable. It still requires:

- one active channel with enough local spendable outbound liquidity;
- a different active channel with enough remote receivable inbound liquidity;
- a viable public Lightning route between those exact peers; and
- enough headroom for channel reserves, in-flight HTLCs, and the routing fee.

If inbound liquidity is missing, obtain it separately before rebalancing. An
LSP or JIT channel, a channel opened from another node, or a reviewed
Lightning-to-on-chain/on-chain-to-Lightning service can change the liquidity
position. Third-party availability, custody, privacy, limits, and fees must be
checked at the time of use.

## 1. Choose the exact channels

1. Open **Channels**.
2. From the channel that should gain outbound liquidity, choose
   **Rebalance In**. This becomes the exact return channel.
3. Under **Leave through this channel**, choose the different channel that
   should lose outbound liquidity.
4. Verify both full peer public keys and channel IDs.
5. Enter the principal and maximum routing fee.
6. Select **Find local route**.

Finding a route is non-paying. It creates no invoice, probe, HTLC, or payment.

![Pre-quote dialog with synthetic identifiers](https://raw.githubusercontent.com/hudhaifahz/hub/e5442b98bcc21752d8362ec54348cf6ee9acff2b/doc/assets/exact-channel-rebalance/prequote-375.jpg)

## 2. Review the quote

Verify the exact outgoing and incoming channel IDs; principal, estimated fee,
fee cap, and maximum source debit; projected balances and zero expected
on-chain change; quote expiry; first and final hop; route fingerprint; and the
warning that no fallback route exists.

If the quote expires or the channel state changes, discard it and find a fresh
route.

![Quote review with synthetic identifiers](https://raw.githubusercontent.com/hudhaifahz/hub/e5442b98bcc21752d8362ec54348cf6ee9acff2b/doc/assets/exact-channel-rebalance/quote-review-1280.jpg)

## 3. Execute once

Execution is a real Lightning payment. Type the displayed confirmation exactly
and select **Execute exact route** only after a fresh owner review. Do not submit
another rebalance while the operation is pending or being reconciled.

![Exact confirmation with synthetic identifiers](https://raw.githubusercontent.com/hudhaifahz/hub/e5442b98bcc21752d8362ec54348cf6ee9acff2b/doc/assets/exact-channel-rebalance/confirmation-375.jpg)

The final record must report success and reconciliation, failure before
submission, submitted but still reconciling, or recovery required. Treat a
pending or recovery state as unresolved. Read the saved operation before doing
anything else; never infer success from a changed balance alone.

## Limits

Exact endpoint selection cannot create missing network liquidity. A route can
still fail because an intermediate channel lacks capacity, a peer is offline,
a fee or CLTV policy changed, or another payment consumed liquidity. The fixed
route is not retried through different channels.

The hosted Grynvault companion is informational only. It cannot connect to a
wallet, find a live route, quote, sign, pay, or rebalance.

## Screenshot provenance

The images use the real dialog component and synthetic fixture data. They
contain no live wallet public key, channel ID, balance, quote, fingerprint, or
timestamp. The complete viewport/state set, source notes, and SHA-256 values are
pinned at engineering commit
[`e5442b98bcc21752d8362ec54348cf6ee9acff2b`](https://github.com/hudhaifahz/hub/tree/e5442b98bcc21752d8362ec54348cf6ee9acff2b/doc/assets/exact-channel-rebalance).

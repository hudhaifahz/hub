# Grynvault Rebalance release notes

Status: **PREPARED — NOT SIGNED — NOT PUBLISHED**

Grynvault Rebalance adds a fail-closed exact-channel circular rebalance flow to
Alby Hub. The owner chooses the exact channel sats must leave through and the
exact channel through which they must return. Route discovery is non-paying;
execution requires a fresh fingerprint-bound quote and exact typed
confirmation.

## Why this exists

The built-in Boltz swap rebalance path is no longer available for this use
case. This release provides a local circular-payment alternative when the node
already has:

- one channel with enough outbound spendable balance for principal plus fees;
- a different channel with enough inbound receiving capacity; and
- a viable Lightning route connecting the selected peers.

The software does not manufacture liquidity and cannot guarantee a route.
Inbound capacity can be obtained separately, for example through an LSP/JIT
channel or a third-party Lightning-to-on-chain exit, but those are independent
services with their own fees, custody, privacy, and counterparty risks.

## Safety model

- exact first-hop and receiving-channel enforcement occurs inside the local
  Hub/LDK process;
- no automatic fallback route or automatic payment retry exists;
- preparation is persisted before the one fixed-route send;
- both durable payment legs must reconcile before terminal success;
- bounded owner-only diagnostics omit invoices, preimages, payment secrets, and
  serialized routes;
- the hosted Grynvault HTTP/MCP companion is informational only and cannot
  access a wallet, quote, sign, route, pay, or rebalance.

## Initial platform

Universal macOS only. Public availability requires a Grynvault Developer ID
signature, Apple notarization and stapling, a PGP-signed artifact manifest,
fresh-download verification, and independent readback of every published URL.

## Attribution

Built from Alby Hub and licensed under Apache-2.0. This is an experimental
independent release and is not endorsed by Alby.

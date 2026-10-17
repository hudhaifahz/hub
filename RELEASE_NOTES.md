# Grynvault Rebalance release notes

Status: **GUIDED MIGRATION SYNTHETIC ACCEPTANCE PASSED — OWNER ACCEPTANCE
PENDING — NOT TAGGED — NOT SIGNED — NOT PUBLISHED**

Grynvault Rebalance adds a fail-closed exact-channel circular rebalance flow to
Alby Hub. The owner chooses the exact channel sats must leave through and the
exact channel through which they must return. Route discovery is non-paying;
execution requires a fresh fingerprint-bound quote and exact typed
confirmation.

Seamless guided migration from an existing Alby Hub embedded-LDK wallet is a
core release requirement. The successor importer implements custom-`WORK_DIR`
discovery, fail-closed competing-copy detection, source quiescence, byte-exact
recovery and staging, deterministic crash recovery, and a persistent network
quarantine before an imported node can contact peers.

A disposable regtest acceptance creates a real LDK wallet with two channels
and inbound/outbound payment history. It passes migration from both the normal
macOS container and a custom path with spaces and Unicode, proves the copied
node remains network-dark before activation, verifies identity and continuity,
then activates and reconnects. An intentionally stale matching clone is blocked
before process or network checks. No production wallet or funds are used.

Publication remains blocked on owner-supervised non-paying acceptance and the
remaining gates in `docs/grynvault-release/GUIDED_MIGRATION_PLAN.md`.

This candidate is replayed onto Alby Hub `v1.24.3`. It also makes long route
identifiers and balance evidence readable at constrained widths and continues
polling a submitted operation after a transient status-refresh failure.

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

Universal macOS 15.4 or newer only. The owner-selected package is intentionally
ad-hoc signed, not Apple Developer ID signed, and not notarized or stapled. Its
filename and in-image warning say `UNSIGNED-EXPERIMENTAL`; Gatekeeper is expected
to warn or block it. Public availability still requires a PGP-signed artifact
manifest, fresh-download verification, and independent readback of every
published URL.

See `MIGRATION.md` for the guided import and recovery flow and
`docs/grynvault-release/USAGE.md` for the non-paying quote flow, exact
confirmation boundary, failure interpretation, and privacy-safe screenshots.

## Attribution

Built from Alby Hub and licensed under Apache-2.0. This is an experimental
independent release and is not endorsed by Alby.

# Grynvault Rebalance release notes

## v0.1.0-rc.8 — Lightning reconnection and security updates

This macOS release integrates the locally verified Tor reliability repair with
Alby Hub updates through `5a7f1d661f248a48659d19f9a4b3299ee282187e`, including
Wails 2.17.1 and the current patched Go and Rust dependencies.

- Opt-in native SOCKS transport waits for Tor to connect before starting the
  Lightning handshake. The existing default direct connection path is preserved.
- Reconnection attempts run independently per peer, use bounded elapsed-time
  delays, and clear pending state when interrupted. Successful connections reset
  the delay. With the existing polling interval, retries occur within about five
  minutes after a completed failure rather than waiting hours.
- An optional outbound-only Tor setup helper is included in the DMG. It starts
  after login and restarts on exit. No personal routes or wallet data are included.
- The build uses Go 1.27.4 without disabling its native-library security checks.
  Bark v0.26.1's embedded LLVM IR is removed from verified build copies; retained
  native code, data, relocations and symbols are independently checked unchanged.

This is an unsigned experimental universal macOS 15.4+ release with a PGP-signed
artifact manifest. It is not Apple-notarized. The patch adds no wallet database
migration. Existing upstream migrations still apply when updating older versions;
follow `MIGRATION.md` for imports and keep normal backups. No release validation
uses a production wallet or sends a payment.

Peer availability is not a guarantee of payment success: remote outages, routes,
liquidity and invoice conditions still matter. A sleeping or shut-down laptop
cannot maintain Lightning service.

Grynvault's existing exact-channel rebalance and guided migration features remain
part of this release. Exact-channel execution requires a fresh bound quote and
explicit confirmation. The earlier migration development records below remain
historical evidence rather than the release status of RC.8.

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

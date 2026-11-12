# Coinos inbound-liquidity plan

Status: design and read-only provider discovery only. No Coinos order, invoice,
HTLC, payment, or channel operation has been created.

## Problem and product boundary

Grynvault Rebalance currently solves one specific problem: move existing
liquidity in a circle through an exact outgoing channel and an exact returning
channel. That requires two usable channels with complementary liquidity.

A user who has outbound liquidity but no useful inbound channel needs a
different operation. A Lightning-to-Bitcoin reverse submarine swap spends
Lightning through a selected channel and returns bitcoin to an on-chain
address. The selected channel gains inbound capacity while the user's total
Lightning balance falls and their on-chain balance rises.

The product must present these as separate tools:

1. **Move liquidity between channels** — exact-channel circular rebalance.
2. **Create inbound on one channel** — Coinos non-custodial swap out to the
   Hub's own on-chain wallet, with an exact outgoing channel.
3. **Buy or open inbound** — LSPS2/JIT or another channel-provider workflow.
   This is future work and must not be described as a swap or rebalance.

Sending to a custodial Coinos account is not part of this feature. If it is
ever offered, it must be separately labelled as a custodial transfer.

## Provider evidence recorded 2026-10-21

- `https://swap.coinos.io/` identifies itself as a non-custodial Bitcoin bridge
  and links to the open-source Boltz web application and backend.
- Its public API responds to the Boltz v2 reverse-swap and fee endpoints.
- The API reported backend version `3.13.0-4d131ef8` during inspection.
- At the time of inspection, the BTC-to-BTC reverse pair advertised a 25,000
  sat minimum, 600,000 sat maximum, 0.5% provider fee, and 530 sats of combined
  lockup and claim miner fees.
- Limits and fees are discovery evidence, not constants. Every quote must fetch
  and bind the current values.

The existing Alby Hub swap service already uses the Boltz v2 client, but it is
single-provider and pays the swap invoice with ordinary pathfinding. Changing
only the API URL would therefore not meet Grynvault's exact-channel promise.

## Required user flow

1. Select **Create inbound** from the liquidity screen.
2. Select the exact channel whose local balance should become inbound.
3. Enter the principal and choose the Hub's fresh internal on-chain address.
4. Review a non-paying quote that shows:
   - provider identity and exact API origin;
   - exact source peer and channel ID;
   - principal and maximum Lightning debit;
   - provider, miner, and maximum routing fees;
   - expected on-chain receipt and estimated inbound created;
   - current and projected channel balances and reserves;
   - quote expiry and a fingerprint of every consequential field.
5. Re-enter a short confirmation bound to the quote fingerprint.
6. Create one provider order and pay its fixed invoice through only the
   selected first-hop channel.
7. Show persistent progress through invoice payment, lockup, claim, and final
   confirmation. A restart must resume the same order rather than create a new
   one.
8. Verify the exact channel delta and the on-chain receipt before reporting
   success.

## Fail-closed controls

- Support embedded LDK only until another backend implements exact first-hop
  payment. Never fall back to ordinary pathfinding.
- Require the selected channel to exist, be usable, and cover principal plus
  the absolute routing-fee cap while preserving its reserve.
- Require an absolute routing-fee cap and show its ppm equivalent. The payment
  must fail if either the selected first hop or cap cannot be honored.
- Bind provider origin, pair hash, amount, destination, source channel, fee
  components, maximum debit, expiry, and wallet identity into the quote
  fingerprint.
- Expire a quote on time, provider-quote change, channel-balance change, wallet
  change, or application restart unless a durable reviewed operation already
  exists.
- Persist an idempotency/operation record before contacting the provider.
- Pin each swap to the API origin that created it. Never silently retry against
  a different provider.
- Allow only the reviewed HTTPS Coinos origin by default. Validate all API,
  invoice, script, amount, timeout, and key material already checked by the
  Boltz client integration.
- Never request or store a Coinos login, password, API token, or wallet seed.
- Keep secrets, invoices, preimages, addresses, and full route data out of
  ordinary logs and support exports.
- Disable automatic Coinos swap-out in the first release. Every value-moving
  operation requires a fresh quote and explicit confirmation.

## Recovery requirements

- A paid invoice with a pending lockup must remain visible after restart and
  resume provider/websocket or polling reconciliation.
- A provider timeout before order creation must leave no payable operation.
- An ambiguous create-order response must be reconciled by the durable
  operation identifier before any retry.
- An invoice payment failure must retain the order for diagnosis and must not
  retry through another channel.
- Refund and cooperative-claim paths must be exercised with fixtures before a
  live acceptance payment.
- The UI must distinguish pending, held HTLC, payment failed, lockup seen,
  claim broadcast, confirmed, recoverable, and manual-attention states.

## Implementation checklist

### A. Provider and persistence

- [ ] Add Coinos to a local provider registry without depending on Alby's
      hosted `GetInfo` provider selection.
- [ ] Refactor swap recovery so each stored swap uses its own recorded API URL
      and websocket/client.
- [ ] Add a durable inbound-liquidity quote/operation model and append-only DB
      migration.
- [ ] Prove existing Boltz and SatsRouting swaps still resume unchanged.

### B. Exact-source payment

- [ ] Route swap invoices through `PinnedPaymentClient` using the reviewed
      channel ID and absolute fee cap.
- [ ] Record the outgoing transaction with the same duplicate-payment,
      balance, metadata, and settlement protections as ordinary payments.
- [ ] Unit-test wrong, missing, offline, ambiguous, and insufficient channels;
      changed invoices; over-cap fees; duplicate calls; restart recovery; and
      the absence of fallback routing.

### C. Quote and API

- [ ] Add read-only quote creation and retrieval endpoints.
- [ ] Add a separate execute endpoint accepting only an unexpired quote ID,
      fingerprint, and confirmation.
- [ ] Revalidate provider terms and channel state immediately before provider
      order creation.
- [ ] Redact quote and route records exposed to logs/support bundles.

### D. User interface

- [ ] Add a top-level Liquidity choice between circular rebalance and creating
      inbound.
- [ ] Explain that swap-out moves value from Lightning to the user's on-chain
      wallet and does not create or reopen a channel.
- [ ] Show exact before/after balances, all fee ceilings, provider attribution,
      quote expiry, and recovery state on desktop and small screens.
- [ ] Link to source, provider terms, recovery guidance, and risk disclosure.

### E. Acceptance and release

- [ ] Go tests, frontend lint/typecheck/build, race-focused tests, and migration
      tests pass without weakening existing exact-channel controls.
- [ ] Quote-only acceptance against the live Coinos API creates no order.
- [ ] Offline fixtures prove order creation, pinned payment, restart recovery,
      claim, and failure states without moving value.
- [ ] Owner separately approves one minimum-size live Coinos acceptance swap,
      exact channel, amount, destination, and maximum routing fee.
- [ ] Independently verify the on-chain receipt and exact channel balance delta.
- [ ] Complete security review, update user documentation/screenshots, then use
      the normal immutable tag, artifact, PGP-signing, and publication gates.

## Non-goals for the first release

- No custodial Coinos account integration.
- No automatic swaps or unattended value movement.
- No promise that a swap repairs an offline or closing channel.
- No new-channel purchase bundled into the swap flow.
- No claim that provider availability or quoted fees are guaranteed.


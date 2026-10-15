# Grynvault Rebalance operator-assisted service

Status: **PREPARED FOR PILOT VALIDATION — NOT YET FOR SALE**

This document defines the narrow service that may be offered only after the
readiness gates below close. It does not authorize a customer payment, access
to a customer wallet, a rebalance, or a public sales claim.

## Offer

Help a self-hosting Alby Hub or compatible LDK operator safely use Grynvault
Rebalance when the operator already has:

- one exact channel with sufficient outbound spendable balance;
- another exact channel with sufficient inbound receiving capacity;
- a viable public route between those counterparties; and
- a current backup and tested recovery plan.

The service may include release verification, backup and migration preparation,
compatibility checks, channel-readiness review, interpretation of one non-paying
quote, an exact owner action packet, and post-operation evidence reconciliation.
It is an on-node circular alternative to an external swap-based rebalance; it
does not create liquidity, perform an on-chain swap, or guarantee a route.

## Non-custodial operating boundary

The customer keeps exclusive control of the machine, wallet, seed, signer,
password, private keys, NWC connections, and the final typed confirmation. The
operator must never request or receive those secrets and must never remote-click
the value-moving confirmation for a customer.

The public service at `https://rebalance.grynvault.com` remains read-only. It
may publish release metadata, compatibility information, and sanitized
diagnostic explanations. It must not accept wallet credentials or NWC, create
an invoice, build or submit a customer route, execute a payment, or deduct a
service fee.

## Proposed pilot price

After all readiness gates close, validate one simple price:

- **1,000 sats, success only**, invoiced after the customer's Hub records a
  reconciled successful operation;
- Lightning routing fees are separate and shown before execution;
- no service fee for `no_route`, expiry, owner cancellation, or a failed
  operation;
- minimum suggested principal: **100,000 sats**; and
- no subscription, automatic debit, percentage-of-funds fee, or bundled retry.

At 100,000 sats, a 1,000-sat service fee is 1% before network fees. At 10,000
sats it is 10% and should not be offered. Record support time and actual routing
fees before deciding whether 1,000 sats is sustainable.

## Required customer disclosure

Before the customer installs or pays, show all of the following in plain text:

- experimental, independently maintained Alby Hub fork;
- not an official Alby release or endorsement;
- macOS package is unsigned by Apple and unnotarized;
- Gatekeeper may warn or block it;
- verify the PGP-signed manifest and exact source tag;
- real sats can be lost through software, backup, channel, routing, or operator
  error;
- no guarantee of route availability, completion, savings, income, or uptime;
- not independently audited;
- the operator never takes custody and never asks for wallet secrets; and
- each attempt requires a fresh quote and customer confirmation.

## Intake and compatibility checklist

- [ ] Customer confirms self-hosted Alby Hub or a specifically supported build.
- [ ] Current version, operating system, architecture, and package origin are
      recorded without collecting secrets.
- [ ] Customer independently verifies the public key fingerprint and manifest.
- [ ] Current backup and rollback path are confirmed before installation.
- [ ] Exact outgoing and incoming channel IDs and full peer pubkeys are read
      from the customer's own UI; labels alone are never accepted.
- [ ] Both channels are online and show enough spendable/receiving capacity.
- [ ] Requested principal, routing-fee cap, and risk tolerance are recorded.
- [ ] The customer understands that route discovery is non-paying and a quote
      can fail without proving a software defect.
- [ ] The customer—not the operator—reviews and types the exact confirmation.

## One-at-a-time action packet

For every attempt, record and show:

- exact principal in sats and eight-decimal BTC;
- exact outgoing/incoming channel IDs and full peer pubkeys;
- current spendable and receiving capacity;
- quote time, expiry, quote ID, and route fingerprint;
- first/final SCIDs and every quoted path;
- exact estimated routing fee, fee cap, and maximum source debit;
- projected local/remote channel balances and expected on-chain change;
- explicit statement that there is no fallback route or automatic retry; and
- exact confirmation phrase, without typing it for the customer.

If the quote expires or any field changes, discard it and prepare a new packet.
Never batch approvals across attempts.

## Post-operation evidence

- [ ] Read the operation state to terminal success or a specific terminal
      failure; do not infer success from a disappearing modal.
- [ ] Match both durable payment records, payment hash/preimage, operation ID,
      exact channels, actual fee, and terminal evidence hash.
- [ ] Read channel balances again and distinguish temporary HTLC commitment,
      reserve changes, and spendable balance.
- [ ] Record no-route or remote-failure evidence without exposing invoices,
      payment secrets, wallet history, or unrelated channel identifiers.
- [ ] Create the 1,000-sat service invoice only after reconciled success; settle
      it separately from the rebalance and never retry ambiguous payment state.

## Gates before the first paid customer

- [ ] Clean-install public RC.2 on a clean macOS 15.4+ profile.
- [ ] Test migration using only an approved copy of wallet data; verify unlock,
      database integrity, channels, balances, history, NWC permissions, backup,
      embedded hashes, UI, and rollback.
- [ ] Complete at least three fresh proof pilots against the intended exact
      outgoing/incoming channels, one owner approval at a time.
- [ ] Exercise at least one safe quote-expiry or no-route case and prove no sats
      moved and no fee was owed.
- [ ] Publish the exact support scope, disclosures, refund rule, privacy rule,
      security contact, and incident-stop procedure.
- [ ] Recheck current dependency, secret, source-boundary, and release-artifact
      evidence immediately before opening sales.
- [ ] Obtain owner approval for the exact public offer and payment mechanism.

## Separate routing-node business

This service cannot force ordinary public payments through Grynvault-operated
nodes. Routing income requires separate public nodes with committed liquidity,
useful counterparties, competitive fees, uptime, rebalancing costs, monitoring,
and real forwarding demand. Evaluate that as a distinct business with its own
capital limits, node/privacy model, accounting, and security review. Do not use
customer credentials or the read-only companion to manufacture routing flow.

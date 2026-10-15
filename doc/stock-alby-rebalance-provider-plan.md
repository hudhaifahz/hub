# Stock Alby rebalance provider plan

Status: architecture and acceptance plan only. No provider is deployed, no node
or wallet credential is configured, and no payment is authorized by this file.

This plan defines a second Grynvault product path for people who use an
unmodified Alby Hub. It does not replace or weaken the exact-channel feature in
the Grynvault Alby/LDK fork.

## Decision

Ship and describe two different modes with different guarantees:

| Mode | Client | Customer channel guarantee | Intended use |
| --- | --- | --- | --- |
| Grynvault Exact Rebalance | Grynvault Alby/LDK build, and later stock Alby if the upstream changes are accepted | Exact outgoing and exact incoming channel IDs; fail closed if either cannot be honored | Operators who need deterministic local-channel movement |
| Grynvault Rebalance Provider | Stock Alby Hub with `REBALANCE_SERVICE_URL` configured | Peer-targeted, best effort; no exact customer channel-ID guarantee | Easier real rebalancing without installing a custom Hub |

Never call the stock-provider mode `exact-channel`, `channel-pinned`, or
`guaranteed`. Its UI, documentation, receipts, pricing, and support terms must
say `peer-targeted` and `best effort`.

## Will it actually rebalance sats?

Yes, if both Lightning payments settle. The stock flow creates a real circular
liquidity movement:

1. The customer's Alby Hub creates a receiving invoice for the principal.
2. The customer Hub asks the configured provider for an order.
3. The provider returns a fee-bearing invoice.
4. The customer Hub pays that provider invoice. This reduces the customer's
   local balance on whichever outbound channel Hub selects.
5. A separate provider Lightning node pays the customer's receiving invoice.
   This increases the customer's local balance on whichever eligible inbound
   channel receives the payment.
6. The customer's wallet ends with approximately the same principal, less the
   provider fee and both routing costs, but with liquidity moved between its
   channels.

The provider must use a Lightning node that is operationally and cryptographically
separate from the customer's node. Using the customer's own node connection as
the provider payment engine is out of scope: it adds dangerous wallet authority,
does not add route control, and is not representative proof of a public service.
An owner self-test is meaningful only when the Grynvault provider node and the
owner's customer Hub are different Lightning node identities.

## Why stock Alby cannot guarantee both customer channel IDs

Alby Hub v1.24.1 sends the provider only:

```json
{
  "token": "alby-hub",
  "pay_request": "<customer receiving invoice>",
  "pay_through_this_public_key": "<requested receiving peer pubkey>"
}
```

The request contains no outgoing customer channel ID and no incoming customer
channel ID. The response contains only an order ID and provider invoice. Stock
Hub then uses its ordinary payment operation, which has no provider-supplied
first-hop channel field.

Consequences:

- A remote provider cannot choose the first local channel inside the customer's
  LDK process.
- The peer public key can target a counterparty but cannot distinguish parallel
  customer channels to that same peer.
- Invoice route hints guide route finding; they do not force the payer's local
  first hop or atomically prove the final local channel.
- Customer NWC credentials add payment authority, not a missing route-injection
  primitive. The provider must not request or store them.
- A provider may control routes on its own node, but those are provider channel
  IDs, not the customer's local channel IDs.

The exact-channel feature therefore remains local to Hub/LDK until Alby accepts
the upstream changes that expose and enforce exact local channel identities.

## Stock compatibility contract

The compatibility endpoint is:

`POST /api/rebalance/v1/create_order`

The first implementation must accept only the three stock fields shown above,
reject unknown fields, and return only:

```json
{
  "order_id": "<opaque id>",
  "pay_request": "<provider invoice>"
}
```

The provider invoice face amount must stay within stock Hub's enforced maximum:
the requested principal plus 0.5 percent (with the implementation's one-msat
rounding allowance). A flat 1,000-sat fee therefore cannot work for every order;
under the stock cap it first fits at a 200,000-sat principal. Pricing must be
amount-aware and independently verified against the exact supported Hub
versions before publication.

Stock Hub uses a 60-second HTTP timeout for order creation. The endpoint must
return the order and provider invoice within that window; settlement continues
asynchronously after the response.

## Proposed safe settlement protocol

The production design should use a provider hold invoice, subject to proving
that the selected provider-node backend exposes safe create, observe, settle,
cancel, and recovery operations:

1. Decode and validate the customer's receiving invoice: supported network,
   fixed positive amount, expiry margin, payment hash uniqueness, and configured
   minimum/maximum principal.
2. Validate the requested peer as a compressed public key and apply per-order,
   per-origin, and global limits.
3. Persist an idempotent order before returning anything.
4. Return a provider hold invoice for principal plus the disclosed provider fee.
5. Wait for the customer's HTLC to be accepted but do not settle it yet.
6. Pay the customer's invoice from the separate provider node.
7. Settle the held customer payment only after the customer invoice is proven
   settled. Cancel the hold invoice on a definitive route failure or timeout.
8. Reconcile both legs after every restart or ambiguous timeout before any
   retry. Never automatically create a second payment for an unknown outcome.

This is a proposed atomicity pattern, not a completed safety claim. It protects
the customer from the provider collecting principal before returning it, but it
creates provider-side loss and liquidity risks that must be bounded and tested.
No public or paid order is allowed until crash recovery, expiry races,
idempotency, duplicate requests, partial/MPP behavior, and terminal reconciliation
have passed deterministic and controlled live tests.

## Isolation and trust boundaries

- Keep `rebalance.grynvault.com` and the existing
  `grynvault-rebalance-companion` repository read-only and stateless.
- Build the payment provider in a separate repository, deployment, hostname,
  credential set, logs, alerts, and incident-stop control.
- Proposed hostname: `provider.rebalance.grynvault.com`. DNS and deployment need
  separate owner approval.
- Do not accept customer seeds, passwords, recovery material, macaroons, NWC
  connection strings, API tokens, private keys, wallet databases, or logs.
- The provider holds only its own tightly scoped node credential in managed
  secret storage. It never returns or logs invoices, preimages, payment secrets,
  channel IDs, request bodies, headers, or query strings.
- Use an edge Worker only as the validated public front door. Keep the node
  connector/payment engine isolated behind an authenticated service boundary.
- Give each order a strongly consistent state owner. Current Cloudflare guidance
  favors a SQLite-backed Durable Object per coordination atom; final platform
  choice remains gated on payment-engine connectivity and recovery testing.
- The public companion MCP may document, validate release metadata, and explain
  reason codes. It must never become a wallet or payment proxy.

## Provider state machine

Minimum durable states:

```text
CREATED
  -> CUSTOMER_PAYMENT_PENDING
  -> CUSTOMER_HTLC_ACCEPTED
  -> RETURN_PAYMENT_IN_FLIGHT
  -> RETURN_PAYMENT_SUCCEEDED
  -> CUSTOMER_PAYMENT_SETTLED
  -> SUCCEEDED
```

Terminal or recovery states must include:

```text
EXPIRED_UNPAID
CANCELLED_NO_ROUTE
FAILED_CONFIRMED
OUTCOME_UNKNOWN
RECOVERY_REQUIRED
```

Transitions must be monotonic and idempotent. `OUTCOME_UNKNOWN` and
`RECOVERY_REQUIRED` stop new value-moving work until independent reconciliation.

## Acceptance checklist

### Contract and non-paying tests

- [ ] Freeze the exact v1 request/response fixtures from Alby Hub v1.24.1.
- [ ] Reject non-POST methods, unsupported content types, unknown fields,
      oversized bodies, malformed JSON, and duplicate JSON keys.
- [ ] Decode and validate supported BOLT 11 networks, amount, expiry, signature,
      and payment hash uniqueness.
- [ ] Reject zero, amountless, expired, near-expiry, wrong-network, over-limit,
      replayed, and already-paid invoices.
- [ ] Validate the requested peer key without claiming a channel guarantee.
- [ ] Prove provider fee calculations remain below the client's enforced cap.
- [ ] Prove idempotent retries return the same order and never create a second
      provider invoice or return payment.
- [ ] Test every state transition, restart point, alarm, timeout, and terminal
      reconciliation path with a fake node connector.
- [ ] Run type, lint, unit, integration, dependency, secret, and Worker dry-run
      checks with generated binding types.

### Private regtest/testnet acceptance

- [ ] Use distinct customer and provider node identities.
- [ ] Prove a completed two-leg liquidity shift and exact balance accounting.
- [ ] Prove no-route, expiry, duplicate request, delayed callback, connector
      outage, Worker restart, state-owner restart, and ambiguous-result cases.
- [ ] Prove held funds are cancelled or settled exactly once and that an
      ambiguous state cannot trigger an automatic retry.
- [ ] Confirm stock Hub displays an honest best-effort warning and total cost.

### Owner-gated mainnet pilot

- [ ] Complete security review with zero unresolved Critical or High findings.
- [ ] Publish privacy, fee, support, refund, incident-stop, and known-limitation
      terms before accepting a paid customer.
- [ ] Prepare one fresh action packet with exact sats, eight-decimal BTC amount,
      provider fee, routing-fee limits, expiry, balances, node identities, and
      rollback/stop conditions.
- [ ] Obtain separate owner approval immediately before each value-moving pilot.
- [ ] Reconcile both payment hashes and customer/provider balance changes before
      considering the pilot successful or attempting another.

## Implementation sequence

1. Create a separate private repository and threat model; do not modify the
   live companion.
2. Implement only the strict stock-Alby contract, state machine, fake node
   connector, and deterministic tests.
3. Select and prove a provider-node backend with hold-invoice semantics in an
   isolated non-mainnet environment.
4. Add authenticated edge-to-engine communication, secret rotation, limits,
   monitoring, reconciliation, and an operator kill switch.
5. Complete a private deployment and non-paying acceptance. Deployment is not
   production activation.
6. Complete regtest/testnet two-node acceptance.
7. Stop for the owner's exact mainnet pilot approval packet.
8. Only after controlled pilots and current security evidence, stop again for
   separate DNS, pricing, and public-launch approvals.

## Public wording

Use wording equivalent to:

> Grynvault Rebalance Provider performs a real two-leg Lightning rebalance from
> stock Alby Hub. It targets your selected receiving peer, but stock Alby does
> not expose exact local channel pinning, so either local channel may differ.
> Use Grynvault Exact Rebalance when both channel IDs must be guaranteed.

Do not advertise the provider as available, tested, live, atomic, secure,
audited, or paid until the matching acceptance evidence above exists.

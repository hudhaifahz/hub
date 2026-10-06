# Pinned Channel Rebalance Reference Plan

Status: planning and implementation only. No live rebalance is authorized by this plan.

This document is the durable checklist and learning log for adding fail-closed channel routing to the Alby Hub rebalance flow. Update it whenever implementation evidence changes an assumption. Do not mark an item complete from code presence alone; require the stated verification evidence.

## Objective

Allow an owner to request a circular rebalance with:

- one exact local outgoing channel;
- one exact local incoming channel;
- an exact principal amount;
- explicit provider-fee and routing-fee limits;
- no automatic substitution of either selected channel;
- a quote/review boundary before any value-moving call; and
- independently recorded route and balance evidence after execution.

## Safety invariants

- [ ] Channel labels are display-only. Bind decisions to stable channel identity plus full counterparty pubkey.
- [ ] Every outgoing MPP part uses the selected local channel as its first hop.
- [ ] Retries remain pinned. A retry must never fall back to another local channel.
- [ ] Restart recovery remains pinned or fails closed; it must never resume as an unrestricted payment.
- [ ] Selected-channel offline, unusable, ambiguous, or insufficient states stop execution.
- [ ] A changed or expired quote stops execution and requires a new review.
- [ ] Provider fees and routing fees have separate exact limits.
- [ ] No unknown or unbounded fee is payable.
- [ ] Incoming channel enforcement is not claimed until it is verified before settlement or the provider protocol is proven atomic.
- [ ] Only one rebalance operation may execute at a time.
- [ ] A live rebalance always requires a fresh, action-specific owner approval packet.
- [ ] Tests, builds, deployment, UI presence, live execution, and economic success are reported as separate states.

## Confirmed starting point

- [x] Upstream Alby Hub base is tag `v1.24.0` at commit `d8ef0e70`.
- [x] Local customization branch is `fix/frontier-crown-nwc-metadata`.
- [x] Existing NWC metadata and app-icon changes were local and uncommitted when this plan was created.
- [x] Existing `origin` points to `https://github.com/getAlby/hub.git`, not an owner-controlled fork.
- [x] The local branch had no configured upstream branch.
- [x] Stock rebalance selects only an incoming peer pubkey in the API.
- [x] Stock LDK payment code uses ordinary pathfinding with no outgoing first-hop constraint.
- [x] The current LDK Go binding exposes fee, CLTV, path-count, and saturation controls but no first-hop selector.
- [x] Lower-level LDK routing supports restricting route finding to a supplied `first_hops` list.
- [x] Lower-level LDK claimable-payment events expose receiving channel IDs, but the current Go binding omits them.

## Version-control checklist

- [ ] Preserve the existing customization as its own reviewed commit before routing work.
- [ ] Exclude unrelated dependency-lock drift unless it is proven necessary and intentionally committed.
- [ ] Add an owner-controlled GitHub fork or repository remote without overwriting upstream history.
- [ ] Keep official Alby as `upstream` and the owner-controlled repository as `origin` or `fork`.
- [ ] Push the preserved metadata branch and configure its upstream tracking branch.
- [ ] Create a separate feature branch for pinned rebalancing.
- [ ] Keep the reference plan updated and committed with each material design change.
- [ ] Record the exact Alby base commit and exact LDK dependency commit in the implementation notes.
- [ ] Keep LDK dependency changes in a separately versioned fork/commit; do not depend on an uncommitted module-cache edit.
- [ ] Use small Conventional Commits so future Alby upgrades can replay or drop each customization independently.
- [ ] Tag or otherwise record the exact source commit used for any installed desktop build.

## Required design

### 1. Explicit request model

- [ ] Add exact outgoing and incoming channel identifiers to quote requests.
- [ ] Include both full counterparty pubkeys as identity cross-checks.
- [ ] Use millisatoshis internally; reject rounding and overflow.
- [ ] Require separate maximum provider-fee and routing-fee values.
- [ ] Reject identical source and destination channels.
- [ ] Reject multiple local channels to the same selected peer unless the exact channel can be enforced.

### 2. Quote-only stage

- [ ] Create a provider order and decode its invoice without paying it.
- [ ] Validate network, amount, provider-fee ceiling, expiry, and payment identity.
- [ ] Snapshot selected channel identities, usability, capacities, reserves, and balances.
- [ ] Calculate the maximum possible source debit and expected post-action ranges.
- [ ] Persist a single-use quote with an expiry and a hash of all material parameters.
- [ ] Return a human-readable review packet; do not initiate payment.

### 3. Fail-closed execution stage

- [ ] Accept only an unexpired, unused quote identifier.
- [ ] Re-read channel identity, state, capacity, and fee limits immediately before sending.
- [ ] Reject any material difference rather than silently refreshing the quote.
- [ ] Atomically mark the quote executing before calling the Lightning backend.
- [ ] Prevent duplicate execution and concurrent rebalance operations.
- [ ] Persist provider order ID, invoice/payment hash, exact limits, and selected channel IDs.

### 4. Outgoing first-hop enforcement

- [ ] Add a pinned-payment capability to the LDK dependency and its Go binding.
- [ ] Supply only the selected `ChannelDetails` entry as `first_hops` for route finding.
- [ ] Permit downstream MPP when useful while requiring every part to share the selected first hop.
- [ ] Apply the same constraint to all automatic retries.
- [ ] Persist the constraint with the payment ID before the payment can start.
- [ ] On missing constraint, unavailable channel, or route failure: fail with no fallback.
- [ ] Return or publish actual successful-path evidence including each first-hop channel ID.
- [ ] Do not implement routing control by disconnecting peers, disabling channels, manipulating fees, or temporarily hiding channels.

### 5. Incoming-channel enforcement and atomicity gate

- [ ] Expose `receiving_channel_ids` through the LDK language binding and Alby event model.
- [ ] Determine whether the rebalance provider supports a compatible hold/claim or shared-preimage atomic flow.
- [ ] Prove that rejecting a misrouted inbound HTLC cannot leave the provider invoice settled without the principal returning.
- [ ] If atomicity is proven, claim only when every incoming MPP part uses the selected channel.
- [ ] If any part arrives elsewhere, fail the inbound payment and record the reason.
- [ ] If atomicity cannot be proven, do not label destination routing as enforced and do not enable the value-moving feature.

### 6. Owner-facing UI

- [ ] Show exact outgoing and incoming peer names, full pubkeys, and stable channel identifiers.
- [ ] Show channel status, spending/receiving capacity, reserve, and projected post-action ranges.
- [ ] Show principal, provider fee, routing-fee cap, and maximum total debit separately.
- [ ] Make quote creation visibly non-paying.
- [ ] Require a distinct final confirmation for the exact, still-valid quote.
- [ ] Disable execution when safety preconditions are not met.
- [ ] Never offer or imply an automatic fallback route.
- [ ] Show a final evidence record, not only a success toast.

## Verification checklist

### Unit and interface tests

- [ ] Exact selected channel is passed to the LDK routing layer.
- [ ] A cheaper alternative first hop is never used.
- [ ] All MPP parts share the selected first hop.
- [ ] Retry remains pinned.
- [ ] Restart recovery remains pinned or fails closed.
- [ ] Offline, unusable, insufficient, stale, expired, ambiguous, and mismatched channels are rejected.
- [ ] Provider fee above limit is rejected.
- [ ] Routing fee above limit is rejected.
- [ ] Duplicate execution is rejected.
- [ ] Wrong inbound channel is rejected before claim when atomic mode is enabled.

### Deterministic integration tests

- [ ] Build a local test topology with at least three local channels.
- [ ] Make an unselected route cheaper and more attractive than the selected route.
- [ ] Prove the selected source is still the only first hop.
- [ ] Prove insufficient selected capacity fails rather than splitting across other local channels.
- [ ] Prove downstream MPP remains allowed while the first hop stays pinned.
- [ ] Interrupt and restart during a pending attempt; prove no unrestricted retry occurs.
- [ ] Deliver an inbound test payment through the wrong channel; prove it is not claimed.
- [ ] Run Go tests, frontend lint/type checks, production build, and desktop build checks.

### Live rollout gates

- [ ] Source is committed and pushed to owner-controlled repositories.
- [ ] Exact installed build commit is recorded.
- [ ] UI controls and backend enforcement are independently verified.
- [ ] No live-value test occurs without a fresh approval packet.
- [ ] Start with the smallest useful explicitly approved test amount.
- [ ] Independently reconcile actual routes, fees, balances, and persisted operation state.
- [ ] Keep the feature disabled if any route evidence is unavailable or ambiguous.

## Upgrade procedure

1. Fetch the new official Alby release into `upstream`.
2. Create an upgrade branch from the new exact tag.
3. Replay the metadata customization commit.
4. Replay the pinned-rebalance commits and the exact compatible LDK dependency commit.
5. Resolve conflicts without weakening any safety invariant.
6. Run the complete deterministic test matrix.
7. Build and record the source commit.
8. Install only after review; perform no live rebalance without separate approval.

## Decision log

### 2026-10-05 — No UI-only routing control

Decision: routing selection must be enforced in the Lightning routing layer. UI state, channel names, route preference, and peer disconnection are not accepted as enforcement.

### 2026-10-05 — Destination enforcement is gated on atomicity

Decision: observing the incoming channel after an automatically claimed payment is audit evidence, not prevention. The feature remains non-value-moving until wrong-channel rejection is shown not to create principal-loss risk.

### 2026-10-05 — Preserve customizations as replayable commits

Decision: metadata customization, pinned-payment dependency work, Alby backend work, and UI work remain separate commits so future upstream upgrades can replay them independently.

## Learning log

Append dated entries here. Each entry must distinguish observed evidence from inference and record any resulting checklist change.

### 2026-10-05 — Initial source audit

Observed:

- Stock Alby pays the provider invoice immediately after order creation.
- Stock Alby passes no outgoing-channel selector to `SendPaymentSync`.
- The LDK `RouteParametersConfig` exposed to Alby contains no first-hop field.
- The rebalance service receives the requested incoming peer pubkey, but local LDK invoice creation does not enforce it.
- Lower-level LDK exposes both restricted `first_hops` route finding and receiving channel IDs.

Consequence:

- The implementation requires a versioned LDK dependency change in addition to Alby Hub changes.
- Incoming enforcement remains blocked until an atomic provider or self-routed design is proven.

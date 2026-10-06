# Pinned Channel Rebalance Reference Plan

Status: implementation in progress; quote review is enabled in source, payment execution is hard-locked. No live rebalance is authorized by this plan.

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

## Current upgrade baseline

- [x] Official Alby Hub release `v1.24.1` is tag commit `7b4488571c5717939d17b199cdd69e93c2a671e8`.
- [x] Official `v1.24.1` updates `github.com/getAlby/ldk-node-go` from `af22e238c194` to `5ba434093284`; the older custom binding must not replace that security-update baseline.
- [x] Custom Rust routing enforcement is replayed on `getAlby/ldk-node` commit `74daaf9` as owner-fork commit `057b73d`.
- [x] Generated Go bindings and a release-mode arm64/x86_64 macOS library are replayed on `getAlby/ldk-node-go` commit `5ba434093284` as owner-fork commit `e7dd77f`.
- [x] Hub resolves the upgraded owner-fork binding as pseudo-version `github.com/hudhaifahz/ldk-node-go v0.0.0-20261006232835-e7dd77fda90e`.
- [x] The obsolete `v1.24.0` frontend lockfile snapshot was intentionally skipped so it cannot overwrite `v1.24.1` dependency/security updates.

## Version-control checklist

- [x] Preserve the existing customization as its own reviewed commit before routing work.
- [x] Exclude unrelated dependency-lock drift unless it is proven necessary and intentionally committed.
- [x] Add an owner-controlled GitHub fork or repository remote without overwriting upstream history.
- [x] Keep official Alby as `upstream` and the owner-controlled repository as `origin` or `fork`.
- [x] Push the preserved metadata branch and configure its upstream tracking branch.
- [x] Create a separate feature branch for pinned rebalancing.
- [x] Keep the reference plan updated and committed with each material design change.
- [x] Record the exact Alby base commit and exact LDK dependency commit in the implementation notes.
- [x] Keep LDK dependency changes in a separately versioned fork/commit; do not depend on an uncommitted module-cache edit.
- [x] Use small Conventional Commits so future Alby upgrades can replay or drop each customization independently.
- [ ] Tag or otherwise record the exact source commit used for any installed desktop build.

## Required design

### 1. Explicit request model

- [x] Add exact outgoing and incoming channel identifiers to quote requests.
- [x] Include both full counterparty pubkeys as identity cross-checks.
- [x] Use millisatoshis internally; reject rounding and overflow.
- [x] Require separate maximum provider-fee and routing-fee values.
- [x] Reject identical source and destination channels.
- [x] Reject multiple local channels to the same selected peer unless the exact channel can be enforced.

### 2. Quote-only stage

- [x] Create a provider order and decode its invoice without paying it.
- [x] Validate network, amount, provider-fee ceiling, expiry, and payment identity.
- [ ] Snapshot selected channel identities, usability, capacities, reserves, and balances. (Identity, usability, spendable, and receivable values are captured; explicit reserve fields remain.)
- [ ] Calculate the maximum possible source debit and expected post-action ranges. (Maximum debit is implemented; explicit post-action ranges remain.)
- [x] Persist a single-use quote with an expiry and a hash of all material parameters.
- [x] Return a human-readable review packet; do not initiate payment.

### 3. Fail-closed execution stage

- [x] Accept only an unexpired, unused quote identifier.
- [ ] Re-read channel identity, state, capacity, and fee limits immediately before sending.
- [ ] Reject any material difference rather than silently refreshing the quote.
- [ ] Atomically mark the quote executing before calling the Lightning backend.
- [ ] Prevent duplicate execution and concurrent rebalance operations.
- [x] Persist provider order ID, both invoices/payment hashes, exact limits, and selected channel IDs.

### 4. Outgoing first-hop enforcement

- [x] Add a pinned-payment capability to the LDK dependency and its Go binding.
- [x] Supply only the selected `ChannelDetails` entry as `first_hops` for route finding.
- [x] Permit downstream MPP when useful while requiring every part to share the selected first hop.
- [x] Disable automatic retries and submit the already-verified fixed route, eliminating fallback route-finding.
- [ ] Persist the constraint with the payment ID before the payment can start.
- [x] On missing constraint, unavailable channel, or route failure: fail with no fallback.
- [ ] Return or publish actual successful-path evidence including each first-hop channel ID.
- [ ] Do not implement routing control by disconnecting peers, disabling channels, manipulating fees, or temporarily hiding channels.

### 5. Incoming-channel enforcement and atomicity gate

- [x] Expose `receiving_channel_ids` through the LDK language binding and Alby event model.
- [ ] Determine whether the rebalance provider supports a compatible hold/claim or shared-preimage atomic flow.
- [ ] Prove that rejecting a misrouted inbound HTLC cannot leave the provider invoice settled without the principal returning.
- [ ] If atomicity is proven, claim only when every incoming MPP part uses the selected channel.
- [ ] If any part arrives elsewhere, fail the inbound payment and record the reason.
- [ ] If atomicity cannot be proven, do not label destination routing as enforced and do not enable the value-moving feature.

### 6. Owner-facing UI

- [ ] Show exact outgoing and incoming peer names, full pubkeys, and stable channel identifiers.
- [ ] Show channel status, spending/receiving capacity, reserve, and projected post-action ranges.
- [ ] Show principal, provider fee, routing-fee cap, and maximum total debit separately.
- [x] Make quote creation visibly non-paying.
- [ ] Require a distinct final confirmation for the exact, still-valid quote.
- [x] Disable execution when safety preconditions are not met.
- [x] Never offer or imply an automatic fallback route.
- [ ] Show a final evidence record, not only a success toast.

## Verification checklist

### Unit and interface tests

- [ ] Exact selected channel is passed to the LDK routing layer.
- [ ] A cheaper alternative first hop is never used.
- [ ] All MPP parts share the selected first hop.
- [ ] Retry remains pinned.
- [ ] Restart recovery remains pinned or fails closed.
- [ ] Offline, unusable, insufficient, stale, expired, ambiguous, and mismatched channels are rejected. (All except execution-time stale re-read have source checks; execution is currently locked.)
- [x] Provider fee above limit is rejected.
- [x] Routing fee above limit is rejected by the pinned LDK route parameters and post-success assertion.
- [ ] Duplicate execution is rejected. (No execution is currently permitted.)
- [ ] Wrong inbound channel is rejected before claim when atomic mode is enabled.

### Deterministic integration tests

- [ ] Build a local test topology with at least three local channels.
- [ ] Make an unselected route cheaper and more attractive than the selected route.
- [ ] Prove the selected source is still the only first hop.
- [ ] Prove insufficient selected capacity fails rather than splitting across other local channels.
- [ ] Prove downstream MPP remains allowed while the first hop stays pinned.
- [ ] Interrupt and restart during a pending attempt; prove no unrestricted retry occurs.
- [ ] Deliver an inbound test payment through the wrong channel; prove it is not claimed.
- [x] Run Go tests, frontend lint/type checks, production build, and desktop build checks.

### Live rollout gates

- [x] Source is committed and pushed to owner-controlled repositories.
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
4. If the official release changes its LDK binding commit, first replay the custom Rust and generated-binding commits on the new official LDK baselines; never point the upgraded Hub back to an older custom binary.
5. Replay the pinned-rebalance Hub commits and pin the exact upgraded owner-fork binding pseudo-version.
6. Resolve conflicts without weakening any safety invariant. Drop obsolete lockfile snapshots instead of overwriting the new release lockfiles.
7. Run the complete deterministic test matrix.
8. Build and record the source commit.
9. Install only after review; perform no live rebalance without separate approval.

## Decision log

### 2026-10-05 — No UI-only routing control

Decision: routing selection must be enforced in the Lightning routing layer. UI state, channel names, route preference, and peer disconnection are not accepted as enforcement.

### 2026-10-05 — Destination enforcement is gated on atomicity

Decision: observing the incoming channel after an automatically claimed payment is audit evidence, not prevention. The feature remains non-value-moving until wrong-channel rejection is shown not to create principal-loss risk.

### 2026-10-05 — Preserve customizations as replayable commits

Decision: metadata customization, pinned-payment dependency work, Alby backend work, and UI work remain separate commits so future upstream upgrades can replay them independently.

### 2026-10-05 — Fixed route with no automatic retry

Decision: the pinned LDK payment computes a route from a singleton first-hop list, verifies every MPP path uses that channel, and submits that fixed route. Automatic LDK route-finding retries are disabled. A failed attempt requires a new quote and explicit owner review; it cannot select another first hop.

### 2026-10-05 — Quote enabled, execution locked

Decision: the source may create and persist a non-paying provider quote, but both the legacy one-step API and the new execute endpoint fail closed. The UI displays the exact action packet and a locked execution control until incoming-channel atomicity is independently proven.

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

### 2026-10-05 — Versioned dependency implementation

Observed:

- `getAlby/ldk-node` base `e10861c` is preserved on local branch `feat/pinned-first-hop`; implementation commit is `fee9e17`.
- `getAlby/ldk-node-go` base `af22e238c194` is preserved on local branch `feat/pinned-first-hop`; generated-binding commit is `5996237` and matching stripped universal macOS library commit is `75a041e6b777`.
- The Rust layer compiles with UniFFI enabled, and the receiving-channel persistence test passes with `RUSTFLAGS='--cfg no_download'`.
- The generated Go binding compiles. Alby LDK tests pass when using the matching local binding and native library.
- Non-macOS generated libraries have not been rebuilt; CI generation is a release gate.

Consequence:

- A source-only binding change is insufficient. Every shipped platform library must be generated from the exact Rust commit before a release is eligible.
- The committed Hub `go.mod` replaces `github.com/getAlby/ldk-node-go` with exact owner-fork pseudo-version `github.com/hudhaifahz/ldk-node-go v0.0.0-20261006064620-75a041e6b777`; the local `go.work` is no longer required to reproduce the focused Hub tests.

### 2026-10-05 — Owner forks created; push credentials mismatched

Observed:

- Owner forks exist at `aya-skaur/hub`, `aya-skaur/ldk-node`, and `aya-skaur/ldk-node-go`.
- Each local repository now keeps official `getAlby` as `upstream` and the owner fork as `origin`.
- Browser authentication belongs to `aya-skaur`, while the terminal HTTPS credential belongs to `hudhaifahz`; GitHub rejected all three pushes with HTTP 403. No permission or credential workaround was attempted.

Consequence:

- Local commits are durable and replayable, but the remote-push checklist remains open until the owner chooses the GitHub identity or grants the intended access.

### 2026-10-06 — Repositories consolidated under `hudhaifahz`

Observed:

- Authenticated terminal access was verified as GitHub account `hudhaifahz` without exposing the credential.
- Proper forks now exist at `hudhaifahz/hub`, `hudhaifahz/ldk-node`, and `hudhaifahz/ldk-node-go`.
- Each local repository uses the `hudhaifahz` fork as `origin`, retains official `getAlby` as `upstream`, and retains the former `aya-skaur` fork as non-primary remote `aya`.
- Hub branches `fix/frontier-crown-nwc-metadata` and `feat/pinned-channel-rebalance`, plus both dependency branches named `feat/pinned-first-hop`, were pushed and their remote hashes checked.
- The initial universal macOS library was 165,474,432 bytes and exceeded GitHub's 100 MB hard file limit. Stripping non-runtime symbols reduced the same arm64/x86_64 library to 96,563,872 bytes. The unpublished artifact-only commit was amended to `75a041e6b777`; `go test ./...` still passes in the binding repository.
- With `GOWORK=off`, focused Hub tests pass against the exact `hudhaifahz/ldk-node-go` pseudo-version. The only output beyond passing tests is the pre-existing Bark macOS deployment-target warning.

Consequence:

- The three-repository implementation is remotely durable and a fresh Hub checkout can resolve the matching custom Go/native binding without relying on this workstation's ignored `go.work`.
- The former `aya-skaur` forks were not deleted; they are retained only as recoverable backup remotes. Official Alby remains the upgrade source through `upstream`.

### 2026-10-06 — `v1.24.1` security-update rehearsal

Observed:

- Official release `v1.24.1` is based on Hub commit `7b4488571c5717939d17b199cdd69e93c2a671e8` and includes security and dependency updates.
- Metadata override and icon commits replay cleanly. The old lockfile snapshot conflicts with and is superseded by the official `v1.24.1` lockfile, so that snapshot was skipped.
- `v1.24.1` advances the official Go/LDK binary from `af22e238c194` to `5ba434093284`; retaining the previous owner-fork pseudo-version would silently discard the updated native binaries.
- Custom LDK routing enforcement compiles on Rust base `74daaf9` as commit `057b73d`. The receiving-channel persistence test passes against the updated dependency lockfile.
- Custom Go binding commit `baa90a2` and release-mode universal macOS library commit `e7dd77f` are based on official binding commit `5ba434093284`. The library is 50,621,744 bytes, contains arm64 and x86_64 slices, and exports the pinned first-hop symbol on both architectures.
- With `GOWORK=off`, focused Hub LDK, API, database, and HTTP tests pass against exact pseudo-version `v0.0.0-20261006232835-e7dd77fda90e`.
- Frontend ESLint, TypeScript, and production HTTP build pass. The existing Lottie `eval`, large-chunk, stale Browserslist, and Bark macOS deployment-target warnings remain warnings rather than test failures.
- A universal Wails desktop bundle builds in staging, embeds the upgraded library, contains only the portable `@executable_path/../Frameworks` LDK runtime path, and passes deep signature verification after ad-hoc signing.
- The installed app was not replaced or restarted. Its process and port `21420` were present, but a five-second HTTP root probe timed out, so live continuity was not proven.

Consequence:

- The source upgrade rehearsal passes, but live installation remains a separate owner-controlled quit, replace, restart, unlock, and continuity checkpoint.
- No rebalance execution gate is opened by this upgrade. Incoming-channel atomicity and the remaining deterministic routing tests are still required before any value-moving trial.

### 2026-10-05 — Hub quote and review implementation

Observed:

- Hub commit `9aebc8e1` adds the optional pinned-payment capability and exact LDK call.
- Hub commit `8e672e4c` carries exact receiving channel IDs into the Hub hold-invoice event model.
- Hub commit `fe3f92cf` adds persisted five-minute quotes, exact channel-plus-pubkey validation, separated fee limits, legacy-flow rejection, and a hard-locked execute endpoint.
- Hub commit `47782062` adds the exact-channel quote/review UI and visibly locked execution state.
- Focused LDK, API, database, frontend lint, and TypeScript checks pass.
- The HTTP frontend production build passes; it reports existing dependency warnings for Lottie `eval`, large chunks, and stale Browserslist data.
- Hub commit `7ab500d7` adds the missing NWC metadata config expectation; the broad HTTP package test now passes.

Consequence:

- The current source can be reviewed and can produce a non-paying quote, but it is intentionally not a value-moving feature.
- The focused backend and broad HTTP package tests are clean; the complete repository test matrix remains a release gate.

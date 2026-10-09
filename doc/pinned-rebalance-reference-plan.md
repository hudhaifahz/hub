# Pinned Channel Rebalance Reference Plan

Status: source implementation and deterministic verification are complete through owner confirmation, exact-route submission, restart recovery, and two-leg terminal reconciliation. The exact build from Hub commit `defe4b32c68044c35318e979dbea6e740d9a39cd` is installed and running with migration, wallet continuity, embedded-library, and UI safety checks recorded below. No live rebalance is authorized by this plan.

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

- [x] Channel labels are display-only. Bind decisions to stable channel identity plus full counterparty pubkey.
- [x] Every outgoing MPP part uses the selected local channel as its first hop. (A real two-part circular settlement uses the same selected first-hop SCID for both successful paths.)
- [x] Retries remain pinned. The same deterministic operation and outbound payment IDs can recover only the same stored fixed route; no pathfinding fallback is available.
- [x] Restart recovery remains pinned or fails closed; it must never resume as an unrestricted payment. (A real fixed-route HTLC survives ChannelManager/ChannelMonitor reload, retransmits only its already-committed selected route, and rejects a second executor send.)
- [x] Selected-channel offline, unusable, ambiguous, or insufficient states stop execution.
- [x] A changed quote stops execution. An expired unacquired quote cannot execute; an already-acquired expired quote can only recover a preparation that LDK previously persisted and cannot create a new invoice.
- [x] Provider fees and routing fees have separate exact limits. (The provider flow remains locked; the enabled local circular path has no provider fee and has a distinct exact routing-fee cap.)
- [x] No unknown or unbounded fee is payable.
- [x] Incoming channel enforcement is not claimed until it is verified before settlement or the provider protocol is proven atomic. (The local self-payment rejects a wrong or mixed incoming channel before claim; the provider flow remains locked.)
- [x] Only one rebalance operation may execute at a time. (A database-level partial unique index permits at most one local quote in `executing`, including races between different valid quote IDs.)
- [ ] A live rebalance always requires a fresh, action-specific owner approval packet.
- [x] Tests, builds, deployment, UI presence, live execution, and economic success are reported as separate states.

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
- [x] Crash-safe prepared-send recovery is versioned as owner-fork LDK commit `77c0454`. Idempotent preparation recovery and recovery-only expired preparation are completed by commits `104945e`, `5ad80b0`, and `4096ee5`. Two-leg circular failure persistence and terminal readback are completed by LDK commit `16630be`. The exact final binary is owner-fork Go binding commit `a3dd5b1`, pinned by the Hub as pseudo-version `github.com/hudhaifahz/ldk-node-go v0.0.0-20261008205838-a3dd5b1fe3bb`; its universal macOS library is `43,195,848` bytes with SHA-256 `3399f9446240441f406b7a307b6845415bbf5ea1282b3cebf9f5a689826a7554`.
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
- [x] Tag or otherwise record the exact source commit used for any installed desktop build.

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
- [x] Snapshot selected channel identities, usability, capacities, reserves, and balances. (Both channels' local/remote balances and local/counterparty reserves are fingerprint-bound.)
- [x] Calculate the maximum possible source debit and expected post-action ranges. (The review shows the fixed-route debit and projected outgoing/incoming local, remote, and spendable balances.)
- [x] Persist a single-use quote with an expiry and a hash of all material parameters.
- [x] Return a human-readable review packet; do not initiate payment.

### 2A. Local exact-route quote record

- [x] Persist each successful local route quote as a separate five-minute review record.
- [x] Give the record a random quote ID and a deterministic SHA-256 route fingerprint.
- [x] Bind the fingerprint to exact channel IDs, peer pubkeys, first/final SCIDs, every path and hop, principal, quoted fee, fee cap, maximum debit, and channel balance snapshots.
- [x] Store the complete reviewed path as JSON and verify the record survives supported database migration.
- [x] Persist the complete encoded LDK route separately from the UI path, bind it into the route fingerprint, reject missing or altered route bytes, and never expose those opaque bytes in the review response.
- [x] Keep invoice, payment hash, preimage, probe, HTLC, and payment authorization material out of the quote record.
- [x] Display quote ID, fingerprint, quote time, expiry, exact channels, route, fee, balance/reserve snapshot, and projections in the review UI.
- [x] Require the executor to reload this record, reject expiry or fingerprint mismatch, and consume it at most once. (The Hub acquisition primitive re-fingerprints persisted route material and uses an atomic `quoted` to `executing` compare-and-set.)

### 3. Fail-closed execution stage

- [x] Accept only an unexpired, unused quote identifier.
- [x] Re-read channel identity, state, capacity, and fee limits immediately before sending. (The fixed-route executor performs the execution-time checks after owner confirmation.)
- [x] Reject any material difference rather than silently refreshing the quote. (The executor rejects channel, SCID, peer, capacity, fee, amount, invoice, expiry, payment identity, snapshot, and serialized-route mismatches.)
- [x] Atomically mark the quote executing before calling the Lightning backend. (The owner-facing orchestration assigns a deterministic operation ID and persists `acquired`, `prepared`, and `submitted` phases.)
- [x] Prevent duplicate execution and concurrent rebalance operations. (Hub retries reuse the exact operation, payment hash, and outbound payment ID; LDK returns an already tracked payment or retries only the same persisted fixed route and ID. A duplicate LDK result is proof of the existing operation, not permission to create another.)
- [x] Reconcile durable terminal Lightning payment-store records into one final Hub success or failure state before releasing the single-operation execution slot. (HTTP/Wails status calls are read/reconcile-only and the UI polls only after exact submission.)
- [x] Persist provider order ID, both invoices/payment hashes, exact limits, and selected channel IDs.

### 3A. Terminal reconciliation proof and acceptance criteria

The durable LDK payment store, not event delivery or UI state, is the terminal authority. Production handling writes `PaymentClaimed`, `PaymentSent`, and `PaymentFailed` payment-store updates before adding the corresponding user event to the durable event queue. A restarted Hub can therefore reconcile from records even if notification delivery was interrupted.

The reconciler must reload both records by their independently persisted IDs and accept them only when all immutable bindings agree: operation ID, inbound payment hash, deterministic outbound payment ID, amount, exact outgoing and incoming user-channel IDs, routing-fee limit, BOLT 11 type, and opposite inbound/outbound directions. Success additionally requires both records to be `Succeeded`, the same nonempty preimage, and an actual outbound fee equal to the fixed reviewed route fee and no greater than its cap.

Allowed status matrix:

| Inbound | Outbound | Hub result |
| --- | --- | --- |
| Pending | missing | remain `executing`; preparation exists but no durable send record is proven |
| Pending | Pending | remain `executing`; settlement is unresolved |
| Succeeded or Failed | Pending | remain `executing`; the other durable leg has not reached terminal state |
| Pending | Succeeded or Failed | remain `executing`; the other durable leg has not reached terminal state |
| Succeeded | Succeeded | atomically record `succeeded` with exact amount, fee, identifiers, and terminal timestamp |
| Failed | Failed | atomically record `failed` with a bounded reason and terminal timestamp |
| Succeeded | Failed | fail closed as contradictory; retain the active lock for investigation |
| Failed | Succeeded | fail closed as contradictory; retain the active lock for investigation |

The LDK failure handler must make a circular failure two-legged before Hub may accept it: after validating the outbound record's circular metadata, persist the matching inbound record as `Failed` in the same replayable event-handling turn. A persistence error causes event replay rather than partial terminal acceptance.

Hub terminal writes must compare-and-set only the exact `executing` operation in `prepared` or `submitted`. Replaying identical evidence returns the existing terminal row unchanged. Changed or contradictory evidence is rejected. The database's single-active-operation index is released only by a successfully persisted terminal transition.

### 4. Outgoing first-hop enforcement

- [x] Add a pinned-payment capability to the LDK dependency and its Go binding.
- [x] Supply only the selected `ChannelDetails` entry as `first_hops` for route finding.
- [x] Permit downstream MPP when useful while requiring every part to share the selected first hop.
- [x] Disable automatic retries and submit the already-verified fixed route, eliminating fallback route-finding.
- [x] Persist the constraint with the payment ID before the payment can start. (The executor persists the distinct outbound payment record before invoking the fixed-route send action.)
- [x] On missing constraint, unavailable channel, or route failure: fail with no fallback.
- [ ] Return or publish actual live successful-path evidence including each first-hop channel ID. (Deterministic real-HTLC tests prove every tested path's first/final SCIDs, but no live operation has occurred.)
- [x] Do not implement routing control by disconnecting peers, disabling channels, manipulating fees, or temporarily hiding channels.

### 5. Incoming-channel enforcement and atomicity gate

- [x] Expose `receiving_channel_ids` through the LDK language binding and Alby event model.
- [x] Add a backward-compatible, persisted exact receiving-channel constraint to an inbound BOLT 11 payment record.
- [x] Add a fail-closed claim decision that requires pending state, exact amount, a persisted preimage, and a nonempty set of MPP parts all reporting the exact selected channel.
- [x] Add a non-sending preparation primitive that validates both exact local channels, persists the inbound invoice/preimage plus operation binding and fee limit, and reserves a distinct outbound payment ID.
- [x] Determine whether the rebalance provider supports a compatible hold/claim or shared-preimage atomic flow. (It was not adopted; the provider execution path stays disabled, and the local circular self-payment removes the external-provider settlement gap.)
- [x] Prove that rejecting a misrouted inbound HTLC cannot leave the provider invoice settled without the principal returning. (For the selected local architecture there is no provider invoice; deterministic wrong-channel and mixed-MPP tests fail the same circular outbound payment before preimage disclosure.)
- [x] If atomicity is proven, claim only when every incoming MPP part uses the selected channel.
- [x] If any part arrives elsewhere, fail the inbound payment and record the reason.
- [x] If atomicity cannot be proven, do not label destination routing as enforced and do not enable the value-moving feature. (Only the proven local circular architecture is enabled in source.)

### 6. Owner-facing UI

- [x] Show exact outgoing and incoming peer names, full pubkeys, and stable channel identifiers.
- [x] Show channel status, spending/receiving capacity, reserves, and projected post-action ranges.
- [x] Show principal, quoted routing fee, routing-fee cap, and maximum total debit separately. (The local path has no provider fee.)
- [x] Make quote creation visibly non-paying.
- [x] Require a distinct final confirmation for the exact, still-valid quote.
- [x] Disable execution when safety preconditions are not met.
- [x] Never offer or imply an automatic fallback route.
- [x] Show a final evidence record, not only a success toast.

## Verification checklist

### Unit and interface tests

- [x] Exact selected channel is passed to the LDK routing layer. (A real three-node circular HTLC test asserts the successful path's first and final SCIDs.)
- [ ] A cheaper alternative first hop is never used.
- [x] All reviewed MPP paths are required to share the selected first hop before submission, and a real executor test settles two downstream branches through that same local first hop.
- [x] Retry remains pinned. (The executor submits the preserved fixed route exactly once and performs no automatic route retry.)
- [x] Restart recovery remains pinned or fails closed. (The reloaded in-flight payment is recognized by deterministic ID, hash, amount, invoice, exact channel constraints, and ChannelManager state; it returns the existing ID without route finding. A persisted-but-untracked send may resubmit only the identical serialized route and payment ID, and LDK rejects any duplicate HTLC.)
- [x] Offline, unusable, insufficient, stale, expired, ambiguous, and mismatched channels are rejected by quote and execution-time checks.
- [x] Provider fee above limit is rejected.
- [x] Routing fee above limit is rejected by the pinned LDK route parameters and post-success assertion.
- [x] Duplicate execution is idempotent without creating a second HTLC. Matching pending or succeeded state returns the deterministic outbound ID; mismatched hash, amount, channel binding, invoice, fee cap, contradictory status, abandoned state, or unknown payment type fails closed.
- [x] Hub preparation and submission recovery persist deterministic operation, hash, and outbound IDs across every Hub-side phase. Retrying `submitted` makes no backend call; retrying after simulated post-send/pre-database uncertainty reuses the same IDs; expired acquisition passes an explicit recovery-only expiry of zero; missing prior preparation stops before send.
- [x] Terminal reconciliation reloads and validates both exact payment records, keeps missing/pending/transient combinations locked, rejects contradictory or tampered evidence, persists terminal evidence by compare-and-set, replays idempotently without another backend read, and releases the single-operation slot only after a matching terminal transition.
- [x] Wrong inbound channel is rejected before claim by the exact-channel decision and real HTLC tests. (The owner-facing source orchestration is not installed or live-authorized.)
- [x] The channel-constrained claim decision rejects mixed MPP, unidentified-channel, empty-part, missing-preimage, missing-amount, underpayment, overpayment, and previously-failed cases.
- [x] The production claim/fail action adapter is shared with a controlled test proving claim, fail, and unconstrained decisions dispatch exactly one, one, and zero ChannelManager actions respectively.
- [x] A rejected constrained claim updates the inbound payment record to `Failed`, and a fresh payment-store read from persisted bytes returns that failed state.

### Deterministic integration tests

- [ ] Build a local test topology with at least three local channels.
- [ ] Make an unselected route cheaper and more attractive than the selected route.
- [x] Prove the selected source is still the only first hop for a real single-path circular settlement.
- [x] Prove insufficient selected capacity fails before the executor invokes its send action.
- [x] Prove downstream MPP remains allowed while the first hop stays pinned. (A five-node, two-part circular payment branches only after the selected source peer, reconverges before the selected destination peer, and reports the same exact first and final SCIDs for both successful paths.)
- [x] Interrupt and restart during a pending attempt; prove no unrestricted retry occurs. (The node is reloaded after the fixed-route sender durably commits the outbound HTLC but before the first peer receives it; recovery identifies that exact in-flight payment, a direct same-ID resend returns `DuplicatePayment` without adding a monitor or HTLC, channel reestablishment retransmits the original HTLC, and settlement uses the selected final channel.)
- [x] Restart a genuinely prepared but unsent circular operation; prove the exact pending record, preimage, channel identities, and ChannelManager state survive, no outbound payment appears, channels remain unusable until reconnection, and operation reuse is still rejected.
- [x] Deliver a real in-memory HTLC with a mismatched required incoming channel; prove it is failed backward and the sender observes failure.
- [x] Deliver a second real in-memory HTLC with the exact reported incoming channel; prove it is claimed and the sender observes success.
- [x] Split one real in-memory MPP across two distinct incoming channels; prove the mixed channel set is rejected before preimage disclosure, both HTLCs fail backward, and the sender observes terminal `RecipientRejected`.
- [x] Run Go tests, frontend lint/type checks, production build, and desktop build checks.

### Live rollout gates

- [x] Source is committed and pushed to owner-controlled repositories.
- [x] Exact installed build commit is recorded.
- [x] UI controls and deterministic backend enforcement are independently verified. Live route availability and value-moving acceptance remain separate gates.
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

### 2026-10-08 — Terminal truth comes from two durable payment records

Decision: neither a single Lightning event nor one successful leg can close a circular operation. Hub may record success only after exact-bound, persisted inbound and outbound records both report success with the same preimage and exact reviewed fee. It may record failure only after both records report failure. Pending, missing, or contradictory combinations retain the active execution lock. Terminal writes are idempotent compare-and-set transitions, and event delivery is diagnostic rather than authoritative.

### 2026-10-08 — Owner execution is action-specific and recoverable

Decision: the local circular route may become value-moving only after the owner reviews a fresh fingerprint-bound packet and types the quote-specific confirmation. The execute call may prepare and submit exactly that deterministic operation once. Status polling never sends; an interrupted `acquired` or `prepared` phase can advance only through the same confirmation, operation ID, payment IDs, and serialized route. A terminal replay returns stored evidence without another backend call.

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
- The exact staged build source is Hub commit `4129cc80e1cdc3eb3d7130ed21831b1b498c0270`. The archived bundle is `/Users/kode/Development/albyhub-builds/v1.24.1/Alby-Hub-v1.24.1-custom-4129cc80.zip` with SHA-256 `a8b1bd150e989afe3791f5a5b44ce88702c83d92139677c3ec89fb23b1a068c3`.
- The installed app was not replaced or restarted. Its process and port `21420` were present, but root, `/api/health`, and `/api/node/status` probes all timed out. The running process had loaded the old `af22e238c194` LDK library from the local Go module cache, so live continuity and portable dependency loading were not proven.

Consequence:

- The source upgrade rehearsal passes, but live installation remains a separate owner-controlled quit, replace, restart, unlock, and continuity checkpoint.
- No rebalance execution gate is opened by this upgrade. Incoming-channel atomicity and the remaining deterministic routing tests are still required before any value-moving trial.

### 2026-10-06 — `v1.24.1` installed with rollback and live continuity proof

Observed:

- Before replacement, the running `v1.24.0` desktop process was unresponsive to both the macOS quit request and `SIGTERM`. Transaction-safe SQLite backups and a copy of the installed application were completed first; only that exact process was then force-stopped.
- The recoverable pre-upgrade application and clean-stop data copy are stored at `/Users/kode/Development/albyhub-rollbacks/2026-10-06-before-v1.24.1`. Both live databases and both transaction-safe pre-quit snapshots returned `integrity_check = ok` before installation.
- `/Applications/Alby Hub.app` now contains the bundle built from Hub source commit `4129cc80e1cdc3eb3d7130ed21831b1b498c0270`. Its executable SHA-256 is `9a1ddd7d49e912fd05f54e81656ad181cf1c8e062dc9f390552bf31c770d1afb`; its embedded signed LDK library SHA-256 is `6e12519aa3ebdc88399990f6f46aa1054b636d48f78d19e7978ff2f9a1e4e21d`.
- Deep signature verification passes after installation. The executable is universal arm64/x86_64, records `version.Tag=v1.24.1`, and resolves LDK only through `@executable_path/../Frameworks`.
- The restarted process runs from `/Applications/Alby Hub.app/Contents/MacOS/Alby Hub`, listens on port `21420`, and has `/Applications/Alby Hub.app/Contents/Frameworks/libldk_node.dylib` loaded rather than a Go module-cache library.
- The live Settings screen reports `v1.24.1`. The Node screen reports all three channels online and shows Lightning balance `1,606,117 sats`, receive limit `2,321,902 sats`, and on-chain balance `41,883 sats` at the verification checkpoint.
- The live Kraken channel menu exposes `Rebalance In`. Its dialog exposes exact outgoing-channel selection, the exact return peer and channel ID, separate provider and routing fee caps, `Create non-paying quote`, and a disabled `Execute locked` control with the incoming-channel atomicity warning.
- No quote was created and no payment was attempted. Post-start checks still return `ok` for both databases, and Hub continuity counts remain `apps=12`, `app_permissions=94`, and `user_configs=16`.
- Plain HTTP and HTTPS probes to root, `/api/health`, and `/api/node/status` still do not return ordinary HTTP responses on the desktop listener. This behavior was present before and after the upgrade; live UI state, process/library inspection, channel status, and direct database checks are the accepted evidence for this installation checkpoint, not those endpoints.

Consequence:

- The customized `v1.24.1` desktop bundle is installed, running, unlocked, and reading the existing wallet state with a complete local rollback path.
- Live UI continuity is proven, but the backend value-moving route is still intentionally locked. The independent backend-enforcement checklist item remains open until incoming-channel atomicity and the remaining deterministic routing tests pass.
- This installation does not authorize or execute a rebalance. Any future value-moving trial still requires a fresh exact-channel approval packet.

### 2026-10-06 — Desktop quote-router failure found and corrected

Observed:

- A live non-paying quote attempt selected outgoing peer `030ef18b788bdfaf899071bb975f258306f83eae0a83d9e52aee93ae894296a42c`, outgoing channel `255349693497905514942651862984333935477`, incoming Kraken peer `02437c00ef5de2686a6bd60f8acb5c83d17010916010a15f479d5ef84c04f04485`, incoming channel `86157859664272214382561858939519142638`, and principal `500,000 sats` with provider-fee cap `2,500 sats` and routing-fee cap `1,000 sats`.
- The installed desktop returned `Unhandled route: POST /api/channels/rebalance/quote`. The HTTP router contained both new endpoints, but the separate Wails desktop request router contained only the legacy one-step route. No provider order, invoice, quote row, or payment was created.
- Hub commit `27321e2ae9a43722cfd56b8803891a32d0e74e87` adds the quote and execute routes to the Wails router. Regression tests prove quote dispatch and locked-execution error propagation. Wails, API, and HTTP test packages pass; only the pre-existing Bark deployment-target warnings remain.
- The replacement universal bundle preserves `version.Tag=v1.24.1`, exact owner-fork binding pseudo-version `v0.0.0-20261006232835-e7dd77fda90e`, and embedded LDK library SHA-256 `6e12519aa3ebdc88399990f6f46aa1054b636d48f78d19e7978ff2f9a1e4e21d`. The installed executable SHA-256 is `9606ef5781d9707e3df348fccb4a2799677f4e468a8f7e530d156f242a252779`.
- The first ad-hoc re-sign omitted Alby's sandbox entitlement and therefore opened the separate non-wallet path `/Users/kode/Library/Application Support/albyhub`. No onboarding or wallet action was intentionally completed. That process was stopped, and the invalid app and archive were retained only in the rollback folder.
- The corrected bundle is signed with Alby's sandbox, download/user-selected file, and client/server network entitlements. It opens the canonical container database at `/Users/kode/Library/Containers/com.getalby.Alby-Hub/Data/Library/Application Support/albyhub/nwc.db`; integrity is `ok` and continuity remains `apps=12`, `app_permissions=94`, and `user_configs=16`.
- The corrected archived bundle is `/Users/kode/Development/albyhub-builds/v1.24.1/Alby-Hub-v1.24.1-custom-27321e2a.zip` with SHA-256 `e602993f3cb331edfee9ab8c93d6b0acb1184646f42a3640eccc20480a0ae5dc`. Its rollback is `/Users/kode/Development/albyhub-rollbacks/2026-10-06-before-wails-route-fix-27321e2a`.
- The corrected app is installed and stopped at the owner password screen after a clean restart. A fresh live quote retest is pending owner unlock. No sats moved.

Consequence:

- Desktop-specific route coverage is a required upgrade test; HTTP route tests alone are insufficient for a Wails build.
- Ad-hoc post-build signing must explicitly preserve `build/darwin/entitlements.plist`; signature validity alone does not prove the app will use the existing sandbox container.
- The quote and backend-enforcement rollout gate remains open until the owner unlocks the corrected build and the same exact-channel non-paying quote succeeds. Payment execution remains disabled independently of that pending test.

### 2026-10-06 — Exact-channel quote reached the provider and was rejected

Observed:

- The owner unlocked the installed diagnostic build from Hub commit `3ddb44d9147166d119e52c360951b419e9e570f4`. Its executable SHA-256 is `ffbfbd8ce73346365d14de1771384bccc21c8cbe6079fe41d6f45e61b95df835`; the archived bundle is `/Users/kode/Development/albyhub-builds/v1.24.1/Alby-Hub-v1.24.1-custom-3ddb44d9.zip` with SHA-256 `b70224f28d7cf964e38b8bf42a838a0945f6974da32600fc9d5f26214a774f0a`.
- All three channels were online immediately before the test. The exact outgoing channel `255349693497905514942651862984333935477` for peer `030ef18b788bdfaf899071bb975f258306f83eae0a83d9e52aee93ae894296a42c` showed `989,340 sats` spendable. The exact incoming Kraken channel `86157859664272214382561858939519142638` for peer `02437c00ef5de2686a6bd60f8acb5c83d17010916010a15f479d5ef84c04f04485` showed `840,498 sats` receiving capacity.
- One non-paying quote request used principal `500,000 sats`, provider-fee cap `2,500 sats`, routing-fee cap `1,000 sats`, and maximum possible debit `503,500 sats`. The UI's `Execute locked` control remained disabled throughout.
- The request passed the Wails router and local channel/capacity checks, reached the external rebalance provider, and returned HTTP `422`. The installed diagnostic build displayed no provider reason because the provider's known error shape nests its message below `error.message`, while commit `3ddb44d9` only accepted top-level string fields.
- Historical responses from the same provider endpoint use the nested reason `no_route_found`. That makes route unavailability the leading explanation for this `422`, but it is an inference rather than a captured reason from this exact request.
- After the rejection, the source and destination channel balances remained `989,340 sats` spendable and `840,498 sats` receivable. The database returned `integrity_check = ok`, `rebalance_quotes=0`, `apps=12`, `app_permissions=94`, and `user_configs=16`. No provider invoice was paid, no quote was persisted, and no sats moved.
- Hub commit `e88a9e4067c907072823d9a35b2e58d29e0e9426` safely parses `error.message`, `error.detail`, or `error.reason` while retaining invoice redaction, control-character filtering, and length limits. API, Wails, and HTTP tests pass; the pre-existing Bark macOS deployment-target warnings remain. This parser commit is source-only and is not installed in the running app.

Consequence:

- The desktop routing defect is closed: a quote request now reaches the provider. The current exact `500,000-sat` Kraken-return quote is unavailable at the provider boundary, so no reviewable quote exists.
- Do not reduce the amount, change either channel, increase either fee cap, or retry automatically. Any follow-up quote must be a separately reasoned non-paying test using fresh channel state.
- Payment execution remains hard-locked. The provider rejection does not change the unresolved incoming-channel atomicity gate and does not authorize a value-moving rebalance.

### 2026-10-06 — Nested provider-reason parser installed; owner unlock pending

Observed:

- A universal macOS bundle was built from pushed Hub source commit `cc76beb56715a9e25483d452c1fe08df738c8231`, which includes nested provider-error parser commit `e88a9e4067c907072823d9a35b2e58d29e0e9426`.
- The staged and installed executable SHA-256 is `bb431b3a378690ed055697704865be0039e0e56d01b545f2c6c58706ec6ea89c`. The embedded signed LDK library remains SHA-256 `6e12519aa3ebdc88399990f6f46aa1054b636d48f78d19e7978ff2f9a1e4e21d` and both artifacts are universal arm64/x86_64.
- The executable resolves LDK only through `@executable_path/../Frameworks`. Deep signature verification passes and the app retains sandbox, download/user-selected file, and client/server network entitlements.
- The archived bundle is `/Users/kode/Development/albyhub-builds/v1.24.1/Alby-Hub-v1.24.1-custom-cc76beb5.zip` with SHA-256 `7fdfd92536d80c75efa31fced2ffd1c28e703bdae30c6df9f092a54a29b84fae`.
- Before replacement, transaction-safe snapshots of both SQLite databases and a full app rollback were created at `/Users/kode/Development/albyhub-rollbacks/2026-10-06-before-nested-provider-error-e88a9e40`. The old process again required an exact-PID force-stop after `SIGTERM` timed out; no broad process or filesystem target was used.
- After installation, `nwc.db` returned `integrity_check = ok` and continuity remained `apps=12`, `app_permissions=94`, and `user_configs=16`.
- The new process runs from `/Applications/Alby Hub.app/Contents/MacOS/Alby Hub` and displays `v1.24.1` at the owner password screen. It has not been unlocked, no quote has been retried, and no sats moved.

Consequence:

- Live parser verification is pending owner unlock. After unlock, re-read exact channel state and retry the unchanged non-paying quote once; do not execute or change any parameter.
- The installed parser improves safe diagnostics only. It does not enable payment execution or change the incoming-channel atomicity gate.

### 2026-10-06 — Provider route unavailability confirmed live

Observed:

- After owner unlock, process `36992` loaded `/Applications/Alby Hub.app/Contents/Frameworks/libldk_node.dylib`; the installed executable and library hashes remained `bb431b3a378690ed055697704865be0039e0e56d01b545f2c6c58706ec6ea89c` and `6e12519aa3ebdc88399990f6f46aa1054b636d48f78d19e7978ff2f9a1e4e21d`.
- All three channels were online. Immediately before the retry, outgoing channel `255349693497905514942651862984333935477` for peer `030ef18b788bdfaf899071bb975f258306f83eae0a83d9e52aee93ae894296a42c` had `989,340 sats` spendable; incoming Kraken channel `86157859664272214382561858939519142638` for peer `02437c00ef5de2686a6bd60f8acb5c83d17010916010a15f479d5ef84c04f04485` had `840,498 sats` receiving capacity.
- One unchanged non-paying quote request used principal `500,000 sats`, provider-fee cap `2,500 sats`, routing-fee cap `1,000 sats`, and maximum possible debit `503,500 sats`. The `Execute locked` control remained disabled.
- The provider returned HTTP `422` with the safely parsed exact reason `no_route_found`. The application log recorded only the status and sanitized reason; it did not log the invoice or request body.
- At the owner's request, one bounded smaller non-paying quote used principal `250,000 sats`, the same exact channels, the same fee caps, and maximum possible debit `253,500 sats`. The provider again returned HTTP `422: no_route_found`; execution remained locked.
- At the owner's explicit request, one further non-paying quote used principal `20,000 sats` (`0.00020000 BTC`), the same exact channels and fee caps, and maximum possible debit `23,500 sats` (`0.00023500 BTC`). The provider again returned HTTP `422: no_route_found`; the sanitized log independently preserved the reason after the short-lived UI notification disappeared.
- After rejection, the selected channel balances remained `989,340 sats` spendable and `840,498 sats` receivable. The database returned `integrity_check = ok`, `rebalance_quotes=0`, `apps=12`, `app_permissions=94`, and `user_configs=16`. No reviewable quote was created, no provider invoice was paid, and no sats moved.

Consequence:

- The nested-reason parser is proven in the installed desktop application. The current provider cannot quote the exact `500,000-sat` return path through Kraken at this time.
- The same rejection at `250,000 sats` and `20,000 sats` shows that the original amount alone was not the limiting factor at the three observed timestamps. This remains provider/path availability evidence, not proof that another fee cap, provider, or time would succeed. Do not continue retrying or alter other parameters automatically.
- Because no quote exists, there is no action-specific payment packet to approve. Payment execution remains hard-locked and the incoming-channel atomicity gate remains unresolved.

### 2026-10-06 — `no_route_found` is provider-specific, not a global route proof

Observed:

- Local LDK state reports the Kraken channel online, public, and able to receive `840,498 sats`. The funding outpoint is `d6d600818fcebec415ae7cc3ed403677f6e3374bd9f59184937a43597048ab15:0`; its public short-channel ID is `969161x1096x0` (`1065603788758843392`).
- The Kraken peer pubkey matches Kraken's currently published post-migration node pubkey. The channel opened after Kraken's August 2026 node migration, so this is not the old Kraken node identity.
- Hub's LDK `MakeInvoice` implementation accepts `throughNodePubkey` but does not use it; it creates an ordinary public invoice. The provider separately receives only `pay_through_this_public_key` with the Kraken pubkey, not the exact Kraken channel ID or short-channel ID.
- The current node has one Kraken channel, but the provider controls route construction for the return leg. Its `no_route_found` response therefore proves only that the provider could not build its required route at that moment. It does not prove that no route exists from this node's selected source channel through the wider network to Kraken.
- Every provider rejection occurred after Hub created a five-minute local incoming invoice. Database rows `224` through `228` remain `PENDING` for the diagnostic attempts even though no provider order or rebalance quote was created. They are unpaid receive invoices, not payments, but expiry-state cleanup is a separate correctness item and they must not be silently deleted.

Proposed next design checkpoint:

- Add a quote-only local circular route builder that creates a self-invoice with a single exact Kraken last-hop hint and asks LDK to find a route using only the selected `030ef18b…` first hop. Return the complete candidate path, exact first and final channel identifiers, estimated routing fee, and failure reason without sending probes or HTLCs.
- Accept a candidate only if every path begins with user channel `255349693497905514942651862984333935477` and ends with Kraken short-channel ID `1065603788758843392`; otherwise fail closed.
- Treat this as a new direct self-payment architecture, not an extension of the current provider quote. Keep execution disabled until circular self-payment behavior, MPP, retry, restart, wrong-channel, fee, and settlement atomicity are proven deterministically.

### 2026-10-06 — Quote-only local circular route implementation

Observed:

- LDK's normal router explicitly rejects a route when payer and payee are the same node (`Cannot generate a route to ourselves`). A literal self-invoice route search therefore cannot implement the quote.
- Owner-fork LDK commit `c96625a` adds a pathfinding-only circular quote. It searches toward a synthetic terminal behind a one-hop route hint for the exact selected inbound channel, restricts first hops to only the exact selected source channel, and then validates every returned path against both channel SCIDs before replacing the synthetic terminal with the real local node in the quote representation.
- The final-hop routing identifier is obtained with LDK's `get_inbound_payment_scid()`. An inbound alias takes precedence when present because that is the identifier the counterparty recognizes for inbound forwarding. The exact local `user_channel_id` remains the primary binding, and the selected route SCID/alias is returned for independent review.
- Quote construction creates no invoice, provider order, database row, probe, HTLC, payment-store entry, or payment. The returned object contains every path, hop pubkey, SCID/alias, per-path fee, and total estimated routing fee.
- A focused Rust regression test proves validation rejects a wrong first-hop SCID or wrong final-hop SCID. The library and UniFFI surface compile. The full Rust unit suite passed; integration tests could not start because this checkout has no `BITCOIND_EXE`, which is an environment prerequisite rather than a product failure.
- Hub now has a separate `/api/channels/rebalance/local-quote` contract. It revalidates exact channel ID plus full peer pubkey, current source spendable capacity, current destination receivable capacity, returned channel IDs, first-hop peer and SCID, penultimate Kraken peer, final-hop SCID/alias, and routing fee cap. SCIDs are serialized as decimal strings so JavaScript cannot lose 64-bit precision.
- The desktop dialog now uses the local quote endpoint and displays the complete candidate route. Provider-fee controls are absent because no provider is involved. The execute control remains disabled and the backend contains no local execution method.

Consequence:

- A successful local quote will prove only that the current gossip graph and scorer can construct a candidate path pinned to both exact local channels. It does not prove live liquidity, settlement, or execution safety.
- No route candidate may be converted into a payment until a separate implementation and deterministic tests prove the synthetic-terminal route can be converted to a valid self-payment route without changing either endpoint, and the owner approves one fresh exact action packet.

Installed verification state:

- Hub source commit `7cd8149189b74a7e2f1211f8cc61359527f6289d`, LDK source commit `c96625a4792d50123878a42db4ac3603b2e804c5`, and Go binding/library commit `77c3d0e05a0cb0cbc1763f563d46f14bb9e68f7b` are pushed to the owner's `codex/upgrade-v1.24.1` branches.
- The installed executable SHA-256 is `c86f7d730ea50d1b5f451cfb70a2e7d592ab5fb93b7a791930e807231a95200f`; the installed signed universal LDK library SHA-256 is `9bf970eb94feb1895754e65dbbffb7181588e19a2ab05183a4dae43645e44430` and exports the quote symbol on both arm64 and x86_64.
- The executable has only `@executable_path/../Frameworks` as its LDK runtime path. Deep signature verification passes with the existing sandbox, download/user-selected file, and client/server network entitlements.
- The archive is `/Users/kode/Development/albyhub-builds/v1.24.1/Alby-Hub-v1.24.1-custom-7cd81491.zip`, SHA-256 `eb85b59ed6bbcd3357f0d10739aef1f802a26fbb15b929a3f7c4696310061749`.
- The recoverable pre-install app and transaction-safe copies of both SQLite databases are in `/Users/kode/Development/albyhub-rollbacks/2026-10-06-before-local-route-7cd81491`. Both database copies passed `integrity_check`.
- Process `49724` loaded the new installed library. Live `nwc.db` passed `integrity_check`; continuity remains `apps=12`, `app_permissions=94`, `user_configs=16`, and `rebalance_quotes=0`.
- Owner unlock completed after the restart. The apparent password failure was only the normal `Starting node...` interval; the existing password was accepted and no password or credential was changed.
- At `2026-10-06 19:25 PDT`, the live channel screen showed the exact source peer `030ef18b788bdfaf899071bb975f258306f83eae0a83d9e52aee93ae894296a42c` offline with `989,340 sats` spendable and the exact Kraken channel `86157859664272214382561858939519142638` inactive with `840,498 sats` receiving capacity. LDK was actively retrying the source peer, while a TCP connection to Kraken's published address was established but the channel had not become active.
- The `20,000-sat` local quote was not submitted because the quote contract fails closed before pathfinding when either selected local channel is inactive. This is a channel-connectivity blocker, not a `no_route_found` result. No invoice, probe, provider order, quote row, HTLC, payment, or balance movement was created.
- At `2026-10-06 20:08 PDT`, Kraken had recovered to `Online`, still showing `138,841 sats` spendable and `840,498 sats` receiving capacity. The exact `030ef18b…` source channel remained offline with `989,340 sats` spendable, and LDK continued logging closed reconnect attempts through `20:07 PDT`; the quote therefore remained unsubmitted.
- Current public explorer data reports that `030ef18b…` advertises only a Tor `.onion:9735` address and had last published a node update two days earlier. Because this is third-party crawler data rather than a live peer handshake, treat Tor-only reachability as a likely explanation for the failed direct reconnects, not a confirmed root cause. The remote peer may still restore the channel by initiating a connection to this Hub.
- At `2026-10-06 21:22 PDT`, all three channels were online and the installed quote-only implementation found a candidate for principal `20,000 sats` (`0.00020000 BTC`) with a `1,000-sat` routing-fee cap. The exact source remained user channel `255349693497905514942651862984333935477` with peer `030ef18b788bdfaf899071bb975f258306f83eae0a83d9e52aee93ae894296a42c`; the exact destination remained Kraken user channel `86157859664272214382561858939519142638` with peer `02437c00ef5de2686a6bd60f8acb5c83d17010916010a15f479d5ef84c04f04485`.
- The quote estimated `17,062 msat` (`17.062 sats`) in routing fees and an exact maximum source debit of `20,017,062 msat` (`20,017.062 sats`). Its one five-hop path was: `030ef18b…` via SCID `1065615883472404481`; `03a5c38d…` via `1058901166039629824`; `026f4620…` via `1066657121068449793`; Kraken `02437c00…` via `1059283795959676929`; and this Hub `02503706…` via exact Kraken final SCID `1065603788758843392`.
- Immediate independent readback showed the source unchanged at `989,340 sats` spendable and Kraken unchanged at `138,629 sats` spendable plus `840,710 sats` receiving capacity. `nwc.db` returned `integrity_check = ok`, `rebalance_quotes=0`, `apps=12`, `app_permissions=94`, and `user_configs=16`. The execute control remained disabled; no invoice, provider order, probe, quote row, HTLC, payment, or sat movement occurred.
- This is proof that the current local graph and scorer can construct a route pinned to both selected local channel identifiers at that moment. It is not proof of live intermediate liquidity, settlement, retry safety, or incoming-channel atomicity, so execution remains locked.

### 2026-10-06 — Durable local quote binding and circular-payment design finding

Observed:

- The quote-only router reaches a synthetic terminal because ordinary LDK pathfinding rejects payer-equals-payee routes. The returned review replaces only the displayed synthetic terminal identity with the real local node; the internal route is not yet a valid payment route and must not be sent as-is.
- LDK's fixed-route send accepts an already constructed route and disables automatic routing retries. This is suitable for preserving the exact reviewed path, but it does not by itself solve circular inbound settlement.
- The current LDK-node event handler deliberately fails a `PaymentClaimable` event when the payment-store entry for that hash is outbound, logging that circular payments are unsupported.
- LDK's manual `receive_for_hash` flow can register an inbound hash without revealing its preimage and exposes all receiving channel IDs at claim time. A distinct outbound payment ID would avoid replacing that inbound payment-store record.

Resulting design constraint:

- A future executor must create and durably register the inbound payment hash before sending, keep the preimage unavailable to the network, submit only the approved fixed route, and claim only after every inbound MPP part reports the exact Kraken channel. Any other incoming channel must fail backward without revealing the preimage.
- The preimage, quote binding, state transition, and restart recovery must be durable before the first HTLC is sent. The storage mechanism must preserve the node's existing encryption and secret-handling boundaries; plaintext application logging or source storage is forbidden.
- A successful quote is now persisted for five minutes in `local_rebalance_quotes` with a random quote ID, full route JSON, request hash, and SHA-256 fingerprint over all execution-relevant route material. The review UI shows its ID, fingerprint, and expiry.
- Fingerprint tests prove deterministic output and mutation detection for hop SCID, exact channels, principal, fee, and final SCID. Database migration tests prove the quote record is copied between supported databases. The complete Go test suite, frontend lint/type checks, and the Wails frontend production build pass; only the pre-existing dependency build warnings remain.
- Owner-fork LDK source commit `45f734a8552716133b51f3310b2e2472296384c2` adds a fail-closed conversion from the synthetic quote route to a real self-payment route object. It rejects empty, blinded, wrong-first-hop, or wrong-final-hop paths; changes only the terminal pubkey to the local node; preserves every SCID and fee; and clears the synthetic route parameters so a future fixed-route send would reconstruct them from the real final hop with zero automatic retries. All 27 Rust library tests pass, including exact conversion and mismatch rejection tests.
- Owner-fork LDK source commit `b8ff3e152226af721f7caba9fd8cb0540547fcf7` moves the existing pinned first-hop payment-store write before `send_payment_with_route`. A persistence failure therefore prevents the value-moving call; an immediate non-duplicate send failure is persisted as failed; and an LDK duplicate/in-flight result remains pending so restart recovery is not falsely marked failed. All 27 Rust library tests pass.
- Owner-fork LDK source commit `eb2ebb689d022c32b23ae505c8a3f7d6f2436a0e` adds a backward-compatible optional exact receiving-channel constraint to persisted inbound BOLT 11 payment records. The claim decision is applied even if ChannelManager unexpectedly knows the preimage and returns `Fail` unless the record is still pending, the received amount exactly equals the persisted amount, the preimage is present, and every reported MPP part has the exact required local user-channel ID. A mismatch fails the HTLC backward without revealing the preimage and marks the inbound record failed, preventing a later retry from becoming claimable. Native Rust tests pass `29/29`; Rust plus UniFFI tests pass `39/39`, with only the two pre-existing binding warnings.
- The `eb2ebb6` tests exercise the complete pure claim/fail decision and persistence compatibility, but do not yet drive a real `PaymentClaimable` event against a controlled ChannelManager to independently observe `claim_funds` versus `fail_htlc_backwards` calls.
- Owner-fork LDK source commit `3087a40a6c09b8d0be4289cc4cb93169668ba1ec` adds a non-sending circular preparation method. It takes a caller-supplied 32-byte operation binding, exact first- and last-hop user-channel IDs, exact amount, expiry, and routing-fee limit; revalidates both channels and maximum source debit; derives a domain-separated outbound payment ID; creates the inbound invoice; and returns only after persisting the invoice, preimage, operation binding, both channel constraints, fee limit, and outbound ID. Preparation is serialized by a shared process lock and persisted operation reuse is rejected after restart. The full invoice is no longer written to logs. Native Rust tests pass `30/30`; Rust plus UniFFI tests pass `40/40`, with only the two pre-existing binding warnings.
- Commit `3087a40` exposes the preparation primitive in source bindings, but no Hub code calls it. It has no fixed-route circular send method and the installed app cannot reach it.
- Owner-fork LDK source commit `5f43f7a98dfa8491fe4aec126233f78d49a6450c` routes the production event-handler claim/fail calls through a small internal action adapter and tests the exact dispatch boundary with controlled counters. Exact decisions call only `claim_funds`; rejection decisions call only `fail_htlc_backwards`; unconstrained payments call neither. Native Rust tests pass `31/31`; Rust plus UniFFI tests pass `41/41`, with only the same two pre-existing binding warnings. This is not yet a multi-node Lightning settlement test and does not independently exercise the handler's failed-record persistence update.
- Owner-fork LDK source commit `39dc42fd5dfd6b7cdf6c7593150efe9b8d40204a` adds a deterministic two-node Lightning test using real in-memory HTLCs. The first payment's actual `PaymentClaimable.receiving_channel_ids` is deliberately mismatched against the persisted requirement; the decision returns `Fail`, the HTLC is failed backward, and the sender observes failure. A second payment uses the exact reported user-channel ID; the decision returns `Claim`, the receiver reveals the preimage, and the sender observes successful settlement. Native Rust tests pass `32/32`; Rust plus UniFFI tests pass `42/42`, with only the same two pre-existing binding warnings. This test moves no live sats and does not use the installed Hub.
- At that checkpoint, the `39dc42f` topology proved real Lightning fail/claim propagation for a single-part payment but did not invoke the event-handler payment-store update, cover real MPP across mixed incoming channels, or reload the pending prepared operation after restart.
- Owner-fork LDK source commit `259d0f4f47a621065fa0f7679530954c31f741a3` moves the rejection action and failed-record update into the same internal helper used by the production event handler. Its test persists a pending constrained inbound record, dispatches rejection, reloads the payment store from the underlying key-value bytes, and verifies the record is `Failed`. Native Rust tests pass `33/33`; Rust plus UniFFI tests pass `43/43`, with only the same two pre-existing binding warnings.
- Owner-fork LDK source commit `fe52dfad0e78e4283302e8ada12dc7865fbde312` adds a deterministic four-node Lightning test that forces one 15,000,000-msat MPP across two distinct receiver channels. The real `PaymentClaimable` event reports two different receiving user-channel IDs; constraining the payment to either one produces `Fail`, no preimage is disclosed, both HTLC parts fail backward, and the sender reaches terminal `RecipientRejected`. Native Rust tests pass `34/34`; Rust plus UniFFI tests pass `44/44`, with only the same two pre-existing binding warnings. This test moves no live sats and does not use the installed Hub.
- Owner-fork LDK source commit `25341602fca8a5ef14c4f5a59645ad00bd270e3d` factors the exact prepared-record construction and operation-reuse test behind the production preparation path, then creates a real fixed-amount invoice against two in-memory channels and restarts from serialized ChannelManager plus both ChannelMonitors. The invoice preimage, exact channel IDs, and byte-equivalent pending payment record survive; both channels remain unusable before reconnection; no outbound payment appears; and the production operation-reuse check still rejects the persisted operation ID. Native Rust tests pass `35/35`; Rust plus UniFFI tests pass `45/45`, with only the same two pre-existing binding warnings. This covers prepared-but-unsent restart, not interruption of an in-flight HTLC.
- Owner-fork LDK source commit `00c82e65b022707bc94e1d6c4c1f86666c593346` closes a route-fidelity gap found before adding any send primitive. The human-readable quote omitted LDK node/channel feature bits, so rebuilding from its visible hops would not reproduce the exact reviewed route. Quotes now also carry the complete finalized LDK route bytes. The decoder requires no retry parameters or blinded tails, cross-checks total amount, total fee, every path, node, SCID, hop amount, fee, and CLTV against the visible quote, rejects trailing bytes, and preserves the original feature bits. Native Rust tests pass `36/36`; Rust plus UniFFI tests pass `46/46`, with only the same two pre-existing binding warnings. The UDL source includes the new field, but Go bindings have not been regenerated and Hub has not been rebuilt or installed.
- Owner-fork LDK source commit `38d9450ea63e70b0df6506cad0efe70daebdad6e` adds a still-unreachable circular executor. It reloads exactly one pending prepared operation, rejects expired or mismatched invoices, re-reads both selected channels and their current SCIDs/capacities, decodes and cross-checks the complete reviewed route, requires every path to start and finish through the selected counterparties and SCIDs, persists a distinct pending outbound record before one fixed-route send action, performs no automatic retry, rejects duplicate submission before a second action, marks both legs failed on an immediate non-duplicate send error, and preserves pending state when LDK reports a recovered duplicate/in-flight payment. A real three-node circular HTLC leaves through the selected first channel, becomes claimable only on the selected incoming channel, settles successfully, and reports the reserved outbound payment ID, exact hash/preimage/amount, `2,000 msat` fee, and the selected first/final SCIDs. Native Rust tests pass `37/37`; Rust plus UniFFI tests pass `47/47`, with only the same two pre-existing binding warnings. The send method is crate-private, absent from the UDL/Go bindings, and unreachable from Hub.
- Owner-fork LDK source commit `f44a8eaacd64f3fdf4fc129e581bc964381b76d8` adds a deterministic five-node real circular MPP test around the dormant executor. One `5,000,000 msat` payment is split into `3,000,000 msat` and `2,000,000 msat` parts only after the selected source peer, reconverges before the selected destination peer, and uses the same exact local outgoing channel and same exact local incoming channel for both parts. The test exercises LDK holding-cell serialization when both additions share the source channel and when both fulfillments share the destination channel; the real `PaymentClaimable` event reports two parts with the exact selected incoming user-channel ID; settlement reports the reserved payment ID, exact hash/preimage/amount, total `6,000 msat` routing fee, and two successful paths whose first and final SCIDs both match the selected channels. Native Rust tests pass `38/38`; Rust plus UniFFI tests pass `48/48`, with only the same two pre-existing binding warnings. The test network is entirely in memory, the executor remains crate-private and unreachable from Hub, and no live sats moved.
- Owner-fork LDK source commit `a579c6ef14730f115f05d935b29ed2c1c553e4d7` strengthens the real single-path circular executor test with an actual in-flight restart. After the one fixed-route send durably creates the outbound HTLC but before the first peer receives it, the local ChannelManager and both selected ChannelMonitors are serialized and reloaded. Both persisted payment-store legs remain pending; a second executor submission is rejected without invoking its send closure; channel reestablishment retransmits the already-committed HTLC; and the payment then becomes claimable on the same selected incoming channel and settles with the same reserved payment ID, hash, preimage, amount, `2,000 msat` fee, and first/final SCIDs. Native Rust tests pass `38/38`; Rust plus UniFFI tests pass `48/48`, with only the same two pre-existing binding warnings. This proves route-pinned in-flight recovery and duplicate-send prevention, not yet replay idempotency of recovered terminal events.
- Owner-fork LDK source commit `746197440551bed4b7a48f1d2aa6481d96d6cd1e` routes the production `PaymentClaimed` and `PaymentSent` payment-store mutations through directly tested update builders. The test begins with the distinct prepared inbound hash ID and reserved outbound payment ID, applies the same terminal updates used by the production event-handler branches, reloads both records from persisted bytes, and replays both updates. The inbound and outbound records remain separate and succeeded; only the outbound record carries the exact routing fee; both replayed mutations return `Unchanged`. Native Rust tests pass `39/39`; Rust plus UniFFI tests pass `49/49`, with only the same two pre-existing binding warnings. This proves idempotent recovered terminal payment-store state; external event-queue delivery remains governed by its existing durable queue.
- Owner-fork LDK source commit `f4d945e` makes the already-tested prepared circular send method public and exposes it through UniFFI. Native Rust tests pass `39/39`; Rust plus UniFFI tests pass `49/49`.
- Owner-fork Go binding commit `87ef9c6` regenerates `PrepareCircularPayment`, `SendPreparedCircularPayment`, and `CircularRouteQuote.RouteBytes` from that exact LDK interface, includes a stripped 43,113,184-byte universal arm64/x86_64 macOS library with SHA-256 `f6fcaa2c235bc1a30776c8693a2932701dc3c6968cf6b9635175230bc4f91789`, and passes `go test ./...`.
- Hub commit `4738efcf` pins pseudo-version `github.com/hudhaifahz/ldk-node-go v0.0.0-20261007072228-87ef9c6e3ed9`, carries the opaque route bytes from LDK, adds migration `202610070200_local_rebalance_route_bytes`, stores those bytes outside the UI response, includes them in the route fingerprint, and rejects missing or tampered route bytes before quote acquisition. The complete Go suite passes; only the pre-existing Bark macOS deployment-target warnings remain.
- The source now has the correct binding and persisted route material, but no Hub backend calls preparation or sending. The execute endpoint remains hard-locked, the installed app still uses the prior quote-only build, and no sats moved.

Consequence:

- Quote persistence closes the review-to-execution identity gap, but it does not authorize payment. Exact inbound decisions, real mixed-channel MPP rejection, prepared-but-unsent and in-flight restart recovery, duplicate pre-send rejection, immediate send-error handling, fixed-route single-path and pinned-MPP settlement, exact first/final channel evidence, reserved outbound-ID settlement, and distinct idempotent terminal payment-store records are now covered.
- The executor must remain unreachable from Hub until restart-safe orchestration closes every persistence boundary. In particular, recovery must distinguish: quote acquired before preparation; preparation persisted before Hub records its result; outbound payment record persisted before ChannelManager receives the send; and an already-committed in-flight payment. A retry may submit the same fixed route and payment ID only when LDK can prove that doing so either starts the previously unsent operation or returns the already-existing in-flight operation without creating a second HTLC. The atomic quote acquisition must then be the mandatory predecessor, followed by a fresh build/install verification and a separate owner action packet.

### 2026-10-05 — Hub quote and review implementation

Observed:

- Hub commit `9aebc8e1` adds the optional pinned-payment capability and exact LDK call.
- Hub commit `8e672e4c` carries exact receiving channel IDs into the Hub hold-invoice event model.
- Hub commit `fe3f92cf` adds persisted five-minute quotes, exact channel-plus-pubkey validation, separated fee limits, legacy-flow rejection, and a hard-locked execute endpoint.
- Hub commit `47782062` adds the exact-channel quote/review UI and visibly locked execution state.
- Hub commit `47057488` adds a dormant acquisition primitive that requires the owner-reviewed route fingerprint, reloads and re-fingerprints every persisted route field, rejects missing, expired, tampered, mismatched, or non-`quoted` records, and atomically transitions exactly one matching quote from `quoted` to `executing`. Eight concurrent attempts against one quote produce exactly one success. The request model now carries the route fingerprint, but the endpoint remains hard-locked and no backend send calls this primitive.
- Hub commit `c60e40ff` adds migration `202610070100_single_local_rebalance_execution`, whose database-level partial unique index permits at most one `executing` local quote across all quote IDs. A second eight-way concurrency test races eight distinct valid quotes and proves exactly one reaches `executing`.
- Focused LDK, API, database, frontend lint, and TypeScript checks pass.
- The HTTP frontend production build passes; it reports existing dependency warnings for Lottie `eval`, large chunks, and stale Browserslist data.
- Hub commit `7ab500d7` adds the missing NWC metadata config expectation; the broad HTTP package test now passes.

Consequence:

- The current source can be reviewed and can produce a non-paying quote, but it is intentionally not a value-moving feature.
- The focused quote-acquisition tests, full API/Wails/HTTP tests, and complete Go repository suite pass. The new acquisition code is committed and pushed but has not been regenerated into bindings, connected to the dormant Rust executor, built, installed, or exercised with live value.

### 2026-10-08 — Crash-safe same-operation send recovery

Observed:

- Owner-fork LDK commit `77c0454` closes the remaining outbound-record/ChannelManager crash boundary for a prepared circular operation. Recovery derives the same domain-separated outbound payment ID and cross-checks the persisted inbound and outbound records against the exact operation ID, payment hash, preimage, secret, invoice, amount, fee cap, outgoing user-channel ID, and incoming user-channel ID.
- If ChannelManager already reports the matching payment as pending or fulfilled, the public prepared-send call returns the existing outbound payment ID before any new send call. Wrong hash, wrong amount, abandoned state, unknown payment type, contradictory persisted status, or any record mismatch fails closed.
- If the outbound record was persisted but ChannelManager does not track it, recovery can reach only the existing fixed-route builder. That builder decodes and revalidates the stored route bytes and both selected channel identities, then submits the identical route with the identical payment ID. An LDK `DuplicatePayment` response is treated as recovery of the existing operation and does not mark either leg failed.
- The real in-flight restart test reloads ChannelManager and both selected ChannelMonitors, confirms the exact deterministic ID/hash/amount is still tracked, and proves a direct same-route/same-ID resend returns `DuplicatePayment` without adding a monitor or second HTLC. The original HTLC then settles through the exact selected final channel.
- Native Rust tests pass `39/39`; Rust plus UniFFI tests pass `49/49`. The only binding-build output is the two pre-existing warnings.
- Owner-fork Go binding commit `958ddc6` contains the rebuilt stripped universal arm64/x86_64 macOS library from LDK commit `77c0454`. Its size is `43,162,496` bytes, SHA-256 is `ad3c898505db401a5fd6dbb8f2b231c8c16bb1d14c6abfa6735f29cb8668c231`, and `go test ./...` passes.
- The Hub pins that artifact as `github.com/hudhaifahz/ldk-node-go v0.0.0-20261008200336-958ddc682bfd`; the complete Hub Go suite passes. The existing Bark macOS deployment-target linker warnings remain unchanged.
- No Hub executor orchestration was wired, no desktop build was installed or restarted, no invoice or HTLC was created outside the in-memory tests, and no sats moved.

Consequence:

- The LDK layer now has a tested, idempotent answer for both sides of the dangerous crash boundary: an already-committed payment is returned without another send, while a durable-but-unsent record may retry only the exact stored route with the exact same deterministic ID.
- Hub execution remains hard-locked. Before it can be considered for installation, the Hub still needs a deterministic operation ID and durable state machine that joins quote acquisition, preparation, exact-route submission, recovery, terminal recording, and rollback/failure transitions without reopening route selection. That orchestration requires its own tests and review; it is not authorized by this source checkpoint.

### 2026-10-08 — Idempotent preparation and dormant Hub orchestration

Observed:

- Owner-fork LDK commits `104945e`, `5ad80b0`, and `4096ee5` make circular preparation idempotently recoverable. The same exact operation returns the previously persisted invoice and identifiers; changed amount, channel bindings, or fee cap fail closed. An expiry of zero is recovery-only: it can return an existing preparation but cannot create a new invoice.
- Native Rust tests pass `39/39` and Rust plus UniFFI tests pass `49/49` at LDK commit `4096ee5`.
- Owner-fork Go binding commit `81c07a4` contains the rebuilt stripped universal arm64/x86_64 macOS library from LDK commit `4096ee5`. Its size is `43,178,936` bytes, SHA-256 is `6c0e34777e14bb3416f736d153ca7fdd7ef2316c4161ba957e11b3815f1c0e46`, and `go test ./...` passes.
- Hub pins that exact artifact as `github.com/hudhaifahz/ldk-node-go v0.0.0-20261008201853-81c07a4acff0` and adds a database migration for deterministic operation ID, execution phase, prepared payment hash, outbound payment ID, and preparation/submission timestamps.
- The Hub runner derives its operation ID from the quote ID and reviewed route fingerprint, re-fingerprints the complete persisted route before every resume, validates LDK's returned amount, fee cap, user-channel IDs, SCIDs, payment hash, and independently derived outbound payment ID, and persists `acquired`, `prepared`, and `submitted` using compare-and-set transitions.
- Tests cover normal preparation/submission, no backend call after a durable `submitted` phase, same-operation retry after simulated post-send/pre-database uncertainty, expired recovery with expiry zero, refusal to create a missing expired preparation, wrong incoming-channel preparation, and tampered submitted identifiers.
- The runner is unexported and referenced only by tests. The existing HTTP, Wails, and UI execution paths remain hard-locked. No desktop app was built, installed, restarted, or unlocked; no live invoice or HTLC was created; no sats moved.

Consequence:

- Hub now has a reviewable crash-safe source model through the submission boundary without exposing a value-moving control. It cannot silently switch channels, create a second operation after expiry, or generate a new route during recovery.
- The remaining source blocker at this checkpoint was terminal reconciliation: consume durable Lightning success/failure evidence, verify the exact operation and both legs, write one terminal Hub state idempotently, and only then release the database's single active-operation slot. Until that was implemented and independently tested, the dormant runner had to stay unreachable and no build containing it could be installed for live use.

### 2026-10-08 — Exact two-leg terminal reconciliation implemented and tested

Observed:

- Owner-fork LDK commit `16630be` validates a failed circular outbound record against its deterministic operation ID, payment hash, invoice, preimage, secret, amount, fee cap, and exact outgoing/incoming user-channel IDs before marking both the distinct outbound and inbound payment records `Failed`. Replay is idempotent; a mismatched binding updates neither record.
- Native Rust library tests pass `41/41`, and Rust plus UniFFI library tests pass `51/51`. A broader integration-fixture invocation also encountered unrelated local Bitcoin regtest directory failures; the complete library matrices covering this behavior passed.
- Owner-fork Go binding commit `a3dd5b1` contains the rebuilt stripped universal arm64/x86_64 macOS library from LDK commit `16630be`. Its size is `43,195,848` bytes, SHA-256 is `3399f9446240441f406b7a307b6845415bbf5ea1282b3cebf9f5a689826a7554`, and `go test ./...` passes.
- Hub commit `59b6e750` pins `github.com/hudhaifahz/ldk-node-go v0.0.0-20261008205838-a3dd5b1fe3bb`, reloads both exact payment records, and validates every immutable binding plus matching terminal preimage and exact reviewed routing fee.
- The Hub migration records actual routing fee, Lightning terminal time, reconciliation time, and a SHA-256 terminal-evidence hash. Terminal writes compare-and-set only the matching `executing` operation in `prepared` or `submitted`; identical replay returns the persisted result without another backend read, while altered evidence is rejected.
- Both-success records produce `succeeded`; both-failed records produce `failed`; missing, pending, or one-terminal/one-pending records retain the active lock; opposite terminal outcomes fail closed as contradictory. A matching terminal transition releases the database's unique active-operation slot, which is proven by acquiring the next quote in the success test.
- Focused Hub tests, the API race test, focused `go vet`, and the complete Hub `go test ./...` suite pass. The existing Bark macOS deployment-target linker warnings remain unchanged.
- The reconciler and runner remain unexported and unreachable from HTTP, Wails, and UI. No desktop build was created or installed, no service was restarted or unlocked, no live invoice or HTLC was created, and no sats moved.

Consequence:

- Terminal source semantics are no longer the blocker: success and failure now require exact, durable, two-leg evidence and release the execution lock only after an idempotent database transition.
- The remaining work is owner-facing execution and reconciliation orchestration, followed by a fresh build/install review and a separate live action packet. Until those are explicitly reviewed and authorized, the existing execute endpoint and UI control stay hard-locked.

### 2026-10-08 — Owner-facing exact-route orchestration implemented in source

Observed:

- Hub commit `587f7837` adds full-access HTTP and Wails routes for exact local execution and read/reconcile status. The legacy provider execute path remains fail-closed.
- The quote response now supplies a quote-specific typed confirmation. A wrong confirmation leaves the quote `quoted` and invokes zero prepare, send, or reconcile backend calls.
- A correct confirmation atomically acquires the quote, prepares and submits the fixed route once, and performs one immediate durable status read. Repeating the execute request after terminal success returns the same stored result without preparing, sending, or reconciling again.
- Status calls never send. An `acquired` operation reports that the same exact confirmation is required; a `submitted` operation reads both durable records; terminal state returns the persisted evidence hash and actual fee without requiring a Lightning backend call.
- Browser recovery stores only the reviewed quote packet, then reloads status after interruption. The database and route fingerprint remain authoritative; altered browser storage cannot alter the operation.
- Migration `202610080300_local_rebalance_balance_evidence` stores both selected channels' local/remote balances and local/counterparty reserves. All eight fields are included in the route fingerprint and covered by migration/fingerprint tests.
- The review screen shows quote time/expiry, exact channel IDs and full pubkeys, path SCIDs, principal, quoted fee, fee cap, fixed maximum debit, spendable capacity, reserves, projected local/remote balances, and expected on-chain change of zero. It distinguishes temporary in-flight HTLC commitment from terminal reconciliation.
- The complete Hub Go suite, API race test, full `go vet`, frontend lint/TypeScript checks, HTTP production build, Wails production build, and database-copy test pass. Output is limited to the previously recorded Bark deployment-target, Lottie direct-eval, bundle-size, and stale Browserslist warnings.
- This commit is source-only. No desktop bundle was built or installed, the running app was not restarted or unlocked, no quote/invoice/HTLC was created by this work, and no sats moved.

Consequence:

- The owner-facing orchestration source gate is closed, but deployment and live acceptance are not. The next permissible step is a separately reviewed build/install checkpoint that proves UI presence, migration, wallet continuity, and the exact embedded source/library hashes without executing a rebalance.
- A live test still requires a fresh action packet from the newly installed build with current exact channel state, exact principal, eight-decimal BTC equivalent, quoted fee, fee cap, maximum debit, projected balances, expiry, quote ID, route fingerprint, and explicit owner approval immediately before the value-moving confirmation.

### 2026-10-08 — Exact owner-execution build installed and non-paying acceptance completed

Observed:

- Hub commit `defe4b32c68044c35318e979dbea6e740d9a39cd` was clean, matched `origin/codex/upgrade-v1.24.1`, and was built as universal arm64/x86_64 Alby Hub `v1.24.1`. It contains owner-execution implementation commit `587f7837` and pins `github.com/hudhaifahz/ldk-node-go v0.0.0-20261008205838-a3dd5b1fe3bb`.
- The pinned unsigned universal LDK library SHA-256 is `3399f9446240441f406b7a307b6845415bbf5ea1282b3cebf9f5a689826a7554`. Its quote, prepare, and prepared-send symbols are exported on both architectures. After embedding and ad-hoc signing, the installed library SHA-256 is `f853168d63280736237e84d40f869fd03fabe3e5560c5e067f026c893b68b19a`; the installed executable SHA-256 is `6b1e522a5fea45d615670b1db98083f7de5fc24d921b7a80effdfd339c770d5a`.
- The installed bundle passes deep signature verification, retains Alby's app-sandbox, client/server network, download, and user-selected-file entitlements, and resolves LDK only through `@executable_path/../Frameworks`.
- The reproducible archive is `/Users/kode/Development/albyhub-builds/v1.24.1/Alby-Hub-v1.24.1-custom-defe4b32.zip`, SHA-256 `604a0ffce06882e4ed05a721a39a392066b2969e97abac7cbfa83c01dd94b68d`.
- Before replacement, transaction-safe pre-install snapshots were created. After the exact old process stopped, post-quit snapshots and the complete prior application were preserved at `/Users/kode/Development/albyhub-rollbacks/2026-10-08-before-owner-execution-defe4b32`. Both database snapshots and both live databases returned `integrity_check = ok`.
- The installed process loads the embedded library and the canonical sandbox databases. Migration `202610080300_local_rebalance_balance_evidence` is recorded, `local_rebalance_quotes` exists, and all eight local/remote balance and reserve evidence columns are non-null `bigint` fields.
- Wallet continuity counts remained `apps=12`, `app_permissions=94`, `user_configs=16`, `transactions=235`, and `rebalance_quotes=0`; the new `local_rebalance_quotes` table is also empty. The live Settings screen reports `v1.24.1`.
- After owner unlock, the existing wallet loaded with four channels, Lightning balance `1,467,297 sats`, receive limit `3,460,722 sats`, and on-chain balance `41,883 sats`. Three channels were online. The Kraken channel for peer `02437c00ef5de2686a6bd60f8acb5c83d17010916010a15f479d5ef84c04f04485`, exact channel ID `86157859664272214382561858939519142638`, showed `138,846 sats` spendable and `840,493 sats` receiving but was offline.
- The current node completed Lightning and on-chain synchronization successfully. Kraken's last recorded direct-connection failures predate this installation, so its offline state is observed peer availability rather than evidence of migration or wallet loss.
- On an online channel, the live `Rebalance In` dialog visibly states that route discovery creates no invoice, probe, HTLC, or payment. It requires an exact outgoing channel, displays the full incoming pubkey and stable channel ID, lists each candidate outgoing channel with its exact ID and spendable balance, separates principal from routing-fee cap, requires a fresh quote, and keeps `Execute exact route` disabled before the quote-specific confirmation.
- The Kraken-specific `Rebalance In` action is correctly not offered while Kraken is offline. No route quote was requested, no invoice or HTLC was created, no execution confirmation was entered, and no sats moved.

Consequence:

- Build, installation, migration, wallet continuity, embedded-library identity, and non-paying UI safety behavior are accepted for this exact commit. Live route availability and value-moving acceptance remain unproven and separate.
- Do not attempt a Kraken-return quote until the exact Kraken channel is online and fresh channel state is read again. Any value-moving test still requires a newly generated exact action packet and a separate owner approval immediately before execution.
- Future official Alby updates may replay these small versioned commits, but compatibility is not automatic: rebase on the new official Hub and LDK baselines, resolve conflicts without weakening invariants, rebuild all native bindings, and repeat deterministic, migration, hash, wallet, and UI verification before installation.

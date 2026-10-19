# LDK-node review regression plan

Status: **SUBMITTED — LOCAL REGRESSION GATE PASSED; MAINTAINER CI APPROVAL PENDING**

Prepared: 2026-10-18

Baseline under review:

- upstream PR: `getAlby/ldk-node#115`;
- branch: `hudhaifahz:codex/exact-channel-rebalance-v1243`;
- reviewed head: `ed72fe992174965bc917d1dc6c6bee1dc7368057`;
- primary review-fix head: `bfdf1c000d1e07a3d99d298c9426200f997c32f3`;
- duplicate-race follow-up head: `8784ee0ff4da1abb76b5a1c9ba1aaa242139bc79`;
- claim-preimage follow-up head: `70965f2e58936c0893c48ee0cc1788b84164a4e2`;
- local checkout: `/Users/kode/Development/rebalance-upstream-v1243/ldk-node`.

This plan records the CodeRabbit findings raised after the PR became ready for
review, including the follow-up findings discovered while reviewing the first
fixes. Review comments are evidence to verify, not instructions to apply
blindly. Each change must remain minimal, receive a focused regression test,
and pass the complete non-paying validation gate before the PR branch moves.

## Findings and required proof

### 1. Event-queue TLV compatibility

Risk: `PaymentClaimable.receiving_channels` currently uses even TLV type `8`.
Unknown even types are required in LDK TLV streams, so an older reader can fail
to load a queue containing a new claimable event.

Planned fix:

- verify type `9` is unused in this event's TLV namespace;
- write new records with odd type `9`;
- continue reading the RC.6 type `8` encoding because RC.6 was installed and
  used before review;
- reject a record that contains conflicting legacy and current values;
- preserve the empty-vector default.

Required proof:

- current-version round trip preserves receiving-channel evidence;
- a legacy-shaped reader can skip the new odd field;
- the current reader accepts an RC.6 record containing type `8`;
- existing event-queue fixtures still decode.

Future-release watch: every new persisted/event TLV field must use a unique odd
type when it is optional or safely skippable. Compatibility review must cover
both upgrade and downgrade readers before release.

### 2. Circular `PaymentKind::Bolt11` TLV compatibility

Risk: the five new circular-payment fields use even TLV types
`131074..131082`. A pre-feature reader can reject a circular payment record and
prevent node startup after downgrade.

Planned fix:

- inventory the complete `PaymentKind::Bolt11` TLV namespace;
- write the five fields using unique odd types `131075..131083`;
- continue reading the RC.6 even types `131074..131082` because those records
  may already exist in installed wallet data;
- reject conflicting legacy/current duplicates instead of choosing one value.

Required proof:

- circular payment records round trip with all five fields populated;
- legacy payment details continue to deserialize;
- a legacy-shaped reader skips the new optional odd fields;
- the current reader accepts the five RC.6 even fields;
- no type collision exists in the variant.

Future-release watch: persisted TLV numbers become compatibility commitments
after release. Never renumber a published field without a migration strategy
and fixture coverage.

### 3. Non-retryable circular-record mismatch

Risk: the `PaymentFailed` handler currently returns `ReplayEvent` for every
error from paired circular-record reconciliation. A deterministic validation
mismatch can therefore replay forever and block all later LDK events.

Planned fix:

- return `ReplayEvent` only for a verified persistence/write failure;
- on a non-persistence mismatch, log the inconsistency and fail only the
  outbound payment record;
- leave the inconsistent inbound record unchanged for diagnosis;
- preserve successful paired-record failure reconciliation.

Required proof:

- persistence failure requests replay;
- a missing or inconsistent inbound record completes event handling and marks
  only the outbound record failed;
- a valid pair still reaches the expected terminal state idempotently;
- a later queued event is not blocked by a deterministic mismatch.

Future-release watch: replay is only for conditions that can become successful
on retry. Validation, identity, and state mismatches must terminate or be
quarantined without stalling unrelated wallet events.

### 4. Final CLTV correctness

Risk: circular quote construction uses an 18-block final CLTV delta while the
receiving invoice requires LDK's higher minimum. A persisted quote can pass
local structural checks but fail at receipt with an incorrect final expiry.

Planned fix:

- construct quotes with LDK's channel-manager minimum final CLTV delta;
- before submission, compare every decoded route path's final hop against the
  actual invoice minimum;
- fail before payment submission if a path is empty or below the invoice
  requirement.

Required proof:

- new quotes use at least LDK's minimum;
- a deliberately low-CLTV serialized route is rejected before send;
- an exact valid route remains accepted;
- stale quote, amount, fee-cap, channel, and fingerprint checks remain intact.

Future-release watch: quote fingerprints bind a route, but invoice-derived
constraints must still be revalidated at execution. Recheck CLTV constants and
invoice defaults whenever LDK or `lightning-invoice` is upgraded.

### 5. Terminal-state preservation on duplicate claims

Risk: a repeated claim for a payment that is already terminal must still fail
the duplicate HTLC backward, but must not overwrite a prior success/failure or
replace its diagnostics with `PaymentNotPending`.

Fix and proof:

- fail the duplicate HTLC as before;
- skip the payment-store rewrite when the payment was already terminal;
- regression-test that a succeeded record remains succeeded with its existing
  diagnostic fields unchanged.

Future-release watch: idempotent event handling must preserve terminal history.
Protocol cleanup for a duplicate event is distinct from mutating the stored
business outcome.

### 6. Synthetic pathfinding identity

Risk: a fixed, publicly known synthetic payee can collide with a real graph
identity and creates a stable fingerprint across quotes.

Fix and proof:

- generate a fresh secp256k1 identity from operating-system entropy per quote;
- discard the private key immediately;
- retain exact final-hop and local-node collision checks;
- regression-test that successive identities differ.

Future-release watch: synthetic routing identities must be ephemeral and must
never substitute for exact first/last-channel validation.

### 7. LDK routing-fee default preservation

Risk: assigning `None` from Grynvault configuration directly into route
parameters clears LDK's built-in routing-fee guard rather than meaning “use the
LDK default.”

Fix and proof:

- only override `max_total_routing_fee_msat` when configuration contains an
  explicit value;
- use the same helper for exact-channel quotes and first-hop sends;
- regression-test both the absent-config and explicit-cap cases.

Future-release watch: an optional application override must not erase a safer
upstream default. Recheck default-preservation semantics after LDK upgrades.

### 8. ChannelManager payment-ID collision

Risk: the application store can consider a payment ID unused while the
recovered ChannelManager still tracks it. Inserting a new pending record before
LDK returns `DuplicatePayment` can leave a false pending payment forever.

Fix and proof:

- preflight every `RecentPaymentDetails` state before inserting the record;
- reject an ID already tracked as awaiting invoice, pending, fulfilled, or
  abandoned;
- if submission returns `DuplicatePayment`, preserve the record as pending
  because LDK defines that result as an already-live in-flight payment and a
  concurrent sender can race the read-only preflight;
- for other submission failures, mark the newly inserted record failed;
- if that diagnostic status update cannot be persisted, log it without masking
  the original send error;
- regression-test that all four recovered states reserve the ID.

Future-release watch: duplicate checks must cover both the application payment
store and ChannelManager's recovered state before persistence. `DuplicatePayment`
is positive evidence of an in-flight LDK payment, not a terminal failure. Only
non-duplicate send failures should convert the newly created record to failed,
and diagnostic persistence failures must not replace the original send error.

### 9. Claim preimage and live payment-hash binding

Risk: the exact-channel claim gate verified that the stored payment contained a
preimage, but presence alone did not prove that the preimage belonged to the
live claimable payment. Corrupt or inconsistent persisted metadata could reach
`claim_funds` with the wrong preimage.

Fix and proof:

- require the persisted payment hash to equal the live claimable event hash;
- recompute SHA-256 over the stored preimage and require the result to equal the
  live claimable event hash before claiming funds;
- fail the HTLC backward and preserve a structured diagnostic when either
  identity check fails;
- regression-test the valid claim, missing-preimage, corrupt-preimage, and
  mismatched stored/live-hash cases.

Future-release watch: cryptographic material must be validated against the live
protocol identifier immediately before releasing funds. Stored metadata
presence is not sufficient evidence of identity.

## RC.6 continuity rule

RC.6 is not treated as an unpublished throwaway encoding. It was installed and
successfully completed a real owner-approved rebalance before these review
fixes. Therefore all persisted-data changes in this cycle must provide tested
dual-read compatibility for RC.6 while writing only the corrected encoding.
This rule applies to future release candidates as soon as they are installed
against real wallet data, even if the upstream PR has not merged.

## Validation sequence before pushing PR #115

1. Preserve the clean exact-head baseline and inspect every proposed diff.
2. Add focused regression tests for each confirmed finding before or with its
   fix.
3. Run `cargo fmt --all -- --check`.
4. Run the focused event, payment-store, and circular-payment tests.
5. Run the full default-feature library suite.
6. Run the full all-feature library suite.
7. Run warning-denied default, UniFFI development, and UniFFI release builds.
8. Run CodeRabbit locally against upstream `main`; resolve critical and warning
   findings or record a verified reason for declining them.
9. Build the changed Rust library in an isolated downstream binding checkout.
10. Re-run the Grynvault/Alby Hub focused Go tests and frontend production
    build without pointing at a live wallet or creating any Lightning action.
11. Confirm the application starts only against synthetic or empty test data;
    no invoice, probe, HTLC, payment, swap, rebalance, or channel operation is
    permitted in this regression pass.
12. Review the final diff and commit hash, then push one scoped follow-up commit
    to the existing PR branch.
13. Read back PR head, changed files, checks, CodeRabbit state, and conflict
    status from GitHub.

## Acceptance evidence — 2026-10-18

- Final diff: exactly `src/event.rs`, `src/payment/bolt11.rs`, and
  `src/payment/store.rs`; `git diff --check` and formatting check passed.
- Rust default library suite: 53 passed, 0 failed.
- Rust all-feature library suite: 63 passed, 0 failed.
- Warning-denied default library build passed.
- UniFFI development and optimized release builds passed.
- Final local CodeRabbit review: 0 findings across the full PR file set.
- Native UniFFI release libraries built for both `aarch64-apple-darwin` and
  `x86_64-apple-darwin`.
- The two libraries were combined into a universal Mach-O library and consumed
  from an isolated copy of the exact Hub-pinned `ldk-node-go` commit
  `3228bdbb7973a2743aee5e9f699e79afd217617b`.
- Hub focused tests (`./api`, `./lnclient/ldk`) passed.
- Hub complete Go test suite (`go test ./...`) passed.
- Hub HTTP backend and Wails-tagged desktop executables built successfully.
- Wails frontend TypeScript/Vite production build passed.
- An empty isolated `WORK_DIR` started successfully on a temporary local port;
  `/api/info` reported setup incomplete, node not running, and wallet locked;
  the process then shut down cleanly.
- The canonical application and wallet directory were not opened, copied, or
  modified. No invoice, probe, HTLC, payment, swap, rebalance, or channel
  operation was created.
- PR #115 readback shows head `8784ee0`, open, ready for review, mergeable, and
  conflict-free. CodeRabbit marked the duplicate-payment race addressed in that
  commit; the final local review reported zero findings. Five fork-originated
  upstream workflows reported `action_required`, meaning maintainer approval is
  still required before those jobs can run; none reported a test failure.

## Follow-up acceptance evidence — 2026-10-19

- Commit `70965f2e58936c0893c48ee0cc1788b84164a4e2` binds the persisted hash and
  SHA-256 of the stored preimage to the live claimable payment hash before
  `claim_funds` can run.
- Focused missing, corrupt, mismatched, and valid preimage/hash cases passed.
- Rust default library suite: 53 passed, 0 failed.
- Rust all-feature library suite: 63 passed, 0 failed.
- Warning-denied default library check and UniFFI development/release builds
  passed; local CodeRabbit review reported zero findings.
- Fresh universal macOS native bindings were consumed by the exact pinned Hub
  copy. Focused and complete Go tests, HTTP/Wails backend builds, and both
  frontend production builds passed.
- An isolated empty-work-directory startup returned HTTP 200 with setup
  incomplete, node stopped, and wallet locked, then shut down cleanly.
- GitHub readback shows PR #115 at head `70965f2`, open, ready for review,
  mergeable, and conflict-free. No live wallet or Lightning operation was used.
- CodeRabbit's remote check reports success and all five inline review threads
  are resolved.
- The public PR description was updated and read back with the current safety
  properties, 53/63 Rust test counts, downstream Hub evidence, and
  preimage/hash claim validation.
- Five upstream workflows remain `action_required`; an Alby maintainer must
  authorize those fork-originated CI runs. This is not a test failure and no
  contributor-side retry can replace maintainer authorization.

## Known non-regression warnings

- UniFFI-generated Rust code warns about deprecated
  `FeeRate::from_sat_per_vb_unchecked`; the same warning reproduces at the
  untouched `ed72fe9` baseline.
- Hub links the existing Bark FFI archive built for macOS 15.4 while the local
  target is macOS 15.0; the linker warns but all focused/full tests and both
  executables complete successfully.
- The frontend retains existing third-party/build warnings for lottie direct
  `eval`, an oversized output chunk, old Browserslist data, and a browser-
  externalized `crypto` import. This payment-safety patch does not alter those
  dependencies or frontend code.
- CodeRabbit's generic docstring-coverage notice reports 47.92% because it also
  counts private helpers and tests. Every new public Rust API is documented,
  the crate's default `deny(missing_docs)` gate passes, and `cargo doc
  --no-deps` completes. We do not add low-value comments to private/test code
  solely to satisfy that heuristic.
- Warning-denied rustdoc is currently blocked only by an existing bare URL in
  `src/lib.rs` that is unchanged from upstream and outside this payment-safety
  patch.

## Submission gate

Do not push if any of the following remains true:

- a confirmed finding lacks a regression test;
- an existing exact-channel, recovery, persistence, or backward-compatibility
  test regresses;
- the downstream binding or Hub build cannot consume the change;
- local review reports an unresolved critical or warning finding;
- the working tree contains unrelated changes;
- validation touched a live wallet or produced a Lightning operation.

No tag, package, signature, publication, deployment, or live rebalance is part
of this review-fix cycle.

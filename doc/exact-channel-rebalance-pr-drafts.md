# Exact-channel rebalance pull request drafts

Status: locally prepared for owner review. Do not open any upstream pull request
until the release checklist's explicit approval gate is satisfied.

Open in dependency order: LDK Node, LDK Node Go, then Hub. Replace each
`<dependency PR>` marker with the final upstream URL only after that PR exists.

## 1. `getAlby/ldk-node`

Title: `feat(payments): add exact-channel circular payment primitives`

### Root cause

The existing high-level payment API can route through an unrestricted first
hop, while ordinary invoice claim handling does not bind a circular payment to
one required receiving channel. UI selection or post-payment observation cannot
provide atomic enforcement.

### Summary

- add a non-paying circular route quote that binds every path to one exact local
  first-hop channel and one exact local receiving channel;
- persist a constrained inbound payment before sending and refuse to reveal its
  preimage unless every received MPP part reports the selected channel;
- submit only the reviewed serialized route, with deterministic operation and
  outbound payment IDs and no pathfinding retry;
- recover prepared and in-flight operations without creating a second payment;
- persist bounded rejection diagnostics and reconcile two exact durable payment
  records into one terminal result.

### Security model

The quote is pathfinding-only. Preparation persists the invoice, preimage, exact
channel constraints, fee cap, and operation binding without sending an HTLC.
Execution re-reads both channels, validates the complete serialized route and
visible path summary, persists the outbound record, and calls fixed-route send
once. The recipient claims only after amount, metadata, preimage state, and all
receiving-channel identities match. Missing or mixed identity fails closed.

### Compatibility and persistence

The new payment-store fields use optional high-numbered TLVs so existing records
remain readable. Legacy terminal records remain valid; diagnostics are additive.
The public UniFFI surface is extended without changing existing methods.

### Verification

- native library tests: `47 passed`;
- UniFFI library tests: `57 passed`;
- real in-memory HTLC tests cover exact success, wrong-channel rejection,
  mixed-channel MPP rejection, fixed-route MPP, prepared restart, in-flight
  restart, duplicate submission, and terminal replay;
- patch secret scan: no findings;
- strict native and UniFFI builds pass with warnings denied (apart from the two
  already documented UniFFI deprecations);
- the production `rustls` dependency is updated to `0.23.46`, closing the two
  2026 RustSec advisories in the production tree; six remaining advisories and
  four maintenance warnings are confined to existing development/build-only
  paths (`ldk-node 0.6.2`, `electrsd`, and `corepc-node`);
- the full integration suite requires the repository's external
  `BITCOIND_EXE`/Electrs fixtures; without them, its 35 tests fail at fixture
  startup rather than exercising this change;
- upstream baseline also fails current `cargo fmt --check` and all-target Clippy
  at the existing benchmark `Arc<Arc<Node>>` mismatch; the feature library and
  test targets compile and pass.

### Review focus

- pre-claim receiving-channel enforcement in `src/event.rs`;
- fixed-route decoding and summary cross-checks in `src/payment/bolt11.rs`;
- restart and duplicate-payment behavior;
- backward-compatible payment-store serialization.

### Changed-file scope

The change is confined to the UniFFI definition, node builder/data store,
payment/event implementation, payment-store serialization, public types, and
their focused tests: `bindings/ldk_node.udl`, `src/builder.rs`,
`src/data_store.rs`, `src/event.rs`, `src/lib.rs`, `src/payment/`, and
`src/types.rs`.

Base: `553a2105b38da88fc5d0900007d30e9bd9669900`

Branch head: `ed72fe992174965bc917d1dc6c6bee1dc7368057`

## 2. `getAlby/ldk-node-go`

Title: `feat(bindings): expose exact-channel circular payment APIs`

Depends on: `<ldk-node PR>`

### Root cause

Hub cannot consume the new exact-channel primitives until the mechanically
generated Go/C bindings and matching native library expose the same reviewed
LDK source interface.

### Summary

- regenerate the Go and C bindings for circular route quoting, preparation,
  fixed-route submission, receiving-channel evidence, and rejection reasons;
- include a universal macOS library built from the exact reviewed LDK source;
- preserve existing exported APIs.

### Reproducibility evidence

- generated Go source matches a fresh `uniffi-bindgen-go` generation byte for
  byte;
- generated header matches apart from generator whitespace;
- universal library architectures: `x86_64 arm64`;
- library size: `43,212,288` bytes;
- SHA-256:
  `879ea591a5f9636425ce56097ca0473df3fc05e62ef94135428a78a45087adcf`;
- `go test ./...` passes;
- secret scan reports no findings.

### Changed-file scope

Only generated binding outputs and the matching universal macOS library change:
`ldk_node/ldk_node.go`, `ldk_node/ldk_node.h`, and
`ldk_node/universal-macos/libldk_node.dylib`.

Base: `25822674b495b8b0f7a8882f6b19a64b70d02396`

Branch head: `3228bdbb7973a2743aee5e9f699e79afd217617b`

## 3. `getAlby/hub`

Title: `feat(rebalance): add fail-closed exact-channel circular rebalancing`

Depends on: `<ldk-node PR>`, `<ldk-node-go PR>`

### Root cause

The retired provider-backed flow cannot satisfy this local use case, and the
stock Hub payment surface has no atomic way to bind both the first and final
channels of a circular payment. A UI-only preference would therefore be unsafe.

### Summary

- add a non-paying exact-channel local route quote with a five-minute review
  record and fingerprint-bound route, channels, amount, fee cap, and balance
  evidence;
- require a quote-specific typed confirmation before the only value-moving
  owner entrypoint;
- serialize execution globally, persist acquired/prepared/submitted phases, and
  recover only the same deterministic fixed-route operation;
- reconcile exact inbound and outbound payment-store records before releasing
  the execution lock;
- surface bounded receiving-channel rejection diagnostics without invoices,
  preimages, secrets, or serialized route bytes;
- keep the review dialog usable with long identifiers and narrow viewports.

### Access boundary

Quote, execute, status, and exact diagnostic routes are in the authenticated
full-owner API group and the desktop owner surface. They are absent from the
read-only API and NWC method dispatch. The legacy provider execution endpoint
remains disabled because its incoming-channel atomicity is not proven.

### Database changes

The migrations add reviewed local quote records, a single-active-execution
partial unique index, opaque route bytes, recovery identifiers and timestamps,
terminal evidence, channel balance/reserve snapshots, and bounded failure
diagnostics. SQLite and PostgreSQL migration copy tests pass. A dedicated test
also applies and rolls back every rebalance migration on both engines.

### Verification

- `go test ./...` and the focused owner API/Wails race suite pass with Go
  1.27.2;
- `govulncheck ./...` reports zero reachable vulnerabilities after the targeted
  `x/net` and gRPC updates;
- frontend lint, TypeScript, formatting, and HTTP production build pass;
- production frontend dependency audit reports zero vulnerabilities;
- the patch secret scan has one reviewed false positive: a Go field named
  `Secret` in a runtime equality check, with no literal credential or token;
- `go vet ./...` reaches only the existing upstream Phoenix date-layout typo;
- existing Bark deployment-target linker warnings remain unchanged;
- a local Wails application build cannot be repeated while the owner's running
  Hub holds the single-instance port because Wails executes the app before
  generating bindings; the PR checks must therefore repeat that build in an
  isolated CI runner;
- a complete final CodeRabbit review found one transient-status polling issue;
  after the retry fix and regression checks, the second complete review reported
  zero findings.

### Screenshots

Attach privacy-safe fixture screenshots for 320, 375, 768, and 1280 CSS pixel
widths showing pre-quote and quote-review states. Do not attach live wallet
identifiers or balances.

Screenshot provenance and the user flow are recorded in
`doc/exact-channel-rebalance-usage.md` and
`doc/assets/exact-channel-rebalance/README.md` on the release-documentation
branch. The real component also passes explicit 200% browser zoom with synthetic
identifiers and no payment action.

### Changed-file scope

The change is confined to owner API models/handlers/tests, rebalance database
models and append-only migrations/tests, authenticated HTTP and Wails routing,
the LDK adapter/interface, exact dependency pins, and the two channel-dialog
components listed in the branch diff. It adds no NWC method or read-only API
route.

Base: Alby Hub `v1.24.3`,
`5d5e696dfe7d430a365d48939c394376e078a56b`

Branch head: `cd6bc1e399ba256adb113bbaf44d659c5b22a170`

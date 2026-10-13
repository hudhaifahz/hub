# Exact-channel rebalance pull request drafts

Status: locally prepared for owner review. Do not open any upstream pull request
until the release checklist's explicit approval gate is satisfied.

Open in dependency order: LDK Node, LDK Node Go, then Hub. Replace each
`<dependency PR>` marker with the final upstream URL only after that PR exists.

## 1. `getAlby/ldk-node`

Title: `feat(payments): add exact-channel circular payment primitives`

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

- native library tests: `41 passed`;
- UniFFI library tests: `51 passed`;
- real in-memory HTLC tests cover exact success, wrong-channel rejection,
  mixed-channel MPP rejection, fixed-route MPP, prepared restart, in-flight
  restart, duplicate submission, and terminal replay;
- secret scan: no findings;
- `cargo audit`: four advisories and three maintenance warnings are confined to
  existing development/test-only dependency paths, not the production library
  tree;
- upstream baseline also fails current `cargo fmt --check` and all-target Clippy
  at the existing benchmark `Arc<Arc<Node>>` mismatch; the feature library and
  test targets compile and pass.

### Review focus

- pre-claim receiving-channel enforcement in `src/event.rs`;
- fixed-route decoding and summary cross-checks in `src/payment/bolt11.rs`;
- restart and duplicate-payment behavior;
- backward-compatible payment-store serialization.

Branch head: `2dc742a8b2057519e22d757f510e5cc880731b2c`

## 2. `getAlby/ldk-node-go`

Title: `feat(bindings): expose exact-channel circular payment APIs`

Depends on: `<ldk-node PR>`

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
- library size: `43,360,288` bytes;
- SHA-256:
  `cf43f66815c44c0796402faafaaa3ed70201d036124438e63f541ec6a05208bd`;
- `go test ./...` passes;
- secret scan reports no findings.

Branch head: `a381cacb18562dff960f24a12f43232633f8bc25`

## 3. `getAlby/hub`

Title: `feat(rebalance): add fail-closed exact-channel circular rebalancing`

Depends on: `<ldk-node PR>`, `<ldk-node-go PR>`

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

- `go test ./...` and `go test -race ./...` pass with Go 1.27.2;
- `govulncheck ./...` reports zero reachable vulnerabilities after the targeted
  `x/net` and gRPC updates;
- frontend lint, TypeScript, HTTP build, and Wails production build pass;
- production frontend dependency audit reports zero vulnerabilities;
- the full branch secret scan has one reviewed false positive: a Go field named
  `Secret` in a runtime equality check, with no literal credential or token;
- `go vet ./...` reaches only the existing upstream Phoenix date-layout typo;
- existing Bark deployment-target linker warnings remain unchanged.

### Screenshots

Attach privacy-safe fixture screenshots for 320, 375, 768, and 1280 CSS pixel
widths showing pre-quote and quote-review states. Do not attach live wallet
identifiers or balances.

Screenshot provenance and the user flow are recorded in
`doc/exact-channel-rebalance-usage.md` and
`doc/assets/exact-channel-rebalance/README.md` on the release-documentation
branch.

Branch head: `1fb6c8cd0fcb2fbee48c888c6274e2659d74c169`

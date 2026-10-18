# Exact-channel upstream contribution package

Status: **PREPARED — NO PULL REQUEST CREATED**

Prepared: 2026-10-18

This is a three-repository contribution and must be reviewed in dependency
order. The prepared branches already exist on the contributor fork from prior
work, but they are behind current upstream and are not ready to submit as-is.
No branch was rewritten or pushed while preparing this package.

Grynvault branding, guided migration, release packaging, PGP workflow, updater
policy, and companion service are deliberately excluded from the upstream
feature contribution.

## Current branch audit

| Order | Repository            | Base branch | Prepared head                                           | Upstream at audit | Divergence                                 | Current merge conflict                                                     |
| ----- | --------------------- | ----------- | ------------------------------------------------------- | ----------------- | ------------------------------------------ | -------------------------------------------------------------------------- |
| 1     | `getAlby/ldk-node`    | `main`      | `hudhaifahz:feat/exact-channel-rebalance` at `2dc742a`  | `553a2105`        | 45 upstream-only / 24 feature-only commits | `src/event.rs`                                                             |
| 2     | `getAlby/ldk-node-go` | `master`    | `hudhaifahz:feat/exact-channel-rebalance` at `a381cac`  | `25822674`        | 3 upstream-only / 10 feature-only commits  | generated `ldk_node/universal-macos/libldk_node.dylib`                     |
| 3     | `getAlby/hub`         | `master`    | `hudhaifahz:feat/exact-channel-rebalance` at `1fb6c8cd` | `5d5e696d`        | 14 upstream-only / 22 feature-only commits | `db/migrations/migrate.go`, `go.mod`, `go.sum`, `lnclient/ldk/ldk_test.go` |

The conflict audit used Git's three-way merge machinery without checking out a
merge or changing a branch. The counts and conflicts must be refreshed
immediately before any rebase or PR.

## Current feature-branch baseline

These checks ran against the existing feature heads before rebasing them onto
current upstream:

- LDK Node: `cargo test --lib` passed all 41 tests. `cargo fmt --check` did not
  pass; it reports formatting drift in `src/lib.rs` and `src/liquidity.rs`.
  Formatting must be corrected on the refreshed contribution branch without
  mixing unrelated source changes.
- LDK Node Go: `go test ./...` passed; the repository currently contains no Go
  test files, so target-library binding smoke tests remain a release and PR
  requirement.
- Hub: focused `api`, `lnclient/ldk`, `wails`, and database-migration tests
  passed. Frontend ESLint, TypeScript, Prettier, and the Wails production build
  passed. The build retained the inherited oversized-chunk and `lottie-web`
  direct-`eval` warnings. Local Go linking also reported the known Bark native
  objects' macOS 15.4 deployment target while the ad-hoc test link used 15.0.

These results establish a baseline only. They do not clear the rebase
conflicts, dependency provenance, full CI, or maintainer-review gates.

## Submission order and gates

### 1. LDK Node

Proposed title:

```text
feat(payments): add recoverable exact-channel circular payments
```

Proposed body:

```text
## Summary

Add a fail-closed circular-payment flow that can quote an exact local first and last channel without sending a probe, durably prepare the reviewed route, submit it idempotently, and reconcile both payment legs after restart.

## Safety properties

- no fallback to a different local first or last channel;
- quote construction is non-paying;
- serialized route data is persisted and bound to reviewed parameters;
- duplicate or conflicting operation IDs fail closed;
- restart recovery distinguishes prepared, submitted, pending, succeeded, failed, and recovery-required outcomes;
- receiving-channel observations and rejection diagnostics are recorded without exposing route bytes.

## Testing

- unit coverage for exact-hop route selection and rejection;
- persistence and restart recovery tests;
- duplicate/concurrent submission tests;
- terminal reconciliation tests for both circular legs;
- formatting, clippy, and the repository's full supported test suite.

## Dependency order

This is the first PR in a three-part stack: ldk-node, generated ldk-node-go bindings, then Alby Hub integration.
```

Before submission:

- rebase onto fresh `getAlby/ldk-node:main` in a disposable local branch;
- resolve `src/event.rs` semantically, preserving current upstream event
  behavior and exhaustive matching;
- split or squash the 24-commit development history into maintainer-sized
  commits if requested;
- run `cargo fmt --check`, clippy with warnings denied, focused payment tests,
  the full library suite, and repository integration tests available in CI;
- regenerate UniFFI definitions only after the Rust API is final.

### 2. LDK Node Go bindings

Proposed title:

```text
feat(bindings): expose exact-channel circular payment APIs
```

Proposed body:

```text
## Summary

Regenerate the Go bindings and native libraries for the accepted ldk-node exact-channel circular-payment API.

The bindings expose non-paying route quotes, durable preparation, idempotent submission, reconciliation, receiving-channel observations, and structured rejection diagnostics needed by Alby Hub.

## Provenance

- generated from the exact accepted ldk-node commit;
- native artifacts built on their target operating systems;
- architecture, exported-symbol, binding-smoke, and SHA-256 evidence recorded for every library.

## Dependency

Depends on: <accepted ldk-node PR/commit>
```

Before submission:

- do not merge the conflicting macOS binary;
- regenerate every committed binding and native library from the accepted Rust
  commit, including Linux x86-64 and Windows x86-64—not only macOS;
- verify clean generated diffs, Go compilation, binding smoke tests, exported
  symbols, architecture, and hashes;
- replace the placeholder dependency line with the accepted Rust PR and exact
  commit.

### 3. Alby Hub

Proposed title:

```text
feat(ldk): add fail-closed exact-channel circular rebalance
```

Proposed body:

```text
## Why

The existing peer-targeted rebalance path cannot guarantee which local channel is used when a peer has parallel channels, and ordinary pathfinding may select a different first or returning channel. Node operators need a reviewed, fail-closed way to move liquidity between two exact channels on the same embedded-LDK node.

## What

- choose exact outgoing and incoming channel IDs;
- find a non-paying local circular route;
- review principal, estimated fee, fee cap, projected channel balances, exact first/final hop, expiry, and a route fingerprint;
- execute only a fresh persisted quote with exact confirmation;
- serialize execution globally and prevent duplicate submission;
- persist crash-safe preparation/submission phases and reconcile terminal outcomes;
- record receiving-channel evidence and fail closed on an incorrect or unidentified return channel;
- disable automatic fallback to different channels;
- keep route bytes and payment secrets out of API/UI/log output.

## Limits

This cannot create liquidity or guarantee a public route. It is implemented only for embedded LDK backends that expose the required exact-channel capabilities; other backends fail closed.

## Testing

- quote and execution validation, stale-state, expiry, fingerprint, fee-cap, and exact-channel mismatch tests;
- concurrency and duplicate-submission tests;
- crash/restart reconciliation and database migration rollback tests;
- LDK capability and receiving-channel tests;
- desktop and HTTP route coverage;
- responsive review-dialog validation.

## Dependencies

- Depends on: <accepted ldk-node PR/commit>
- Depends on: <accepted ldk-node-go PR/commit>
```

Before submission:

- rebase onto fresh `getAlby/hub:master` after both dependency PRs are accepted
  or maintainers explicitly agree to a stacked review;
- resolve migration ordering without renumbering or modifying any accepted
  upstream migration;
- replace the forked module pin with the accepted upstream binding revision;
- separate unrelated dependency-security bumps if upstream already contains
  them;
- resolve the LDK test conflict against current upstream behavior;
- run Go tests, race tests, PostgreSQL migration tests, `go vet`, frontend lint,
  TypeScript, formatting, HTTP/Wails builds, and all-platform desktop CI;
- give maintainers screenshots that contain synthetic identifiers only.

## Maintainer review notes

The Hub diff is currently about 5,000 added lines across 27 files. Before
submission, offer maintainers either:

1. the complete end-to-end PR after the two dependencies land; or
2. a short reviewed stack: capability/model changes, persistence/recovery,
   API/LDK integration, then UI.

Do not publish three overlapping Hub PRs without maintainer agreement. Keep the
invariants visible in every review unit: exact channels, non-paying quote,
route-byte opacity, one reviewed execution, idempotency, and crash-safe
reconciliation.

## Final owner gate

Creating any of these pull requests is a public action. Before creation, show
the owner the refreshed base and head SHAs, final diff stats, dependency links,
test evidence, exact PR title/body, and whether the PR is draft or ready for
review. Opening one PR does not authorize either dependent PR.

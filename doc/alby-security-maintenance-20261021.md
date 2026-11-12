> Later maintenance snapshot: [2026-10-22](/Users/kode/Development/alby-security-maintenance-20261021/hub/doc/alby-security-maintenance-20261022.md). The record below preserves the October 21 evidence.

# Alby security maintenance — 2026-10-21

**PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Recorded 2026-10-21T09:23:46.319998-07:00. This is a maintenance snapshot, not publication or wallet-operation approval.

## Detected upstream changes

The official latest release and newest tag remain [v1.24.3](https://github.com/getAlby/hub/releases/tag/v1.24.3), published 2026-10-16T08:00:53Z, peeling to `5d5e696dfe7d430a365d48939c394376e078a56b`. The fully enumerated GitHub API returned 51 releases and 71 tags, without next-page links.

The official default branch is `master`. Its captured snapshot is [60dbcc3befe7629dba2a14a2e60df1d4620fe72e](https://github.com/getAlby/hub/commit/60dbcc3befe7629dba2a14a2e60df1d4620fe72e), seven commits newer than the prior baseline. It includes the Go CA-distrust mitigation, rebuilt upstream LDK TLS bindings, Postgres settled-transaction lookup, LND fee-policy selection, Alby SDK/lightning-tools updates, a CLN hold-plugin startup fix, and the x/crypto dependency update. Two commits arrived during this run and were included before final validation.

## Exact local candidates

All three worktrees use branch `codex/alby-security-20261021`.

| Component | Exact source | Location |
| --- | --- | --- |
| Hub tested code/build | `5a6b05f80f853998214bd432d9b59163fc159a71` | `/Users/kode/Development/alby-security-maintenance-20261021/hub` |
| LDK TLS source | `22c66d60ffe13e6dd0532d7f22501aa55b6a769f` | `/Users/kode/Development/alby-security-maintenance-20261021/ldk-node` |
| Rebuilt Go/native binding | `ea7860a8558930f501df672616b88c0a72c6c584` | `/Users/kode/Development/alby-security-maintenance-20261021/ldk-node-go` |

The original fork was clean at `4c5dc1c1ce7e8284cd5ad49b87d5dab95f09776c` on `codex/coinos-inbound-liquidity`; it was not switched, merged, or edited. The maintenance Hub includes that whole owner-fork head and the exact upstream snapshot. Exact-channel APIs, Hub rebalance/migration logic, desktop route-redaction implementation, Rust payment/event source, and the generated Go/C API remain unchanged from their reviewed bases. The source ancestry and targeted byte-for-byte Git diffs were verified.

Hub requires Go 1.26.10 and selects patched Go 1.27.3; the Docker builder is also pinned to 1.27.3. Both native slices were rebuilt with Rust 1.91.1, `--locked --release --features uniffi`, and macOS deployment target 15.4, then combined with `lipo`. The source preserves the exact-channel feature and uses rustls-webpki 0.103.16 from upstream security commit `87ea53f2e1f32cc7487521776ae10f9e3850a891`.

Universal raw library SHA-256: `bb487c9c2640e4c053aa508998f459b6b03216be4592ea080a64a4e9b5476f78`. Each architecture retains all six circular-payment FFI/checksum symbols and all eight public distrusted-SPKI entries. Install name remains `@rpath/libldk_node.dylib`. Other platform binding binaries were inherited, not rebuilt or validated; macOS remains the supported fork target.

The replacement pin is `github.com/hudhaifahz/ldk-node-go v0.0.0-20261021161128-ea7860a85589`, with module checksum `h1:OjpkVT4fzISDV4Ej79z9v1hEJ11oWoGf3yFJfmtsJ5o=`. This commit is local and unavailable remotely. Validation used a temporary file-based Go module proxy built from that clean binding worktree, with a process-local sumdb exclusion for this unpublished fork module. No global Go environment or production security setting was changed. Before any downstream public build, obtain owner approval for dependency publication and independently verify the remote module commit/checksum/native-library hash.

## Validation and limitations

| Check | Evidence/result |
| --- | --- |
| Complete Go suite | Pass against final source, Go 1.27.3 |
| Go race tests | Pass: API, Wails, guided migration, DB migrations, transactions |
| PostgreSQL regressions | Pass: transactions, DB migrations, db_migrate, using a new disposable local server; server stopped afterward |
| Wails-tag root/Wails tests | Pass with one explicit exclusion described below |
| Go vet/module verification/HTTP build | Pass |
| Go vulnerability scan | Zero reachable vulnerabilities; two imported-package and one required-module non-reachable findings still reported |
| Frontend | Frozen install, lint, types, formatting, HTTP and Wails builds pass; production audit: zero vulnerabilities across 135 dependencies |
| Rust exact-channel and persistence tests | 10 route/duplicate/CLTV tests and 2 legacy/RC.6 persistence tests pass |
| Native TLS verifier regression | Upstream `distrusted_spki` test passes; 30 verifier source files byte-match the registry crate used in the build |
| Rust formatting / locked production graph | Pass; production uses fixed verifier 0.103.16 and not legacy verifier 0.101.7/native TLS |
| Universal macOS candidate | Build and static inspection pass: both architectures, independent bundle/data identity, macOS 15.4 floor, source commit/label, pinned library metadata, reviewed rpath, ad-hoc integrity only |
| Packaging/source boundary checks | Pass; unapproved-tag packaging is rejected; source/workflow checks and shell syntax pass |
| Secret scans | Hub source, LDK patch, and supported artifact text scans pass; binary scanning has normal scanner-coverage limits |
| Private source archive/SBOM | Prepared from exact source; CycloneDX contains 1107 components |

The full Wails-tag invocation failed only at `TestReleaseAcceptanceNetworkSandboxIntegration`: its required loopback listener `127.0.0.1:21421` was already occupied. The existing service was not stopped, inspected, or changed. The remaining Wails-tag tests pass with `-skip '^TestReleaseAcceptanceNetworkSandboxIntegration$'`. This integration gate must run on an isolated host with the required port available; it is not accepted as passing here. No installation or runtime wallet acceptance was performed.

The raw Rust audit exits nonzero: 12 advisory entries and four unmaintained warnings remain. Two entries are RUSTSEC-2026-0381 for production `webpki-roots` 0.25.4 and 1.0.9; the other flagged package versions are absent from the production normal/build graph. The new intermediate-SPKI verifier is present and its regression passes, but that does not make the root-store package audit clean. Retain this security hold for review/remediation. No zero-unresolved-Critical/High, complete TLS-stack remediation, independent audit, or public-release-ready claim is made. Go scanning does not assess native Bark or OS TLS trust.

CodeRabbit reviewed the initial Hub merge and correctly found that the old replacement would mask the new LDK pin; the rebuilt library pin resolves that issue. A second Hub pin/provenance review completed with zero findings across six files, and the LDK TLS patch review completed with zero findings across two files. The final incremental crypto/CLN review was blocked by the free OSS quota (reported 49-minute wait); its source diff was inspected manually and all affected Go checks were rerun. There is no completed CodeRabbit review of that final incremental diff.

## Prepared artifacts and next gates

The untagged, ad-hoc-sealed acceptance wrapper at `/Users/kode/Development/alby-security-maintenance-20261021/hub/build/bin/Grynvault Rebalance.app` embeds `owner-acceptance-5a6b05f80f85` and the exact tested source above. It has no publisher identity or notarization, intentionally lacks the production app entitlements, and must not be installed, uploaded, published, or treated as a release artifact. It was not launched. No DMG, release tag, PGP signature, public release, PR, push, deployment, or installation was created.

- Source archive SHA-256: `c6dba6051975acc85a4cb6d939c3102a4aa7b1e2f8d9f1449df97e7ab937f9e9`.
- SBOM SHA-256: `1fb6848e926708ddac192fc85593d038426a1cf5fb7a8c848860aa23630c49d4`.
- Full local evidence and command results: `/Users/kode/Development/alby-security-maintenance-20261021/evidence`.
- Evidence index with file hashes: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/evidence-hashes-final.json`.
- Final Go matrix: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/go-matrix-final.json`.
- Package inspection: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/macos-build-snapshot-inspection.log`.
- Rust audit and production-graph assessment: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/rust-audit-assessment.json`.

Review the remaining Rust dependency findings, close the isolated sandbox-test gate, and refresh the final incremental review before proposing release promotion. All standing RC.4/guided-import/RC.5 acceptance gates remain in force; their recorded public/installed state was not independently refreshed by this source-only monitor. Stop for separate owner approval before each push, dependency publication, tag, PGP signature, PR, installation, deployment, or public change.

No wallet secrets, real/copied wallet data, wallet processes, invoices, network probes, HTLCs, payments, swaps, rebalances, or channel operations were accessed or created. Mock/fixture-only migrations and test data were used.

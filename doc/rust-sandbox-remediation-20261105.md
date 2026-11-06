# Rust and sandbox remediation — 2026-11-05

**PREPARED AND TESTED LOCALLY — NOT PUSHED, RELEASED OR INSTALLED.**

Recorded 2026-11-06T01:39:24.629299+00:00. Owner authorized resolving Rust findings and the sandbox test, explicitly requiring approval before any push. Existing wallet/signing/publication boundaries remain. This is not a claim that every raw audit entry is cleared or that a release is ready.

## Outcome and exact source

| Repository | Branch | Exact code/binary commit |
| --- | --- | --- |
| LDK source | codex/alby-rust-sandbox-20261105 | d8c68f5534e5aaa9d4f994751bc763e3938c2c22 |
| Go binding/native library | codex/alby-rust-sandbox-20261105 | 1551acb4e724cab38f641dde391f0a608201f590 |
| Hub tested/package source | codex/alby-rust-sandbox-20261105 | 3e4ed99996f3f3669569bcd9a1e67448e201986c |

All three are separate worktrees under `/Users/kode/Development/alby-rust-sandbox-20261105`. Original monitor/source branches were not modified. A later Hub documentation-only commit records this report.

### Production Rust remediation

Updated locked Tokio1.53.2→1.53.3, Electrum0.24.1→0.24.2 (uses roots1.x), root store1.0.9→1.0.10 and pulldown-cmark0.13.4→0.13.6. Raised Tokio/Electrum manifest minimums. Removed obsolete automatic test binary download/archive features, eliminating vulnerable zip/minreq TLS downloader dependencies; existing CI already obtains checksum-verified test binaries through its explicit script and BITCOIND_EXE/ELECTRS_EXE. Integration coverage remains, rather than deleting the old VSS compatibility test. No downloader or wallet binary was run/downloaded during this task.

[Root-store advisory](https://rustsec.org/advisories/RUSTSEC-2026-0381.html) and [Tokio advisory](https://rustsec.org/advisories/RUSTSEC-2026-0386.html) fixed versions are selected. The normal/build graph with UniFFI across all targets contains ZERO advisory-listed vulnerabilities, and the Tokio unsoundness warning is gone. A direct root-store regression excludes the distrusted emSign G3 root at1.0.10, and fails against1.0.9, proving it detects the original trust-store defect. Previous intermediate-key verifier mitigation stays intact.

**Raw Rust audit remains NONZERO:**8 entries all flow only through intentionally pinned `ldk-node-062 = =0.6.2` under `cfg(vss_test)`. They cover old h2/rustls/webpki/roots in that historical compatibility reference, absent the production normal/build graph.5 unmaintained warnings remain (basic-toml/bincode/paste and test-only pin-utils/rustls-pemfile). These are explicitly retained residuals, not suppressed/waived or claimed repaired. Retiring/replacing that historical reference is separate from fixing the shipped library; do not claim a clean repository-wide audit or zero residual findings. No audit ignore configuration added. Fresh audit stderr has no yanked-registry timeout; earlier incomplete registry check is superseded by this result.

### Sandbox test remediation

Changed ONLY the synthetic helper to bind IPv6 `::1` at the SAME production lock ports. The production IPv4 lock addresses and production sandbox profile are byte-for-byte unchanged. This avoids the occupied IPv4 sockets and does not stop, inspect or connect to the owner's existing process. The child still runs under the real sandbox profile and proves external-network/non-loopback bind denial, local ownership-port communication and Unix IPC. No skip/exclusion or permissive policy was introduced.

The isolated integration, complete Wails-tag root/package suite and sandbox race suite PASS, including after the rebuilt native dependency. This closes the former occupied-port TEST gate. It does not close real copied-wallet, guided-import, installation or runtime acceptance gates. An initial alternative127.0.0.2 synthetic bind was unavailable on this host; no host network interface was changed. IPv6 loopback required no host configuration changes.

## Native and Hub verification

Rebuilt arm64 and x86_64 with Rust1.91.1, locked release +UniFFI and macOS15.4. Export sets match491 FFI/metadata symbols in EACH architecture; generated Go/C APIs and all Rust payment/event/persistence/binding source unchanged. Fresh10 exact-channel route/duplicate/CLTV tests +2 persistence tests pass, and all integration tests compile (`cargo check --tests` only, not executed). Rust formatting passes.

New universal raw library SHA256 **19f9cbea84af10665fe10370a7c1d6e586a2cf8f9cd23a0458ab7e069e411d3f**, size50,571,104 bytes; `@rpath/libldk_node.dylib` install name preserved. Binding version **v0.0.0-20261106012124-1551acb4e724**, local module h1 **h1:v0oaspypTOoU23I0KzPJDsBuw1N/iSZ/SOX85Qi6v+I=**. Exact source/binary/API proof recorded in native-provenance.json. All other platform library bytes remain inherited/unvalidated; no non-macOS binary remediation claim.

Hub pins that exact replacement version and expected native hash in BOTH builders and inspector. Private process-local file proxy/GONOSUMDB exception remains narrow to unpublished custom binding; no global/production security config changed. This local resolution is not remote-build reproducibility acceptance.

Full final Go suite, API/HTTP/db/migration/import/LDK/Wails race, vet/module verify/HTTP build, full Wails-tag suite WITHOUT exclusions all pass. Rust source does not alter database/import schema; fresh SQLite checks pass and November5 PostgreSQL fixture evidence is retained for unchanged migration code, not claimed rerun here. Desktop builder repeats frozen install/lint/types/format; HTTP frontend build and production audit134 deps/zero findings pass. Final universal macOS build/static inspection passes exact `3e4ed99996f3f3669569bcd9a1e67448e201986c`/untagged acceptance label, independent identity/data folder, macOS15.4 floor, reviewed rpath, new native hash/ad-hoc integrity, new binding/compression provenance and MPTCP-disabled default in both slices. App/HTTP binary never launched. Acceptance wrapper is not a distributable installer and intentionally lacks production sandbox entitlements.

CodeRabbit completed Rust Cargo changes and sandbox helper changes with ZERO findings. Final automated review of5 native pin/package files hit OSS rate quota (51-minute notice); manual review verified the exact computed hash/module pin. Subsequent two-file compression pin change is manually validated/tested, not claimed automatically reviewed. No billing/policy override or independent-audit claim. This review limitation remains recorded before Hub publication.

### Newly surfaced Go advisories

During validation, three Go advisories published2026-11-05T22:54:32Z became visible. Updated klauspost/compress1.18.7→1.20.2 for GO-2026-6865, then reran full/race/Wails/vet/module/HTTP and rebuilt the final package. Earlier c3e package/test evidence retained separately; FINAL package is `3e4ed99996f3f3669569bcd9a1e67448e201986c`.

Raw final govulncheck still exits3 on two Echo v4 whole-module entries, GO-2026-6872 and GO-2026-6875 (no patched v4 version listed). Application assessment: unlock endpoints share a constant global limiter bucket, with passing existing rotating-X-Forwarded-For regression; no Echo Proxy/ProxyWithConfig middleware is used. Those two vulnerable default/proxy conditions are absent current configuration. This is a documented scope assessment, not a vendor fix or clean raw scanner claim. Preserve raw results and revisit if limiter/proxy configuration changes.1 imported-package/3 required-module nonreachable findings also remain. Details/official JSON and echo-advisory-assessment.json retained.

Source/commit-range secret scans and boundaries pass. Broader artifact strings/text scan rawexit1 has4 exact known public-literal false positives (2 Go h1 checksums and same concatenated Go DNS/decimal/poll literal in2 slices); exact-byte comparison, redacted findings and assessment retained, no literal credential found/blanket suppression. Not exhaustive native binary analysis. Source archive SHA256 `918a91ee2b42b0afdcc6d4a40504025d542d8f16a8a1c5862bf586c2369a2b68`; SBOM SHA256 `0263006658a11cc1bf831807fa181207c2255126e43a8b91779a84ccf1b1806f`.

## Compatibility with Moritz's main-LDK direction

Read owner-referenced thread `codex://threads/01a16a03-1b40-78b3-905e-83afbc8d9c44`: it is preparing a focused contribution to **lightningdevkit/ldk-node**, rather than only Alby's fork, and excludes Hub packaging. That work is complementary. This remediation touches Cargo dependencies/test downloader policy, native builds/pins and synthetic sandbox isolation; it does not change the core exact-channel feature source. Separate topic branches/worktrees preserve the upstream preparation branch and original maintenance refs. No direct overwrite/conflict was introduced. Future upstream API/dependency adoption may require a rebase and binding regeneration; do not promise every future merge will be conflict-free. No message or PR was sent to that thread/upstream as part of this work.

## Concrete source-push packet — OWNER APPROVAL REQUIRED

This replaces the OLD unapproved two-dependency packet for this newly rebuilt candidate; do not execute the stale October22 SHAs.

1. Push exact LDK source **d8c68f5534e5aaa9d4f994751bc763e3938c2c22** to `https://github.com/hudhaifahz/ldk-node.git`, NEW branch `codex/alby-rust-sandbox-20261105`.
2. Push exact binding/native source **1551acb4e724cab38f641dde391f0a608201f590** to `https://github.com/hudhaifahz/ldk-node-go.git`, same NEW branch. Includes the rebuilt50.6MB macOS universal library and provenance note; generated APIs/other-platform binaries unchanged.

Both proposed remote refs were absent (exit2/no errors); recheck before execution, never force. LDK source push triggers existing CI, including synthetic/regtest integration jobs; automatic binding-publication job is guarded to main. Binding repository has no workflow. No source push executed here. Approval of these two actions would NOT authorize Hub push, main-LDK PR, tag/PGP, release/upload, installation/deployment or wallet operations. Require fresh remote exact-SHA and standard Go proxy/sumdb/module/native-hash verification after approved publication before proposing a Hub push. Publication approval does not waive retained raw-audit/review/runtime gates.

No wallet secrets/data/process/log access, copied wallet, real invoice, Lightning probe/HTLC/payment/swap/rebalance/channel operation, new chat, push/tag/PGP/PR/release/install/deploy/public change occurred in this work. Synthetic fixtures only. All evidence is private under `/Users/kode/Development/alby-rust-sandbox-20261105/evidence`; evidence-hashes.json indexes results. Original monitor/source worktrees and main-LDK preparation source untouched.

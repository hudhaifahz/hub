# Alby v1.24.4 maintenance — 2026-10-22

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Recorded 2026-10-22T11:06:10.591717-07:00. This supersedes the prior monitor snapshot without closing any installation, migration, signing, publication, or wallet-operation gate.

## Official upstream snapshot

- Latest official release: [v1.24.4](https://github.com/getAlby/hub/releases/tag/v1.24.4), published `2026-10-22T08:59:21Z`; tag object `feba0c0c1795e768564de42cd62b674f19b90d60`, peeled commit `c401324c065d50aea2a8b61006b6bb03b3ef776c`.
- Default branch: `master`, exact snapshot `6873fa512bde916c5c9214fa25d3a1258d9404fb`, one commit beyond the release tag.
- Official GitHub API fully enumerated 52 releases and 72 tags without next-page links. API refs and independently read git refs agree. No upstream wallet binary was downloaded, installed, or launched.
- Two new commits since the prepared `60dbcc3b` snapshot: CI chooses the latest Node 22 patch, and NWC `create_connection` names unsupported methods in its error. The latter is after v1.24.4. Existing upstream Go/LDK TLS fixes were already prepared on October 21.

## Exact prepared state

The existing isolated `codex/alby-security-20261021` branch was updated in `/Users/kode/Development/alby-security-maintenance-20261021/hub`. Tested code/build source is `40857768fb63b99a769b3b5e17139c3dd895f8b6`; a later documentation-only commit records this evidence. It contains the original owner-fork head `4c5dc1c1ce7e8284cd5ad49b87d5dab95f09776c` and the exact current upstream snapshot. The original `grynvault-rebalance-rc3` checkout remains clean and unchanged on `codex/coinos-inbound-liquidity`.

Workflow conflicts were resolved by preserving the fork's pinned actions, read-only permissions, guarded package flow, cache configuration, and no automatic release path, while adding `check-latest: true` to all three Node 22 consumers. The UI upstream-version label now says v1.24.4 and release-manifest provenance names exact master snapshot `6873fa51`.

The custom dependency remains local binding commit `ea7860a8558930f501df672616b88c0a72c6c584` from LDK source `22c66d60ffe13e6dd0532d7f22501aa55b6a769f`, pseudo-version `v0.0.0-20261021161128-ea7860a85589`, expected module hash `h1:OjpkVT4fzISDV4Ej79z9v1hEJ11oWoGf3yFJfmtsJ5o=`. Raw universal library SHA-256 remains `bb487c9c2640e4c053aa508998f459b6b03216be4592ea080a64a4e9b5476f78`. The library was reused byte-for-byte, not rebuilt unnecessarily. Rust payment/event source, generated Go/C API, Hub exact-channel/migration logic, and desktop route-redaction source were compared with their reviewed bases and remain unchanged.

An isolated Node **22.23.4 security release** was downloaded from nodejs.org and SHA-256 checked against its official SHASUMS256 manifest before extraction/use. Archive hash: `c9e1dbc748b33ed2d14f42445191ecaf14c56259c7a9a26d0042c49ec2b66428`. This is a private build tool under the evidence directory, not a wallet installation or global tool upgrade. Builds/tests use Go 1.27.3 and the existing process-local module proxy for the unpublished dependency.

## Current validation

| Check | Result |
| --- | --- |
| Complete Go suite, count=1 | Pass |
| Race tests | Pass: NWC controllers, API, Wails, guided migration, DB migrations |
| Fixture migration/continuity checks | Pass in the Go suite; no schema or migration changes. Prior PostgreSQL evidence retained; PostgreSQL was not rerun today |
| Wails-tag root/Wails tests | Full invocation fails only the sandbox integration's occupied port 21421; remaining tests pass with that explicit exclusion |
| Go vet, module verify, HTTP build | Pass |
| Go vulnerability scan | Zero reachable vulnerabilities; two imported-package and one required-module non-reachable findings remain |
| Frontend | Frozen install, lint, types, format, HTTP/Wails compilation pass; corrected frontend audit examines 135 production dependencies with zero findings |
| Rust routing and persistence | 10 exact-channel/duplicate/CLTV tests and 2 legacy/RC.6 persistence tests rerun successfully |
| Native dependency graph | Fixed verifier 0.103.16 retained; unchanged native checksum verified |
| Universal macOS package inspection | Pass: x86_64/arm64, macOS 15.4, independent identity/data directory, reviewed rpath, exact source/tag/library metadata, ad-hoc integrity only |
| Source/workflow/secret checks | Actionlint, shell syntax, source boundaries, source and supported artifact-text scans pass |
| Tag/publication gate | Packaging rejects a missing/unapproved release tag. No tag-creating regression script, DMG, PGP signing, upload, or release was performed |
| Private source/SBOM | Prepared from exact tested source, hashed in the evidence index |

The initial audit from the repository root saw zero dependencies and is excluded from valid audit evidence. The corrected `frontend` audit above is authoritative for this run. No real or copied wallet data was used; migration tests used fixtures. The app and HTTP binary were built and statically inspected only, never launched.

## Open gates

1. **Critical remote-build reproducibility finding.** CodeRabbit's complete 22-file maintenance review completed with one critical finding: the replacement dependency is unpublished. This is valid for clean remote builds. GitHub returns `422 No commit found` for binding `ea7860a...`, while the prior published binding control returns 200; the Go module proxy returns 404 for the candidate version. Local tests succeed through the documented private module proxy and do not prove public dependency availability. Resolve through separately approved dependency publication and independent clean module/checksum readback. Do not treat the completed review as zero-findings approval.
2. **Rust dependency audit remains nonzero.** Today's advisory database reports 12 entries and four unmaintained warnings. The two production entries remain RUSTSEC-2026-0381 for root stores 0.25.4/1.0.9, despite the fixed intermediate-SPKI verifier. Other flagged versions are absent from the production normal/build graph. No clean raw-audit, zero-unresolved-Critical/High, or complete TLS-stack remediation claim is made.
3. **Sandbox integration remains unaccepted.** The required loopback ownership port 21421 is occupied. The existing service was not stopped, inspected, or changed. Run this test on an isolated runner before closing its gate. The test exclusion is recorded, not accepted as a pass. No wallet-runtime acceptance was attempted.

All prior release/installation and guided-import acceptance gates remain separate. Public or installed Grynvault state was not independently refreshed by this source monitor.

## Proposed dependency publication packet — requires owner approval

Purpose: make the already tested native dependency and its exact source downloadable for reproducible remote builds. This is public source/binary branch publication for review, not a release or installation. Both proposed remote branches were independently checked and are absent. Recheck before executing; do not force-push.

| Order | Owner origin | Remote branch | Exact local commit |
| --- | --- | --- | --- |
| 1 | `https://github.com/hudhaifahz/ldk-node.git` | `codex/alby-security-20261021` | `22c66d60ffe13e6dd0532d7f22501aa55b6a769f` |
| 2 | `https://github.com/hudhaifahz/ldk-node-go.git` | `codex/alby-security-20261021` | `ea7860a8558930f501df672616b88c0a72c6c584` |

The second branch includes the rebuilt macOS library and its scope/limitations record. Inherited non-macOS binaries are not newly validated or advertised as supported. A source push may trigger existing CI builds; the tracked LDK workflow restricts automatic Go-binding publication to `main`, so this proposed codex branch does not select that publication job. The Go-binding worktree has no tracked workflow directory.

After these two approved pushes, independently read back each branch SHA, download the candidate module into a fresh cache using ordinary proxy/sumdb verification, and match the expected module/native hashes. A mismatch stops further work. The Hub push, any PR, tag, PGP signature, GitHub release, install, deployment, or production/public release change is excluded and still needs its own approval. Publishing the dependency does not resolve the Rust-audit or sandbox acceptance gates.

**This packet has not been executed.** No push, release, install, wallet secret/data access, wallet launch, copied-wallet connection, invoice, Lightning probe, HTLC, payment, swap, rebalance, or channel operation occurred.

## Local evidence

- Evidence directory: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261022`.
- Hash index: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261022/evidence-hashes.json`.
- Source archive SHA-256: `e665794c227d9a898e262955e4210f542640857f56acc5d4b81c11f2c4d336a7`.
- SBOM SHA-256: `b198b3184a6be7f73da906f904c6463a31cc245473d7a01ca96a5b801279969c`.
- Fresh review: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261022/hub-coderabbit-full.log`.
- Remote module resolution/control: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261022/remote-module-resolution.json`.
- Exact proposed actions/ref checks: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261022/proposed-dependency-publication.json`.

# Approved dependency pushes — 2026-11-05

**TWO APPROVED BRANCHES PUSHED AND READ BACK — NO RELEASE OR INSTALLATION.**

Recorded 2026-11-06T05:52:32.230171+00:00. Owner explicitly approved the two dependency branches in the remediation packet, required main-LDK PR separation, and prohibited release/installation. That approval was executed exactly; no Hub/main-LDK PR/tag/PGP/release/deployment/install action was taken.

## Published source refs

| Repository | Branch | Exact verified remote commit |
| --- | --- | --- |
| [hudhaifahz/ldk-node](https://github.com/hudhaifahz/ldk-node/tree/codex/alby-rust-sandbox-20261105) | codex/alby-rust-sandbox-20261105 | d8c68f5534e5aaa9d4f994751bc763e3938c2c22 |
| [hudhaifahz/ldk-node-go](https://github.com/hudhaifahz/ldk-node-go/tree/codex/alby-rust-sandbox-20261105) | codex/alby-rust-sandbox-20261105 | 1551acb4e724cab38f641dde391f0a608201f590 |

Preflight verified clean worktrees, exact packet commits, exact origin URLs and absent proposed remote refs. Pushes used explicit SHA:branch refspecs, follow-tags disabled, submodule pushes disabled and push signing disabled; no force. Both Git ls-remote and GitHub Git-ref API independently read back exact approved SHAs. Main-LDK contribution source/PR and default branches were not modified by these pushes. No PR was opened. All prior raw-audit/configuration assessments and installation gates remain separate.

## Public module verification and checksum correction

Fresh empty module cache downloaded `github.com/hudhaifahz/ldk-node-go@v0.0.0-20261106012124-1551acb4e724` through **https://proxy.golang.org** with **sum.golang.org**, GONOSUMDB/GOPRIVATE/GONOPROXY explicitly empty, no private file proxy. Origin SHA is1551acb4e724cab38f641dde391f0a608201f590. Signed checksum-database lookup retained. Canonical public module sum:

`h1:orT2l6DJJuCxT+NqALn6njuxn0AB4ESQzIsqC3oE7kU=`

Downloaded raw universal native library verifies SHA256 `19f9cbea84af10665fe10370a7c1d6e586a2cf8f9cd23a0458ab7e069e411d3f`, identical to the approved/tested binary.

The earlier locally generated ZIP mistakenly contained a worktree .git metadata FILE. This produced old private checksum `h1:v0oaspypTOoU23I0KzPJDsBuw1N/iSZ/SOX85Qi6v+I=`. Exact ZIP comparison proves that .git is the ONLY extra entry; ALL14 tracked files, including native bytes, are identical. The metadata file was never pushed or included in the public module. Earlier private checksum claims are historical and superseded, not accepted public provenance.

Corrected only local Hub go.sum to the canonical sum, commit **3bcd8fce294031da73ea6549807c5816d6cfe3d9**; this Hub commit was NOT pushed. Quarantined old task-cache metadata/archives, replaced this candidate's local ZIP/hash records with canonical public files, restored original read-only cache permissions, and left other module versions/global Go settings unchanged. Original evidence/helper remains historical; new `canonical-module-proxy.go` refuses worktree inputs, requiring an exact Git export to prevent recurrence. Do not reuse the old metadata-bearing archive/helper for future candidate provenance.

With standard public proxy/sumdb and NO checksum exceptions, module verify, full Go suite, full Wails-tag suite WITHOUT exclusions and exact-source universal macOS build/static inspection PASS. Both package slices embed canonical module checksum, new native hash and Hub source3bcd8fce294031da73ea6549807c5816d6cfe3d9. App never launched/installed. This closes module availability/canonical-checksum readback; it does not close CI/runtime/release gates.

## CI failure and concrete local follow-up

[The VSS integration run](https://github.com/hudhaifahz/ldk-node/actions/runs/42169123523) failed in `Run VSS Integration tests`. Bounded diagnostic processing confirmed the known missing BITCOIND_EXE fixture error; no raw runtime logs were retained/displayed. Removing obsolete Cargo downloader features left this workflow without explicit test binaries. The earlier broad statement that all CI already provisioned them was incomplete: Rust/benchmark jobs do, but VSS did not.

Prepared exact local LDK follow-up **7f3fac35f1414b501364a45a46cc16de362c4b45**. It adds three lines to `.github/workflows/vss-integration.yml`: two explanatory comments and `source ./scripts/download_bitcoind_electrs.sh` before the existing VSS test commands. The existing script verifies pinned checksums. Sourcing and testing in the same shell retains the temporary binaries until both commands finish. No tests skipped, no obsolete vulnerable downloader restored, no core feature/native/API change.

Validation: actionlint1.7.7 PASS; shell syntax PASS; diff/secret scan PASS; CodeRabbit completed the one-file workflow review with ZERO findings. No downloader, backend binary or wallet was executed locally. Runtime CI rerun remains necessary after a separately approved push. Other CI jobs were running; CLN/LND completed successfully at the recorded snapshot. Cross-build workflow started, but its binding-publication job is guarded to main and cannot publish from this topic branch. Do not claim all GitHub CI passed.

### Additional approval requested — NOT EXECUTED

One FAST-FORWARD-ONLY follow-up source push to **hudhaifahz/ldk-node**, branch **codex/alby-rust-sandbox-20261105**, from expected remote d8c68f5534e5aaa9d4f994751bc763e3938c2c22 to **7f3fac35f1414b501364a45a46cc16de362c4b45**. Recheck expected remote head first; no force. This reruns existing CI including fixture/regtest integration jobs. No binding branch update is needed.

This new commit was not in the two-commit packet the owner approved. It is held locally until new explicit approval. Approval would not authorize Hub push, main-LDK PR work, tag/signing/release/install/deployment. Two original pushes are DONE; do not request or perform them again as if pending. Raw Rust8 legacy-test entries/maintenance warnings, conditional Echo assessments and other review/runtime gates remain documented in the remediation report.

## Evidence and separation

Evidence: `/Users/kode/Development/alby-rust-sandbox-20261105/evidence/push-20261105`. Public-download/sumdb record, canonical ZIP comparison, Git/API remote readbacks, CI metadata, bounded VSS diagnostic summary, workflow review/lint, standard-proxy tests and static package proof retained. No wallet secrets/data/process access, copied wallet, real invoice, Lightning probe/HTLC/payment/swap/rebalance/channel operation, release or installation was performed locally. Main-LDK direction remains separate. Local Hub work and the CI follow-up are unpublished.

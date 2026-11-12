# Exact-channel rebalance maintenance execution brief

Canonical release-state brief: `/Users/kode/Development/albyhub-v1.24.1/doc/exact-channel-rebalance-execution-brief.md`. Canonical release gate: `/Users/kode/Development/albyhub-v1.24.1/doc/exact-channel-rebalance-release-gate.md`.


## Upstream security maintenance snapshot — 2026-10-21

**PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Official release/tag remains v1.24.3 (`5d5e696dfe7d430a365d48939c394376e078a56b`). Upstream master snapshot `60dbcc3befe7629dba2a14a2e60df1d4620fe72e` contains seven newer commits, including Go CA-distrust and LDK intermediate-SPKI mitigations. Isolated branch `codex/alby-security-20261021` in `/Users/kode/Development/alby-security-maintenance-20261021/hub` includes the original owner-fork source and this snapshot. Exact tested Hub source: `5a6b05f80f853998214bd432d9b59163fc159a71`; patched LDK: `22c66d60ffe13e6dd0532d7f22501aa55b6a769f`; rebuilt binding: `ea7860a8558930f501df672616b88c0a72c6c584`.

Go 1.27.3, full Go/race/vet/module/HTTP checks, synthetic SQLite/Postgres migration regressions, frontend lint/types/builds/audit, 12 exact-channel/persistence tests, the TLS verifier distrust regression, and universal macOS static package inspection pass. Raw library SHA-256: `bb487c9c2640e4c053aa508998f459b6b03216be4592ea080a64a4e9b5476f78`. Go reports zero reachable vulnerabilities. The binding dependency pin is unpublished and was tested through a local module proxy; remote module/hash readback remains an owner-gated prerequisite.

Open holds: raw Rust audit still flags production root stores under RUSTSEC-2026-0381 despite the fixed/tested verifier; one Wails sandbox integration test is excluded because port 21421 is already occupied; the final incremental CodeRabbit review hit its OSS quota. No security hold, installation/migration acceptance, release tag, PGP signing, publication, or installed-state gate is closed by this monitor. The untagged acceptance wrapper was built and statically inspected only. No wallet was launched or accessed; no invoice, probe, HTLC, payment, swap, rebalance, or channel operation occurred. No push, tag, PGP signature, PR, installation, deployment, or public change was made.

Exact commands, check results, limitations, artifact hashes, local dependency provenance, and separate approval gates are recorded in `/Users/kode/Development/alby-security-maintenance-20261021/hub/doc/alby-security-maintenance-20261021.md` and `/Users/kode/Development/alby-security-maintenance-20261021/evidence/evidence-hashes-final.json`. All standing release and owner approval gates remain in force.


## Alby v1.24.4 maintenance refresh — 2026-10-22

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Official v1.24.4 was published `2026-10-22T08:59:21Z`, tag object `feba0c0c1795e768564de42cd62b674f19b90d60`, peeled source `c401324c065d50aea2a8b61006b6bb03b3ef776c`. Default master is `6873fa512bde916c5c9214fa25d3a1258d9404fb`, one further NWC error-reporting fix. The existing isolated maintenance branch includes both; tested Hub source is `40857768fb63b99a769b3b5e17139c3dd895f8b6`. CI now checks the latest Node 22 patch while preserving all fork safeguards; local validation uses SHA-verified Node 22.23.4 and Go 1.27.3. Exact-channel and route-redaction source/native invariants remain unchanged.

Full Go/race/fixture-migration/vet/module/HTTP checks, frontend checks/audit, 12 Rust route/persistence tests, source/workflow/secret checks, and universal macOS build/static inspection pass. Wails sandbox integration still fails because ownership port 21421 is occupied; the remaining Wails-tag tests pass with that explicit exclusion. The raw Rust audit still has 12 entries/four warnings, including production root stores under RUSTSEC-2026-0381. CodeRabbit's full 22-file review now completes and reports one critical, verified remote-build blocker: the exact native dependency is still unpublished (GitHub candidate 422 vs published control 200, Go proxy candidate 404). Local module-proxy tests do not close this gate.

A concrete two-dependency-branch publication packet, exact SHAs, binary scope, CI effects, clean remote readback requirements, and all exclusions are prepared in `/Users/kode/Development/alby-security-maintenance-20261021/hub/doc/alby-security-maintenance-20261022.md`. It requires owner approval before any push; it does not authorize the Hub push, a tag, PR, PGP signature, release, install, deployment, or wallet action. No packet action was executed. The untagged app was built/inspected only, never launched. Detailed evidence and hashes: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261022/evidence-hashes.json`. All prior release/installed acceptance gates remain in force.


## Upstream security and migration maintenance — 2026-11-02

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Latest release remains v1.24.4, peeled c401324c065d50aea2a8b61006b6bb03b3ef776c; exact official master is b0be1e8b41d1616be1c0339e77e7b3eecf3b42f4. Existing isolated codex/alby-security-20261021 incorporates upstream migration repair a58445ab (unlink destination transactions from deleted apps to prevent balance inheritance on app-ID reuse), MPTCP-default mitigation b0be1e8b (`godebug multipathtcp=0`, upstream-labeled CVE-2026-98432 mitigation; host kernels still require patching), and intervening catalog removal. Exact tested Hub source: c3d6d6afda7a5fca6dc0f4336282730bd22237e8. Patched Go toolchain 1.27.3, exact-channel/redaction invariants, both Grynvault quote tables and existing native dependency/hash are preserved. Original rc3 checkout remains clean/unchanged.

Full Go/race/vet/module/HTTP and frontend checks pass; fresh SQLite/PostgreSQL migration/continuity tests and migration races pass with fixtures only, and the disposable PostgreSQL server was stopped. Universal macOS build/static inspection passes, including MPTCP-disabled build metadata in both slices; never launched. Go has zero reachable vulnerabilities; frontend production audit has zero findings across 135 dependencies. Incremental CodeRabbit review of all 6 changed files completes with zero findings; the earlier full-maintenance unpublished-dependency finding remains unresolved. Source secret scan passes; broader artifact-text scan has 4 reviewed false positives (module checksums and Go formatting/constant strings), with raw exit/results retained.

Open holds: exact dependency remains unpublished (GitHub 422/proxy retry 404, absent remote codex refs); Wails sandbox test still fails only on occupied port 21421, with remaining tests passing under explicit exclusion; refreshed raw Rust audit now has 15 entries/four unmaintained warnings and one Tokio I/O-safety warning. Production roots 0.25.4/1.0.9 and Tokio 1.53.2 remain flagged; affected Tokio-function reachability is unproven, and no clean raw audit or complete remediation is claimed. Existing two-dependency publication approval packet is unchanged/unexecuted. No push, tag, PGP, PR, release, install, deploy, public change or wallet operation occurred. All prior owner/release/installed acceptance gates remain open.

Exact evidence, artifact hashes and limitations: `/Users/kode/Development/alby-security-maintenance-20261021/hub/doc/alby-security-maintenance-20261102.md`; `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261102/evidence-hashes.json`.


## Desktop security and dependency maintenance — 2026-11-03

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Official latest release remains v1.24.4; exact master a71bd10aafd23afad286b8d03d89e96f66752606 adds seven commits, including Wails 2.17.0 development IPC hardening, Echo/LND/Bark/btcec updates and CI least-privilege settings. Existing isolated codex/alby-security-20261021 preserves pinned/read-only/guarded CI and removed auto-publication workflows, Go1.27.3/gRPC1.83.2/MPTCP-disabled default, native exact-channel replacement/hash, route redaction and both Grynvault quote tables. Both packaging scripts and CI CLI pin match Wails2.17.0; tool compiled only in private evidence folder, no global installation. Exact final tested source: 037beaae929ab7e9269b880ba57c2d61b25130c3. Original rc3 checkout clean/unchanged.

Initial CodeRabbit review found a valid major upstream Bark regression: TotalSpendableMsat included funds awaiting refresh despite feeding NWC/send controls. Corrected to native SpendableSats only; added pure fixture regression for zero/nonzero spendable with refresh/pending funds. Follow-up review of both changed Go files completed with zero findings. Full final Go/race/vet/module/HTTP, frontend, source/actionlint/secret and universal macOS static checks pass; fresh PostgreSQL migration/continuity fixture checks passed before the balance-only correction (DB code unchanged), and fixture server stopped. Go has zero reachable vulnerabilities (1 imported-package/3 required-module non-reachable); frontend audit135 deps/zero findings. App built/inspected only, never launched; both slices retain MPTCP-disabled default and new Wails/Bark dependency provenance.

Open gates unchanged: unpublished native dependency (absent refs/GitHub422; today's proxy timeout000 inconclusive), occupied21421 sandbox integration failure with explicit exclusion remainder passing, raw Rust audit15 entries/four unmaintained warnings/one Tokio unsoundness warning. Production roots0.25.4/1.0.9 and Tokio1.53.2 remain flagged, affected-function reachability unproven; no clean raw audit or complete remediation claim. Artifact text scan4 known false positives are triaged with raw results retained. Existing dependency-push approval packet remains unchanged/unexecuted. No push/tag/PGP/PR/release/install/deploy/public change or wallet access/operation occurred; all owner/release/installed gates remain separate.

Exact evidence and artifact hashes: `/Users/kode/Development/alby-security-maintenance-20261021/hub/doc/alby-security-maintenance-20261103.md`; `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261103/evidence-hashes.json`.


## Alby v1.24.5 tag maintenance — 2026-11-04

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Official latest published release remains v1.24.4. New annotated v1.24.5 tag object da9b6b5b32aff46f495815d4fc984ed8bf8a6952 peels to exact default master 708e3669d69362575d5e20ab4d94731e4d45b6ef; release-by-tag API404,52 releases/73 tags, independent configured Git refs agree. Nine new commits include on-curve app-key validation, notification redaction/per-app delivery failure handling, relay acknowledgement reporting/new go-nostr, shutdown-safe LDK subscriptions, address-network validation and frontend dependency updates. Reused isolated codex/alby-security-20261021; tested/build source 074f381f6fc094fb4868c03dd5a01e20b2b4c3bf. Only checksum conflict resolved. Manifest provenance updated; exact-channel APIs/redaction/migration/schema/Bark correction/native library/toolchain/pinned workflows preserved; original rc3 clean/unchanged.

Recovered full Go suite, expanded race suite, SQLite/Postgres migration/continuity fixtures, vet/module/HTTP, frontend checks/audit134 deps/zero findings,12 Rust route/persistence tests, source secrets/boundaries and universal macOS static package inspection PASS. Initial Go setup failed on proxy ZIP EOF; official transport recovery preserved new module checksum verification and initial error evidence. Fixture PostgreSQL stopped. Both macOS slices retain new go-nostr/Wails2.17/MPTCP-disabled default. Built/inspected only; never launched. Go zero reachable,1 imported-package/3 required-module nonreachable findings. Artifact scan4 exact-public-literal false positives triaged, raw redacted results retained.

Open gates: CodeRabbit two connection failures (no completed review claim); unpublished native dependency branches still absent and unchanged two-push packet unapproved; occupied21421 sandbox integration failure with explicit exclusion remainder passing; raw Rust audit15 entries/four unmaintained/one Tokio unsoundness warning unchanged. Production roots/Tokio remain flagged, reachability unproven; no clean audit/full remediation/release-ready claim. No push/tag/PGP/PR/release/install/deploy/public change/wallet access or operation. Owner/public/installed acceptance gates remain separate.

Exact evidence and limitations: `/Users/kode/Development/alby-security-maintenance-20261021/hub/doc/alby-security-maintenance-20261104.md`; `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261104/evidence-hashes.json`.


## Alby v1.24.6 maintenance — 2026-11-05

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Official v1.24.6 published2026-11-05T09:45:44Z; annotated objecte6f9ac07c505776f417e83534ee796dee9c9fe3f peels to exact default master921003b65ece004551612647ce1df22327a2486d, API/configured Git refs agree (53releases/74tags). Two new commits0489c7cf/921003b6 fix LNURL retry duplicate-payment risk and malformed metadata backend/UI crashes. Reused isolated codex/alby-security-20261021, conflict-free merge and manifest provenance update. Tested/build source9c04ab63029bab4e88436b0d49581793063f89b6. Exact-channel/LDK/redaction/migrations/schema/import/Bark correction/native hash/dependency pins/read-only workflows preserved; original rc3 remains clean/unchanged.

Full Go/race/SQLite/Postgres migration fixtures/vet/module/HTTP, frontend checks/audit134deps zero findings,12 Rust route/persistence tests, source secrets/boundaries and universal macOS static inspection PASS. Offline actual-component mock regression proves same-fixture-invoice retry/invalidation/disabled inputs, rejects old source, plus malformed metadata/CSV guards; no wallet/network/value operation. Initial Wails build raced generated frontend files; rerun after package completion fails ONLY existing occupied21421 sandbox integration, explicit-exclusion remainderPASS. Initial failure retained. App/HTTP never launched, fixture PostgreSQL stopped. CodeRabbit service recovered and completed all34 changed files vs2f58734a (November4+5), ZERO findings; closes connection hold only, no independent-audit claim.

Open holds: unpublished native dependency branches absent and unchanged approval packet unapproved; occupied-port sandbox integration; raw Rust audit15 vulnerabilities/5 unmaintained/1Tokio-unsound. Vulnerability IDs/versions unchanged; new pin-utils0.1.1 unmaintained warning absent production tree. Production roots/Tokio still flagged with reachability unproven; yanked-status registry checks timed out although advisory scan JSON completed. Go zero reachable,1 imported-package/3 required-module nonreachable. Four artifact scan findings are exact prior public-literal falsepositives. No clean raw audit/full remediation/release-ready claim. No push/tag/PGP/PR/release/install/deploy/public change or wallet access/operation; all owner and installed acceptance gates remain separate.

Exact evidence and limitations: `/Users/kode/Development/alby-security-maintenance-20261021/hub/doc/alby-security-maintenance-20261105.md`; `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261105/evidence-hashes.json`.


## Owner-authorized Rust and sandbox remediation — 2026-11-05

**PREPARED/TESTED LOCALLY — NOT PUSHED/RELEASED/INSTALLED.**

Separate worktrees `/Users/kode/Development/alby-rust-sandbox-20261105` on codex/alby-rust-sandbox-20261105 preserve original monitor/main-LDK preparation source. Rust coded8c68f5534e5aaa9d4f994751bc763e3938c2c22, rebuilt binding1551acb4e724cab38f641dde391f0a608201f590, final tested Hub3e4ed99996f3f3669569bcd9a1e67448e201986c. Patched Tokio1.53.3/Electrum0.24.2/roots1.0.10/pulldown0.13.6, removed obsolete automatic test download features (CI still supplies verified test binaries). Production normal/build graph0 advisory vulnerabilities; raw audit8 entries remain ONLY in retained LDK0.6.2 VSS compatibility reference plus5 maintenance warnings. No suppression/clean repository-wide audit claim. Root-store regression passes patched/rejects old store,12 routing/persistence tests and compile-only integration checks pass.

Production sandbox policy/IPv4 lock addresses unchanged; synthetic helper uses isolatedIPv6loopback at same ports. Complete Wails-tag suite and sandbox race pass WITHOUT exclusions, closing occupied-port TEST hold only. New macOS universal raw native SHA19f9cbea84af10665fe10370a7c1d6e586a2cf8f9cd23a0458ab7e069e411d3f,491 exported binding symbols unchanged in each slice; generated APIs/route/redaction/import/schema source unchanged. Full final Go/race/vet/module/HTTP/SQLite, frontend and universal static package checks pass. Other-platform binaries inherited/unvalidated; wallet/install acceptance separate.

Fresh Go database added3 advisories: compressor patched1.20.2, final rawscan still2 Echo v4 whole-module flags assessed against constant-global unlock limiter/passing spoofed-header regression and absent Proxy middleware; no clean rawscan/vendor-fix claim. CodeRabbit Rust/sandbox reviews0 findings; final native-pin review hit OSS quota, subsequent pin-only changes manually verified/tested.4 known artifact-public-literal falsepositives retained, source secrets pass. No wallet touched/app launched or push/tag/PGP/PR/release/install/deploy/public change.

New exact two-dependency source-push packet is in `/Users/kode/Development/alby-rust-sandbox-20261105/hub/doc/rust-sandbox-remediation-20261105.md`. Both new remote refs absent; old October22 packet is stale for this candidate. Owner approval required BEFORE any push; Hub/main-LDKPR/tag/sign/release/install remain separate. Main-LDK upstream contribution is complementary and untouched, with future rebase/API alignment potentially needed. Evidence `/Users/kode/Development/alby-rust-sandbox-20261105/evidence/evidence-hashes.json`.


## Owner-approved dependency pushes — 2026-11-05

**DEPENDENCIES PUSHED/VERIFIED — HUB AND FOLLOW-UP CI FIX LOCAL — NOT RELEASED/INSTALLED.**

Owner approved exact two-branch packet. Remote hudhaifahz/ldk-node codex/alby-rust-sandbox-20261105=d8c68f5534e5aaa9d4f994751bc763e3938c2c22; binding samebranch=1551acb4e724cab38f641dde391f0a608201f590. Git+API readbacks agree; noforce/tags/signing/mainLDKPR/Hubpush. Public Go proxy +sumdb fresh empty-cache download verifies canonical module h1:orT2l6DJJuCxT+NqALn6njuxn0AB4ESQzIsqC3oE7kU= and raw native SHA19f9cbea84af10665fe10370a7c1d6e586a2cf8f9cd23a0458ab7e069e411d3f. Prior private h1v0o archive erroneously included .git worktree metadata; ALL14 tracked files/native bytes match public ZIP exactly. Old checksum superseded, task cache quarantined/canonicalized without other-version/globalconfig changes; future proxy helper requires exported Git tree. Local Hub checksum-only commit3bcd8fce294031da73ea6549807c5816d6cfe3d9 NOT pushed. Standard proxy/sumdb/no exceptions module verify/fullGo/fullWails/universal static package pass; never launched.

GitHub VSS integration failed missing BITCOIND_EXE: this workflow lacked explicit binary provisioning after downloader removal. Reviewed/linted local fix7f3fac35f1414b501364a45a46cc16de362c4b45 adds existing checksum-verified download script in SAME test shell, preserving both test commands. CodeRabbit0findings/actionlint/shell/secret checks pass. This follow-up push is NOT approved/executed; new commit exceeds original exact packet. OtherCI pending, no all-CI-green claim. Main-LDK PR work separate/untouched. Raw audit/configuration/runtime/release gates remain.

Current evidence and single follow-up approval packet: `/Users/kode/Development/alby-rust-sandbox-20261105/hub/doc/dependency-push-20261105.md`; `/Users/kode/Development/alby-rust-sandbox-20261105/evidence/push-20261105`. No release/install/wallet operation. Original two approved pushes complete; do not repeat approval prompts for them.

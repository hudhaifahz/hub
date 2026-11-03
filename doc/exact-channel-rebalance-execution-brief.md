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

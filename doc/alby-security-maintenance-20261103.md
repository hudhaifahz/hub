# Desktop security and dependency maintenance — 2026-11-03

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Recorded 2026-11-03T16:12:46.302777+00:00. This supersedes the November 2 maintenance validation, preserving every owner-controlled publication, signing, installation and wallet-operation gate.

## Official upstream and scope

Latest official Alby Hub release remains [v1.24.4](https://github.com/getAlby/hub/releases/tag/v1.24.4), published `2026-10-22T08:59:21Z`, tag object `feba0c0c1795e768564de42cd62b674f19b90d60`, peeled `c401324c065d50aea2a8b61006b6bb03b3ef776c`. All 52 releases and 72 tags were enumerated. Official default master is **`a71bd10aafd23afad286b8d03d89e96f66752606`**, independently confirmed by API/configured upstream git refs and final readback.

Seven commits since the integrated `b0be1e8b41d1616be1c0339e77e7b3eecf3b42f4` are incorporated: Hopp catalog `0c160732`, Echo 4.16.2 `92b71a7c`, LND 0.21.6-beta `c945c1f9`, Bark bindings 0.26.1/API adaptation `46b8b52e`, CI token restrictions `e7ba94aa`, btcec 2.5.1 `211800c1`, and Wails 2.17.0 `a71bd10a`.

Qualifying security changes include [Wails 2.17.0](https://github.com/wailsapp/wails/releases/tag/v2.17.0) development-server IPC origin/Host/capability-cookie hardening, removal of Windows notification PowerShell fallback, and template trust warnings. These protect their documented scopes; no installed-wallet acceptance or universal runtime security claim follows. [Echo 4.16.2](https://github.com/labstack/echo/releases/tag/v4.16.2) contains authentication-scheme/CSRF validation repairs and prior 4.16.1 security changes. [LND 0.21.6-beta](https://github.com/lightningnetwork/lnd/releases/tag/v0.21.6-beta) documents its Linux MPTCP mitigation, not a replacement for kernel patching. No host/kernel configuration was examined or changed. No wallet binary was downloaded, installed or launched.

## Exact local preparation

Existing isolated branch `codex/alby-security-20261021` in `/Users/kode/Development/alby-security-maintenance-20261021/hub` contains merge `e8cf5deb` and balance correction **`037beaae929ab7e9269b880ba57c2d61b25130c3`**, the exact final tested code/build source. A later documentation-only commit records this checkpoint. Original `/Users/kode/Development/grynvault-rebalance-rc3` remains clean/unchanged at `4c5dc1c1ce7e8284cd5ad49b87d5dab95f09776c`, branch `codex/coinos-inbound-liquidity`.

An initial partial-clone merge could not retrieve a missing upstream blob. A read-only official-upstream `fetch --no-filter --refetch` supplied it; no reset, force-push or source loss occurred. Merge conflicts were resolved by retaining the fork's existing pinned actions, read-only token permissions, checkout credential removal, guarded packaging flow and removed automatic Docker/release publication workflows. Those safeguards already implement the relevant upstream least-privilege intent; write-permission publication jobs were not reintroduced. Wails CLI pin in the active workflow and both local packaging guards now match module 2.17.0. Manifest provenance names exact upstream master. Patched Go toolchain 1.27.3, newer fork gRPC 1.83.2, `godebug multipathtcp=0` and the custom LDK replacement are retained.

Hub exact-channel APIs/LDK adapter, route-log redaction, guided import, schema migrations and November 2 migration repair are unchanged. Both Grynvault quote tables remain included. Exact-channel parent `cd6bc1e399ba256adb113bbaf44d659c5b22a170` and redaction `59bb47a41eacaccc5a95ed717d87befb6759bbe5` remain ancestors of original and maintenance heads.

### Verified review finding and correction

CodeRabbit's complete 9-file incremental review against prior documentation head `4599d221c01b924861024257cf3ce097a9e3a0cc` reported one major issue: the new upstream Bark balance mapping added NeedsRefreshSats into TotalSpendableMsat. Inspection verified this field also feeds NWC get_balance and UI send maxima, so keeping only NextMaxSpendable bounded was insufficient. The corrected mapping exposes only native SpendableSats in all spendable fields. Refresh/pending funds are not claimed lost; the existing response does not expose them as a separate total-assets balance. A broader pending/asset display would require separate API/UI work.

A pure native-balance-value mapping helper and regression test cover zero/nonzero spendable funds with refresh and other pending buckets; no wallet is created/opened or called. Final complete Go/race suites and package were rerun after the correction. Follow-up CodeRabbit review of both changed Go files completed with zero findings. This is automated incremental review, not an independent audit, and does not close the earlier unpublished-dependency Critical finding.

## Build and dependency provenance

Custom native LDK source/binding remain clean at `22c66d60ffe13e6dd0532d7f22501aa55b6a769f` / `ea7860a8558930f501df672616b88c0a72c6c584`. Replacement pseudo-version is `v0.0.0-20261021161128-ea7860a85589`, expected Go h1 `OjpkVT4fzISDV4Ej79z9v1hEJ11oWoGf3yFJfmtsJ5o=`. Raw universal library SHA-256 remains `bb487c9c2640e4c053aa508998f459b6b03216be4592ea080a64a4e9b5476f78`; exact-channel generated APIs/payment/event source unchanged, library reused byte-for-byte. Custom non-macOS binaries remain inherited/unvalidated. Updated Bark native code is a separate upstream dependency and is not assessed by the custom LDK Rust audit.

Build tools: Go 1.27.3, previously SHA-verified isolated Node 22.23.4, Wails 2.17.0, Syft 1.41.1. Official Wails module download was checked against `h1:pJqX9ogqW4PPE/gSeCQUcYmpPghxEpHajONw9rPj1pg=`; CLI was compiled read-only from a private source copy into this evidence folder, SHA-256 `449a686174dba38565a918f7fce6ae714776f95e29ab3fa2a497e1b22810d19b`. No global tool installation/replacement occurred. Local unpublished LDK binding resolution uses the documented process-local file proxy/narrow GONOSUMDB exception only; no global or production security settings changed.

## Validation and limits

| Check | Result |
| --- | --- |
| Final Go suite | Full count=1 suite passes at `037beaae929ab7e9269b880ba57c2d61b25130c3`, including Bark mapping regression and SQLite migration/continuity tests |
| Race/vet/module/HTTP | Pass: migration/db/API/Wails/guided-import/Bark race suite; vet, module verify and HTTP build |
| PostgreSQL fixtures | Fresh PostgreSQL 16.13 db/migrations/transactions suite and SQLite↔PostgreSQL migration races pass at merge e8cf5deb. Database/import code has no diff between that merge and final balance-only correction; SQLite checks repeat in final full suite. Disposable fixture server stopped and short socket directory removed |
| Wails-tag tests | Full final invocation fails only the sandbox integration on occupied 127.0.0.1:21421; explicit exclusion passes the remainder. Existing listener/process never inspected/stopped/changed |
| Go vulnerability scan | Zero reachable findings; 1 imported-package/3 required-module non-reachable findings remain. Does not assess native/OS vulnerabilities |
| Frontend | Frozen install, lint/type/format, HTTP/Wails builds pass; unchanged lockfile production audit covers 135 dependencies/zero findings. Final package reruns desktop compilation and frontend checks |
| Rust routing/persistence | 10 exact-channel/duplicate/CLTV tests and 2 persistence tests pass, fixtures only, unchanged native source |
| Raw Rust audit | Exit 1; 15 vulnerability entries, 4 unmaintained warnings and 1 unsoundness warning. Production roots and Tokio holds remain below |
| Source/workflows | Source-boundary, shell syntax, fresh actionlint 1.7.7 and full maintenance-range secret scans pass. No auto-publication or permission weakening |
| Package | Universal final-source macOS build/static inspection pass: 15.4 floor, independent identity/data folder, reviewed rpath, exact source/untagged label/native provenance, ad-hoc integrity only. Both slices embed Wails 2.17.0, Bark bindings 0.26.1 and MPTCP-disabled Go default |
| Tag/release boundary | Nonexistent/unapproved tag rejected before creating output/tag/DMG. No tag-creating regression script, DMG, PGP signature or upload |
| Artifact secrets | Broad supported text/strings scan raw exit 1 has the same 4 independently verified false positives: 2 public Go module checksums and one concatenated Go DNS-format/decimal-constant/poll-error literal in each slice. Redacted findings/assessment retained; no literal credential found, no blanket suppression. This is not exhaustive native-binary secret analysis |
| Private source/SBOM | Final source archive SHA-256 `d58ec2cc434625fac175ff9caf09588d78c37e2594decfce204ba8fb81b626b2`; CycloneDX 1,103 components, SHA-256 `8ff0b9accf60d715089a6e850e9f9c3fbf4accc7fd7e4c653816d0b3cba9b313` |

Initial e8cf5deb test/build evidence is retained separately under `initial-e8cf5deb` and is not the final package. Final matrices and package explicitly name `037beaae929ab7e9269b880ba57c2d61b25130c3`. No wallet data, password, seed, owner log, NWC connection, real/copied wallet or wallet process was accessed. Go/Rust tests use synthetic fixtures/mocks; the balance regression uses values only. App and HTTP binary were built/statically inspected only, never launched. No invoice, Lightning probe, HTLC, payment, swap, rebalance or channel operation occurred.

## Unchanged open release gates

- **Remote native dependency:** proposed LDK/binding codex branches remain absent; candidate GitHub commit still returns 422. Today's proxy request timed out (000), which is inconclusive rather than a 404; the prior confirmed 404 remains historical evidence. Existing October 22 two-dependency push packet/SHAs/CI effects/exclusions remain unchanged, unapproved and unexecuted. Fresh standard proxy/sumdb SHA/hash verification remains mandatory after any approved publication. No Hub push or other public action is included.
- **Rust findings:** production root stores 0.25.4/1.0.9 remain flagged by RUSTSEC-2026-0381 despite fixed verifier 0.103.16; production Tokio 1.53.2 remains flagged by RUSTSEC-2026-0386. Affected-function reachability is neither established nor dismissed. Other flagged vulnerability versions are absent normal/build production tree. No clean raw audit, zero unresolved Critical/High, complete remediation or release-ready claim.
- **Sandbox/runtime:** occupied-port integration gate needs a separate isolated runner; exclusion is not acceptance. No existing process may be stopped/inspected to satisfy this monitor.
- **Publication/installation:** no push, tag, PGP, PR, release, install, deploy or public change. All owner-controlled and previous installed/guided-import acceptance gates remain separate. Private untagged acceptance wrapper lacks production sandbox entitlements by design; do not install/publish it. Public/installed Grynvault state was not independently reaccepted.

Detailed commands, official snapshots, review findings/fix, provenance, test matrices and hashes are retained in `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261103`; hash index `evidence-hashes.json`.

# Upstream security and migration maintenance — 2026-11-02

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Recorded 2026-11-02T16:10:40.385276+00:00. This source-monitor checkpoint supersedes the October 22 maintenance validation without closing any owner-controlled release, installation, guided-import or wallet-operation gate.

## Official upstream and prepared source

The latest official release remains [v1.24.4](https://github.com/getAlby/hub/releases/tag/v1.24.4), published `2026-10-22T08:59:21Z`, annotated tag `feba0c0c1795e768564de42cd62b674f19b90d60`, peeled `c401324c065d50aea2a8b61006b6bb03b3ef776c`. All 52 releases and 72 tags were enumerated. Official API and independent configured-upstream git refs agree on master `b0be1e8b41d1616be1c0339e77e7b3eecf3b42f4`; final refs were checked again.

Three upstream commits since the prior integrated `6873fa512bde916c5c9214fa25d3a1258d9404fb` are included:

- `a1dfc61d2b28d5793ecad6e1e0c9c367777651c4`: previously assessed Solvocard catalog removal.
- `a58445ab229e222968a02f6608738c0228427f29`: [migration repair #2657](https://github.com/getAlby/hub/pull/2657), unlinking migrated transactions from deleted apps so later app-ID reuse does not inherit their isolated balances. The source transaction amounts/history remain; the orphaned destination app reference becomes NULL.
- `b0be1e8b41d1616be1c0339e77e7b3eecf3b42f4`: [MPTCP default #2672](https://github.com/getAlby/hub/pull/2672). The upstream module comment labels `godebug multipathtcp=0` a mitigation for CVE-2026-98432 and explicitly says host kernels still require patching. This monitor did not inspect or patch host kernels. Runtime configuration can override a Go default; no complete host or native-network remediation is claimed.

Existing isolated branch `codex/alby-security-20261021` was updated in `/Users/kode/Development/alby-security-maintenance-20261021/hub`. Exact tested code/build source is **`c3d6d6afda7a5fca6dc0f4336282730bd22237e8`**, a merge of prior maintenance documentation head `be4179dbcd19ef99bb88eefcfad0fd95a231cbc5` and exact upstream master. The go.mod conflict was resolved by retaining patched `toolchain go1.27.3`, adding the upstream MPTCP default, and preserving the custom native replacement. Release-manifest provenance now names exact upstream master. No new release version was invented.

The original `/Users/kode/Development/grynvault-rebalance-rc3` checkout remains clean and unchanged at `4c5dc1c1ce7e8284cd5ad49b87d5dab95f09776c` on `codex/coinos-inbound-liquidity`. Original exact-channel and route-redaction commits remain ancestors. Changes since October 22 affect only migration code/tests, the card catalog, go.mod and manifest provenance; Hub exact-channel APIs, route redaction, guided-import implementation, schema migrations and both Grynvault quote tables remain preserved. The existing migration regression still checks local-rebalance table continuity.

## Native dependency provenance

LDK source remains `22c66d60ffe13e6dd0532d7f22501aa55b6a769f`; binding remains `ea7860a8558930f501df672616b88c0a72c6c584`, pseudo-version `v0.0.0-20261021161128-ea7860a85589`. Expected Go module hash remains `h1:OjpkVT4fzISDV4Ej79z9v1hEJ11oWoGf3yFJfmtsJ5o=` and raw universal native SHA-256 remains `bb487c9c2640e4c053aa508998f459b6b03216be4592ea080a64a4e9b5476f78`. Native payment/event code and generated APIs are unchanged; the same library was reused, not rebuilt. Non-macOS binaries remain inherited and unvalidated.

Tests/builds use Go 1.27.3, previously SHA-verified isolated Node 22.23.4, Wails 2.16.0 and Syft 1.41.1. The unpublished binding is resolved through the previously documented process-local file Go proxy and narrow GONOSUMDB exception. No global or production security configuration was changed; local test success does not prove remote reproducibility.

## Validation against the tested source

| Check | Result and evidence |
| --- | --- |
| Complete Go suite | Pass, `go test -count=1 -timeout 240s ./...`; go-matrix.json and go-tests.log |
| Migration and continuity | Verbose SQLite suite passes, including deleted-app balance regression and local-rebalance continuity. Fresh PostgreSQL 16.13 fixture suite passes for cmd/db_migrate, db/... and transactions; SQLite→PostgreSQL and PostgreSQL→SQLite migration cases and migration race suite pass |
| Go race/vet/module/HTTP | Pass: cmd/db_migrate, db/..., API, Wails, guidedmigration; vet, module verification and HTTP build pass |
| Wails-tag tests | Full invocation fails only TestReleaseAcceptanceNetworkSandboxIntegration because 127.0.0.1:21421 is occupied; remainder passes with explicit exclusion. Existing listener/process never inspected, stopped or changed |
| Go vulnerability scan | Zero reachable findings; 2 imported-package and 3 required-module non-reachable findings remain. Go scan does not assess native/OS trust |
| Frontend | Frozen install, lint, type/format checks, HTTP and Wails builds pass. Production audit covers 135 dependencies with zero findings |
| Native invariants | Raw library hash matches, clean Rust/binding worktrees and exact heads read back; 10 exact-channel/duplicate/CLTV tests and 2 persistence tests pass, fixture-only |
| Rust dependency audit | Raw exit 1: 15 vulnerability entries, 4 unmaintained warnings, 1 unsoundness warning. See open gates below; no clean raw audit claim |
| Review | CodeRabbit CLI 0.7.1 completed incremental committed review against be4179db across all 6 changed files, zero findings. This does not supersede the earlier full-maintenance unpublished-dependency finding or constitute an independent audit |
| Source and package safeguards | Source-boundary script and changed-shell syntax pass. Workflows unchanged; earlier actionlint evidence retained rather than rerun. Missing/unapproved tag rejected before any DMG, output directory or tag creation; tag-creating regression script deliberately not run |
| macOS build/static inspection | Universal app built and inspected only, never launched. Both architectures embed `multipathtcp=0` in Go build metadata, exact source/untagged acceptance label/native provenance; macOS 15.4 floor, independent identity/data directory, reviewed rpath and ad-hoc integrity checks pass |
| Secret checks | Full maintenance commit-range scan passes. Broader supported artifact-text scan exits 1 with 4 triaged false positives: 2 public module h1 checksums and the same concatenated Go DNS-format/decimal-constant/poll-error string in each architecture. Redacted findings and exact assessment retained; no blanket suppression and no literal credential found |
| Private source/SBOM | Archive SHA-256 `ea224471b5b802f289e4af92a7543380283fff03475c0540bba1c26bc3d31a3a`; CycloneDX 1107 components, SHA-256 `673db93b689a57595470d665deb7af9a0644f3a623918d4a0c27f1e76769b626`. These are private preparation evidence, not a release artifact set |

The first disposable PostgreSQL startup failed because its Unix socket path exceeded macOS's limit. It shut down, was retried with a short private temporary socket directory, and all checks passed. The test-only server was stopped and that socket directory removed; no wallet data was used. No real or copied wallet, seed, password, NWC connection or wallet log was accessed. Only synthetic fixtures and mocks were used. The application/HTTP binary was built, never run. No invoice, Lightning probe, HTLC, payment, swap, rebalance or channel operation occurred.

## Open release and owner gates

1. **Unpublished dependency remains blocked.** Both proposed codex dependency branch refs are absent; GitHub commit readback returns 422 and a retried ordinary Go proxy request returns 404. One earlier proxy request timed out (000), retained as such rather than treated as absence. The two-dependency publication packet in the October 22 report is unchanged and has not been approved or executed. It includes a rebuilt macOS binary, can trigger CI, and excludes Hub push, PR/tag/PGP/release/install/deploy. Following approval, remote SHA and fresh standard proxy/sumdb module/native hash verification remain mandatory. Nothing was pushed.
2. **Rust audit and new warning need resolution before release.** Production normal/build tree still contains webpki-roots 0.25.4 and 1.0.9 flagged by RUSTSEC-2026-0381 despite previously fixed verifier 0.103.16. It also contains Tokio 1.53.2 flagged by new-to-this-monitor unsoundness warning RUSTSEC-2026-0386 (AsyncFd I/O safety; advisory lists 1.53.3 as a patched version). A package being present is not proof that an affected function is reachable; this run does not establish or dismiss reachability, and does not modify/rebuild native dependencies. Other flagged vulnerability versions, including the three additional pulldown-cmark/zip entries, are absent from the production normal/build tree. Exact assessment is in rust-audit-assessment.json. No zero-unresolved-Critical/High, complete-remediation or release-ready claim.
3. **Sandbox integration hold remains.** Close it on a separate isolated runner without touching the existing owner process. The explicit test exclusion is not a pass.
4. **Owner-controlled publication/installation gates remain separate.** No Hub push, tag, PGP signing, PR, release, installation, deployment or public change. No installed or public Grynvault state was independently reaccepted by this monitor. The acceptance wrapper is private/untagged/ad-hoc-only, with production sandbox entitlements intentionally absent, and must not be installed or published.

Evidence directory: `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261102`. Commands, test matrices, raw scanner results, provenance, fixtures-only PostgreSQL lifecycle, upstream snapshots and artifact hashes are retained there. Hash index: `evidence-hashes.json`.

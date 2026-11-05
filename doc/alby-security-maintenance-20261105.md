# Alby v1.24.6 maintenance — 2026-11-05

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Recorded 2026-11-05T16:08:49.807403+00:00. This checkpoint supersedes November4 maintenance validation. It does not authorize publication, installation or wallet operations.

## Official release and integration

Official [Alby Hub v1.24.6](https://github.com/getAlby/hub/releases/tag/v1.24.6) was published 2026-11-05T09:45:44Z, not draft/prerelease. Enumerated53 releases/74 tags. Annotated tag object `e6f9ac07c505776f417e83534ee796dee9c9fe3f` peels to **921003b65ece004551612647ce1df22327a2486d**, also upstream default master. API release/tag object/compare and configured upstream Git refs independently agree; final refs unchanged. v1.24.5 remains the previously observed source tag; v1.24.6 is the newer published release. No wallet binary was downloaded.

Two new commits since integrated708e3669: [0489c7cf](https://github.com/getAlby/hub/pull/2681) reuses the fetched LNURL invoice on retry after an ambiguous response, relying on existing same-invoice pending/settled checks, invalidates it on amount/comment changes and disables those inputs while sending; [921003b6](https://github.com/getAlby/hub/pull/2682) guards malformed metadata across backend app/get_info handlers and frontend transaction/app/avatar/messageboard/CSV paths. These prevent duplicate-payment risk and malformed-metadata crashes in their documented scopes. This source repair does not prove installed-wallet behavior or authorize any payment.

Existing isolated branch `codex/alby-security-20261021` in `/Users/kode/Development/alby-security-maintenance-20261021/hub` merges this exact upstream source without conflicts. Manifest provenance updated. Exact final tested/build commit **9c04ab63029bab4e88436b0d49581793063f89b6**; subsequent documentation-only commit records evidence. Existing precommit hook formatted the merged messageboard file before the tested commit. Original `/Users/kode/Development/grynvault-rebalance-rc3` stays clean/unchanged at `4c5dc1c1ce7e8284cd5ad49b87d5dab95f09776c` on `codex/coinos-inbound-liquidity`.

## Preserved safety boundaries

All exact-channel API files, LDK adapter/native replacement, Wails route-log redaction, DB/schema/migrations, guided import, Bark spendable-only correction, dependency lockfiles and pinned/read-only/guarded workflows have zero diff against November4 headba298359. Both Grynvault quote tables remain; no channel substitution, fee/expiry/fingerprint weakening or automatic route retry. Exact-channel parent `cd6bc1e399ba256adb113bbaf44d659c5b22a170`, redaction `59bb47a41eacaccc5a95ed717d87befb6759bbe5` and official master are ancestors of tested head. The ordinary-send LNURL invoice cache is separate from exact-channel quote execution.

Go1.27.3/Wails2.17.0/gRPC1.83.2 and MPTCP-disabled default retained. LDK `22c66d60ffe13e6dd0532d7f22501aa55b6a769f` and binding `ea7860a8558930f501df672616b88c0a72c6c584` stay clean/unchanged. Universal native library reused byte-for-byte SHA256 `bb487c9c2640e4c053aa508998f459b6b03216be4592ea080a64a4e9b5476f78`; non-macOS libraries inherited/unvalidated. Same private process-local file proxy and narrow custom-binding GONOSUMDB exception; no global/production changes or checksum bypass. Isolated previously SHA-verified Node22.23.4 and private Wails CLI reused, no tool installation.

## Exact validation

| Check | Evidence/result |
| --- | --- |
| Full Go | count=1 complete suite PASS at 9c04ab63029bab4e88436b0d49581793063f89b6; SQLite migration and continuity verbose tests PASS |
| Race/vet/module/HTTP | Expanded migration/db/API/Wails/guidedimport/Bark/apps/NIP47/LDK/swaps race PASS; vet, module verify, HTTP build PASS |
| PostgreSQL | Fresh PostgreSQL16.13 fixture migration/db/transactions and migration race PASS; server stopped and private socket removed |
| Wails-tag | Initial build overlapped frontend dist generation, causing missing embedded files. Repeated AFTER completed frontend/package build: fails ONLY existing sandbox integration on occupied127.0.0.1:21421. Explicit-exclusion remainder PASS. Both attempts retained; no source repair/bypass. Existing listener/process never inspected/stopped |
| Frontend | Frozen install/lint/types/format/HTTP and Wails builds PASS; production audit134 dependencies/zero findings |
| Offline behavior | `offline-regressions.cjs` executes actual transpiled LNURL component with mocked React hooks and requests only. Ambiguous-response retry uses same fixture invoice; changed amount/comment invalidates cache; inputs disabled during pending request. Same test rejects November4 source. Actual metadata-normalization/CSV functions tested with malformed values and inherited constructor labels. No valid invoice, wallet/network request or payment created; fixture strings only. First harness run had an insufficient microtask wait, corrected to an event-loop turn; final PASS |
| Automated review | CodeRabbit0.7.1 recovered; completed full34-file incremental review versus2f58734a (covers November4 unreviewed delta AND today's changes), ZERO findings. Closes connection hold only. Earlier full-maintenance Critical unpublished-dependency finding remains unresolved; not independent audit |
| Rust | Fresh10 exact-channel/duplicate/CLTV +2 persistence fixture tests PASS. Unchanged source/library. Fresh raw audit exit1:15 vulnerability entries,5 unmaintained warnings,1 unsoundness warning. Vulnerability IDs/versions unchanged; new dev-only pin-utils warning described below |
| Go scan | Zero reachable;1 imported-package/3 required-module nonreachable findings. Native/OS trust not covered |
| Source/package guards | Full maintenance-range and staged-diff secret scans PASS; source boundaries PASS. Workflows and release shell code unchanged except manifest base string; prior actionlint/syntax evidence retained. Nonexistent/unapproved tag rejected before artifact/tag/DMG creation |
| Universal macOS | Exact-source private untagged build and static inspection PASS: independent identity/data path, macOS15.4 floor, both architectures, reviewed rpath, exact source/acceptance label/native hash/ad-hoc integrity. Both slices retain Wails2.17/MPTCP-disabled metadata; app never launched. Acceptance wrapper lacks production sandbox entitlements by design, must not be installed/published |
| Artifact secrets | Raw exit1, four verified EXACT SAME public literals as November4:2 Go h1 checksums and identical concatenated Go DNS-format/decimal-constant/poll-error literal in both slices. Redacted raw findings/assessment retained, no literal credential found or blanket suppression. Text/strings scanning is not exhaustive native analysis |
| Private archive/SBOM | Source archive SHA256 `4376665fcc81d3f005824f125ae5a81fd71e6a96ae7d69a76db66f3cea2572b3`; CycloneDX1100 components SHA256 `c344abe06e1fd7a0ded7f0cdf485d6672b4398fc21c8d2cb30cf0055f0faa1c5` |

## Audit and release holds

RustSec database last commit `26fd2d43f7280a1fb925a3743ec463017afe7119` (2026-11-05T09:50:49+01:00) reports unchanged15 vulnerability IDs/versions. Production webpki-roots0.25.4/1.0.9 (RUSTSEC-2026-0381) remain flagged despite previous verifier fix; production Tokio1.53.2 (RUSTSEC-2026-0386) remains flagged, affected-function reachability neither proven nor dismissed. New unmaintained warning RUSTSEC-2026-0441 pin-utils0.1.1 is absent normal/build production tree. Registry requests to check yanked packages timed out; advisory JSON completed but yanked-status check is incomplete. No clean raw audit, zero unresolved Critical/High or full-remediation claim. Native Bark/OS trust outside these scans.

Both proposed dependency remote codex branches still absent (exit2/no errors). October22 two-dependency publication packet/SHAs/CI effects/exclusions unchanged, unapproved/unexecuted. No repeated approval question solely due heartbeat; silence/time is not approval. Fresh standard remote module/sumdb/native hash readback mandatory after any separately approved dependency publication. Hub push, upstream PR, tag, PGP signing, release, installation and deployment remain separately gated. Occupied-port integration and prior guided-import/runtime acceptance remain open. CodeRabbit recovery does not close these holds.

No push, tag, PGP signing, PR, install, deploy, publication/public release change or new chat. No wallet secrets/data/process/log access, copied-wallet connection, real invoice, Lightning probe, HTLC, payment, swap, rebalance or channel operation. Fixture/mocked tests only; app and HTTP executable built/statically inspected only, never launched. Historical public/installed state not reaccepted.

Exact official snapshots, commands, matrices, recovered Wails attempts, offline harness, review, scan assessments and static provenance are in `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261105`; `evidence-hashes.json` indexes the evidence. Prior evidence is preserved.

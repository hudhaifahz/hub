# Alby v1.24.5 tag maintenance — 2026-11-04

**DETECTED AND PREPARED LOCALLY — TESTED WITH OPEN GATES — NOT PUSHED, RELEASED, OR INSTALLED.**

Recorded 2026-11-04T16:15:07.279795+00:00. This is a local maintenance checkpoint, not release readiness, independent audit or installed-wallet acceptance.

## Official state

Official API enumerated 52 releases and 73 tags. Latest published release remains [v1.24.4](https://github.com/getAlby/hub/releases/tag/v1.24.4), published 2026-10-22T08:59:21Z, confirmed again at completion. New annotated tag v1.24.5 has object `da9b6b5b32aff46f495815d4fc984ed8bf8a6952`, peeled commit **708e3669d69362575d5e20ab4d94731e4d45b6ef**, also the default master HEAD. Configured upstream Git refs, API tag object, tags list and compare agree. Release-by-tag endpoint returns HTTP404: do not describe v1.24.5 as a published release. The original v1.24.4 tag remains unchanged.

[Nine upstream commits](https://github.com/getAlby/hub/compare/a71bd10aafd23afad286b8d03d89e96f66752606...708e3669d69362575d5e20ab4d94731e4d45b6ef) since the prior integrated snapshot are included: bip39 2.4.0 f115c3fc; Sonner 2.0.8 cdf0306c; commitlint CLI21.2.3 10a5ef00; on-curve app-key validation a147e078; notification redaction/delivery handling 6dab5ae4; commitlint config21.2.4 1be0245f; no-relay-ack failure reporting and go-nostr update 0ce942aa; LDK event-subscription shutdown caefe882; and swap/refund address-network validation before remote calls 708e3669.

Notification logs now use allowlisted fields rather than full transaction/notification data, excluding preimages and metadata. This is source-level hardening, not evidence that installed logs are repaired. No wallet process, credentials, data or logs were inspected.

## Preparation and invariants

Reused isolated `codex/alby-security-20261021` in `/Users/kode/Development/alby-security-maintenance-20261021/hub`. Exact tested/build source **074f381f6fc094fb4868c03dd5a01e20b2b4c3bf** merges the official snapshot; a subsequent documentation-only commit records this report. Only go.sum conflicted; selected new official go-nostr checksums while preserving the existing custom native replacement and sums. Manifest upstream provenance now names 708e3669d69362575d5e20ab4d94731e4d45b6ef. No reset, force push or tag creation.

Original `/Users/kode/Development/grynvault-rebalance-rc3` remains clean/unchanged at `4c5dc1c1ce7e8284cd5ad49b87d5dab95f09776c`, branch `codex/coinos-inbound-liquidity`. Both original/maintenance retain exact-channel ancestor `cd6bc1e399ba256adb113bbaf44d659c5b22a170` and route-redaction ancestor `59bb47a41eacaccc5a95ed717d87befb6759bbe5`. Exact-channel API, Wails route redaction, guided import, schema/migration code and Bark spendable-only correction have no diff against November 3. LDK adapter change only handles closed event subscriptions during shutdown; no exact-route fallback/retry changes. Existing quote tables, constraints, fee/expiry/fingerprint checks and pinned/read-only/guarded workflows remain intact.

Go1.27.3, gRPC1.83.2, Wails2.17.0 and MPTCP-disabled default remain. Custom native LDK/binding trees stay clean at `22c66d60ffe13e6dd0532d7f22501aa55b6a769f` / `ea7860a8558930f501df672616b88c0a72c6c584`; raw universal library SHA256 `bb487c9c2640e4c053aa508998f459b6b03216be4592ea080a64a4e9b5476f78` is unchanged. No native rebuild/replacement; non-macOS libraries remain inherited/unvalidated.

New go-nostr pseudo-version `v0.0.0-20261104074428-45060cc10054` verifies public h1 `cWThPhGqugukEIMa1vZKZ4TRjzlxneKZ7JyHt4GFrbg=`. Initial Go suite failed at dependency setup due to official proxy ZIP unexpected EOF. Retrieved unchanged official proxy mod/info/zip with verified TLS through curl, cached in the existing private file proxy, then successfully ran Go module verification and the complete suite. No checksum/security bypass: the process-local narrow GONOSUMDB exception remains only for unpublished custom binding, never go-nostr. Initial errors and recovered results retained separately. Initial uncompressed GitHub API timeouts recovered using compressed HTTPS requests; incomplete JSON was never accepted.

## Validation

| Check | Exact result |
| --- | --- |
| Go | Recovered full count=1 suite PASS; SQLite migration/continuity verbose checks PASS; expanded race suite includes migrations/db/API/Wails/guided import/Bark/apps/NIP47/LDK/swaps and PASS; vet/module verify/HTTP build PASS |
| PostgreSQL | Fresh PostgreSQL16.13 synthetic fixture db/migrations/transactions tests and migration race PASS; server stopped, short socket removed; no existing database/service touched |
| Wails tests | Full root test invocation fails ONLY sandbox integration on occupied127.0.0.1:21421; explicit exclusion remainder PASS. Existing listener/process never inspected or stopped; exclusion is not acceptance |
| Frontend | SHA-verified isolated Node22.23.4 reused; frozen install/lint/types/format/HTTP and Wails builds PASS; production audit134 dependencies/zero findings |
| Go vulnerabilities | Zero reachable;1 imported-package/3 required-module nonreachable findings remain; native and OS security not covered |
| Native Rust |10 exact-channel/duplicate/CLTV tests and2 persistence tests PASS with fixtures only; unchanged source/binary. Fresh raw cargo audit exit1:15 vulnerability entries,4 unmaintained,1 unsoundness warning, same versions/advisories as November3 |
| Source | Boundaries, changed-shell syntax, staged diff secrets, full maintenance-range git secrets PASS. Workflows unchanged; prior actionlint1.7.7 evidence retained rather than repeated |
| Automated review | CodeRabbit0.7.1 authenticated; two incremental review attempts vs2f58734a failed connecting with Unknown error/TRPCClientError. NO completed review or zero-finding claim. Earlier Critical unpublished-dependency finding remains open |
| Packaging | Exact-source universal macOS private untagged build and static inspection PASS: independent identity/data folder,15.4 floor, reviewed rpath, native hash, source/owner-acceptance label and ad-hoc integrity. Both slices embed new go-nostr, Wails2.17 and multipathtcp=0. No app/HTTP execution |
| Tag guard | Nonexistent/unapproved tag rejected before output/tag/DMG creation. No tag-creating regression script executed |
| Artifact secrets | Raw exit1; four findings verified exact same public literals as November3: two module h1 checksums and duplicated concatenated Go DNS-format/decimal-constant/poll-error literal in both slices. No literal credential found; raw redacted findings retained, no blanket suppression. Supported text/strings only, not exhaustive native analysis |
| Private source/SBOM | Archive SHA256 `9ca8e34524752d94975aa812e4b9641840a03af3ec4cc95d860b68f89e1b2ced`; CycloneDX1067 components, SHA256 `41e4cbea4d8fe42f93e97596f4829a5d4ee2655b022b3964f47ee1106549c819` |

## Open gates and owner control

Production root stores0.25.4/1.0.9 remain flagged by RUSTSEC-2026-0381 despite the previously fixed verifier; production Tokio1.53.2 remains flagged by RUSTSEC-2026-0386. Affected-function reachability neither proven nor dismissed. Other flagged vulnerability versions absent normal/build production tree. No clean raw audit, zero unresolved Critical/High or complete remediation claim. Updated native Bark/OS trust is outside these scans.

Both proposed dependency remote `codex/alby-security-20261021` branches remain absent (exit2, no errors). The October22 concrete two-dependency publication packet, exact SHAs, CI effects and exclusions remain unchanged/unapproved/unexecuted. Standard proxy/sumdb/remote-native-hash readback remains mandatory after separately approved publication. No repeat approval question merely because the heartbeat fired; silence/time is not approval.

CodeRabbit connection outage, native dependency publication/reproducibility, raw Rust findings, occupied-port sandbox integration and prior guided-import/runtime/release gates remain open. Private untagged acceptance wrapper intentionally lacks production sandbox entitlements and must not be installed/published. No hub push/PR/tag/PGP/release/install/deploy authorization follows from a passing build.

No push, tag, PGP signing, PR, installation, deployment, publication/public release change, new chat, wallet secrets/data/process/log access, copied-wallet connection, invoice, Lightning probe, HTLC, payment, swap, rebalance or channel operation occurred. Tests use synthetic fixtures/mocks only. The app/HTTP executable were built and statically inspected only, never launched.

Exact commands, initial failures/recovery, official snapshots, test matrices, scan assessments, package provenance and hashes are in `/Users/kode/Development/alby-security-maintenance-20261021/evidence/20261104`; `evidence-hashes.json` indexes the evidence. Earlier November3 evidence remains unchanged.

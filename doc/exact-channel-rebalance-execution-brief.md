# Exact-channel rebalance execution brief

Use this brief to resume and complete the Grynvault exact-channel rebalance
release. It supplements, but does not replace, the detailed release checklist.

## Reusable goal prompt

Complete the Grynvault exact-channel rebalance release by following this brief.
Prioritize the installed-app UI repair, then cut and verify immutable RC.3,
prove clean installation and copied-data migration with rollback, refresh the
security/test evidence, prepare the current upstream PR chain, update every
public surface, and only then prepare Nostr and Stacker News publication.
Do not implement the deferred stock-Alby payment provider described in
`doc/stock-alby-rebalance-provider-plan.md`. The owner concluded on 2026-10-15
that peer-targeted best-effort routing does not satisfy the product's purpose.
Preserve that document as the rationale and future compatibility analysis only;
focus implementation on exact local-channel enforcement and upstream adoption.
Keep all records current and stop only at the explicit owner approval,
value-moving, signing, deployment, PR-opening, or publication gates below.

## Read first

Read these files completely before acting:

- `AGENTS.md`
- `doc/exact-channel-rebalance-release-plan.md`
- `doc/exact-channel-rebalance-release-gate.md`
- `doc/pinned-rebalance-reference-plan.md`
- `doc/exact-channel-rebalance-launch-drafts.md`
- `doc/macos-package-safety.md`
- `doc/stock-alby-rebalance-provider-plan.md`
- `/Users/kode/Development/btc_jv/AGENTS.md`

Update the release plan and engineering record throughout the work. Record exact
commits, tests, findings, artifact hashes, deployments, approvals, failures, and
independent readback evidence. Commit and push documentation after meaningful
milestones.

## Repository map

- Hub: `/Users/kode/Development/albyhub-v1.24.1`
  - GitHub: `hudhaifahz/hub`
  - Branch: `codex/upgrade-v1.24.1`
  - Planning baseline: `a6c77555db7a216e7ff76cd59b3371c0e119907e`
- LDK: `/Users/kode/Development/ldk-node-v1.24.1`
  - Required feature commit: `05d86840f2184dbe846e260b71e7eeed2bea998a`
- LDK Go: `/Users/kode/Development/ldk-node-go-v1.24.1`
  - Required binding commit: `2d4b8cae5b4188451fbccdb347113838a9981dbf`
- Frontier Crown portal: `/Users/kode/Development/CrownFrontier`
  - GitHub: `Grynvault/frontierCrown`
  - Preserve unrelated dirty changes.
  - Isolated Portal 04 worktree:
    `/Users/kode/.codex/worktrees/rebalance-portal-04/CrownFrontier`
  - Portal baseline branch/head: `codex/rebalance-portal-04` /
    `873833e9364ce693e02e86f129125b712a84d579`
  - Security branch/head: `codex/rebalance-portal-security` /
    `7645db5a813a0c28926b4e780078aa76adbe5421`
- Grynvault root site: `/Users/kode/Development/btc_jv`
  - GitHub: `Grynvault/btc_jv`
  - This repository serves `grynvault.com`. Do not deploy the Frontier Crown
    frontend over it.
- Hosted companion: `/Users/kode/Development/grynvault-rebalance`
  - Dedicated read-only Cloudflare Worker repository; no wallet source,
    credentials, storage, database, or execution binding.
  - Private GitHub: `hudhaifahz/grynvault-rebalance`
  - Main head: `71d3c4dbfd159efc45b4761f1cf36ca4bc75f99c`
  - Prepared unsigned-metadata branch/head:
    `codex/unsigned-release-metadata` /
    `24fc2475c18b5e69559bfd2c757154e30f2afd3c`
  - Clean GitHub Actions run: `38482199397`
- Independent release worktree:
  `/Users/kode/Development/grynvault-rebalance-release`
  - Branch: `codex/grynvault-rebalance-release`
  - Base: `1fb6c8cd0fcb2fbee48c888c6274e2659d74c169`

## Current resume checkpoint

The independent release source remains immutable at annotated tag
`grynvault-rebalance-v0.1.0-rc.2`, tag object
`c4270fbaa40a183947562c0460b6e3c466f9a8eb`, and exact commit
`29a34b97fdd890f1895ca9827ba0e59f8457f3ee`. GitHub Actions run
`38521293804` passed the supported-macOS source/release checks. The public
prerelease at
`https://github.com/hudhaifahz/hub/releases/tag/grynvault-rebalance-v0.1.0-rc.2`
contains the explicitly unsigned/unnotarized universal macOS 15.4+ DMG and its
source, SBOM, manifest, detached signature, public key, checksum, Gatekeeper,
and preparation evidence. A fresh external download verified DMG SHA-256
`b329d7bec494cfff6b6c4f2865f2aa13e08a11e96726696b869cd0309c6dfe1a`
and a good signature from exact PGP fingerprint
`D131A3D0EAB0CF98A8413853CD9FDBF9D075C126`. Read-only mounting and static
inspection confirmed exact rc.2/source metadata, independent bundle identity,
both architectures, macOS 15.4+, the reviewed runtime path, ad-hoc integrity
signature, no Team Identifier, and expected Gatekeeper rejection. It was not
launched or installed.

The canonical companion at `https://rebalance.grynvault.com` is live with
current metadata and exactly five read-only MCP tools. Negative acceptance
rejected a fabricated execution tool, hostile Origin, and extra diagnostic
fields. It has no wallet, NWC, invoice, route, payment, or execution binding.
Frontier Crown PR `Grynvault/frontierCrown#6` was reviewed, tested, and
squash-merged with exact-head protection at
`18b1e6efdc381c3f0ea516150e865c42c2a8eb30`. Netlify production deploy
`6ad0542ffa83daccd5046625` is ready from that exact commit; the live portal
shows RC.2, the full PGP fingerprint, explicit warnings, and exact immutable
release/source/docs links. The final CodeRabbit CLI review is recorded in the
release gate; the repository was not recognized as connected to a CodeRabbit
organization, so no GitHub-App or independent-audit claim is permitted.

The running `/Applications/Alby Hub.app` is still the earlier owner diagnostic
build, not the branded public rc.2 package. Resume with the owner-controlled
clean-install and copied-data migration gate only after repairing the
installed-app dialog and producing a new immutable release candidate. RC.2 is
immutable and must never be replaced in place.

RC.3 immutable source now exists separately on owner-fork branch
`codex/grynvault-rebalance-rc3-prep` at exact commit
`9583e7a230b57d02fe4bb187f47a1fcb9804ddab`. It is based directly on the
reviewed Alby Hub v1.24.3 contribution head
`cd6bc1e399ba256adb113bbaf44d659c5b22a170`, is zero commits behind that base,
and carries only the independently branded release/package layer above it. On
2026-10-16, the owner approved and the repository published annotated tag
`grynvault-rebalance-v0.1.0-rc.3`, tag object
`613a9a63a0842d2895ba1d0461bae6a9dd819d45`, peeling to that exact commit.
GitHub source-compatibility run `38857778185` completed successfully, and a
fresh CodeRabbit CLI review of the complete RC.3 release layer against exact
parent `cd6bc1e399ba256adb113bbaf44d659c5b22a170` completed with zero findings
across 64 reviewed files. A tag-specific universal macOS build, raw and mounted
bundle inspections, ad-hoc package inspection, source archive, 1,106-component
SBOM, unsigned manifest, and preparation index all passed. The private DMG has
SHA-256 `253941da436a04bb5a314b45b915544b8fcac9e65f8f3be8fe59cdc560c1511e`.
The owner separately approved a detached armored PGP signature over manifest
SHA-256 `731b31e877c664beb1afe35f8cd1ebe5368e57511b9874fba18a869c191df850`.
An isolated-keyring verification returned `GOODSIG` and `VALIDSIG` for full
fingerprint `D131A3D0EAB0CF98A8413853CD9FDBF9D075C126`; the signature SHA-256 is
`c064ac2807a9eab41f7f90bb56ac27ec5fb2a29df9f7d66f4f73f8d827a7c9e3`.
The owner-approved offline RC.3 clean-install, real copied-wallet migration,
and rollback acceptance completed on 2026-10-16. The run corrected the canonical
macOS work-directory record to the live Alby container path, created and tested
a restricted backup, proved password/seed identity and storage continuity under
external-network denial, and restored the untouched canonical Hub. The copied
node intentionally could not complete Electrum synchronization, so no second
copy was ever online. No public RC.3 GitHub release, portal/companion deployment,
upstream PR, Nostr event, Stacker News submission, or Lightning operation was
created by this acceptance.

One fresh owner-approved `10,000 sats` (`0.00010000 BTC`) Kraken-return proof
pilot has reconciled successfully on the installed owner build. Durable
readback reported `succeeded / succeeded`, exact actual routing fee
`8,981 msat` (`8.981 sats`), zero operations left in `executing`, database
integrity `ok`, the intended outgoing/Kraken balance shifts, and all four
channels online. This proves one operation on the installed owner build; it is
not public-package clean-install or migration acceptance and authorizes no
retry or later payment.

The owner rejected the installed narrow-window dialog on usability grounds.
Long identifiers and quote values can be clipped, labels can separate from
their values, and the footer consumes too much of the available height. Earlier
synthetic geometry evidence is not installed-app visual acceptance. Fix and
verify this before the main public launch.

Never execute an expired quote, batch pilots, or move sats without the owner's
action-time approval of the exact channels, principal, eight-decimal BTC
equivalent, fee, cap, maximum debit, balances, expiry, quote ID, and route
fingerprint. Upstream PRs, Nostr signing, and Stacker News
submission/payment remain separately gated.

## Priority completion sequence

Treat this order as the active priority for the resumed goal. Do not skip ahead
from a prepared or locally passing state to a public-completion claim.

1. **Repair and accept the installed UI.** Implement a responsive stacked
   label/value layout at constrained widths; use compact copyable monospace
   fields with `overflow-wrap: anywhere` for exact identifiers; keep the active
   review/status content independently scrollable; and reduce the action
   footer's footprint without obscuring evidence. Verify the real component and
   installed desktop app at 320, 375, 768, and 1280 CSS pixels plus 200% zoom in
   pre-quote, quoted, executing, succeeded, failed, and recovery states. Require
   owner visual acceptance.
2. **Create immutable RC.3.** Because RC.2 cannot be changed, record a new exact
   source commit and annotated tag. Rebuild the universal macOS package, source
   archive, SBOM, provenance, checksums, Gatekeeper evidence, manifest, and
   release notes. Preserve the explicit unsigned/unnotarized classification.
   Stop for the owner's exact tag and PGP-signing approvals.
3. **Prove the downloadable package.** On a clean macOS 15.4+ profile, verify a
   clean install with no wallet data. Using only an owner-approved copy of the
   wallet data, verify migration, unlock, database integrity, channels,
   balances, transaction history, NWC apps and permissions, backup state,
   embedded source/library hashes, UI behavior, and rollback. Do not move sats
   during installation acceptance.
4. **Refresh security and regression evidence.** Run the complete Go, Rust,
   binding, frontend, migration, package, dependency, secret, and release-tree
   checks against the exact RC.3 source. Require zero unresolved Critical or
   High findings and document all residuals. Do not call this an independent
   audit.
5. **Prepare the current upstream PR chain.** Rebase the isolated LDK, LDK Go,
   and Hub contributions on current upstream heads, rerun their full matrices,
   refresh the PR bodies and screenshots, and stop for owner review before
   opening the three public PRs. Never merge on behalf of upstream maintainers.
6. **Refresh every public surface.** Publish only after the separate release
   approval, then update the read-only companion and Portal 04 to exact RC.3
   metadata and immutable links. Independently read back every download, hash,
   signature, source commit, certificate, endpoint, portal action, and PR URL.
7. **Publish the launch posts last.** Prepare the final Nostr kind-1 event and
   show its exact content, tags, relay set, and timestamp policy for separate
   signing/publication approval. Read it back from a target relay. Then perform
   a fresh exact-URL Stacker News duplicate check and show the exact title, URL,
   territory, description, and current fee for a separate submission/payment
   approval. Reconcile one posting attempt before any retry.

### Paid-service boundary

The public open-source announcement does not authorize selling an operator
service. Before the first paid customer, separately complete at least three
fresh intended-channel proof pilots one approval at a time, prove one safe
expiry or no-route case moved no sats and incurred no fee, publish the support
scope/refund/privacy/security/incident-stop terms, refresh the release evidence,
and obtain owner approval for the exact offer and payment mechanism. Never
operate a customer's wallet or remote-click the customer's value-moving
confirmation.

## Proposed defaults

Use these unless current evidence reveals a conflict:

- Product: Grynvault Rebalance
- Initial platform: universal macOS
- Independent name, icon, bundle identifier, data directory, and updater
- Immutable release assets in GitHub Releases under `hudhaifahz`
- Companion service at `rebalance.grynvault.com`
- Isolated stateless Cloudflare Worker on an unadvertised provider hostname,
  followed only after private acceptance by narrowly scoped canonical DNS
- MCP SDK v2 with modern revision `2026-07-28` and stateless Streamable HTTP
  compatibility revision `2025-11-25` explicitly pinned
- Open-source release followed by a paid installation, migration, recovery, or
  diagnostic-support experiment
- Operator-service boundary and paid-pilot gates:
  `doc/exact-channel-rebalance-operator-service.md`

## Product boundary

The local Hub/LDK application performs exact first-hop and receiving-channel
enforcement. The hosted HTTP/MCP companion publishes verified releases,
compatibility results, prerequisites, manifest verification, documentation, and
sanitized diagnostic explanations.

The hosted service cannot replace the local patch for forced routing. It must
not execute payments or rebalances, proxy Hub owner APIs, accept an NWC
connection, or claim it can force routes in an unmodified Alby Hub.

The proposed `provider.rebalance.grynvault.com` payment service is deferred and
must not be implemented or deployed under the current goal. Stock Hub does not
transmit both local channel IDs or accept a provider-selected first hop, so its
peer-targeted best-effort flow does not satisfy the exact-channel product
purpose. Reconsider it only after explicit owner direction or upstream protocol
changes. Keep the existing `rebalance.grynvault.com` companion read-only.

Never request or accept seeds, passwords, recovery material, macaroons, NWC
connection strings, API tokens, private keys, invoices, preimages, payment
secrets, serialized routes, wallet databases, or unredacted wallet logs.

## Execution order

Follow the waves in `doc/exact-channel-rebalance-release-plan.md`:

1. Reconcile current source, installed application, upstream, artifacts, portal,
   DNS, and deployment topology.
2. Close non-paying owner-build acceptance. Verify the canonical database,
   migrations, embedded library, sync, channels, balances, and responsive UI.
3. Prepare clean upstream contributions for LDK, LDK Go, and Hub.
4. Rebase on current upstream and repeat the complete test and security matrix.
5. Create an independently branded, source-pinned, explicitly unsigned and
   unnotarized release candidate with prominent Gatekeeper warnings, a
   PGP-signed manifest, and explicit migration and rollback.
6. Build and privately deploy the bounded HTTP/MCP companion; complete its
   security tests before canonical DNS.
7. Prepare Portal 04 with immutable release, source, verification, release-note,
   and security links. Keep binaries out of the portal repository.
8. Capture privacy-safe screenshots using fixture identifiers.
9. Prepare the GitHub release, upstream PRs, portal deployment, Nostr post, and
   Stacker News submission.
10. Independently read back every public URL, hash, signature, certificate,
    commit, PR head, portal link, and publication before declaring completion.

Do not equate configured, built, tested, deployed, reachable, signed, published,
or independently accepted states.

## Hard safety and approval boundaries

- Do not move or commit sats during release work.
- Every future payment, rebalance, swap, or channel operation requires a fresh
  single-action approval packet under the Lightning routing skill.
- Keep all credentials, MFA, signing secrets, private keys, passwords, and
  wallet secrets outside chat, source, screenshots, logs, and commands.
- Apple Developer ID, notarization, and stapling are omitted by owner decision.
  PGP and Nostr signing and MFA remain owner-controlled.
- Never publish the current ad-hoc owner-test archive.
- Require zero unresolved Critical or High security findings.
- Do not call the work independently audited unless an independent reviewer has
  completed and documented that audit.
- If a finding may endanger users or funds, stop public disclosure and prepare a
  private report for Alby's security contact.

Stop for separate owner approval immediately before each of these actions:

- opening public upstream pull requests;
- publishing a GitHub release;
- changing production DNS;
- deploying Portal 04 publicly;
- signing or publishing a Nostr event; or
- submitting or paying for a Stacker News post.

When owner-controlled authentication or signing blocks progress, complete every
safe preparatory task, preserve a resumable state, and present one concise owner
action request.

## Completion

Completion means every item in the release plan's `Completion evidence` section
is satisfied with current, independently read-back evidence. A local build,
successful test, deployment indicator, or health check alone is not completion.

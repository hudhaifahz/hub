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

The acceptance also exposed a release-blocking log-sanitization defect: a
failed Wails OAuth callback logged its route with the one-time authorization
code. Owner-fork commit `59bb47a4` on
`codex/fix-redact-wails-route-secrets` removes query values from desktop route
logs/errors, redacts payment-request paths, and passes the focused Wails and
complete Go test suites. RC.3 is immutable and must not be modified or promoted
as the final candidate. Prepare a new immutable candidate from the reviewed
security-fix source and repeat the package/signature/install gates before any
public RC update.

RC.4 is now the private accepted successor. Annotated tag
`grynvault-rebalance-v0.1.0-rc.4`, tag object
`df8597eb314217d3b40e89135079e1a0ca2da4cc`, peels to exact source
`4de3883f4d76af2a312962c2d5e7546963bf9261`. Its private universal DMG has
SHA-256 `534d258c9c59b4ac63758de306899ee932f31457d54eb9a849757c931c41345e`;
the owner-controlled manifest signature verified in a public-key-only keyring.
The owner-approved offline clean-profile and copied-wallet acceptance completed
on 2026-10-16 with hash-verified backups, database integrity `ok`, canonical
public node identity, all NWC apps and permissions, transaction snapshot,
static backups, and channel-monitor identities preserved. External networking
was denied throughout the disposable copy's run, and startup stopped at the
expected Electrum boundary. A narrow scan of new RC.4 runtime lines found no
credential-shaped route data. The disposable copy is stopped and canonical
Alby Hub is the only network-connected wallet process. No Lightning operation
was created. RC.4 GitHub publication and portal/companion production updates
still require separate owner approvals.

The later live installation exposed a separate release-blocking first-launch
trust failure. Grynvault correctly used its independent
`com.grynvault.rebalance` container, but the clean welcome flow did not explain
that the owner's existing Alby Hub channels remained in the separate
`com.getalby.Alby-Hub` container or offer an integrated import. This looked like
lost channels even though no wallet data or channel state had been lost. A
manual, hash-verified copy restored the live owner wallet in Grynvault: both
SQLite integrity checks returned `ok`, 12 apps, 94 permissions, 266
transactions, 16 configuration rows, 35 migrations, and 13 local rebalance
records were present, and owner UI readback showed the same four online
channels, 1,080,605 sats Lightning balance, 4,796,754 sats receive limit, and
41,883 sats on-chain. The original Alby Hub data remains untouched. This proves
continuity for this owner-controlled migration only; it does not make RC.4 an
acceptable general installer.

RC.4 is immutable and must not be relabeled or silently replaced. The active
successor work is owner-fork head
`952a6042fb85fbf5fddc4814602e152a13e75ebe` on branch
`codex/guided-alby-import`, based on exact RC.4 source commit
`4de3883f4d76af2a312962c2d5e7546963bf9261`.
Before any RC.5 tag, package, publication, portal update, or promotion:

1. first run must prominently say that channels are not missing and that the
   independent app uses a separate data folder;
2. existing Alby Hub embedded-LDK users must be offered a guided desktop import
   before new-wallet setup;
3. the importer must require the official Hub to be closed, preserve the
   original, create recovery copies of both source and target, reject
   incomplete/non-LDK/symlinked inputs, hash-verify every copied file, validate
   both SQLite databases, record non-secret continuity counts, and stage the
   import for a clean restart;
4. Grynvault and official Alby Hub desktop must hold mutually exclusive
   process locks so copied node data cannot run in both apps simultaneously;
5. a lock conflict must produce a native explanation that the wallet and
   channels are safe instead of an apparently failed launch;
6. clean-install acceptance must exercise cancel, wrong folder, incomplete
   wallet, stock database without Grynvault tables, valid import, restart,
   unlock, identity/channel/balance/history/NWC/backup continuity, rollback,
   and the two-app lock conflict; and
7. installation acceptance must remain non-paying: no invoice, probe, HTLC,
   payment, swap, rebalance, or channel operation.

The current source tests cover verified staging, container-parent discovery,
fresh-target backup, compatibility with a stock database that has no local
rebalance table, incomplete/non-LDK rejection, symlink rejection, and dual-lock
cleanup/retention. Focused Go/Wails tests and frontend lint/type/format checks
pass. The complete Go suite, `go vet`, HTTP build, and Wails build also pass;
only the already recorded Bark macOS-floor, Lottie direct-eval, bundle-size,
crypto-externalization, stale Browserslist, and Node deprecation warnings
remain. A source-level browser preview at 320 CSS pixels measured document
width `320/320` with no horizontal overflow, and the 375-pixel first slide
showed the import action before new-wallet setup. This preview does not exercise
the Wails directory chooser or native restart path. A reviewed commit, guarded
universal owner-test build, package inspection, and owner-visible wallet
continuity readback are complete. A disposable clean-profile import, native
restart/conflict exercise, rollback, and exact tagged RC.5 artifact remain open
before publication can be proposed.
The source head and remote branch were independently read back after push. It
contains implementation commit `7fa5309c324b54752d4139ab8857d887e5888f54`
and follow-up UX commit `78e44201ad67b1b3d865c1ae3aaf5cfbce69994a`,
plus native binding-generation and release-inspection repairs through
`952a6042fb85fbf5fddc4814602e152a13e75ebe`. The installed owner test used
runtime commit `f7bf3881fea11f19c8fc4d6ae9de7e74b88d4ecc`; after local-only unlock it
showed all four channels online, preserved balances, and passed all three SQLite
integrity checks with exactly one Grynvault process. No tag, release, pull
request, public artifact, deployment, publication, or Lightning operation was
created.

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

## Current RC.4 checkpoint — 2026-10-16

Step 1 of the owner-requested successor sequence is complete at exact owner-fork
commit `4de3883f4d76af2a312962c2d5e7546963bf9261`. The route-redaction fix,
malformed-payment-path regression, shipped security disclosure, complete local
matrix, zero-finding final CodeRabbit review, universal macOS preflight, mounted
DMG inspection, expected Gatekeeper rejection, 1,106-component SBOM, source
archive, and secret scans all pass. Disposable tag `preflight-4de3883f` was
deleted after proof.

Proposed immutable tag `grynvault-rebalance-v0.1.0-rc.4` is absent locally and
remotely. Stop for exact owner approval before creating or pushing it. After tag
approval, rebuild tag-specific artifacts and stop again for approval of the
exact manifest SHA-256 before PGP signing. Installation/migration, GitHub
publication, and portal/companion updates remain later, separate gates.

The owner subsequently approved that exact tag action. Annotated tag object
`df8597eb314217d3b40e89135079e1a0ca2da4cc` is now pushed and independently
peels to the approved commit. Tag-specific private artifacts passed raw and
mounted inspection, expected Gatekeeper rejection, hash verification, SBOM
generation, and artifact secret scanning. The unsigned 1,564-byte manifest has
SHA-256 `8f2f5d921249b279993300a88925eb6ebeec476f8290bd80b8ef97ca36b17c7b`.
No signature or GitHub release exists. Stop for separate PGP-signing approval.

The owner subsequently approved and completed that local signing gate. The
265-byte detached armored signature has SHA-256
`21d5aebc2e07917e6ba1653343086ca41a5f7b70399ddc3256bb8d02f51cf129`.
A fresh public-key-only keyring returned `GOODSIG` and `VALIDSIG` for full
fingerprint `D131A3D0EAB0CF98A8413853CD9FDBF9D075C126` and contained no secret
key. Nothing was uploaded or published. Stop for fresh action-time approval
before the offline installation/migration check.

# Exact-channel rebalance execution brief

Use this brief to resume and complete the Grynvault exact-channel rebalance
release. It supplements, but does not replace, the detailed release checklist.

## Read first

Read these files completely before acting:

- `AGENTS.md`
- `doc/exact-channel-rebalance-release-plan.md`
- `doc/pinned-rebalance-reference-plan.md`
- `doc/macos-package-safety.md`
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
    `7054c902890772f34e1d1a9533ec5463862d2e45`
- Grynvault root site: `/Users/kode/Development/btc_jv`
  - GitHub: `Grynvault/btc_jv`
  - This repository serves `grynvault.com`. Do not deploy the Frontier Crown
    frontend over it.
- Hosted companion: `/Users/kode/Development/grynvault-rebalance`
  - Dedicated read-only Cloudflare Worker repository; no wallet source,
    credentials, storage, database, or execution binding.

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

## Product boundary

The local Hub/LDK application performs exact first-hop and receiving-channel
enforcement. The hosted HTTP/MCP companion publishes verified releases,
compatibility results, prerequisites, manifest verification, documentation, and
sanitized diagnostic explanations.

The hosted service cannot replace the local patch for forced routing. It must
not execute payments or rebalances, proxy Hub owner APIs, accept an NWC
connection, or claim it can force routes in an unmodified Alby Hub.

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
5. Create an independently branded, signed, notarized, stapled, source-pinned
   release candidate with explicit migration and rollback.
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
- Developer ID, notarization, PGP, Nostr signing, and MFA remain owner-controlled.
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

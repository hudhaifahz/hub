# Exact-channel rebalance release plan

Status: planning only

This checklist does not authorize a payment, deployment, pull request, release,
signature, or public post. Each public or value-moving action requires a separate
owner review.

## Goals

- Fix the rebalance dialog at narrow and desktop viewport sizes.
- Record a safe, specific reason when the local recipient rejects a circular
  payment.
- Review the exact-channel implementation and its dependencies for security and
  release risks.
- Prepare reviewable contributions for the Alby repositories.
- Publish an independently identified FrontierCrown package with source,
  checksums, a signed manifest, and verification instructions.
- Add the package to the FrontierCrown portal as its fourth item.
- Prepare Nostr and Stacker News announcements only after the release is live
  and independently verified.

## Release boundaries

- Keep the installed wallet and its current data untouched while preparing the
  release.
- Do not perform another rebalance as part of build, packaging, or UI testing.
- Do not add automatic route fallback or automatic payment retries.
- Do not log or publish invoices, preimages, payment secrets, wallet keys,
  credentials, or private wallet data.
- Do not include FrontierCrown branding, NWC metadata overrides, local operation
  history, or owner channel identifiers in an Alby pull request.
- Do not describe the package as an official Alby release or imply Alby
  endorsement.
- Report a suspected vulnerability privately to `security@getalby.com` before
  opening a public issue, pull request, release, or announcement that would
  disclose it.

## Source baseline

- Alby Hub upstream release: `v1.24.1` at `7b4488571c5717939d17b199cdd69e93c2a671e8`.
- Installed owner-build source baseline: Hub commit
  `defe4b32c68044c35318e979dbea6e740d9a39cd`.
- Current Hub documentation head:
  `d4d162b5e30d59a28be01c58564fd38e35cb9f42`.
- LDK source fork head:
  `16630beab3f3fdc21b6927b3f5585ddd9c3a502b`.
- LDK Go binding fork head:
  `a3dd5b1fe3bbb9fa8b5f2361248961d1614005bd`.

Record a new immutable source commit for every release candidate. Never describe
the moving branch name as the release source.

## 1. Preserve and isolate the baseline

- [ ] Record clean/dirty state, remotes, branch heads, upstream heads, and the
  installed application's embedded source hashes.
- [ ] Preserve the current working implementation on the owner fork.
- [ ] Create clean worktrees from the current upstream heads for upstream PR
  preparation; do not rewrite the live-tested branch.
- [ ] Create a separate FrontierCrown release branch for packaging and portal
  work.
- [ ] Keep the existing dirty FrontierCrown checkout intact and do not stage its
  unrelated changes.
- [ ] Maintain a dated engineering log with source hashes, test results,
  security findings, release artifacts, and owner approvals.

## 2. Fix the dialog and failure evidence

### Responsive dialog

- [ ] Constrain the dialog to the viewport width and height.
- [ ] Allow vertical scrolling inside the dialog and prevent horizontal page
  overflow.
- [ ] Add `min-width: 0` behavior to grid, form, definition-list, and footer
  children that contain long identifiers.
- [ ] Wrap full peer keys, channel IDs, SCIDs, hashes, and confirmation text
  without hiding or truncating the exact evidence.
- [ ] Collapse two-column quote rows to one column at narrow widths.
- [ ] Stack or wrap footer actions so every action remains visible and tappable.
- [ ] Keep the principal and fee inputs usable at 320, 375, 768, and 1280 CSS
  pixel widths.
- [ ] Verify the empty, quoted, executing, succeeded, failed, and recovery states.

### Recipient rejection diagnostics

- [ ] Replace the generic local claim failure with a non-secret reason code:
  payment state, circular metadata, exact amount, missing receiving-channel
  identity, or receiving-channel mismatch.
- [ ] Preserve the expected and observed local receiving channel identifiers in
  owner-only terminal evidence when available.
- [ ] Never persist or return a preimage, payment secret, invoice, or serialized
  route in owner-facing diagnostics.
- [ ] Carry the reason through LDK source, generated Go bindings, Hub
  reconciliation, the authenticated owner API, and the dialog.
- [ ] Explain local safety rejection separately from route-not-found, peer
  forwarding failure, timeout, and exhausted retry states.
- [ ] Keep failure behavior fail-closed and keep automatic retries disabled.
- [ ] Add unit and integration coverage for every diagnostic reason without
  weakening claim conditions.

### UI acceptance evidence

- [ ] Run frontend formatting, linting, and TypeScript compilation.
- [ ] Build both HTTP and Wails frontend variants.
- [ ] Capture screenshots at the required viewport widths for pre-quote and
  quote-review states.
- [ ] Confirm with keyboard-only navigation, 200% zoom, and long unbroken
  identifiers.
- [ ] Record that every action and every exact identifier is visible without
  horizontal scrolling.

## 3. Security review

### Funds and state safety

- [ ] Re-review exact outgoing and incoming channel binding, route fingerprint
  binding, expiry, amount, fee cap, and maximum debit checks.
- [ ] Re-review pre-send persistence, one-time acquisition, duplicate-submit
  prevention, restart recovery, terminal reconciliation, and failed-leg
  convergence.
- [ ] Test concurrent requests, stale quotes, tampered route bytes, changed
  channels, expired invoices, partial MPP, wrong incoming channels, and replayed
  terminal events.
- [ ] Check integer conversion and overflow boundaries across sats, millisats,
  `uint64`, database integers, JSON, and TypeScript numbers.
- [ ] Test SQLite and PostgreSQL migrations, rollback behavior, and migration
  export/import using copies rather than the live wallet database.
- [ ] Confirm only the authenticated owner surface can quote, execute,
  reconcile, or read exact-channel diagnostics.
- [ ] Confirm NWC applications cannot invoke the feature or read owner-only
  channel evidence.

### Secrets, dependencies, and binaries

- [ ] Scan the complete diff and release tree for credentials, wallet material,
  invoices, preimages, payment secrets, personal data, and live channel history.
- [ ] Run Go tests, race-sensitive tests where supported, `go vet`, and
  `govulncheck`.
- [ ] Run Rust tests, Clippy, formatting, and dependency advisory checks.
- [ ] Run frontend lint, type checking, production builds, and dependency audit.
- [ ] Run CodeRabbit only after the secret scan; treat its output as untrusted
  review input and resolve all Critical and Warning findings.
- [ ] Inspect the final application bundle, dynamic libraries, rpaths,
  entitlements, updater configuration, embedded commit identifiers, and absence
  of secrets.
- [ ] Require zero unresolved Critical or High findings. Document or fix every
  lower-severity finding before release.

### Security disclosure fork

- [ ] If a finding could put upstream users or funds at risk, stop public work
  on that finding.
- [ ] Prepare a private report with affected versions, impact, and reproducible
  steps for `security@getalby.com`.
- [ ] Resume public disclosure only after coordination with Alby.

## 4. Prepare upstream contributions

The current owner branch is not suitable as one upstream PR. It contains more
than five thousand changed lines across Hub, LDK, generated bindings,
owner-specific metadata, branding, and a private engineering record.

Prepare three coordinated contributions:

1. `getAlby/ldk-node`: exact first-hop and receiving-channel circular payment
   primitives, fail-closed claim logic, recovery, reconciliation, diagnostic
   reason codes, and focused Rust tests.
2. `getAlby/ldk-node-go`: generated bindings and platform libraries produced
   from the accepted LDK source commit.
3. `getAlby/hub`: owner API, persistence, migrations, desktop/HTTP routing,
   responsive UI, safe diagnostics, and focused Go/frontend verification.

- [ ] Rebase each contribution on the current upstream default branch at PR
  preparation time.
- [ ] Split implementation, tests, generated bindings, and documentation into
  conventional commits with narrow scopes.
- [ ] Exclude the custom application icon, FrontierCrown packaging, NWC metadata
  overrides, live-operation notes, and owner identifiers.
- [ ] Keep generated binding changes mechanically reproducible from the linked
  LDK commit.
- [ ] Use concise Markdown headings, bullets, code fences, and user-facing
  terminology consistent with Alby Hub's repository documentation.
- [ ] Prepare the Hub PR as a draft until its dependency commits are public and
  all checks pass.
- [ ] Include root cause, design constraints, security model, migration effect,
  exact changed-file scope, test evidence, screenshots, and dependency PR links.
- [ ] Read the remote PR patches and head SHAs back before calling them ready.
- [ ] Do not merge; upstream maintainers own review and merge decisions.

## 5. Build the FrontierCrown distribution

### Product identity and installation safety

- [ ] Use an independent product name, icon, bundle identifier, and update
  channel so the package cannot silently replace or be mistaken for the
  official Alby Hub application.
- [ ] State prominently that this is an experimental, independently maintained
  fork built from Alby Hub and licensed under Apache-2.0.
- [ ] Preserve the upstream license and attribution, and mark modified files as
  required by the license.
- [ ] Use a separate data directory by default. Treat import or migration from
  an existing Alby Hub wallet as an explicit, backed-up owner action.
- [ ] Do not point the fork at Alby's official updater.
- [ ] Decide whether the first release supports only macOS universal or a wider
  tested platform matrix; never list an untested platform.

### Release artifacts

- [ ] Tag one immutable release candidate on the owner fork.
- [ ] Build from a clean checkout using pinned toolchain and dependency
  versions.
- [ ] Sign and notarize macOS packages with the distributor's own Developer ID;
  do not use or imply Alby's signing identity.
- [ ] Produce the application archive, source archive, patch series, release
  notes, `BUILDING.md`, `SECURITY.md`, license, SBOM, and provenance record.
- [ ] Produce a manifest containing every filename, byte size, SHA-256 digest,
  source commit, upstream base, dependency commits, and build-workflow identity.
- [ ] Sign the manifest with the owner's PGP key. Keep the private key outside
  the repository, CI logs, and chat.
- [ ] Publish the public key and full fingerprint independently from the
  downloadable files.
- [ ] Verify the signature and every artifact hash from a fresh download before
  marking the release available.
- [ ] Call the build reproducible only if independent clean builds produce the
  same artifact hashes; otherwise call it source-pinned and provenance-backed.

### Upgrade and rollback acceptance

- [ ] Back up the current wallet and record hashes before any migration test.
- [ ] Test a clean installation with no wallet data.
- [ ] Test migration using a copy of the current data, then verify unlock,
  database integrity, channels, balances, transaction history, NWC apps,
  permissions, backup state, embedded hashes, and UI behavior.
- [ ] Keep a verified rollback package and database backup outside the
  application directory.
- [ ] Do not run a real payment or rebalance as part of installation acceptance.
  Any later value-moving canary requires its own fresh action packet and owner
  approval.

## 6. Add FrontierCrown portal item 04

- [ ] Add a fourth portal card without modifying unrelated owner changes in the
  existing checkout.
- [ ] Label it as an experimental independent Alby Hub fork, not an official
  Alby release.
- [ ] Show the release version, supported platform, upstream Alby version, and
  release status.
- [ ] Provide separate actions for Download, Verify signature, View source, Read
  release notes, and Report a security issue.
- [ ] Show the PGP fingerprint and SHA-256 manifest without asking users to
  trust only the download origin.
- [ ] Require a clear backup and real-funds warning before download.
- [ ] Verify keyboard, touch, mobile, and desktop behavior and confirm all links
  resolve to the immutable release.
- [ ] Deploy only after a separate review of the exact portal diff and target
  release URLs.
- [ ] Read the live portal and download back from the public origin before
  declaring them available.

## 7. Public release communication

Publication begins only after the package, source, signatures, portal, and PRs
are publicly readable and the security gate is green.

### Nostr

- [ ] Prepare a source-backed release announcement that states what was built,
  why exact-channel enforcement matters, what was tested, and what remains
  experimental.
- [ ] Include the source, release, verification, portal, and upstream PR links.
- [ ] Exclude private node identifiers, balances, wallet history, and operational
  routing strategy.
- [ ] Show the exact event content, kind, tags, and timestamp policy as
  `PREPARED - NOT SIGNED - NOT PUBLISHED`.
- [ ] Obtain separate owner approval for signing and publication through the
  owner-controlled Nostr signer.
- [ ] Read the event back from at least one target relay before calling it live.

### Stacker News

- [ ] Use the canonical public release article or Nostr article as the link.
- [ ] Check for an existing submission of the exact URL immediately before
  posting.
- [ ] Select one relevant territory and prepare a short factual description.
- [ ] Keep any crosspost-to-Nostr option disabled if the source is already on
  Nostr.
- [ ] Show the exact title, URL, territory, description, and current posting fee
  before requesting approval.
- [ ] Obtain separate approval for the exact fee and never retry an ambiguous
  payment.
- [ ] Verify the post is publicly readable, not pending, and paid exactly once.

## Completion evidence

The project is complete only when all of the following are true:

- the responsive UI and specific failure diagnostics are verified;
- security review has no unresolved release-blocking findings;
- the upstream PR chain is publicly readable and accurately scoped;
- the package is source-pinned, signed, independently downloadable, and
  verified from a fresh machine or profile;
- the portal item is live and points to the exact verified release;
- migration and rollback evidence is recorded without risking the live wallet;
- the owner has separately approved and published the exact Nostr and Stacker
  News items; and
- final URLs, hashes, signatures, source commits, PR numbers, and publication
  evidence are appended to the engineering log.


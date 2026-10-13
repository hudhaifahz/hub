# Exact-channel rebalance release plan

Status: owner-build acceptance is complete; isolated upstream preparation and
security verification are in progress; public release is blocked on the gates
recorded below

Execution-agent instructions: `doc/exact-channel-rebalance-execution-brief.md`.

This checklist does not authorize a payment, deployment, pull request, release,
signature, or public post. Each public or value-moving action requires a separate
owner review.

## Current public-release snapshot (2026-10-13)

| Surface               | Observed state                                                                                                                                                                        | Release meaning                                                                                                                     |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| Exact-channel source  | Public on the `hudhaifahz` Hub, LDK, and LDK Go forks                                                                                                                                 | Reviewable source exists, but no immutable release tag or upstream PR chain exists yet.                                             |
| Owner package         | Guarded owner-test ZIP exists locally for source commit `4a82cbf40cc9c2aa60255737e2a4b1f968001e9f`                                                                                    | It is ad-hoc signed, uses the official Alby bundle identity and data container, and must not be published.                          |
| Independent preflight | Unsigned universal bundle built from exact clean commit `c41a5e1773030da5fccc174386edc8730423b4a3`                                                                                    | Identity and embedded-code inspection pass; it was not installed or launched and is not a public release candidate.                |
| GitHub release        | The `hudhaifahz/hub` fork has zero releases                                                                                                                                           | There is no public binary, immutable release page, signed manifest, or public download URL.                                         |
| Frontier Crown portal | The live portal contains Portal 01 through Portal 03; isolated security branch `codex/rebalance-portal-security` stages the locked Portal 04 card with 0 High/Critical audit findings | No rebalance card or download is live. Production remains blocked by the final release, preview-review, and owner deployment gates. |
| Grynvault endpoint    | `rebalance.grynvault.com` does not resolve; an unadvertised Cloudflare provider hostname serves the reviewed fixture                                                                  | Temporary acceptance is complete, but canonical DNS and public launch are not.                                                      |
| macOS signing         | No Developer ID Application identity is available on the release Mac                                                                                                                  | Public macOS signing, notarization, stapling, and Gatekeeper acceptance cannot be completed yet.                                    |
| Manifest signing      | No local PGP secret key is available to the release workflow                                                                                                                          | A public key, fingerprint, owner-controlled signing ceremony, and detached signature are still required.                            |
| Upstream baseline     | `v1.24.1` remains the latest tagged Alby Hub release; upstream `master` has advanced beyond it                                                                                        | Rebase and compatibility checks must be repeated immediately before PR and release candidate creation.                              |

The local owner-test archive is evidence, not a downloadable product. The portal
must not link to it, and the endpoint must not advertise it as a release.

## Product split and non-custodial boundary

Ship two related but deliberately different products:

1. **Local exact-channel software.** The independently named desktop fork (and,
   if accepted, the upstream Alby/LDK changes) performs quote creation,
   first-hop binding, final-channel binding, invoice/payment handling, execution,
   and reconciliation inside the owner's Hub process.
2. **Hosted release and diagnostic service.** `rebalance.grynvault.com` publishes
   verified release metadata, compatibility information, documentation, and
   sanitized diagnostic explanations through HTTP and MCP. It never receives
   wallet credentials or executes a payment.

A hosted service cannot truthfully replace the local patch for stock Alby Hub.
The remote service does not control the user's LDK route builder or recipient
claim logic and therefore cannot atomically enforce both exact channels. Do not
offer an `execute_rebalance` tool, accept an NWC connection, or market the hosted
service as forced routing. If remote execution is ever explored, it requires a
separately audited local companion/upstream capability and a new threat model.

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
  `4a82cbf40cc9c2aa60255737e2a4b1f968001e9f`.
- Current Hub documentation head:
  `4e515474bf67c2f8cc5f8ada7d9bfbedcd313592` before this plan update.
- LDK source fork head:
  `05d86840f2184dbe846e260b71e7eeed2bea998a`.
- LDK Go binding fork head:
  `2d4b8cae5b4188451fbccdb347113838a9981dbf`.
- Isolated Hub PR-preparation base/head:
  `b7d0b0be829ba387d6c2f95c54a1449b71f6cd0e` / `1fb6c8cd`.
- Isolated LDK PR-preparation base/head:
  `74daaf98c7eeee7bca58c703c103d57affbd928a` / `2dc742a8`.
- Isolated LDK Go PR-preparation base/head:
  `5ba434093284d0dc8291e45d2a3d2c1c5a879698` / `a381cacb`.
- Isolated Frontier Crown Portal 04 base/head:
  `c6d2691dfa80d58391e5271cb583528f0be51a4d` /
  `873833e9364ce693e02e86f129125b712a84d579`.
- Isolated Frontier Crown security-remediation head:
  `7054c902890772f34e1d1a9533ec5463862d2e45` on
  `codex/rebalance-portal-security`.
- Independent Grynvault release-preparation head:
  `c41a5e1773030da5fccc174386edc8730423b4a3` on
  `codex/grynvault-rebalance-release`; the owner-fork remote readback matches and
  macOS GitHub Actions run `38483312899` passed.
- Reviewed universal macOS LDK library: `43,360,288` bytes, SHA-256
  `cf43f66815c44c0796402faafaaa3ed70201d036124438e63f541ec6a05208bd`.

Record a new immutable source commit for every release candidate. Never describe
the moving branch name as the release source.

## 1. Preserve and isolate the baseline

- [x] Record clean/dirty state, remotes, branch heads, upstream heads, and the
      installed application's embedded source hashes.
- [x] Preserve the current working implementation on the owner fork.
- [x] Create clean worktrees from the current upstream heads for upstream PR
      preparation; do not rewrite the live-tested branch.
- [x] Create a separate FrontierCrown release branch for packaging and portal
      work.
- [x] Keep the existing dirty FrontierCrown checkout intact and do not stage its
      unrelated changes.
- [x] Maintain a dated engineering log with source hashes, test results,
      security findings, release artifacts, and owner approvals.

## 2. Fix the dialog and failure evidence

### Responsive dialog

- [x] Constrain the dialog to the viewport width and height.
- [x] Allow vertical scrolling inside the dialog and prevent horizontal page
      overflow.
- [x] Add `min-width: 0` behavior to grid, form, definition-list, and footer
      children that contain long identifiers.
- [x] Wrap full peer keys, channel IDs, SCIDs, hashes, and confirmation text
      without hiding or truncating the exact evidence.
- [x] Collapse two-column quote rows to one column at narrow widths.
- [x] Stack or wrap footer actions so every action remains visible and tappable.
- [x] Keep the principal and fee inputs usable at 320, 375, 768, and 1280 CSS
      pixel widths.
- [x] Verify the empty, quoted, executing, succeeded, failed, and recovery states.

### Recipient rejection diagnostics

- [x] Replace the generic local claim failure with a non-secret reason code:
      payment state, circular metadata, exact amount, missing receiving-channel
      identity, or receiving-channel mismatch.
- [x] Preserve the expected and observed local receiving channel identifiers in
      owner-only terminal evidence when available.
- [x] Never persist or return a preimage, payment secret, invoice, or serialized
      route in owner-facing diagnostics.
- [x] Carry the reason through LDK source, generated Go bindings, Hub
      reconciliation, the authenticated owner API, and the dialog.
- [x] Explain local safety rejection separately from route-not-found, peer
      forwarding failure, timeout, and exhausted retry states.
- [x] Keep failure behavior fail-closed and keep automatic retries disabled.
- [x] Add unit and integration coverage for every diagnostic reason without
      weakening claim conditions.

### UI acceptance evidence

- [x] Run frontend formatting, linting, and TypeScript compilation.
- [x] Build both HTTP and Wails frontend variants.
- [x] Capture screenshots at the required viewport widths for pre-quote and
      quote-review states.
- [x] Confirm with 200% browser zoom. Keyboard-only navigation and long
      unbroken identifiers pass; 320 CSS-pixel testing is additional narrow-screen
      stress evidence, not a substitute for the explicit zoom check.
- [x] Record that every action and every exact identifier is visible without
      horizontal scrolling.

Privacy-safe usage instructions and screenshot provenance are recorded in
`doc/exact-channel-rebalance-usage.md` and
`doc/assets/exact-channel-rebalance/README.md`. Geometry checks passed at 320,
375, 768, and 1280 CSS pixels: the body, dialog, and scroll region had no
horizontal overflow, and all three actions remained inside the viewport. The
exact-confirmation action became keyboard-reachable only after the fixture text
matched; it was not invoked.

The explicit browser-zoom check used Brave at 200% on the real component's
synthetic quote and recovery fixtures. Retina `devicePixelRatio` increased from
2 to 4; the effective viewport was 756 CSS pixels wide, and document client and
scroll widths both remained 756 pixels. The 510-pixel-wide internal review
region scrolled from 0 through its 1,181-pixel maximum while the three footer
actions remained visible inside the viewport. Full quote, operation, payment,
peer, and channel identifiers used `break-all` wrapping. Keyboard-only focus
moved between the labelled confirmation field and the enabled Close control;
disabled payment controls were correctly skipped. No execution control was
invoked. The localhost origin was then restored to 100% zoom and the temporary
browser tab was closed.

Synthetic saved-operation fixtures also verified executing, succeeded, failed,
and exact-recovery states at 320 and 1280 CSS pixels. Every state kept the body,
dialog, and scroll region within bounds. Ordinary terminal states disabled both
payment actions; recovery exposed a separately labelled exact-retry action that
remained disabled until its exact confirmation was supplied.

## 3. Security review

### Funds and state safety

- [x] Re-review exact outgoing and incoming channel binding, route fingerprint
      binding, expiry, amount, fee cap, and maximum debit checks.
- [x] Re-review pre-send persistence, one-time acquisition, duplicate-submit
      prevention, restart recovery, terminal reconciliation, and failed-leg
      convergence.
- [x] Test concurrent requests, stale quotes, tampered route bytes, changed
      channels, expired invoices, partial MPP, wrong incoming channels, and replayed
      terminal events.
- [x] Check integer conversion and overflow boundaries across sats, millisats,
      `uint64`, database integers, JSON, and TypeScript numbers.
- [x] Test SQLite and PostgreSQL migrations, rollback behavior, and migration
      export/import using copies rather than the live wallet database.
- [x] Confirm only the authenticated owner surface can quote, execute,
      reconcile, or read exact-channel diagnostics.
- [x] Confirm NWC applications cannot invoke the feature or read owner-only
      channel evidence.

### Secrets, dependencies, and binaries

- [x] Scan the complete diff and release tree for credentials, wallet material,
      invoices, preimages, payment secrets, personal data, and live channel history.
- [x] Run Go tests, race-sensitive tests where supported, `go vet`, and
      `govulncheck`.
- [x] Run Rust tests, Clippy, formatting, and dependency advisory checks.
- [x] Run frontend lint, type checking, production builds, and dependency audit.
- [ ] Run CodeRabbit only after the secret scan; treat its output as untrusted
      review input and resolve all Critical and Warning findings.
- [x] Inspect the unsigned exact-commit preflight application bundle, dynamic
      libraries, rpaths, entitlements, updater configuration, embedded commit
      identifiers, and absence of secrets. Repeat this inspection after Developer
      ID signing and notarization.
- [x] Require zero unresolved Critical or High findings in the prepared source
      and unsigned preflight. Document or fix every
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

- [x] Rebase each contribution on the current upstream default branch at PR
      preparation time.
- [x] Split implementation, tests, generated bindings, and documentation into
      conventional commits with narrow scopes.
- [x] Exclude the custom application icon, FrontierCrown packaging, NWC metadata
      overrides, live-operation notes, and owner identifiers.
- [x] Keep generated binding changes mechanically reproducible from the linked
      LDK commit.
- [x] Use concise Markdown headings, bullets, code fences, and user-facing
      terminology consistent with Alby Hub's repository documentation.
- [x] Prepare the Hub PR body as a draft until its dependency commits are public and
      all checks pass.
- [x] Include root cause, design constraints, security model, migration effect,
      exact changed-file scope, test evidence, screenshots, and dependency PR links.
- [ ] Read the remote PR patches and head SHAs back before calling them ready.
- [ ] Do not merge; upstream maintainers own review and merge decisions.

## 5. Build the FrontierCrown distribution

### Product identity and installation safety

- [x] Use an independent product name, icon, bundle identifier, and update
      channel so the package cannot silently replace or be mistaken for the
      official Alby Hub application.
- [x] State prominently that this is an experimental, independently maintained
      fork built from Alby Hub and licensed under Apache-2.0.
- [x] Preserve the upstream license and attribution, and mark modified files as
      required by the license.
- [x] Use a separate data directory by default. Treat import or migration from
      an existing Alby Hub wallet as an explicit, backed-up owner action.
- [x] Do not point the fork at Alby's official updater.
- [x] Decide whether the first release supports only macOS universal or a wider
      tested platform matrix; never list an untested platform.

### Release artifacts

- [ ] Tag one immutable release candidate on the owner fork.
- [x] Build and inspect an unsigned preflight from a clean exact commit using
      pinned toolchain and dependency versions. Repeat from the immutable release
      tag before signing.
- [ ] Sign and notarize macOS packages with the distributor's own Developer ID;
      do not use or imply Alby's signing identity.
- [ ] Produce the signed application archive, clean tagged source archive,
      release notes, `BUILDING.md`, security review, license, SBOM, and provenance
      record. Do not publish patch-series attachments because they can reproduce
      credential-shaped values removed from upstream source.
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

- [x] Add a fourth portal card without modifying unrelated owner changes in the
      existing checkout.
- [x] Label it as an experimental independent Alby Hub fork, not an official
      Alby release.
- [ ] Show the release version, supported platform, upstream Alby version, and
      release status.
- [x] Provide separate actions for Download, Verify signature, View source, Read
      release notes, and Report a security issue.
- [ ] Show the PGP fingerprint and SHA-256 manifest without asking users to
      trust only the download origin.
- [x] Require a clear backup and real-funds warning before download.
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

## 8. Deploy `rebalance.grynvault.com` as a non-custodial companion

### Version-one scope

The first hosted service is intentionally read-only and stateless with respect
to wallets. Its public HTTP surface is:

- `GET /` — human-readable product boundary, installation guide, risk notice,
  source links, security contact, and release status;
- `GET /healthz` — process health only, with no wallet or release secrets;
- `GET /v1/releases/latest` — immutable artifact URLs, byte sizes, SHA-256
  hashes, source/upstream/dependency commits, supported platforms, signing
  fingerprint, and release status;
- `GET /v1/compatibility/{hubVersion}` — tested/untested/unsupported status and
  the exact evidence record; and
- `POST /v1/diagnostics/explain` — explanation of a closed, documented reason
  code without uploading invoices, routes, keys, balances, or wallet logs.

Expose the same bounded functions at `/mcp` using the current stable MCP
Streamable HTTP specification:

- `get_latest_release`;
- `check_compatibility`;
- `verify_release_manifest`;
- `explain_rebalance_failure`; and
- `get_rebalance_prerequisites`.

MCP tool descriptions must explicitly say that calls are informational and
cannot quote, sign, pay, route, or rebalance. Results must identify observed,
inferred, and unsupported states instead of inventing wallet state.

### Prohibited inputs and behavior

- Never request or accept seeds, passwords, recovery material, macaroons, NWC
  connection strings, API tokens, private keys, preimages, payment secrets,
  invoices, serialized routes, or complete wallet databases/logs.
- Never proxy Hub owner APIs or expose a user's Hub to the public Internet.
- Never retain full pubkeys, channel IDs, balances, or failure payloads by
  default. If a future opt-in support flow needs them, design separate consent,
  redaction, retention, deletion, and access controls first.
- Never let release metadata point to a mutable branch, `latest` object, or
  unsigned artifact. Each item must resolve to an immutable release asset and
  be covered by the signed manifest.
- Never execute a Lightning payment, rebalance, swap, channel open/close, or
  signer action. Those remain local and require action-time owner approval.

### Hosting and DNS

`grynvault.com` is currently served from the separate `Grynvault/btc_jv`
Netlify project and its authoritative DNS is in Route 53. Keep the root site
untouched. Deploy the companion as an isolated stateless Cloudflare Worker. Use
an unadvertised provider hostname for private acceptance, then create only the
exact Route 53 delegation or record required for
`rebalance.grynvault.com` after a separate owner approval.

This supersedes the initial Railway default. Current Cloudflare Agents guidance
uses SDK v2 `createMcpHandler` for new stateless MCP servers; the companion
needs neither a database nor Durable Objects, and the Worker boundary provides
native request limits, rate-limit bindings, preview versions, and a small
deployment surface. This is an implementation choice, not permission to change
production DNS.

Do not configure a Cloudflare Worker custom domain unless the zone or delegated
subdomain is first placed under Cloudflare DNS; the current AWS nameservers do
not satisfy that prerequisite. A temporary provider hostname may be used for
private acceptance, but all public manifests and portal links must use the
canonical Grynvault hostname after TLS is valid.

### Hosted-service security gate

- [x] Put the service in a dedicated repository or isolated workspace with no
      wallet source tree, release signing secrets, or production database access.
- [x] Pin the MCP specification version and publish supported protocol versions.
- [x] Validate method, content type, origin, request size, schema, reason-code
      enum, and all release identifiers; reject unknown fields where practical.
- [x] Add per-IP and per-tool rate limits, timeouts, structured secret-redacting
      logs, generic errors, and request IDs that contain no submitted data.
- [x] Set HSTS, CSP, `X-Content-Type-Options`, restrictive CORS, and a minimal
      permissions policy; do not permit browser credentials on public endpoints.
- [x] Generate an SBOM, run dependency and secret scans, review the deploy diff,
      and require zero unresolved Critical or High findings.
- [x] Test malformed JSON, oversized payloads, header spoofing, method confusion,
      SSRF/path traversal attempts, cache poisoning, rate-limit bypass, and MCP tool
      name/argument confusion.
- [x] Keep `/healthz` shallow; separately verify every public endpoint, TLS,
      security headers, and response schema from outside the deployment platform.
- [x] Publish a security contact, privacy statement, uptime disclaimer, and a
      clear statement that no wallet credentials should ever be submitted.

Implementation evidence on 2026-10-13:

- isolated repository: `/Users/kode/Development/grynvault-rebalance`;
- private owner repository: `hudhaifahz/grynvault-rebalance`, with remote
  `main` read back at `71d3c4dbfd159efc45b4761f1cf36ca4bc75f99c`;
- clean companion GitHub Actions run `38482199397` passed the source checks,
  dependency audit, SBOM generation, and tracked-source secret scan;
- exact packages: `agents@0.28.1`, MCP server/client `2.3.1`,
  `@modelcontextprotocol/sdk@1.32.1`, `wrangler@4.150.0`, and
  `@cloudflare/vitest-plugin@1.4.1`;
- HTTP and MCP schemas accept no wallet connection, credentials, invoices,
  routes, channel identifiers, balances, or logs, and release status remains
  truthfully `unpublished` with no artifact URL;
- strict formatting, typed lint, generated bindings, three TypeScript
  compilations, 13 Worker-runtime boundary tests, and the Wrangler dry-run
  bundle pass;
- a clean install/check under Node `22.18.0`, dependency audit, CycloneDX SBOM,
  and Gitleaks `8.29.1` scan pass with no vulnerability or secret finding;
- `docs/SECURITY-REVIEW.md` records the trust boundary, tests, and residuals;
  `docs/PROVIDER-ACCEPTANCE.md` records the external provider readback.
- active temporary provider version
  `2867ad4b-9d5d-4136-add7-4bb4bbf3b04e` serves companion source commit
  `72febac0ce14123a12079b9cd3903602b5214a5a` on
  `grynvault-rebalance-companion.hudhaifahz.workers.dev`; external TLS, HTTP
  schemas, security headers, hostile-request handling, actual rate limiting,
  and modern/legacy MCP negotiation passed, and the release endpoint reads
  back exact reviewed Hub commit
  `1fb6c8cd0fcb2fbee48c888c6274e2659d74c169`;
- the first acceptance version revealed that Cloudflare invocation logs include
  ordinary request-envelope metadata. No wallet data was sent. Commit
  `0c44a0e` disabled persisted invocation logs and enabled query-string
  redaction before the accepted version was deployed. Real-time operator
  tailing remains an explicit metadata-bearing diagnostic action.

## 9. Execution sequence and approval gates

### Wave A — close owner-build acceptance

- [x] Unlock the currently installed diagnostic build through the owner UI.
- [x] Read back chain sync, all channel identities/statuses, balances, canonical
      database path, embedded library hash, migration, and responsive dialog states.
- [x] Perform no payment or rebalance during this acceptance wave.

### Wave B — prepare clean upstream contributions

- [x] Create isolated clean worktrees for the three PRs in Section 4.
- [x] Rebase against the then-current upstream commits, rerun the complete test
      and audit matrix, and prepare draft PR bodies and screenshots.
- [ ] Stop for owner review before opening any public PR or security disclosure.

### Wave C — create the independently identified release candidate

- [x] Prepare and verify the product name, bundle identifier, icon, data
      directory, updater policy, universal-macOS platform matrix, and
      migration/rollback UX. Public release approval remains separate.
- [ ] Make a clean release tag and build all artifacts from that tag.
- [ ] Obtain an owner-controlled Developer ID signing path and PGP signing key;
      keep their private material outside source, logs, and chat.
- [ ] Sign, notarize, staple, hash, manifest, PGP-sign, and verify from a fresh
      download and a clean macOS profile before calling it a release candidate.

### Wave D — build and privately accept the companion

- [x] Implement only the HTTP/MCP functions in Section 8 with fixture data.
- [x] Deploy to a temporary provider hostname, run the hosted-service security
      gate, and verify no wallet-execution path or credential field exists.
- [ ] Add the exact Route 53 record only after private acceptance, then verify
      canonical TLS and all endpoints externally.

### Wave E — stage the portal and release pages

- [x] Stage Portal 04 on isolated branch `codex/rebalance-portal-04` at
      `873833e9364ce693e02e86f129125b712a84d579`; keep every release action locked
      unless status, immutable source/download/manifest/notes/docs URLs, and the
      full signing fingerprint are present.
- [x] Remediate the inherited Frontier Crown dependency findings on isolated
      branch `codex/rebalance-portal-security` at
      `7054c902890772f34e1d1a9533ec5463862d2e45`; full and production audits now
      contain 0 Critical and 0 High findings, complete tests/builds pass, and
      the remaining Moderate/Low development-tool findings are documented.
- [ ] Store binaries on the immutable GitHub release, not in the portal Git
      repository or Netlify bundle.
- [ ] Add Portal 04 in an isolated branch and point it only to the verified
      GitHub assets, signed manifest, source tag, release notes, security contact,
      and `rebalance.grynvault.com` documentation.
- [ ] Review the exact portal diff and a deploy preview. Stop for owner approval
      before merging or deploying the public card.

### Wave F — public launch

- [ ] Publish the immutable release and read every artifact back from its public
      URL; then activate Portal 04 and read the live portal back.
- [ ] Open the upstream PR chain only after its separate owner review.
- [ ] Prepare exact Nostr and Stacker News payloads with final URLs and claims.
      Signing a Nostr event and paying a Stacker News posting fee each require their
      own final owner approval and independent readback.

## 10. Small-business experiment

Treat this as validation-stage infrastructure, not yet a routing-liquidity
business. The first customer is a self-hosting Alby/LDK operator who already has
one channel with enough outbound liquidity and another with enough inbound
liquidity, and wants auditable control over which two channels a circular
rebalance uses. Exclude custodial-wallet users and operators seeking guaranteed
routes or guaranteed profit.

Start with one open-source offer: verified builds, compatibility records,
diagnostic help, and migration/rollback documentation. Test willingness to pay
with a bounded support membership or paid installation/recovery session after
the free release is trustworthy. Do not charge per rebalance or operate customer
wallets in version one. Evidence of repeated support demand can justify a later
managed monitoring product; evidence of demand for inbound liquidity belongs in
a separate LSPS2/LSP product with its own capital, legal, security, and unit-
economics review.

Track weekly: verified downloads, successful clean installs, compatible upgrade
tests, support requests, diagnostic resolutions, upstream review progress, paid
support conversions, support time per user, and security incidents. Stop or
change the offer if operators will not install a third-party fork, support cost
exceeds revenue, or upstream adoption makes the fork unnecessary.

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

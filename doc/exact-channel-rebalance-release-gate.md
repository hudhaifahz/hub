# Exact-channel rebalance release gate

Status: **PUBLIC RC.2 PRERELEASE — PGP VERIFIED — APPLE UNSIGNED/UNNOTARIZED — CLEAN INSTALL AND MIGRATION PENDING**

This packet records the current decision point after source, package, CI,
immutable-tag, PGP, public-release, canonical-endpoint, and portal verification.
Those completed actions had separate owner approval. It does not authorize a
new installation, migration, upstream pull request, public post, or Lightning
value movement.

## Current reviewed source candidate

- repository: `hudhaifahz/hub`
- branch: `codex/grynvault-rebalance-release`
- exact source commit:
  `29a34b97fdd890f1895ca9827ba0e59f8457f3ee`
- immutable annotated tag: `grynvault-rebalance-v0.1.0-rc.2`
- tag object: `c4270fbaa40a183947562c0460b6e3c466f9a8eb`
- peeled tag commit: `29a34b97fdd890f1895ca9827ba0e59f8457f3ee`
- exact successful GitHub Actions run: `38521293804`
- supported package target: universal macOS 15.4 or newer
- distribution class: explicitly `UNSIGNED-EXPERIMENTAL`
- Apple Developer ID signature: absent by owner decision
- Apple notarization/stapling: absent by owner decision
- release URL:
  `https://github.com/hudhaifahz/hub/releases/tag/grynvault-rebalance-v0.1.0-rc.2`
- PGP primary fingerprint:
  `D131A3D0EAB0CF98A8413853CD9FDBF9D075C126`
- reviewed unsigned LDK SHA-256:
  `cf43f66815c44c0796402faafaaa3ed70201d036124438e63f541ec6a05208bd`

Historical immutable annotated tag `grynvault-rebalance-v0.1.0-rc.1` has tag
object `88cd63322c9f842a271571d621bba51393c9e709` and still peels to
`8a5478f379788e0c7978a4b8dfcfd2afebf29426`. It must never be moved. The
owner separately approved rc.2 at the newer package-path commit. Both tags are
immutable and must never be moved.

## Evidence closed at the current commit

- The owner-fork branch and remote ref both resolve to exact commit
  `29a34b97fdd890f1895ca9827ba0e59f8457f3ee`.
- GitHub Actions run `38521293804` passed checkout, pinned-tool installation,
  frontend and workflow checks, Go and race tests, vet, current vulnerability scanning,
  tracked-tree secret scanning, and release-script checks.
- Local source-boundary checks, release-script regressions, actionlint 1.7.7,
  focused Go tests, Wails binding-only compilation, `go vet`, Gitleaks 8.30.1,
  and govulncheck 1.9.0 passed. Govulncheck reports zero reachable
  vulnerabilities; the two previously documented non-reachable dependency
  advisories remain visible.
- A clean universal application built with Node 22, Yarn 1.22.22, Wails 2.16.0,
  and Go 1.27.2. Both executable slices and `LSMinimumSystemVersion` declare
  macOS 15.4, matching the inherited Bark native-object floor.
- Raw inspection passed product/bundle identity, independent data identity,
  updater/source boundary, both architectures, exact runtime path, embedded
  exact source/tag, and the pinned unsigned LDK hash.
- The packager applied only ad-hoc integrity signatures, verified required
  entitlements and absence of any certificate authority or Team Identifier,
  created an unmistakably named `UNSIGNED-EXPERIMENTAL` DMG, embedded
  `README-UNSIGNED.txt`, verified the disk image and mounted app, and recorded
  Gatekeeper rejection with `source=no usable signature` and exit status 3.
- A clean source archive and CycloneDX 1.6 SBOM were created from rc.2. The SBOM
  contains 1,100 components: 283 Go, 815 npm, and two normalized lockfile
  components; it contains no local checkout path.
- The local-only test tag `grynvault-unsigned-local-29a34b97` was deleted after
  verification.
- On 2026-10-14, the owner approved immutable annotated tag
  `grynvault-rebalance-v0.1.0-rc.2` at the exact reviewed commit. Git and the
  GitHub Git-object API independently read back tag object
  `c4270fbaa40a183947562c0460b6e3c466f9a8eb`, peeled commit
  `29a34b97fdd890f1895ca9827ba0e59f8457f3ee`, and unsigned-tag verification
  reason `unsigned`, as expected.
- Frontier Crown Portal 04 commit
  `82aff42b327b2609bb4b5783aba79b6ea8bbcba9` adds explicit unsigned,
  unnotarized, Gatekeeper, macOS 15.4+, and PGP-manifest disclosure plus
  desktop/mobile screenshots. Documentation head
  `7645db5a813a0c28926b4e780078aa76adbe5421` records successful full workspace
  checks, refreshed audits, secret scans, responsive browser readback, and exact
  remote SHA. That historical staging evidence is preserved below; the final
  portal publication is recorded in the current public acceptance section.
- Companion branch `codex/unsigned-release-metadata` was pushed and read back at
  `24fc2475c18b5e69559bfd2c757154e30f2afd3c`. It prepares the same explicit
  signing/platform/source metadata; formatting, lint, generated types, all
  TypeScript checks, 13 Worker-runtime tests, a dry-run bundle, zero-vulnerability
  dependency audit, and secret scan passed. It was not deployed, so the accepted
  provider version remained unchanged at that historical checkpoint.

## Current public acceptance (2026-10-14)

- GitHub prerelease rc.2 is public at the exact tag above. A fresh independent
  download produced the same 79,082,856-byte DMG and SHA-256
  `b329d7bec494cfff6b6c4f2865f2aa13e08a11e96726696b869cd0309c6dfe1a`.
- An isolated GnuPG home imported the published 721-byte public key and read
  exact primary fingerprint `D131A3D0EAB0CF98A8413853CD9FDBF9D075C126`
  plus encryption subkey `1F4E227EAED1809E009067B1E8251E01DC70A7A9`.
  The 265-byte detached signature verified as good over the public manifest.
- Public artifact SHA-256 values independently read back as: manifest
  `86fc987effeae42d7a24ae6f7e3a454a71af39d9c08fdc50ca6ae2299e2121c3`,
  signature `1b6516610b9b24c36dd43aa571560b0153ea4692cee51a5a5c7270153ce89475`,
  public key `5f70e12180ccf1120b27e15eae214cb27ca3e82edd8afab4dd40e61755ab8714`,
  source archive `71da411e271d033da2a5bcd1be43eccb8e9d3595df3c917b4a346c97d23b002d`,
  SBOM `e771bb19bc842a94781a2334cedc070699de0a8400f7e04cc83dab089fe9550c`,
  preparation index `250a5d23c9fb9dd37c24f086c375ea4e689e6dcb4958ba67a6f89620c134450a`,
  and Gatekeeper record `be4d0f340f2d09aa58fcbe2f8ec2492e10e6b6c43a783406a5cb81c58545b4b0`.
- The downloaded DMG was mounted read-only and never launched. Static
  inspection confirmed product `Grynvault Rebalance`, bundle
  `com.grynvault.rebalance`, macOS 15.4+, `x86_64` and `arm64` application and
  LDK slices, exact rc.2/source metadata, the sole reviewed runtime path,
  ad-hoc signature, no Team Identifier, successful deep integrity verification,
  and expected Gatekeeper rejection. It detached cleanly.
- `https://rebalance.grynvault.com` is live over canonical TLS with current rc.2
  release/compatibility metadata and exactly five read-only MCP tools. It
  rejected a fabricated `execute_rebalance` method, a hostile Origin, and
  unapproved diagnostic fields. It has no wallet, NWC, invoice, route, payment,
  or execution binding.
- Frontier Crown PR `#6` was locally reviewed with CodeRabbit CLI against exact
  base `2aa0c7e0fc1d0b4972225027e1118c1f6af28a51`. Two valid minor findings were
  fixed and regression tested; the final pass reported one minor suggestion
  that was rejected because explicit empty configuration is intentionally a
  fail-closed release disable and the test exercises every empty field. The CLI
  also reported that this repository is not connected to a CodeRabbit
  organization visible to the signed-in account; this is not a GitHub-App or
  independent-audit claim.
- PR `Grynvault/frontierCrown#6` was squash-merged with exact-head protection at
  `18b1e6efdc381c3f0ea516150e865c42c2a8eb30`. Netlify production deploy
  `6ad0542ffa83daccd5046625` is ready from that exact commit, scanned 4,309
  files with no secret matches, and deployed no functions or edge functions.
  Live `/portal` browser readback shows the full warnings, rc.2, macOS floor,
  upstream version, fingerprint, and exact download/verification/source/docs
  links.
- The branded public rc.2 package has not been clean-installed or accepted
  against a copied wallet. The running `/Applications/Alby Hub.app` remains the
  earlier owner diagnostic build. No current evidence supports a claim of
  independent audit, Apple trust, public-package migration continuity, or a
  production hosted executor.

## Final rc.2 local artifact evidence

These artifacts were rebuilt from the immutable rc.2 tag with Node 22.23.3,
Yarn 1.22.22, Wails 2.16.0, and Syft 1.41.1. They are now published as the
explicitly unsigned experimental prerelease and freshly verified as above.
The DMG and its mounted application passed bundle, architecture, minimum-macOS,
embedded-source, updater-boundary, entitlement, ad-hoc-signature, and disk-image
verification. Gatekeeper rejected the DMG with `source=no usable signature` and
exit status 3, which is the required result for this unsigned path.

- unsigned experimental DMG: 79,082,856 bytes,
  `b329d7bec494cfff6b6c4f2865f2aa13e08a11e96726696b869cd0309c6dfe1a`
- DMG checksum file: 160 bytes,
  `ab65e3fbd5f9719efb2494aca4c6c9d460871129a9a83bbc802f88ed7293c359`
- Gatekeeper evidence: 221 bytes,
  `be4d0f340f2d09aa58fcbe2f8ec2492e10e6b6c43a783406a5cb81c58545b4b0`
- source archive: 2,495,913 bytes,
  `71da411e271d033da2a5bcd1be43eccb8e9d3595df3c917b4a346c97d23b002d`
- CycloneDX 1.6 SBOM: 2,225,015 bytes,
  `e771bb19bc842a94781a2334cedc070699de0a8400f7e04cc83dab089fe9550c`
- unsigned preparation index: 1,072 bytes,
  `250a5d23c9fb9dd37c24f086c375ea4e689e6dcb4958ba67a6f89620c134450a`

The unsigned preparation index records `UNSIGNED_PREPARATION_ONLY` and absent
Apple signing/notarization. It remains build evidence; authenticity is supplied
separately by the PGP-signed final manifest.

## Historical non-release local artifact evidence

These hashes prove the local test pipeline only. They are not final public
release hashes and must not be placed on the portal as downloads:

- unsigned experimental DMG: 79,082,617 bytes,
  `5ecfb2e237a8579f9ebcff1f15efb43a0e077b0b461fbf4e68823b5a6bd7d1bf`
- source archive: 2,496,076 bytes,
  `7ede0c39b7555ccfb749f1916effde26dff0d18c67a9f87dfd24c02b1ecca2d2`
- CycloneDX SBOM: 2,047,656 bytes,
  `fecace7fb1cb85e72348cf789a77329ab8373287f67d3ec40c7000e53ee9ea35`

## Review status

Three earlier complete CodeRabbit CLI reviews against base
`1fb6c8cd0fcb2fbee48c888c6274e2659d74c169` produced two findings, then one,
then zero findings at historical rc.1 commit
`8a5478f379788e0c7978a4b8dfcfd2afebf29426`. Every finding was reproduced,
fixed, regression tested, committed, and checked remotely.

The current CodeRabbit CLI is authenticated and reviewed the final Portal 04 PR
diff as recorded above. It also reports that `Grynvault/frontierCrown` is not
connected to a CodeRabbit organization accessible to that account. Therefore
the evidence is a local CLI review, not a GitHub App review and not an
independent security audit.

## Owner decisions and remaining gates

The owner explicitly declined Apple Developer ID signing, notarization, and
stapling. Those are omitted rather than pending. The optional private workflow
mode remains isolated but is not part of the chosen release path. Documentation,
filenames, UI, and portal text must never imply Apple publisher identity or
Gatekeeper acceptance.

PGP signing is complete with the owner-controlled key and full 40-hex
fingerprint above. Private PGP material and the passphrase remain outside
source, CI logs, and chat.

## Separate approval sequence

1. **Complete:** verify and push the unsigned-package source commit and pass its
   exact GitHub CI run.
2. **Complete:** the owner approved immutable annotated tag
   `grynvault-rebalance-v0.1.0-rc.2` at exact commit
   `29a34b97fdd890f1895ca9827ba0e59f8457f3ee`; remote Git and GitHub object
   readback match. Never move rc.1 or rc.2.
3. **Complete:** rebuild and verify the full unsigned artifact set from rc.2;
   final local artifact hashes are recorded above.
4. **Complete:** owner selected the public key/fingerprint and performed
   detached signing through the owner-controlled signer.
5. **Complete:** verify the PGP signature, every artifact hash, explicit
   unsigned warnings, and Gatekeeper rejection from a fresh download.
6. **Open owner installation gate:** test clean installation and migration only
   with owner-approved copies and a
   verified rollback; do not move sats.
7. **Complete:** publish the GitHub prerelease, canonical read-only endpoint,
   and Portal 04 with independent public readback.
8. Stop separately before each live pilot, upstream PR, Nostr signing, and
   Stacker News submission/payment.

No action later in this sequence inherits approval from an earlier step.

# Exact-channel rebalance release gate

Status: **UNSIGNED PACKAGE PATH VERIFIED — NEW TAG AND PGP SIGNING PENDING — UNPUBLISHED**

This packet records the exact decision point after non-paying source, package,
and CI verification. It does not authorize a new immutable tag, PGP signature,
installation, migration, GitHub release, deployment, pull request, DNS change,
public post, or Lightning value movement.

## Current reviewed source candidate

- repository: `hudhaifahz/hub`
- branch: `codex/grynvault-rebalance-release`
- exact source commit:
  `29a34b97fdd890f1895ca9827ba0e59f8457f3ee`
- exact successful GitHub Actions run: `38521293804`
- supported package target: universal macOS 15.4 or newer
- distribution class: explicitly `UNSIGNED-EXPERIMENTAL`
- Apple Developer ID signature: absent by owner decision
- Apple notarization/stapling: absent by owner decision
- reviewed unsigned LDK SHA-256:
  `cf43f66815c44c0796402faafaaa3ed70201d036124438e63f541ec6a05208bd`

Historical immutable annotated tag `grynvault-rebalance-v0.1.0-rc.1` has tag
object `88cd63322c9f842a271571d621bba51393c9e709` and still peels to
`8a5478f379788e0c7978a4b8dfcfd2afebf29426`. It must never be moved. The
current package-path commit is newer, so a separately approved new immutable
tag is required before final artifact construction.

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
- A clean source archive and CycloneDX 1.6 SBOM were created. The SBOM contains
  1,073 components: 256 Go, 815 npm, and two normalized lockfile components;
  it contains no local checkout path.
- The local-only test tag `grynvault-unsigned-local-29a34b97` was deleted after
  verification. No new release tag or GitHub release was created.
- Frontier Crown Portal 04 commit
  `82aff42b327b2609bb4b5783aba79b6ea8bbcba9` adds explicit unsigned,
  unnotarized, Gatekeeper, macOS 15.4+, and PGP-manifest disclosure plus
  desktop/mobile screenshots. Documentation head
  `7645db5a813a0c28926b4e780078aa76adbe5421` records successful full workspace
  checks, refreshed audits, secret scans, responsive browser readback, and exact
  remote SHA. No preview or production deployment occurred.
- Companion branch `codex/unsigned-release-metadata` was pushed and read back at
  `24fc2475c18b5e69559bfd2c757154e30f2afd3c`. It prepares the same explicit
  signing/platform/source metadata; formatting, lint, generated types, all
  TypeScript checks, 13 Worker-runtime tests, a dry-run bundle, zero-vulnerability
  dependency audit, and secret scan passed. It was not deployed, so the accepted
  provider version remains unchanged.

## Non-release local artifact evidence

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

The local CodeRabbit CLI is currently signed out. The CodeRabbit GitHub app was
added to the owner's account, but no public pull request was opened merely to
trigger it. CodeRabbit remains review input and is not an independent security
audit. A later owner-approved PR must receive and resolve its current review.

## Owner decisions and remaining gates

The owner explicitly declined Apple Developer ID signing, notarization, and
stapling. Those are omitted rather than pending. The optional private workflow
mode remains isolated but is not part of the chosen release path. Documentation,
filenames, UI, and portal text must never imply Apple publisher identity or
Gatekeeper acceptance.

PGP manifest signing still requires an owner-approved public key and full
40-hex fingerprint. Private PGP material must remain outside source, CI logs,
and chat. This Mac previously reported zero GPG secret keys; that fact must be
rechecked only when the owner chooses a signing path.

## Separate approval sequence

1. **Complete:** verify and push the unsigned-package source commit and pass its
   exact GitHub CI run.
2. **Owner immutable-tag gate:** approve a new tag, expected to be
   `grynvault-rebalance-v0.1.0-rc.2`, at exact commit
   `29a34b97fdd890f1895ca9827ba0e59f8457f3ee` or at a later exact reviewed
   commit if CodeRabbit or CI causes another fix. Never move rc.1.
3. Rebuild the full unsigned artifact set from that exact immutable tag and
   replace all local-test hashes with final hashes.
4. **Owner PGP gate:** select the public key/fingerprint and authorize detached
   signing of the final manifest through an owner-controlled signer.
5. Verify the PGP signature, every artifact hash, explicit unsigned warnings,
   and Gatekeeper rejection from a fresh download.
6. Test clean installation and migration only with owner-approved copies and a
   verified rollback; do not move sats.
7. Stop again before publishing a GitHub release.
8. Stop separately before canonical DNS, Portal 04 deployment, upstream PRs,
   Nostr signing, and Stacker News submission/payment.

No action later in this sequence inherits approval from an earlier step.

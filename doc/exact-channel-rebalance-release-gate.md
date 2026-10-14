# Exact-channel rebalance release gate

Status: **TAGGED — UNSIGNED — UNPUBLISHED**

This packet records the exact decision point after non-paying source and bundle
verification and the separately approved immutable tag. It does not authorize
a signing workflow, release, install, deployment, pull request, DNS change, or
public post.

## Reviewed source candidate

- repository: `hudhaifahz/hub`
- branch: `codex/grynvault-rebalance-release`
- exact source commit:
  `8a5478f379788e0c7978a4b8dfcfd2afebf29426`
- approved immutable tag: `grynvault-rebalance-v0.1.0-rc.1`
- remote annotated-tag object:
  `88cd63322c9f842a271571d621bba51393c9e709`
- supported release target: universal macOS (`x86_64 arm64`) only
- exact successful CI run: `38494419778`
- reviewed embedded LDK SHA-256:
  `cf43f66815c44c0796402faafaaa3ed70201d036124438e63f541ec6a05208bd`

The tag uses a Grynvault-specific namespace so it cannot be confused with
upstream Alby Hub tags. Git and GitHub API readback both show that the remote
annotated tag peels to exact commit
`8a5478f379788e0c7978a4b8dfcfd2afebf29426`. The tag and source commit must not
be moved.

## Evidence already closed

- The owner-fork branch and remote ref both resolve to the exact source commit.
- GitHub Actions passed the complete supported-macOS source, frontend, Go,
  race, vet, vulnerability, and tracked-tree secret matrix at that commit.
- A local unsigned preflight built from the exact commit and passed product
  identity, bundle identity, architecture, runtime-path, embedded-source,
  embedded-library, and Info.plist inspection.
- A second unsigned preflight built directly from immutable tag
  `grynvault-rebalance-v0.1.0-rc.1` passed the same inspection. Its embedded tag
  is the immutable tag, its embedded source commit is the exact reviewed
  commit, and its universal LDK library retains the pinned unsigned SHA-256.
- The preflight was not installed, launched, signed, notarized, stapled, or
  published. The earlier temporary local preflight tag was deleted.
- Fresh remote readback found no upstream-base drift for Hub, LDK Node, or LDK
  Node Go.
- GitHub still reports zero releases, and the tag push triggered no signing or
  publication workflow.

## Review gate closed

CodeRabbit CLI `0.7.1` authenticated as GitHub account `hudhaifahz`, and its
health check passed with nine checks and no warnings. Three complete committed
diff reviews used base commit
`1fb6c8cd0fcb2fbee48c888c6274e2659d74c169`:

1. The first pass reported two valid findings: Bash 4-only uppercase expansion
   in manifest generation and an invalid post-signing comparison of a modified
   signed library against its pinned unsigned whole-file digest.
2. The second pass reported one valid finding: release workflow checkout and
   validation accepted an unqualified tag name that could collide with a
   same-named branch.
3. The final pass on exact commit
   `8a5478f379788e0c7978a4b8dfcfd2afebf29426` reported zero findings.

The first two findings were fixed in commits
`7a321ae8bbcc1ea733d58e3cc569b112dd3088d0` and
`8a5478f379788e0c7978a4b8dfcfd2afebf29426`. The fixes add Bash 3.2 regression
coverage, separate unsigned-byte inspection from signed-identity inspection,
verify the signed app and embedded library share a non-Alby Developer ID Team
Identifier, parse entitlements safely, and require fully qualified
`refs/tags/<tag>` references throughout the workflow and release scripts.
GitHub Actions runs `38493921672` and `38494419778` passed. CodeRabbit output
was treated as untrusted review input and does not constitute an independent
third-party security audit.

## Signing prerequisites still open

The release Mac currently has zero valid code-signing identities and zero GPG
secret keys. Private material must remain owner-controlled and must never be
printed, committed, pasted into chat, or stored in the release source tree.

The protected signing workflow expects these GitHub Actions secrets, whose
values must be supplied by the owner through GitHub's secret controls:

- `GRYNVAULT_APPLE_DEVELOPER_CERTIFICATE_P12_BASE64`
- `GRYNVAULT_APPLE_DEVELOPER_CERTIFICATE_PASSWORD`
- `GRYNVAULT_APPLE_PASSWORD`
- `GRYNVAULT_APPLE_TEAM_ID`
- `GRYNVAULT_APPLE_USERNAME`
- `GRYNVAULT_APPLE_SIGNING_IDENTITY`

PGP manifest signing additionally requires one owner-approved public key and
its full 40-hex fingerprint. The private PGP key stays outside GitHub Actions,
source, logs, and chat unless the owner approves a separately documented secure
signing path.

## Separate approval sequence

The review gate is green, and the owner separately approved exact source commit
`8a5478f379788e0c7978a4b8dfcfd2afebf29426` and immutable tag
`grynvault-rebalance-v0.1.0-rc.1`. Keep the remaining actions distinct:

1. **Complete:** create, push, and independently read back the approved
   immutable tag.
2. **Owner signing gate:** establish the owner-controlled Developer ID,
   notarization, and PGP paths, then explicitly authorize the protected workflow
   as a private signed candidate.
3. Inspect Developer ID signing, notarization, stapling, Gatekeeper acceptance,
   hashes, manifest, PGP signature, and clean-profile behavior.
4. Test clean installation and migration only with owner-approved copies and a
   verified rollback; do not move sats.
5. Stop again before publishing a GitHub release.
6. Stop separately before public DNS, Portal 04 deployment, upstream PRs,
   Nostr signing, and Stacker News submission/payment.

No action later in this sequence inherits approval from an earlier step.

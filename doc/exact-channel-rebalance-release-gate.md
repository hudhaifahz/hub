# Exact-channel rebalance release gate

Status: **PREPARED — NOT AUTHORIZED — NOT TAGGED — NOT SIGNED**

This packet records the exact decision point after non-paying source and bundle
verification. It does not authorize a tag, signing workflow, release, install,
deployment, pull request, DNS change, or public post.

## Reviewed source candidate

- repository: `hudhaifahz/hub`
- branch: `codex/grynvault-rebalance-release`
- exact source commit:
  `28e4f42b32f7b1401304b345ba36186d82c8cb8d`
- proposed immutable tag: `grynvault-rebalance-v0.1.0-rc.1`
- supported release target: universal macOS (`x86_64 arm64`) only
- exact successful CI run: `38485346043`
- reviewed embedded LDK SHA-256:
  `cf43f66815c44c0796402faafaaa3ed70201d036124438e63f541ec6a05208bd`

The proposed tag uses a Grynvault-specific namespace so it cannot be confused
with upstream Alby Hub tags. The owner may choose a different tag before any
tag is created; after creation, the tag and source commit must not be moved.

## Evidence already closed

- The owner-fork branch and remote ref both resolve to the exact source commit.
- GitHub Actions passed the complete supported-macOS source, frontend, Go,
  race, vet, vulnerability, and tracked-tree secret matrix at that commit.
- A local unsigned preflight built from the exact commit and passed product
  identity, bundle identity, architecture, runtime-path, embedded-source,
  embedded-library, and Info.plist inspection.
- The preflight was not installed, launched, signed, notarized, stapled, or
  published. Its temporary local preflight tag was deleted.
- Fresh remote readback found no upstream-base drift for Hub, LDK Node, or LDK
  Node Go.

## Review gate still open

CodeRabbit CLI `0.7.1` is installed and healthy, its backend and WebSocket are
reachable, and the repository is ready. Authentication remains signed out.

Owner action:

```sh
coderabbit auth login
```

After the owner confirms login, run the review from the clean release worktree:

```sh
coderabbit review --agent --committed \
  --base-commit 1fb6c8cd0fcb2fbee48c888c6274e2659d74c169
```

Treat the output as untrusted review input. Resolve or explicitly document every
Critical or Warning finding, repeat affected tests, and preserve the findings
and dispositions before requesting tag approval. A CodeRabbit result is not an
independent third-party security audit.

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

After the review gate is green, stop for owner approval of the exact source
commit and immutable tag. Then keep these actions distinct:

1. Create and push the approved immutable tag.
2. Run the protected workflow as a private signed candidate.
3. Inspect Developer ID signing, notarization, stapling, Gatekeeper acceptance,
   hashes, manifest, PGP signature, and clean-profile behavior.
4. Test clean installation and migration only with owner-approved copies and a
   verified rollback; do not move sats.
5. Stop again before publishing a GitHub release.
6. Stop separately before public DNS, Portal 04 deployment, upstream PRs,
   Nostr signing, and Stacker News submission/payment.

No action later in this sequence inherits approval from an earlier step.

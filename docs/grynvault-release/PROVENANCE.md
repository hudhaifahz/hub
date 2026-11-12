# Grynvault Rebalance provenance

Status: preparation record; final release values are appended only after the
immutable tag, owner-controlled PGP signing, and fresh-download verification.

## Source chain

| Component | Release source |
| --- | --- |
| Alby Hub upstream | `5a7f1d661f248a48659d19f9a4b3299ee282187e` |
| LDK Node | `e270baac5f6f8bf440a8c80f3365b15f9d911e94` |
| LDK Node Go | `fb5c2a997bb4842446e958eae22d58cb7164fa96` |

The final Grynvault packaging commit is intentionally not guessed here. It is
inserted into the release manifest from the exact immutable tag.

## Embedded native library

- file: `libldk_node.dylib`
- architectures: `x86_64 arm64`
- byte size: `50970000`
- SHA-256:
  `2c1ebe99232a306053c586f8a089a9d6546d034b91a77c0d83248c7333cbe5f2`
- install name: `@rpath/libldk_node.dylib`

## Product identity

- product: `Grynvault Rebalance`
- bundle identifier: `com.grynvault.rebalance`
- data directory name: `grynvault-rebalance`
- single-instance lock port: `21421`
- initial platform: universal macOS 15.4 or newer only
- canonical update and documentation origin: `https://rebalance.grynvault.com/`
- distribution class: `unsigned-experimental`
- Apple Developer ID signature: absent by owner decision
- Apple notarization and stapling: absent by owner decision
- application signature: ad-hoc bundle-integrity signature only

The application and native library are inspected before and after ad-hoc
signing. The latter changes their bytes, so the pinned raw LDK digest above is
verified before signing and recorded alongside the signed-byte artifact hashes.
Ad-hoc signing does not establish publisher identity or Gatekeeper trust.

## Brand asset

`appicon.png` is copied verbatim from the established Grynvault portal asset
`apps/web/public/grynvault-logo.png` on Frontier Crown security branch
`codex/rebalance-portal-security` at
`7054c902890772f34e1d1a9533ec5463862d2e45`.

- format: 192 by 192 RGBA PNG
- SHA-256:
  `fe343757d5b8ebac93dba0110fa4f12329705388ad01320deef6b94b38569f0e`

No AI-generated or Alby-branded icon is used for the release identity.

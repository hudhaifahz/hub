# Grynvault Rebalance provenance

Status: preparation record; final release values are appended only after the
immutable tag, owner-controlled PGP signing, and fresh-download verification.

## Source chain

| Component   | Upstream base                              | Reviewed feature head                      |
| ----------- | ------------------------------------------ | ------------------------------------------ |
| Alby Hub    | `5d5e696dfe7d430a365d48939c394376e078a56b` | `cd6bc1e399ba256adb113bbaf44d659c5b22a170` |
| LDK Node    | `553a2105b38da88fc5d0900007d30e9bd9669900` | `ed72fe992174965bc917d1dc6c6bee1dc7368057` |
| LDK Node Go | `25822674b495b8b0f7a8882f6b19a64b70d02396` | `3228bdbb7973a2743aee5e9f699e79afd217617b` |

The final Grynvault packaging commit is intentionally not guessed here. It is
inserted into the release manifest from the exact immutable tag.

## Embedded native library

- file: `libldk_node.dylib`
- architectures: `x86_64 arm64`
- byte size: `43,212,288`
- SHA-256:
  `879ea591a5f9636425ce56097ca0473df3fc05e62ef94135428a78a45087adcf`
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

# Grynvault Rebalance provenance

Status: preparation record; final release values are appended only after the
immutable tag, signing ceremony, and fresh-download verification.

## Source chain

| Component   | Upstream base                              | Reviewed feature head                      |
| ----------- | ------------------------------------------ | ------------------------------------------ |
| Alby Hub    | `b7d0b0be829ba387d6c2f95c54a1449b71f6cd0e` | `1fb6c8cd0fcb2fbee48c888c6274e2659d74c169` |
| LDK Node    | `74daaf98c7eeee7bca58c703c103d57affbd928a` | `2dc742a8b2057519e22d757f510e5cc880731b2c` |
| LDK Node Go | `5ba434093284d0dc8291e45d2a3d2c1c5a879698` | `a381cacb18562dff960f24a12f43232633f8bc25` |

The final Grynvault packaging commit is intentionally not guessed here. It is
inserted into the release manifest from the exact immutable tag.

## Embedded native library

- file: `libldk_node.dylib`
- architectures: `x86_64 arm64`
- byte size: `43,360,288`
- SHA-256:
  `cf43f66815c44c0796402faafaaa3ed70201d036124438e63f541ec6a05208bd`
- install name: `@rpath/libldk_node.dylib`

## Product identity

- product: `Grynvault Rebalance`
- bundle identifier: `com.grynvault.rebalance`
- data directory name: `grynvault-rebalance`
- single-instance lock port: `21421`
- initial platform: universal macOS only
- canonical update and documentation origin: `https://rebalance.grynvault.com/`

## Brand asset

`appicon.png` is copied verbatim from the established Grynvault portal asset
`apps/web/public/grynvault-logo.png` on Frontier Crown security branch
`codex/rebalance-portal-security` at
`7054c902890772f34e1d1a9533ec5463862d2e45`.

- format: 192 by 192 RGBA PNG
- SHA-256:
  `fe343757d5b8ebac93dba0110fa4f12329705388ad01320deef6b94b38569f0e`

No AI-generated or Alby-branded icon is used for the release identity.

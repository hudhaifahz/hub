# Building Grynvault Rebalance

Grynvault Rebalance is an experimental, independently maintained fork built
from Alby Hub. It is not an official Alby release.

## Supported release target

The first tested public target is universal macOS (`arm64` and `x86_64`). Other
upstream targets remain source-compatible but are not release-supported until
they receive their own complete acceptance evidence.

## Pinned inputs

- Alby Hub upstream base: `b7d0b0be829ba387d6c2f95c54a1449b71f6cd0e`
- reviewed exact-channel Hub head:
  `1fb6c8cd0fcb2fbee48c888c6274e2659d74c169`
- LDK Node feature head:
  `2dc742a8b2057519e22d757f510e5cc880731b2c`
- LDK Node Go binding head:
  `a381cacb18562dff960f24a12f43232633f8bc25`
- universal LDK library SHA-256:
  `cf43f66815c44c0796402faafaaa3ed70201d036124438e63f541ec6a05208bd`
- Go toolchain: `go1.27.2`, selected by `go.mod`
- Node.js: `22.x`
- Yarn: `1.22.22`
- Wails CLI: `v2.16.0`

The final source commit and immutable tag are recorded in the signed release
manifest. A moving branch name is never sufficient release provenance.
Public source artifacts contain the clean tagged tree. Patch-series attachments
are intentionally not produced because a patch can reproduce credential-shaped
text removed from upstream source.

## Non-paying local verification

Run these commands only against a clean checkout and never point `WORK_DIR` or
`DATABASE_URI` at a live wallet:

```bash
scripts/release/check-source-boundaries.sh
GOTOOLCHAIN=go1.27.2 go test ./...
GOTOOLCHAIN=go1.27.2 go test -race ./api ./wails
GOTOOLCHAIN=go1.27.2 go vet ./...
GOTOOLCHAIN=go1.27.2 govulncheck -show verbose ./...
gitleaks dir . --redact=100

cd frontend
yarn install --frozen-lockfile
yarn prepare:wails
yarn lint:js
yarn tsc:compile
yarn format
yarn build:http
yarn build:wails
yarn audit --groups dependencies
```

The inherited Bark deployment-target linker warnings are a documented baseline
residual. They do not change a test exit status and are not silently
reclassified as an application vulnerability.

## Release build

The release script refuses a dirty or untagged checkout and injects an immutable
source URL into the frontend:

```bash
scripts/release/build-macos.sh <immutable-tag>
scripts/release/inspect-macos.sh \
  "build/bin/Grynvault Rebalance.app" <immutable-tag> <source-commit>
```

This produces an unsigned internal candidate unless the owner separately uses
the protected signing workflow. Do not distribute an unsigned or ad-hoc signed
candidate as a public release.

## Owner-controlled signing gate

Developer ID signing, notarization, stapling, and PGP manifest signing require
separate owner approval and owner-controlled credentials. The GitHub workflow is
manual-only and cannot use Alby's signing identity. Its `build_release` input
must be explicitly enabled and the protected Grynvault Apple secrets must be
present. Never print, export, or commit private signing material.

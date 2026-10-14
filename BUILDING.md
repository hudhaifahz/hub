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

The release script explicitly links Apple's `UniformTypeIdentifiers` framework
for both macOS architectures. A fresh x86_64 link otherwise leaves Wails'
`UTType` reference unresolved on current SDKs.
It also pins the application and CGO deployment target to macOS 15.4, matching
the newest minimum declared by the inherited Bark native objects. Older macOS
versions are not represented as supported by this experimental package.

The final source commit and immutable tag are recorded in the signed release
manifest. A moving branch name is never sufficient release provenance.
Public source artifacts contain the clean tagged tree. Patch-series attachments
are intentionally not produced because a patch can reproduce credential-shaped
text removed from upstream source.

The tagged source tree also includes the end-user guide at
`docs/grynvault-release/USAGE.md`; its screenshots and provenance are pinned to
an immutable documentation commit rather than copied from a moving branch.

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

Any deployment-target linker warning is a release failure. The build and
bundle metadata must remain aligned at macOS 15.4 unless every native
dependency is rebuilt and re-tested for an older supported target.

## Release build

The release script refuses a dirty or untagged checkout and injects an immutable
source URL into the frontend:

```bash
scripts/release/build-macos.sh <immutable-tag>
scripts/release/inspect-macos.sh \
  "build/bin/Grynvault Rebalance.app" <immutable-tag> <source-commit>
```

This produces an unsigned internal candidate. To produce the owner-selected,
explicitly unsigned experimental disk image after the raw-byte inspection:

```bash
scripts/release/package-unsigned-macos.sh \
  "build/bin/Grynvault Rebalance.app" \
  <immutable-tag> <source-commit> build/out
```

The package script ad-hoc signs the native library and application for local
bundle-integrity checks, embeds a `README-UNSIGNED.txt` warning in the disk
image, verifies the mounted result, records the local Gatekeeper assessment,
and names the artifact `UNSIGNED-EXPERIMENTAL`. It does not create an Apple
publisher identity, notarization ticket, or Gatekeeper-accepted release.

Generate the pinned CycloneDX source SBOM with Syft 1.41.1:

```bash
scripts/release/generate-sbom.sh \
  build/out/Grynvault-Rebalance-<immutable-tag>-source.cdx.json
```

The SBOM generator fails unless it finds both Go and npm components and rejects
an output containing the local checkout path.

## Owner-controlled authenticity and publication gate

The owner has deliberately declined Apple Developer ID signing, notarization,
and stapling for this release path. Those capabilities remain isolated as an
unused private workflow mode and cannot use Alby's signing identity.

PGP manifest signing still requires separate owner approval and an
owner-controlled key. The manual GitHub workflow only produces a private
candidate artifact; it never publishes a GitHub release. Never print, export,
or commit private signing material. Public availability requires a detached
PGP signature, owner approval, and fresh-download verification of every URL.

Do not disable Gatekeeper globally. A user who verifies the manifest and elects
to accept unsigned software may use macOS's per-application Open/Open Anyway
flow. Building from the immutable tag is the safer option.

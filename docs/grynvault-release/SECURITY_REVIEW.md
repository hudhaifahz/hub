# Pre-release security review

Date: 2026-10-13

Status: internal pre-release review; not an independent third-party audit and
not a public-release approval.

## Reviewed boundary

- Grynvault product identity, data-directory isolation, updater boundary, and
  macOS release workflow
- exact-channel Hub, LDK Node, and Go-binding integration at the commits listed
  in `PROVENANCE.md`
- current source tree, production dependencies, generated frontend bundle, and
  non-paying unit/race/static-analysis paths

No wallet database, recovery material, signing key, notarization credential, or
live Lightning payment was used.

## Evidence matrix

| Check | Pinned tool/runtime | Result |
| --- | --- | --- |
| Go unit/integration suite | Go 1.27.2 | Pass |
| Focused race suite (`./api ./wails`) | Go 1.27.2 | Pass |
| Go static analysis | Go 1.27.2 `go vet` | Pass |
| Reachable Go vulnerability scan | `govulncheck` 1.9.0, database updated 2026-10-08 | 0 reachable vulnerabilities |
| Frontend dependency audit | Yarn 1.22.22 | 0 vulnerabilities in production dependencies |
| Current-tree secret scan | Gitleaks 8.30.1 rules, redacted output | 0 findings after remediation |
| GitHub workflow lint | actionlint 1.7.7 | Pass |
| Frontend ESLint, TypeScript, Prettier | versions locked by `frontend/yarn.lock` | Pass |
| HTTP and Wails frontend builds | Node 22 / locked Yarn dependencies | Pass |
| Source-boundary assertions | `scripts/release/check-source-boundaries.sh` | Pass |

An unsigned universal preflight build identified and fixed duplicate runtime
path processing caused by `otool` reporting each path once per architecture.
The release script now deduplicates paths before mutation; a fresh exact-commit
candidate must pass bundle inspection before the signing gate can open.
Generated Wails bindings are excluded from source formatting checks, and the
release build fails if packaging changes any tracked file.
Wails is also required to clean its generated binary directory, skip module-file
rewrites, and remove local build paths before compiling a candidate.
Because Wails applies an ad-hoc signature to the packaged native library, the
release script restores the reviewed pinned LDK bytes after packaging and before
the owner-controlled Developer ID signing step.
Bundle inspection normalizes identical per-architecture Mach-O metadata before
requiring one exact runtime path and one exact native-library install name.

The secret review removed credential-shaped NWC connection strings inherited
from upstream source and made the optional Lightning messageboard explicitly
operator-configured. Those historical values are not reproduced in release
patch attachments. Anyone responsible for the upstream credentials should
treat previously published values as exposed and rotate or revoke them.

## Non-reachable dependency advisories

The final `govulncheck -show verbose ./...` result reports that application code
is affected by zero vulnerabilities. It also identifies two advisories without
a reachable symbol path:

- `GO-2026-4518` in `github.com/jackc/pgproto3/v2` 2.3.3, with no fixed release
  listed by the Go vulnerability database;
- `GO-2026-5932` for the unmaintained `golang.org/x/crypto/openpgp` package;
  the scanner found no call from this application to the affected package.

The separately reported `github.com/klauspost/compress/s2` advisory was removed
by updating the indirect dependency to 1.18.7. These residual module/package
notices are recorded rather than misrepresented as reachable findings.

## Known build warning

The inherited Bark macOS static library was built with a macOS 15.4 deployment
target while the application link target is macOS 15.0. The linker warns but
the builds and tests succeed. This must remain visible in candidate evidence and
be re-evaluated before broadening the supported macOS range.

## Gates still closed

This review does not authorize tagging, Developer ID signing, notarization,
stapling, installation against a live wallet, a public download, DNS changes,
an upstream pull request, or any Lightning value movement. Those steps require
their separate owner and acceptance gates.

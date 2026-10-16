# Pre-release security review

Date: 2026-10-16

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

| Check                                 | Pinned tool/runtime                                 | Result                                       |
| ------------------------------------- | --------------------------------------------------- | -------------------------------------------- |
| Go unit/integration suite             | Go 1.27.2                                           | Pass                                         |
| Focused race suite (`./api ./wails`)  | Go 1.27.2                                           | Pass                                         |
| Go static analysis                    | Go 1.27.2 `go vet`                                  | Pass                                         |
| Reachable Go vulnerability scan       | `govulncheck` 1.9.0, current database on 2026-10-16 | 0 reachable vulnerabilities                  |
| Frontend dependency audit             | Yarn 1.22.22                                        | 0 vulnerabilities in 134 production packages |
| Current-tree secret scan              | Gitleaks 8.30.1 rules, redacted output              | 0 findings after remediation                 |
| GitHub workflow lint                  | actionlint 1.7.7                                    | Pass                                         |
| Frontend ESLint, TypeScript, Prettier | versions locked by `frontend/yarn.lock`             | Pass                                         |
| HTTP and Wails frontend builds        | Node 22 / locked Yarn dependencies                  | Pass                                         |
| Source-boundary assertions            | `scripts/release/check-source-boundaries.sh`        | Pass                                         |

The candidate source replays the exact-channel work onto Alby Hub
`v1.24.3` and pins the reviewed LDK and Go binding heads recorded in
`PROVENANCE.md`. The final upstream-oriented Hub diff received a complete
CodeRabbit review with zero findings after fixing dependency provenance and a
transient reconciliation-status polling gap. CodeRabbit is review input, not an
independent security audit.

The offline RC.3 acceptance test revealed that a failed desktop OAuth callback
could place its one-time authorization-code query value in a Wails route log.
RC.3 was withheld from publication. The successor source removes every query
value from Wails route logs and reflected unhandled-route errors, redacts
payment-request path segments, and applies the payment redaction even when URL
parsing fails on malformed escapes. Focused regression tests cover OAuth query
values, ordinary query values, payment requests, and malformed payment paths.
After the malformed-path fallback was corrected, a complete CodeRabbit review
of the RC.3-to-successor diff reported zero findings. CodeRabbit is review
input, not an independent security audit.

An unsigned universal preflight build identified and fixed duplicate runtime
path processing caused by `otool` reporting each path once per architecture.
The release script now deduplicates paths before mutation; a fresh exact-commit
candidate must pass bundle inspection before the signing gate can open.
Generated Wails bindings are excluded from source formatting checks, and the
release build fails if packaging changes any tracked file.
The clean-checkout release test also found that platform-specific Wails utility
files and generated bindings must exist before TypeScript validation. A
bindings-only build target now creates them without claiming the desktop port,
opening a wallet, or starting the Hub service; the release then runs
`prepare:wails` before lint and type checking instead of relying on files left
behind by an earlier developer session.
The same clean universal build exposed an implicit local linker dependency: the
x86_64 target could resolve Wails' `UTType` reference only when the shell had
previously supplied the framework flag. The release script now explicitly links
Apple's `UniformTypeIdentifiers` framework for both architectures.
Wails is also required to clean its generated binary directory, skip module-file
rewrites, and remove local build paths before compiling a candidate.
Because Wails applies an ad-hoc signature to the packaged native library, the
release script restores the reviewed pinned LDK bytes after packaging. The
unsigned experimental packager verifies those raw bytes first, then applies an
explicit ad-hoc signature to both the library and application and verifies the
mounted disk image.
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

## macOS compatibility floor

The inherited Bark macOS static library was built with a macOS 15.4 deployment
target. The application CGO flags, linked executable, and bundle metadata are
therefore pinned to macOS 15.4 rather than advertising an older unverified
target. Compatibility below 15.4 requires rebuilding and testing every native
dependency and is outside this release boundary.

## Known distribution risk

The owner has declined Apple Developer ID signing, notarization, and stapling.
The resulting package has no Apple-verifiable publisher identity and Gatekeeper
is expected to warn or block it. Ad-hoc signing detects some post-signing bundle
changes but is not an authenticity guarantee. Release filenames, the in-image
warning, provenance, checksum index, and documentation must never imply Apple
trust. PGP manifest signing and independently verified hashes remain mandatory
before any public download is approved.

## Gates still closed

This review does not authorize a new immutable tag, PGP signing, installation
against a live wallet, a public download, DNS changes, an upstream pull request,
or any Lightning value movement. Those steps require their separate owner and
acceptance gates. Apple signing and notarization are intentionally omitted from
the selected path rather than treated as pending approvals.

The guided importer is also a release blocker. A later internal acceptance run
showed that its copy, hash, and SQLite checks could faithfully accept an older
work directory for the same embedded-LDK node. When that stale copy was started
online, LDK detected lost channel state and stopped; a peer subsequently
force-closed the two channels it shared with that instance. The exact-channel
rebalance implementation was not involved, but the migration lifecycle was not
safe enough for publication.

No successor candidate may ship until the requirements in
`GUIDED_MIGRATION_PLAN.md` pass, including competing-copy detection, verified
source quiescence, custom-`WORK_DIR` handling, and a startup-level network
quarantine that prevents an imported node from contacting peers before local
identity and continuity acceptance. Public documentation must keep the incident
sanitized and must not include a real node public key, channel ID, transaction
ID, wallet path, balance, or credential.

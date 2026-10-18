# Pre-release security review

Date: 2026-10-17

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
| Focused race suite                    | Go 1.27.2; migration, service, API, Wails packages  | Pass                                         |
| Go static analysis                    | Go 1.27.2 `go vet`                                  | Pass                                         |
| Reachable Go vulnerability scan       | `govulncheck` 1.9.0, current database on 2026-10-16 | 0 reachable vulnerabilities                  |
| Frontend dependency audit             | Yarn 1.22.22                                        | 0 vulnerabilities in 134 production packages |
| Current-tree secret scan              | Gitleaks 8.30.1 rules, redacted output              | 0 findings after remediation                 |
| GitHub workflow lint                  | actionlint 1.7.7                                    | Pass                                         |
| Frontend ESLint, TypeScript, Prettier | versions locked by `frontend/yarn.lock`             | Pass                                         |
| HTTP and Wails frontend builds        | Node 26.11.1 / locked Yarn dependencies             | Pass                                         |
| Source-boundary assertions            | `scripts/release/check-source-boundaries.sh`        | Pass                                         |

### Guided migration addendum

The 2026-10-16 successor working tree was reviewed and tested separately from
the already-withheld RC.4 artifact. It introduces no release tag or artifact.
The migration security boundary now includes:

- a stable local-only wallet comparison fingerprint and fail-closed matching
  copy detection across known Alby and Grynvault locations;
- macOS filesystem-locality enforcement for both source and target, rejecting
  network-mounted folders before staging while allowing locally attached
  volumes and custom paths;
- two quiescence snapshots plus post-copy source hashing and process checks;
- byte-verified, permission-restricted source recovery and staging trees;
- deterministic resume when interruption occurs before or after target moves;
- a startup-consumed quarantine marker that prevents auto-unlock, LDK client
  construction, peer/relay startup, ordinary setup, and non-migration Wails
  routes;
- offline identity/channel/balance/history/NWC verification using a disposable
  LDK storage copy;
- a fresh continuity and source-process check immediately before activation;
- automatic return to offline quarantine after interrupted or failed
  activation;
- a restart-safe abandon transition that archives the failed target/stage and
  control record, restores prior Grynvault data when present, and records the
  action before permitting a new import; and
- synthetic responsive fixtures that are visibly labeled and contain no real
  wallet data.

The fork also suppresses upstream Alby release banners and release-note widgets.
An upstream version is dependency/provenance input, not a safe binary update for
an installed Grynvault wallet. Users may be directed to a replacement only
after the compatible Grynvault source, artifact, signed manifest, and migration
path pass their own release gates. This candidate adds no automatic updater or
unsigned remote update feed.

Direct regression coverage now includes permission-denied source data,
unrelated known wallet copies, validation-sidecar nonmutation, non-local source
and target rejection, authoritative copy-progress events, interrupted
activation recovery, control-only blocked quarantine, and correct-password
identity mismatch. These complement the existing state-machine, regtest, and
network-dark acceptance coverage; they do not replace owner-supervised
acceptance on the final packaged candidate.

The real disposable acceptance path created two 500,000-sat regtest channels
and both incoming and outgoing history, exercised standard and custom source
paths, proved the destination listener remained unused while quarantined,
activated deliberately, observed peer reconnection, and blocked a matching
stale clone before any network check. No production wallet or public network
was involved.

The current rerun passed the full Go suite with Wails tags, focused race tests
for `guidedmigration`, `service`, `api`, and `wails`, `go vet`, both frontend
build modes, source-boundary assertions, Gitleaks 8.30.1, actionlint 1.7.7,
Yarn's production dependency audit, and govulncheck 1.9.0. Govulncheck again
reported zero reachable vulnerabilities and the same two non-reachable notices
recorded below. The inherited frontend build still reports the existing
`lottie-web` direct-`eval` warning and oversized bundle warning; neither is
introduced by the migration component, and the production dependency audit
reports zero advisories.

CodeRabbit's complete review of the staged migration diff first identified a
misleading default-folder error, unsafe skipping of a corrupt known competing
wallet, the absence of a supported archive/reset path for blocked imports, and
two status/documentation clarity issues. Each was reproduced or confirmed,
fixed, and covered where appropriate by regression tests. The final complete
review reported zero findings. CodeRabbit remains review input rather than an
independent security audit.

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

The first owner-supervised attempt at commit
`e2ead0a0497a5e2b94d0fe5ff0da47593f8680c7` found a separate availability bug:
the destination database created during a normal first launch was re-scanned as
an arbitrary competing wallet and blocked every real guided import before
staging. The stop was safe—external networking was denied, the source remained
byte-identical, and no migration control or staged wallet existed—but the owner
gate did not pass. The remediation permits only a demonstrably pristine target
and retains fail-closed handling for corrupt, partial, configured, or active
wallet state. A new exact-commit owner acceptance is required after the full
matrix and review pass.

The second owner-supervised attempt at commit
`ffe237cd90d8bdcdf8038a1d66c154faac9ac487` passed source/destination hashing,
pristine-target handling, startup quarantine, and the complete offline
continuity inspection. An intentional process-level network denial then caused
activation to fail at LDK connection startup, while the copied wallet stayed
quarantined and no Lightning action was created. A subsequent click exposed a
recoverability defect: `activation_failed` retained valid verification evidence
but was not accepted as a retryable activation state, so the backend emitted a
misleading verification-required error. The successor treats only a failed
activation with preserved verification as retryable, repeats all final
continuity and process-exclusion checks, clears the password field after a
failed attempt, and renders the stored connection failure and offline state.
This change does not weaken quarantine or permit ordinary startup. A new exact
commit and owner-supervised acceptance are required.

The successor diff passed the full Go suite with Wails tags, focused race
coverage, `go vet`, both frontend build modes, formatting, source-boundary
assertions, production dependency audit, a current Gitleaks scan, and a current
govulncheck scan with zero reachable vulnerabilities. CodeRabbit reviewed all
eight changed files and reported zero findings. The synthetic
`activation_failed` screen was inspected at desktop and 390-pixel mobile widths
with no horizontal overflow; it displays the preserved verification, exact
connection error, offline status, recovery controls, and guarded retry action.

The next owner-supervised attempt at exact commit
`31d2c3b807be5d8a8fa69f6caec8a80bca7a480b` passed the copy, hash,
integrity, quarantine, and offline continuity checks. The owner then selected
the activation control unintentionally during review. The external network
sandbox prevented LDK from starting or reaching peers, and the canonical wallet
remained byte-identical and later reconnected normally. No Lightning operation
succeeded, but the workflow did not provide enough human-error resistance.

The successor therefore clears the verification password, hides activation
behind a separate review action, requires password re-entry, the full node ID,
an exact network-and-wallet-bound phrase, an acknowledgement checkbox, and a
final native confirmation. The backend independently validates the phrase,
rejects verification older than five minutes, and requires a new offline
verification after any failed activation. These changes require a fresh exact-
commit owner acceptance; this review still does not authorize tagging, signing,
publishing, installation, or value movement.

That acceptance rerun stopped before staging after exposing an asynchronous
source-selection race: automatic discovery could complete after a manual folder
choice and overwrite the reviewed path. The successor now makes the start of
manual selection authoritative and ignores any later discovery response. This
must be rechecked in the exact-commit owner acceptance.

The later PGP-verified RC.5 package rerun exposed a separate acceptance-launch
failure: macOS LaunchServices did not retain the shell-selected disposable
`WORK_DIR` or the launcher process's network sandbox. The installed app opened
the canonical wallet and briefly started its normal LDK runtime. The operator
stopped it before entering migration or activation; no acceptance action
created a Lightning operation. The canonical wallet was restored from the
fresh pre-run backup, matched its original complete file manifest, and passed
both SQLite integrity checks. RC.5 package acceptance therefore failed and
publication remains blocked.

The successor's release-only launcher runs before wallet locks, database open,
or service initialization. It validates a mode-restricted sidecar bound to the
canonical disposable path, rejects symlink targets and incomplete or duplicate
flags, sets the exact work directory internally, self-reexecutes the packaged
binary under macOS `sandbox-exec`, and proves outbound denial by requiring a
reserved-address dial to fail with `EPERM` or `EACCES`. Guided-migration
replacement processes inherit the sandbox marker and path, revalidate the
sidecar, and repeat the denial proof before service construction. Normal
desktop launches do not enter this acceptance-only path.

# Linux and Windows release plan

Status: **PLANNED — NO PLATFORM BUILD, TAG, SIGNATURE, OR PUBLICATION APPROVED**

Prepared: 2026-10-18

The RC.6 source remains immutable. Linux and Windows support will be delivered
from a new exact source commit and, after acceptance, a new owner-approved
immutable release tag. A successful compile is not platform acceptance.

## Initial supported targets

| Priority | Target                           | Initial package                               | Native LDK library                        | Status  |
| -------- | -------------------------------- | --------------------------------------------- | ----------------------------------------- | ------- |
| 1        | Linux x86-64, Ubuntu 22.04 class | permission-preserving `.tar.bz2`              | `x86_64-unknown-linux-gnu/libldk_node.so` | planned |
| 2        | Linux x86-64, Ubuntu 24.04 class | separate WebKit 4.1 `.tar.bz2`                | same target library, tested separately    | planned |
| 3        | Windows 11 x86-64                | portable `.zip`, then reviewed NSIS installer | `x86_64-pc-windows-msvc/ldk_node.dll`     | planned |

Linux arm64 is a later target. The pinned binding module contains an aarch64
library, but upstream Alby desktop CI currently accepts Linux amd64, Windows
amd64, and macOS. Arm64 must not be advertised until its Wails packaging,
runtime dependencies, migration safety, and real-hardware acceptance pass.

## Current reusable inputs

- Wails `v2.16.0`, Go `1.27.2`, Node.js `22.x`, and Yarn `1.22.22`
- pinned exact-channel binding module
  `github.com/hudhaifahz/ldk-node-go@3228bdbb7973`
- present native library hashes in the reviewed module cache:
  - Linux x86-64 `libldk_node.so`:
    `83fd72fedb6f35bb4cf1bc556283e7193c2bec2ea57c2439a264b797661e2128`
  - Windows x86-64 MSVC `ldk_node.dll`:
    `0e87876c0c5a888b957ca7c6ea2a6b6ca9a98c899ccd4cc3a0e6890e53230420`

These hashes are planning evidence, not release hashes. Final libraries must be
rebuilt or freshly retrieved from the exact accepted binding revision, then
hashed and tested on their target systems.

## Cross-platform blockers that must be fixed first

The current fork compiles platform-neutral stubs but its seamless migration
safety boundary is macOS-specific:

- `guidedmigration.IsLocalPath` currently permits every non-macOS path instead
  of rejecting network or remote filesystems;
- `guidedmigration.OpenWalletProcessIDs` performs no source-process check
  outside macOS;
- opening the recovery folder returns “currently supported in the macOS
  desktop app” on other systems;
- the process-level network-dark acceptance boundary refuses non-macOS
  platforms;
- a desktop lock conflict has no native Linux/Windows user-facing alert;
- standard Alby and Grynvault data-directory discovery needs explicit OS rules
  and tests for XDG/Linux and `%LOCALAPPDATA%`/Windows layouts.

No Linux or Windows package may be described as supporting seamless migration
until these are implemented fail-closed and accepted on the target OS.

## Engineering sequence

### Phase 1: shared platform boundary

1. Refactor process ownership, local-filesystem verification, recovery-folder
   reveal, and acceptance networking into explicit OS adapters.
2. Keep the macOS implementation unchanged behind its adapter and preserve all
   existing regression tests.
3. Add a common capability report so unsupported safety checks disable import
   before copying data rather than silently becoming permissive.
4. Keep exact-channel quote/execution code platform-neutral and preserve its
   database schema and route-redaction invariants.

### Phase 2: Linux x86-64

1. Build on native GitHub Ubuntu runners, not by cross-compiling on macOS.
2. Follow upstream's two-runtime matrix:
   - Ubuntu 22.04/WebKitGTK 4.0;
   - Ubuntu 24.04/WebKitGTK 4.1 with the `webkit2_41` build tag.
3. Bundle `libldk_node.so` beside the application, set a relative executable
   RPATH, verify it with `readelf`/`ldd`, and fail on an unexpected host-library
   resolution.
4. Implement Linux local-filesystem classification, process/file ownership
   detection, XDG data discovery, desktop lock notification, folder reveal,
   and a network namespace/firewall-backed acceptance launcher.
5. Preserve executable permissions in the archive and test extraction on a
   clean supported distribution.

### Phase 3: Windows x86-64

1. Build on a native Windows runner with WebView2 and the reviewed MSVC LDK
   DLL; do not rely on a Linux/macOS cross-build as release evidence.
2. Place `ldk_node.dll` beside the executable and verify PE architecture,
   imports, exported binding symbols, and clean-machine startup.
3. Implement Windows local-volume classification, open-handle/process checks,
   `%LOCALAPPDATA%` discovery, Explorer folder reveal, named-mutex or equivalent
   wallet ownership enforcement, and Windows Firewall-backed acceptance
   isolation.
4. Validate the portable ZIP first. Generate the NSIS installer only after its
   install, upgrade, uninstall, and data-preservation behavior are tested.
5. Decide the Windows authenticity policy before public release. A detached
   PGP-signed manifest authenticates downloaded bytes but does not remove
   Microsoft Defender SmartScreen warnings from an unsigned executable.

## Per-platform non-paying gate

Run from a clean exact-commit checkout with no live wallet path configured:

- source-boundary and secret scans;
- full Go tests plus focused race tests;
- migration rollback and platform-adapter tests;
- `go vet` and current reachable-vulnerability scans;
- locked frontend install, lint, TypeScript, formatting, and both frontend
  builds;
- exact-channel deterministic topology, stale quote, fee cap, concurrency,
  duplicate submission, crash/restart, incorrect return channel, and route-log
  redaction tests;
- native Wails build and target-specific dependency inspection;
- fresh SBOM, artifact hashes, and malware scan;
- responsive UI screenshots using synthetic identifiers at narrow and wide
  viewports.

## Clean-install and migration acceptance

Each OS requires an owner-supervised, documented acceptance run that:

1. starts from a clean machine or VM and verifies the exact candidate hashes;
2. confirms an empty work directory shows no phantom wallet or channels;
3. creates hash-verified backups without modifying the source wallet;
4. tests automatic standard-location discovery and a separately selected custom
   location, including spaces and an attached local volume;
5. rejects a network share, symbolic-link escape, running source, stale copy,
   matching competing copy, corrupted database, identity mismatch, changed
   source, and occupied target;
6. stages byte-exact copies and enters persistent network quarantine before
   unlocking the copy;
7. verifies database integrity, wallet/node identity, channels, balances,
   history, NWC permissions, backup state, route-log redaction, and responsive
   UI while external networking is denied;
8. requires a separate plain-language activation confirmation and rechecks the
   source immediately before networking;
9. verifies deliberate activation, peer reconnection, restart persistence, and
   an abandon/rollback path without touching the preserved source;
10. stops the test copy and proves the canonical wallet can restart normally.

No invoice, probe, HTLC, payment, swap, rebalance, or channel operation is
permitted during this migration acceptance.

## Live exact-channel acceptance

Live execution is a later, separate owner gate for each target. The action
packet must name the exact outgoing and incoming channel IDs and full peer
public keys, principal in sats and BTC, fresh estimated fee, hard fee cap,
maximum source debit, quote expiry, and expected before/after balances.

After one approved small rebalance, independently read back the saved operation
state, actual fee, exact first channel, observed returning channel, unidentified
receiving count, and channel balances. Never retry an ambiguous outcome.

## Release and update policy

- Never add new binaries to or move the immutable RC.6 tag.
- Prefer one new RC tag only when Linux, Windows, and macOS artifacts are all
  built from the same exact source commit. Otherwise use distinct candidates
  until the sources converge.
- Create a platform-specific manifest section for filename, architecture,
  native-library hash, packaging mode, OS signing status, and SBOM.
- Owner approval is required separately for tag creation, PGP signing, GitHub
  publication, portal changes, and any live rebalance.
- Continue suppressing stock Alby update prompts. Advertise an upstream update
  only after a compatible Grynvault build and migration path pass these gates
  on every supported platform.

## Definition of public support

A platform is public only when its source, package, signed manifest, SBOM,
clean-machine installation, seamless migration, rollback, responsive UI,
exact-channel safety tests, and owner-approved live acceptance all have durable
evidence. Until then, label it **planned** or **experimental internal build**—not
supported.

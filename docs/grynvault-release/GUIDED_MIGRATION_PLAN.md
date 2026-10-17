# Guided migration release plan

Status: **IMPLEMENTED — SYNTHETIC ACCEPTANCE PASSED — OWNER-SUPERVISED
ACCEPTANCE AND RELEASE APPROVALS PENDING**

Owner requirement: a seamless migration from an existing Alby Hub embedded-LDK
wallet is a core product capability. Grynvault Rebalance must not be published
as a downloadable Alby fork until this plan is implemented and accepted.

This document is the execution brief for the next release candidate. It is not
permission to tag, sign, publish, connect a copied production wallet to the
Lightning network, or move sats. Those remain separate owner-controlled gates.

## Outcome

A person installing Grynvault Rebalance can bring an existing Alby Hub wallet
over through one guided flow, including when Alby Hub uses a non-default
`WORK_DIR`, without seeing a frightening empty-wallet screen and without ever
starting an unverified copy of an LDK node online.

The normal journey is:

1. Grynvault finds the standard Alby Hub data directory automatically.
2. If it is elsewhere, the person chooses the Alby Hub application-data folder
   or exact `WORK_DIR` using the native folder picker.
3. Grynvault identifies and validates the wallet, checks for competing copies,
   and presents a plain-language summary before copying anything.
4. Grynvault creates a recovery snapshot, stages a byte-verified copy, and
   restarts into migration quarantine.
5. The person unlocks locally and reviews the node identity and continuity
   summary.
6. Only after local acceptance does Grynvault activate networking and open the
   normal wallet interface.

The person never types a seed, password, macaroon, NWC secret, or signing key
into a Grynvault website, hosted service, log, or support form.

## Why the current importer cannot ship

The current importer correctly copies, hashes, and validates the selected
SQLite data. That proves copy integrity, not that the selected directory is the
latest state of the Lightning node.

During internal acceptance, an older work directory for the same node was
copied successfully and then started online. LDK detected lost channel-monitor
or channel-manager state and stopped. A peer subsequently force-closed the two
channels it shared with that stale instance. Other channels remained open
because channel state and peer reactions are independent. No evidence of a
penalty was observed, but the incident proves that a valid historical copy is
not a safe live-migration source.

The release must therefore treat these as separate claims:

- **Copy integrity:** destination bytes equal the quiescent source bytes.
- **Database integrity:** required SQLite databases and records are readable.
- **Wallet identity:** the selected data belongs to the intended wallet/node.
- **Source authority:** no newer or concurrently active copy of the same node
  is known.
- **Network activation:** the verified destination may start its LDK network
  services.

Passing the first two claims must never imply the last three.

## Product contract

### Supported source discovery

The importer must:

- check the standard macOS Alby Hub container location;
- recognize selection of an app container, `Data` directory, Application
  Support directory, `albyhub` directory, or exact custom `WORK_DIR`;
- provide **Choose another Alby data folder** for arbitrary paths, including
  external volumes and paths containing spaces;
- remember no sensitive path remotely and never upload directory metadata;
- avoid recursive whole-disk searches;
- inspect known Grynvault and Alby locations for competing copies of the same
  wallet before allowing activation.

Automatic discovery is convenience, not authority. The confirmation screen
must show the resolved absolute source path before import.

### Source validation

Before a source is eligible, the importer must verify:

- `nwc.db` exists and passes `PRAGMA integrity_check`;
- `ldk/storage/ldk_node_data.sqlite` exists and passes the same check;
- required wallet configuration records exist;
- the backend is embedded LDK;
- the source and target are separate, local paths with no symbolic-link escape;
- relevant data is quiescent across two snapshots;
- Alby Hub is not running against the source;
- Grynvault is not running another service against the source or destination;
- the source did not change during backup, staging, or validation;
- no existing target wallet would be replaced without an explicit, separately
  recoverable migration decision.

The preflight summary must show, where safely available:

- product/source type;
- resolved source path;
- local-only wallet fingerprint;
- network and embedded-LDK backend;
- application, permission, transaction, and migration counts;
- channel count and a non-secret continuity summary;
- most recent durable activity evidence and its limitations;
- whether another matching wallet directory was found.

No checksum, file timestamp, database count, or channel count alone proves that
a Lightning state copy is current. The UI and logs must not claim otherwise.

### Duplicate and stale-state policy

The importer must derive a stable, local-only comparison fingerprint from
durable wallet identity material without logging secrets. It must compare the
selected source with every known Alby/Grynvault candidate directory that can be
read safely.

Fail closed when:

- a matching copy has stronger evidence of newer durable activity;
- two matching copies cannot be ordered confidently;
- a matching copy appears active;
- the selected source changes during the operation;
- the source is explicitly identified as a backup, snapshot, or staged restore
  rather than an authoritative live work directory.

The recovery message must name the conflicting local paths and explain that no
wallet was changed. Historical backups belong in a separate disaster-recovery
workflow; they must not be accepted by the live guided importer.

### Process exclusion and quiescence

The desktop lifecycle must enforce, not merely request, single ownership:

- Grynvault continues to hold its own single-instance lock and the compatible
  Alby desktop lock for its entire lifetime;
- preflight checks the selected directory for active use using the strongest
  platform-supported process/file-lock evidence;
- two full metadata/hash snapshots separated by a bounded stability interval
  must match before copying;
- a final source snapshot must still match after staging and validation;
- any uncertainty aborts before `restore` activation.

Closing a window is not accepted as proof that Alby Hub stopped. The importer
must state whether exclusive ownership was actually established.

### Copy, backup, and evidence

The importer must preserve the good behavior already implemented:

- create a timestamped, permission-restricted source snapshot;
- preserve any existing Grynvault target before modification;
- copy only durable wallet state and static channel backups, not logs;
- refuse symlinks and non-regular wallet files;
- preserve legitimate SQLite WAL/SHM sidecars;
- remove only sidecars generated solely by read-only validation;
- hash source snapshot, staging tree, and activated restore tree;
- require byte-for-byte equality and successful database integrity checks;
- write a redacted local evidence record with schema version, source path,
  counts, hashes, validation results, and state-machine checkpoints;
- never record seeds, unlock passwords, NWC secrets, invoices, preimages,
  payment secrets, serialized routes, or private keys.

The operation must be crash-safe and resumable. A crash at any checkpoint must
leave either the untouched pre-import target or a clearly marked staged import,
never an ambiguous half-activated wallet.

### Network quarantine and activation

This is the mandatory control added after the internal incident.

After a direct import, Grynvault must restart with all LDK peer connection,
gossip, invoice, payment, rebalance, swap, and channel-operation capabilities
disabled. The setup UI may access only the local copied databases needed for
verification.

While quarantined, the app must:

- show **Migration verification — Lightning is offline** instead of the normal
  empty or disconnected wallet;
- accept the existing unlock password only in the local app UI;
- verify that the unlocked node identity matches the preflight fingerprint;
- display the expected channel identities/count, stored balances, history,
  NWC apps/permissions, backup state, and embedded release hashes;
- explain any difference before offering activation;
- keep all value-moving controls unavailable;
- provide **Return to source wallet** and evidence-folder actions that do not
  start networking.

The activation confirmation must state the exact source path, node fingerprint,
channel count, network, and that the source Alby Hub must remain stopped.
Offline verification and network activation are separate screens/actions. The
verification password is cleared before activation review. Activation requires
the password again, the complete verified node ID, an exact typed phrase bound
to the verified network and wallet fingerprint, an acknowledgement checkbox,
and a final native confirmation. The backend accepts only a verification less
than five minutes old; a failed activation consumes that verification and
requires another offline verification before retry. Activation is recorded
without recording credentials.

Only successful activation clears the quarantine marker. A failed or cancelled
verification remains offline and recoverable.

## UX requirements

The flow must remain short even though the controls are strict:

1. **Wallet found** — show the detected location and an option to choose a
   different folder.
2. **Ready to copy** — show identity, continuity, source-authority, and backup
   checks with one clear confirmation.
3. **Copying safely** — show checkpoint progress; never close without first
   showing what will happen next.
4. **Verify your wallet** — after restart, unlock locally while Lightning is
   visibly offline.
5. **Ready to activate** — show the final continuity summary and one activation
   action.
6. **Migration complete** — show the connected node and channels, recovery-copy
   location, and a reminder not to reopen the source installation.

Requirements across all supported window sizes:

- no clipped identifiers, controls, evidence, or footer buttons;
- long paths and identifiers wrap or provide safe copy controls;
- progress survives app restart;
- errors explain what was preserved and what the person should do next;
- the UI never reports that channels are missing merely because migration has
  not activated them yet;
- accessible keyboard navigation, focus order, and status announcements;
- synthetic screenshots for the default path, custom path, stale-copy block,
  quarantine verification, and successful completion.

## Implementation workstreams

### A. Discovery and preflight API

- Split the current one-shot import endpoint into discover, inspect, stage,
  verify, activate, cancel, and status operations.
- Add typed request/response models and a persisted migration state machine.
- Make folder selection explicit and separately inspectable before mutation.
- Implement custom-path resolution and known-candidate enumeration.
- Add local-only wallet fingerprinting and duplicate comparison.

### B. Lifecycle safety

- Add verifiable source-process exclusion and quiescence checks.
- Introduce an import-quarantine marker consumed before service startup.
- Prevent construction/start of the LDK networking service while quarantined.
- Make activation atomic and restart-safe.
- Ensure setup cancellation cannot fall through to ordinary node startup.

### C. Copy and recovery hardening

- Retain byte-exact staging and post-validation hashing.
- Add final-source revalidation and crash checkpoints.
- Add target-conflict handling and rollback metadata.
- Version and document the evidence schema.

### D. Guided interface

- Replace the current single-button screen with the six-step journey above.
- Add custom-directory selection without penalizing the default path.
- Show plain-language freshness ambiguity and safe recovery choices.
- Keep the full experience responsive and unclipped.

### E. Documentation and release integration

- Update `MIGRATION.md`, release notes, security review, usage guide, manifest,
  portal download copy, and repository README.
- Treat Grynvault as a tested fork with its own compatibility boundary. Do not
  show raw upstream Alby releases as actionable updates. Announce and link an
  update only after the upstream release has been rebased into Grynvault, the
  complete safety/release matrix has passed, and the Grynvault artifact and
  signed manifest have been published.
- Publish only synthetic screenshots and redacted fixture evidence.
- Explain the difference between live migration and disaster recovery.
- Document custom `WORK_DIR` selection for macOS and future platforms.

## Verification matrix

No production-channel copy may be placed online during development or release
acceptance.

### Automated tests

- default macOS container discovery;
- custom `WORK_DIR`, parent-container, spaces, Unicode, and external-volume
  path resolution;
- missing, corrupt, incomplete, non-LDK, symlinked, and permission-denied data;
- live source process and changing source files;
- same source/target, nested paths, existing target, and staged import;
- matching newer copy, matching ambiguous copy, unrelated wallet copy, and
  historical backup selection;
- WAL/SHM preservation and validation-generated sidecar cleanup;
- byte/hash equality before and after database validation;
- interruption at every state-machine checkpoint and deterministic resume;
- quarantine surviving restart, crash, and failed unlock;
- zero peer/network initialization while quarantined;
- identity mismatch and continuity mismatch blocking activation;
- successful atomic activation and retained recovery evidence;
- log-redaction regression coverage.

### Synthetic/regtest acceptance

- create two simulated channels and history on a disposable regtest node;
- migrate from the standard location and a custom `WORK_DIR`;
- prove the destination remains network-dark before activation;
- unlock, verify continuity, activate, and prove the original source cannot run
  concurrently;
- create an intentionally stale clone and prove discovery blocks it before any
  peer connection;
- capture responsive screenshots using synthetic identifiers only.

### Owner-supervised release acceptance

The final candidate may inspect real directory metadata only after a separate
approval. It must not start a cloned production LDK node, create an invoice,
probe, HTLC, payment, swap, rebalance, or channel operation.

Acceptance must independently read back:

- exact source and target paths;
- release commit and embedded hashes;
- backup and evidence hashes;
- database integrity;
- wallet/node identity;
- channel identities and stored state;
- balances and history;
- NWC apps and permissions;
- quarantine enforcement and zero network initialization;
- responsive UI and recovery instructions.

## Release gates

Checked boxes below are supported by the working-tree evidence recorded on
2026-10-16. A checked engineering gate does not authorize any later owner gate.

- [x] Plan reviewed against the current code and incident evidence.
- [x] Discovery/preflight and custom-path handling implemented.
- [x] Duplicate/stale-state detection implemented fail-closed.
- [x] Process exclusion and quiescence implemented, including a final
      activation-time source-process check.
- [x] Persistent network quarantine implemented below the UI layer.
- [x] Guided UI implemented and responsive.
- [x] Automated matrix passes from a clean detached checkout.
- [x] Synthetic/regtest migration and stale-clone tests pass.
- [x] Security review updated; reachable vulnerability and secret scans pass.
- [x] Code review has no unresolved safety-critical findings.
- [ ] Owner-supervised non-paying acceptance passes.
- [x] Migration documentation and screenshots are complete.
- [ ] Exact release commit is approved for an immutable tag.
- [ ] Rebuilt artifacts and manifest are independently verified.
- [ ] Owner approves the PGP signature for the exact manifest hash.
- [ ] Owner separately approves GitHub and portal publication.
- [ ] Public URLs and fresh downloads are read back and verified.
- [ ] Upstream PR and promotional posts receive separate approval.

## 2026-10-16 owner acceptance attempt

The owner approved a non-paying offline acceptance at exact commit
`e2ead0a0497a5e2b94d0fe5ff0da47593f8680c7`. The universal candidate embedded
that commit and the reviewed LDK digest. The live application was stopped
cleanly, two source manifests matched, and a permission-restricted isolated
source copy matched the stopped source byte-for-byte. The candidate ran with
external networking denied; independent socket inspection observed only the
two loopback ownership locks.

The real guided UI discovered and validated the isolated embedded-LDK wallet,
then failed closed before staging. Grynvault's normal first-run startup had
created a valid but unconfigured `nwc.db` in the destination. The competing-copy
scanner treated that destination as an unverifiable wallet, even though it was
the pristine target the importer was supposed to replace. No migration control,
recovery copy, staged wallet, invoice, probe, HTLC, payment, swap, rebalance, or
channel operation was created. The isolated source remained byte-identical and
the untouched canonical application was restarted and reconnected.

The owner gate therefore remains unchecked for that commit. The follow-up fix
classifies the target separately: a fresh service database with no setup
identity, LDK storage, app permissions, or transaction history is permitted,
while a corrupt database, partial setup identity, LDK storage, or prior wallet
activity remains fail-closed. It also excludes the already-classified target
from the competing-copy scan. Regression coverage must pass before a new exact
commit is proposed for owner-supervised acceptance.

## 2026-10-17 owner acceptance attempt

The owner approved a new non-paying offline acceptance at exact commit
`ffe237cd90d8bdcdf8038a1d66c154faac9ac487`. Fresh application and wallet-data
copies matched their source manifests byte-for-byte. The real guided UI accepted
the demonstrably pristine target, discovered the isolated source, created the
recovery evidence, and entered startup-level network quarantine. With external
networking denied, offline verification passed for wallet identity, stored
channels, balances, history, NWC apps and permissions, database integrity,
backup state, and embedded source/LDK hashes. Independent socket inspection saw
only the local ownership listeners before activation.

The activation control then correctly reached LDK startup and failed because
the acceptance sandbox denied every external connection. No invoice, probe,
HTLC, payment, swap, rebalance, or channel operation was created. The imported
copy remained quarantined and the isolated source manifest remained unchanged.
The untouched canonical application was restarted and reconnected.

The attempt also exposed a retry-state defect: after that expected network
failure, the UI retained the verified continuity summary and activation form,
but a retry was rejected with the misleading message to complete verification.
The persisted quarantine record already contained the successful verification;
its phase had advanced to `activation_failed`. The successor must preserve this
verification, clearly show the activation failure and offline state, clear the
password field, and permit a retry only after repeating the password,
continuity, node-identity, and source-process checks. The exact-commit owner gate
therefore remains unchecked and requires a successor commit plus a fresh owner
approval.

## 2026-10-17 activation-friction acceptance finding

The owner approved another non-paying offline acceptance at exact commit
`31d2c3b807be5d8a8fa69f6caec8a80bca7a480b`. The empty-workdir check,
duplicate-wallet block, isolated source selection, recovery snapshot, staged
copy, restart quarantine, and offline wallet continuity verification all
passed. The copied wallet confirmed the expected identity, two stored channels,
history, NWC apps and permissions, backup state, and embedded release hashes.

During the owner review, the activation action was selected unintentionally.
The process-level network sandbox denied every external connection, LDK startup
failed with `ConnectionFailed: Network connection closed`, and the disposable
copy remained in `activation_failed`. No invoice, probe, HTLC, payment, swap,
rebalance, or channel operation succeeded. The canonical wallet and isolated
source still matched their pre-test durable-file manifests; the disposable
process was stopped and the untouched installed wallet restarted and
reconnected.

The containment worked, but the human-safety gate did not. A full node-ID field
and password were insufficient protection in the surrounding workflow. The
successor must separate offline verification from activation, clear the prior
password, require an exact typed network-and-wallet confirmation plus explicit
acknowledgement, enforce a short verification lifetime in the backend, and
require re-verification after every failed activation. The exact-commit owner
gate remains unchecked pending a new build and complete rerun.

## Stop conditions

Stop and preserve evidence if any of the following occurs:

- another copy of the same node may be newer or active;
- the source changes at any point;
- exclusive ownership cannot be established;
- database integrity, identity, continuity, or hash checks differ;
- any network socket or peer initialization occurs while quarantined;
- the UI falls through to ordinary wallet startup after a failed migration;
- a test would require placing a cloned production node online;
- the work reaches tagging, signing, publication, or value movement without the
  corresponding explicit owner approval.

## Completion definition

The guided importer is complete only when a new user can install Grynvault,
accept the automatically detected Alby directory or choose a custom work
directory, migrate through the guided flow, verify the same wallet while
network-quarantined, deliberately activate it, and see the expected node and
channels—with recovery evidence available and the former source prevented from
running concurrently.

Copy success, a green unit test, a working folder picker, a build, or an empty
fresh-wallet launch is not completion.

## 2026-10-16 implementation evidence

The current working tree adds a versioned guided-migration state machine,
quiescent-source snapshots, recovery and staging hashes, known-copy conflict
detection, crash-resume checkpoints, startup-level quarantine, offline LDK
continuity inspection, final re-verification, explicit activation, and a
restart-safe archive/reset path for blocked imports. It also keeps routine
database cleanup paused while quarantine is active. A control-only blocked
state remains under the backend route quarantine rather than relying on UI
navigation.

Disposable regtest acceptance used three real LDK nodes and a native regtest
Bitcoin node. The migrating wallet opened two 500,000-sat channels, recorded
both outgoing and incoming payment history, migrated from the standard macOS
container and a custom path containing spaces and Unicode, remained
network-dark before activation, then reconnected after deliberate activation.
A matching stale clone was blocked before process or network checks. No
production wallet, channel, or funds were used.

The responsive documentation fixtures are rendered by the real migration
component and pinned in `MIGRATION.md`. The current checks and their exact scope
are recorded in `SECURITY_REVIEW.md`. The next boundary remains the documented
owner-supervised non-paying acceptance; tagging, signing, publication, and any
Lightning value action remain closed.

The follow-up matrix audit added explicit macOS local-filesystem enforcement,
truthful desktop checkpoint events, and direct regressions for permission
failures, unrelated wallet copies, validation-created sidecars, interrupted
activation, control-only blocked startup, and wallet-identity mismatch. Custom
paths on locally attached volumes remain supported; network-mounted source or
target folders now fail before staging. The transient copy indicator is
replaced after relaunch by the persisted control/quarantine phase, preserving
the fail-closed restart behavior.

# Installation, migration, and rollback

Status: **guided migration redesign required before public release**.

Seamless guided migration is a core product requirement. The implementation
and release acceptance contract is
[`docs/grynvault-release/GUIDED_MIGRATION_PLAN.md`](docs/grynvault-release/GUIDED_MIGRATION_PLAN.md).
The current one-step importer must not be treated as release-ready.

Grynvault Rebalance is deliberately isolated from Alby Hub.

## Clean installation

- bundle identifier: `com.grynvault.rebalance`
- application name: `Grynvault Rebalance`
- default data directory name: `grynvault-rebalance`
- desktop single-instance lock port: `21421`
- updater: Grynvault compatibility and release service, never Alby's official
  updater

On macOS, sandboxing can remap the physical container path. The invariant is
that the application uses the `grynvault-rebalance` directory under its own
data home and never silently opens the default `albyhub` directory.

## Importing an existing wallet

Discovery may be automatic, but migration and network activation are always
explicit. The supported experience must find the normal Alby Hub location and
also allow the owner to choose an arbitrary custom `WORK_DIR`.

Before importing:

1. Quit the source Alby Hub cleanly.
2. Create and independently verify a current backup and record its SHA-256.
3. Preserve the known-good application package and a database copy outside both
   application directories.
4. Check known Alby and Grynvault locations for a newer or active copy of the
   same wallet. A byte-perfect historical backup is not a safe live source.
5. Import only through the guided flow into a clean Grynvault data directory.
6. Restart the imported copy in network quarantine. Do not initialize LDK peer
   networking merely because the copy and database checks passed.
7. Verify unlock, wallet/node identity, every stored channel identity, balances,
   transaction history, NWC apps and permissions, backup state, embedded
   source/library hashes, and responsive UI.
8. Activate networking only after those checks pass and the owner confirms the
   exact source, identity, and channel summary locally.

Installation acceptance is non-paying. Do not create an invoice, probe, HTLC,
payment, swap, rebalance, channel open, or channel close merely to test the
package.

## Rollback

If any identity, balance, channel, permission, or database check differs:

1. Stop Grynvault Rebalance without retrying a payment.
2. Preserve its logs and data directory for local diagnosis without sharing
   secrets.
3. Restore the known-good Alby Hub package and its independently preserved data
   copy.
4. Re-verify unlock, sync, channels, balances, history, apps, and backups.

Never run both copies against the same live LDK storage directory, and never
start two separate copies of the same LDK node state online. A stale clone can
cause peers to defensively force-close channels even when the copy itself is
byte-perfect.

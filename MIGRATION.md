# Installation, migration, and rollback

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

Import is never automatic. Before importing:

1. Quit the source Alby Hub cleanly.
2. Create and independently verify a current backup and record its SHA-256.
3. Preserve the known-good application package and a database copy outside both
   application directories.
4. Import only through the documented backup/migration flow into a clean
   Grynvault data directory.
5. Verify unlock, database integrity, sync, every channel identity and status,
   balances, transaction history, NWC apps and permissions, backup state,
   embedded source/library hashes, and responsive UI before considering the
   migration accepted.

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

Never run both copies against the same live LDK storage directory.

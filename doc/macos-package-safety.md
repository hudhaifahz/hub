# macOS package safety

The desktop wallet's data location is part of its security boundary. A macOS
build can have a valid signature and still open a different, empty wallet
profile if its sandbox entitlement or bundle identity changes. A build can also
silently load a developer-machine copy of the Lightning library if an absolute
runtime search path remains in the executable.

## Required invariants

Before an application is installed or archived, verify all of the following:

- the application has a valid deep code signature;
- the bundle identifier is the intended identifier for that distribution;
- the app sandbox, client/server network, Downloads, and user-selected-file
  entitlements are present;
- the executable has only `@executable_path/../Frameworks` as its runtime search
  path;
- the executable links `libldk_node.dylib` through `@rpath`;
- the executable and embedded Lightning library contain the same architectures;
- a public candidate has a non-ad-hoc signature and a signing Team Identifier.

Run the static gate with the identity expected by that distribution:

```sh
EXPECTED_BUNDLE_ID='com.example.Product' \
REQUIRE_NONADHOC_SIGNATURE=1 \
./scripts/verify-darwin-app.sh './build/bin/Product.app'
```

The release workflow runs the same gate after signing. It fails before creating
the disk image if any invariant is missing.

Do not distribute a raw Wails build. The normal build output can be correctly
compiled yet still lack the sandbox entitlements and retain a developer-machine
runtime path. Only an archive created by the guarded packaging command, or a
release artifact that passed the workflow gate, is eligible for testing.

## Local test packages

Ad-hoc signing is only for an owner's local test build. The command makes this
classification explicit and requires `owner-test` in the filename:

```sh
PACKAGE_CLASS=owner-test \
EXPECTED_BUNDLE_ID='com.example.Product' \
CODESIGN_IDENTITY=- \
ALLOW_ADHOC=1 \
./scripts/package-darwin-app.sh './build/bin/Product.app' \
  './build/out/Product-owner-test.zip'
```

The package command refuses to overwrite prior evidence and writes a SHA-256
file plus a manifest containing the bundle identity, architectures, executable
hash, embedded Lightning-library hash, archive size, and archive hash.

## Public release candidates

A Developer ID signed archive is still only a candidate. Use
`PACKAGE_CLASS=developer-id-candidate`, an independent distribution identity,
and a filename containing `candidate`. Do not publish it until the final package
is notarized, the ticket is stapled, Gatekeeper accepts a fresh download, and
the published manifest and signature verify.

An independently maintained fork must use its own product name, icon, bundle
identifier, signing identity, update channel, and data directory. Importing an
existing wallet is a separate, backed-up owner action; the package must never
silently reuse or replace the official application's data container.

## Runtime acceptance

Static checks prevent the known packaging failure, but they do not prove wallet
continuity. After an owner-approved installation or migration, verify without
making a payment:

1. the process opens the expected data directory;
2. the process loads the embedded `libldk_node.dylib` from the application;
3. database integrity passes and expected configuration counts remain present;
4. unlock and chain sync complete;
5. channels, balances, transaction history, connected apps, permissions, and
   backup status match the pre-install record;
6. the exact-channel dialog remains usable at narrow and wide window sizes.

If chain sync fails, preserve the wallet and report the concrete chain-source or
network error. An empty channel list is not evidence that channels were lost.

## Failure interpretation

The startup sequence may report a network connection failure when DNS or the
configured chain source is temporarily unavailable. Keep the wallet locked,
retry without changing its data directory, and distinguish a later successful
sync from a packaging defect. A package defect is indicated by the process
opening the wrong data directory or loading an LDK library outside the app;
those conditions require immediate rollback and quarantine of the package.

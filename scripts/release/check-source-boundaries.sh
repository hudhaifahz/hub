#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

require_literal() {
  local literal="$1"
  local path="$2"
  if ! grep -F -q -- "$literal" "$path"; then
    echo "required release boundary is missing from $path: $literal" >&2
    exit 1
  fi
}

require_literal '"name": "Grynvault Rebalance"' wails.json
require_literal '"outputfilename": "Grynvault Rebalance"' wails.json
require_literal '<string>com.grynvault.rebalance</string>' build/darwin/Info.plist
require_literal '<string>15.4.0</string>' build/darwin/Info.plist
require_literal 'export MACOSX_DEPLOYMENT_TARGET=15.4' scripts/release/build-macos.sh
require_literal 'acceptance build requires a clean checkout' scripts/release/build-macos-acceptance.sh
require_literal 'version.SourceCommit=${source_commit}' scripts/release/build-macos-acceptance.sh
require_literal './scripts/release/inspect-macos.sh' scripts/release/build-macos-acceptance.sh
require_literal 'must not be uploaded or published' scripts/release/build-macos-acceptance.sh
require_literal '<string>com.grynvault.rebalance.dev</string>' build/darwin/Info.dev.plist
require_literal 'const defaultWorkdirName = "grynvault-rebalance"' service/workdir.go
require_literal '//go:build wails && !bindings' main_wails.go
require_literal 'prepareReleaseAcceptanceLaunch(os.Args)' main_wails.go
require_literal 'verifyReleaseAcceptanceNetworkDenied()' release_acceptance_launch.go
require_literal 'verifyReleaseAcceptanceBoundaryPlatform()' release_acceptance_launch_darwin.go
require_literal '(deny network-outbound (remote ip))' build/darwin/release-acceptance.sb
require_literal '(require-not (local tcp "*:21420"))' build/darwin/release-acceptance.sb
require_literal '(require-not (local tcp "*:21421"))' build/darwin/release-acceptance.sb
require_literal 'codesign --force --sign - "$app_path"' scripts/release/build-macos-acceptance.sh
require_literal '--require-acceptance-adhoc' scripts/release/build-macos-acceptance.sh
require_literal '/usr/bin/nohup /usr/bin/sandbox-exec' scripts/release/launch-macos-acceptance.sh
require_literal '--grynvault-release-acceptance-network-dark' scripts/release/launch-macos-acceptance.sh
require_literal '127.0.0.1:21421' desktop_locks.go
require_literal '127.0.0.1:21420' desktop_locks.go
require_literal '//go:build wails && bindings' main_wails_bindings.go
require_literal 'hubwails.NewBindingsApp()' main_wails_bindings.go
require_literal '//go:build bindings' wails/wails_app_bindings.go
require_literal 'https://rebalance.grynvault.com/' frontend/src/lib/grynvault.ts
require_literal 'VITE_GRYNVAULT_SOURCE_URL' frontend/src/lib/grynvault.ts
require_literal 'experimental, independently maintained fork' frontend/src/components/GrynvaultForkNotice.tsx

if grep -nE 'com\.getalby\.' build/darwin/Info.plist build/darwin/Info.dev.plist wails.json; then
  echo "official Alby bundle identity remains in Grynvault identity files" >&2
  exit 1
fi
if grep -Fq -- '--entitlements' scripts/release/build-macos-acceptance.sh; then
  echo "acceptance process wrapper unexpectedly retains app entitlements" >&2
  exit 1
fi
if grep -R -n -F -- 'https://getalby.com/update/hub' frontend/src; then
  echo "official Alby updater remains reachable from the Grynvault frontend" >&2
  exit 1
fi
if grep -nE 'service\.NewService|net\.Listen|21421' main_wails_bindings.go wails/wails_app_bindings.go; then
  echo "bindings-only build may start a service, open a wallet, or claim the desktop port" >&2
  exit 1
fi
if grep -R -n -F -- 'Developer ID Application: Alby Inc.' .github/workflows build; then
  echo "Alby signing identity remains in an active release configuration" >&2
  exit 1
fi
if grep -R -nE 'gh release create|docker-build-push|packages: write' .github/workflows; then
  echo "an automatic publication path remains in the release branch" >&2
  exit 1
fi

echo "release source boundaries passed"

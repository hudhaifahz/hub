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
require_literal '<string>com.grynvault.rebalance.dev</string>' build/darwin/Info.dev.plist
require_literal 'const defaultWorkdirName = "grynvault-rebalance"' service/workdir.go
require_literal '127.0.0.1:21421' main_wails.go
require_literal 'https://rebalance.grynvault.com/' frontend/src/lib/grynvault.ts
require_literal 'VITE_GRYNVAULT_SOURCE_URL' frontend/src/lib/grynvault.ts
require_literal 'experimental, independently maintained fork' frontend/src/components/GrynvaultForkNotice.tsx

if grep -nE 'com\.getalby\.' build/darwin/Info.plist build/darwin/Info.dev.plist wails.json; then
  echo "official Alby bundle identity remains in Grynvault identity files" >&2
  exit 1
fi
if grep -R -n -F -- 'https://getalby.com/update/hub' frontend/src; then
  echo "official Alby updater remains reachable from the Grynvault frontend" >&2
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

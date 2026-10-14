#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <output-file>" >&2
  exit 64
fi

output_file="$1"
repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

expected_syft_version="1.41.1"
if ! command -v syft >/dev/null 2>&1; then
  echo "syft ${expected_syft_version} is required" >&2
  exit 1
fi
actual_syft_version="$(syft version | awk -F': *' '$1 == "Version" { print $2; exit }')"
if [[ "$actual_syft_version" != "$expected_syft_version" ]]; then
  echo "expected syft ${expected_syft_version}, found ${actual_syft_version:-unknown}" >&2
  exit 1
fi

output_directory="$(dirname "$output_file")"
mkdir -p "$output_directory"
temporary_output="$(mktemp "${output_directory}/.sbom.XXXXXX")"
normalized_output="$(mktemp "${output_directory}/.sbom-normalized.XXXXXX")"
syft_cache="$(mktemp -d "${TMPDIR:-/tmp}/grynvault-syft-cache.XXXXXX")"
cleanup() {
  rm -f "$temporary_output" "$normalized_output"
  rm -rf "$syft_cache"
}
trap cleanup EXIT

XDG_CACHE_HOME="$syft_cache" SYFT_CHECK_FOR_APP_UPDATE=false syft scan dir:. \
  --quiet \
  --source-name 'Grynvault Rebalance' \
  --source-version "$(git rev-parse HEAD)" \
  --override-default-catalogers go-module-file-cataloger,javascript-lock-cataloger \
  --exclude './.git/**' \
  --exclude './build/**' \
  --exclude './frontend/node_modules/**' \
  --output "cyclonedx-json=${temporary_output}"

# Syft 1.41.1 records the two scanned lockfiles as absolute file-component
# names. Normalize only those file names so the published SBOM does not expose
# or depend on the CI checkout path; package bom-refs and dependency edges stay
# untouched.
jq --arg root "${repo_root}/" '
  (.components[]? |
    select(.type == "file" and (.name | startswith($root))) |
    .name) |= ltrimstr($root)
' "$temporary_output" >"$normalized_output"

jq -e '
  .bomFormat == "CycloneDX" and
  (.specVersion | test("^1\\.[0-9]+$")) and
  ([.components[]? | select((.purl? // "") | startswith("pkg:golang/"))] | length > 0) and
  ([.components[]? | select((.purl? // "") | startswith("pkg:npm/"))] | length > 0)
' "$normalized_output" >/dev/null

if rg -F -q "$repo_root" "$normalized_output"; then
  echo "SBOM contains an absolute local checkout path" >&2
  exit 1
fi

mv "$normalized_output" "$output_file"
rm -f "$temporary_output"
rm -rf "$syft_cache"
trap - EXIT
echo "CycloneDX SBOM written to $output_file"

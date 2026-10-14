#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 4 ]]; then
  echo "usage: $0 <app-path> <immutable-tag> <expected-source-commit> <output-directory>" >&2
  exit 64
fi

app_path="$1"
release_tag="$2"
expected_source_commit="$3"
output_directory="$4"
repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ ! "$release_tag" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
  echo "release tag contains unsupported characters" >&2
  exit 1
fi
if [[ ! "$expected_source_commit" =~ ^[a-f0-9]{40}$ ]]; then
  echo "expected source commit must be a 40-character lowercase Git object ID" >&2
  exit 1
fi
if [[ -n "$(git status --porcelain=v1 --untracked-files=no)" ]]; then
  echo "packaging requires a clean tracked checkout" >&2
  exit 1
fi

source_commit="$(git rev-parse HEAD)"
tag_ref="refs/tags/${release_tag}"
tag_commit="$(git rev-parse --verify "${tag_ref}^{commit}" 2>/dev/null || true)"
if [[ -z "$tag_commit" || "$tag_commit" != "$source_commit" || "$source_commit" != "$expected_source_commit" ]]; then
  echo "release tag must exist locally and resolve to HEAD; expected source commit must also match" >&2
  exit 1
fi

app_path="$(cd "$(dirname "$app_path")" && pwd)/$(basename "$app_path")"
output_directory="$(mkdir -p "$output_directory" && cd "$output_directory" && pwd)"
ldk_path="${app_path}/Contents/Frameworks/libldk_node.dylib"
artifact_base="Grynvault-Rebalance-${release_tag}-macOS-universal-UNSIGNED-EXPERIMENTAL"
dmg_path="${output_directory}/${artifact_base}.dmg"
checksum_path="${dmg_path}.sha256"
evidence_path="${output_directory}/${artifact_base}.gatekeeper.txt"
staging_directory="$(mktemp -d "${TMPDIR:-/tmp}/grynvault-rebalance-unsigned-stage.XXXXXX")"
mount_directory="$(mktemp -d "${TMPDIR:-/tmp}/grynvault-rebalance-unsigned-mount.XXXXXX")"
mounted=0

cleanup() {
  if [[ "$mounted" -eq 1 ]]; then
    hdiutil detach "$mount_directory" -quiet >/dev/null 2>&1 || true
  fi
  rm -rf "$staging_directory" "$mount_directory"
}
trap cleanup EXIT

./scripts/release/inspect-macos.sh "$app_path" "$release_tag" "$expected_source_commit"

# Ad-hoc signatures provide local bundle-integrity verification only. They do
# not identify Grynvault, confer Apple trust, or replace notarization.
codesign --force --sign - "$ldk_path"
codesign --force --sign - --entitlements build/darwin/entitlements.plist "$app_path"
./scripts/release/inspect-macos.sh "$app_path" "$release_tag" "$expected_source_commit" --require-adhoc

cp -R "$app_path" "$staging_directory/"
cat >"$staging_directory/README-UNSIGNED.txt" <<EOF
GRYNVAULT REBALANCE — UNSIGNED EXPERIMENTAL macOS BUILD

This application is intentionally NOT signed with an Apple Developer ID and is
NOT notarized by Apple. macOS Gatekeeper is expected to warn or block it. An
ad-hoc signature checks bundle integrity only; it does not identify a publisher.

Release tag: ${release_tag}
Source commit: ${expected_source_commit}
Source: https://github.com/hudhaifahz/hub/commit/${expected_source_commit}

Before opening this application:
1. Compare the DMG SHA-256 with the PGP-signed release manifest.
2. Confirm that the manifest fingerprint matches the separately published owner key.
3. Prefer building from the immutable source tag if you cannot accept unsigned software.
4. Back up the existing wallet before any explicit migration or import.

Do not disable Gatekeeper globally. If you choose to proceed after verification,
use macOS's per-application Open/Open Anyway flow and accept the risk explicitly.
EOF

rm -f "$dmg_path" "$checksum_path" "$evidence_path"
hdiutil create \
  -volname "Grynvault Rebalance UNSIGNED" \
  -srcfolder "$staging_directory" \
  -ov \
  -format UDZO \
  "$dmg_path"
hdiutil verify "$dmg_path"
hdiutil attach -readonly -nobrowse -mountpoint "$mount_directory" "$dmg_path" >/dev/null
mounted=1
./scripts/release/inspect-macos.sh \
  "$mount_directory/Grynvault Rebalance.app" \
  "$release_tag" \
  "$expected_source_commit" \
  --require-adhoc
grep -Fq 'intentionally NOT signed with an Apple Developer ID' "$mount_directory/README-UNSIGNED.txt"
hdiutil detach "$mount_directory" -quiet
mounted=0

set +e
(
  cd "$output_directory"
  spctl --assess --type open --context context:primary-signature --verbose=4 "$(basename "$dmg_path")"
) >"$evidence_path" 2>&1
spctl_status=$?
set -e
{
  echo "spctl_exit_status=${spctl_status}"
  echo "expected_result=not accepted as a notarized Developer ID distribution"
} >>"$evidence_path"
if [[ "$spctl_status" -eq 0 ]]; then
  echo "Gatekeeper unexpectedly accepted the unsigned experimental disk image" >&2
  exit 1
fi

(
  cd "$output_directory"
  shasum -a 256 "$(basename "$dmg_path")" >"$(basename "$checksum_path")"
)

echo "unsigned experimental disk image: $dmg_path"
echo "checksum: $checksum_path"
echo "Gatekeeper evidence: $evidence_path"
echo "WARNING: this artifact has no Apple publisher identity and is not notarized"

#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 3 || $# -gt 4 ]]; then
  echo "usage: $0 <app-path> <expected-tag> <expected-source-commit> [--require-adhoc|--require-acceptance-adhoc|--require-developer-id]" >&2
  exit 64
fi

app_path="$1"
expected_tag="$2"
expected_source_commit="$3"
signing_mode="${4:-}"
expected_bundle_id="com.grynvault.rebalance"
expected_macos_minimum="15.4.0"
expected_ldk_sha="be8ee3aa741b9ef7f017dda80e90e523b4fd7ae6cc7b4d18a7c2314495c7ef9a"
executable_path="${app_path}/Contents/MacOS/Grynvault Rebalance"
ldk_path="${app_path}/Contents/Frameworks/libldk_node.dylib"
plist_path="${app_path}/Contents/Info.plist"
inspection_directory="$(mktemp -d "${TMPDIR:-/tmp}/grynvault-rebalance-inspection.XXXXXX")"
trap 'rm -rf "$inspection_directory"' EXIT
strings_path="${inspection_directory}/executable-strings.txt"
entitlements_path="${inspection_directory}/entitlements.plist"

if [[ ! "$expected_source_commit" =~ ^[a-f0-9]{40}$ ]]; then
  echo "expected source commit must be a 40-character lowercase Git object ID" >&2
  exit 1
fi
if [[ -n "$signing_mode" && "$signing_mode" != "--require-adhoc" && "$signing_mode" != "--require-acceptance-adhoc" && "$signing_mode" != "--require-developer-id" ]]; then
  echo "unsupported signing mode: $signing_mode" >&2
  exit 64
fi

for required_path in "$app_path" "$executable_path" "$ldk_path" "$plist_path"; do
  if [[ ! -e "$required_path" ]]; then
    echo "missing required bundle path: $required_path" >&2
    exit 1
  fi
done

bundle_id="$(plutil -extract CFBundleIdentifier raw -o - "$plist_path")"
bundle_name="$(plutil -extract CFBundleName raw -o - "$plist_path")"
bundle_macos_minimum="$(plutil -extract LSMinimumSystemVersion raw -o - "$plist_path")"
if [[ "$bundle_id" != "$expected_bundle_id" ]]; then
  echo "unexpected bundle identifier: $bundle_id" >&2
  exit 1
fi
if [[ "$bundle_name" != "Grynvault Rebalance" ]]; then
  echo "unexpected bundle name: $bundle_name" >&2
  exit 1
fi
if [[ "$bundle_macos_minimum" != "$expected_macos_minimum" ]]; then
  echo "unexpected minimum macOS version: $bundle_macos_minimum" >&2
  exit 1
fi

for binary_path in "$executable_path" "$ldk_path"; do
  architectures="$(lipo -archs "$binary_path")"
  if [[ "$architectures" != *arm64* || "$architectures" != *x86_64* ]]; then
    echo "bundle is not universal: $binary_path ($architectures)" >&2
    exit 1
  fi
done

for architecture in x86_64 arm64; do
  executable_macos_minimum="$(
    vtool -show-build -arch "$architecture" "$executable_path" |
      awk '
        $1 == "cmd" && $2 == "LC_VERSION_MIN_MACOSX" { legacy = 1; next }
        legacy && $1 == "version" { print $2; exit }
        $1 == "minos" { print $2; exit }
      '
  )"
  if [[ "$executable_macos_minimum" != "15.4" ]]; then
    echo "unexpected ${architecture} executable deployment target: ${executable_macos_minimum:-missing}" >&2
    exit 1
  fi
done

if ! otool -L "$executable_path" | grep -q '@rpath/libldk_node.dylib'; then
  echo "main executable does not use the expected LDK install name" >&2
  exit 1
fi
rpaths="$(otool -l "$executable_path" | awk '$1 == "cmd" && $2 == "LC_RPATH" { getline; getline; print $2 }' | sort -u)"
if [[ "$rpaths" != '@executable_path/../Frameworks' ]]; then
  echo "main executable has an unexpected rpath set: ${rpaths:-none}" >&2
  exit 1
fi
ldk_install_name="$(otool -D "$ldk_path" | awk '$1 ~ /^@/ { print $1 }' | sort -u)"
if [[ "$ldk_install_name" != '@rpath/libldk_node.dylib' ]]; then
  echo "embedded LDK has an unexpected install name: $ldk_install_name" >&2
  exit 1
fi

actual_ldk_sha="$(shasum -a 256 "$ldk_path" | awk '{print $1}')"
if [[ -z "$signing_mode" && "$actual_ldk_sha" != "$expected_ldk_sha" ]]; then
  echo "embedded LDK digest mismatch: $actual_ldk_sha" >&2
  exit 1
fi

strings "$executable_path" >"$strings_path"
if ! grep -Fq "$expected_tag" "$strings_path"; then
  echo "expected release tag is not embedded in the executable" >&2
  exit 1
fi
if ! grep -Fq "$expected_source_commit" "$strings_path"; then
  echo "expected source commit is not embedded in the executable" >&2
  exit 1
fi
if ! grep -Fq "$expected_ldk_sha" "$strings_path"; then
  echo "expected LDK digest is not embedded in the executable" >&2
  exit 1
fi
if ! grep -Fq "grynvault-rebalance" "$strings_path"; then
  echo "independent data-directory identity is not embedded" >&2
  exit 1
fi
if ! rg -a -F -q "https://github.com/hudhaifahz/hub/commit/${expected_source_commit}" "$app_path"; then
  echo "immutable source URL is not embedded in the bundle" >&2
  exit 1
fi

if rg -a -l 'com\.getalby\.Alby-?Hub' "$plist_path" >/dev/null || \
  rg -a -l 'https://getalby\.com/update/hub' "$app_path" >/dev/null; then
  echo "bundle retains an official Alby bundle or updater identity" >&2
  exit 1
fi

if [[ -n "$signing_mode" ]]; then
  codesign --verify --strict --verbose=2 "$ldk_path"
  codesign --verify --deep --strict --verbose=2 "$app_path"
  if [[ "$signing_mode" == "--require-acceptance-adhoc" ]]; then
    codesign -d --entitlements - --xml "$app_path" >"$entitlements_path" 2>/dev/null || true
    if [[ -s "$entitlements_path" ]]; then
      plutil -lint "$entitlements_path" >/dev/null
      for entitlement in \
        com.apple.security.app-sandbox \
        com.apple.security.network.client \
        com.apple.security.network.server; do
        entitlement_key_path="${entitlement//./\\.}"
        if [[ "$(plutil -extract "$entitlement_key_path" raw -o - "$entitlements_path" 2>/dev/null || true)" == "true" ]]; then
          echo "acceptance process wrapper unexpectedly retains entitlement: $entitlement" >&2
          exit 1
        fi
      done
    fi
  else
    codesign -d --entitlements - --xml "$app_path" >"$entitlements_path" 2>/dev/null
    plutil -lint "$entitlements_path" >/dev/null
    for entitlement in \
      com.apple.security.app-sandbox \
      com.apple.security.network.client \
      com.apple.security.network.server \
      com.apple.security.files.downloads.read-write \
      com.apple.security.files.user-selected.read-write; do
      entitlement_key_path="${entitlement//./\\.}"
      if [[ "$(plutil -extract "$entitlement_key_path" raw -o - "$entitlements_path" 2>/dev/null || true)" != "true" ]]; then
        echo "required entitlement is missing or false: $entitlement" >&2
        exit 1
      fi
    done
  fi
  signing_details="$(codesign -dv --verbose=4 "$app_path" 2>&1)"
  ldk_signing_details="$(codesign -dv --verbose=4 "$ldk_path" 2>&1)"
  if [[ "$signing_mode" == "--require-adhoc" || "$signing_mode" == "--require-acceptance-adhoc" ]]; then
    for details in "$signing_details" "$ldk_signing_details"; do
      if ! grep -Fq 'Signature=adhoc' <<<"$details"; then
        echo "unsigned experimental candidate is not ad-hoc signed" >&2
        exit 1
      fi
      if grep -Fq 'Authority=' <<<"$details"; then
        echo "unsigned experimental candidate unexpectedly has a certificate authority" >&2
        exit 1
      fi
      if ! grep -Fq 'TeamIdentifier=not set' <<<"$details"; then
        echo "unsigned experimental candidate unexpectedly has a Team Identifier" >&2
        exit 1
      fi
    done
  else
    if grep -Fq 'Authority=Developer ID Application: Alby Inc.' <<<"$signing_details" ||
      grep -Fq 'Authority=Developer ID Application: Alby Inc.' <<<"$ldk_signing_details"; then
      echo "public candidate must not use Alby's signing identity" >&2
      exit 1
    fi
    if ! grep -Fq 'Authority=Developer ID Application:' <<<"$signing_details" ||
      ! grep -Fq 'Authority=Developer ID Application:' <<<"$ldk_signing_details"; then
      echo "public candidate application or LDK library is missing a Developer ID Application signature" >&2
      exit 1
    fi
    app_team_identifier="$(awk -F= '$1 == "TeamIdentifier" { print $2; exit }' <<<"$signing_details")"
    ldk_team_identifier="$(awk -F= '$1 == "TeamIdentifier" { print $2; exit }' <<<"$ldk_signing_details")"
    if [[ -z "$app_team_identifier" || "$app_team_identifier" == "not set" || "$app_team_identifier" != "$ldk_team_identifier" ]]; then
      echo "public candidate application and LDK library do not share one signing Team Identifier" >&2
      exit 1
    fi
  fi
fi

echo "bundle identifier: $bundle_id"
echo "bundle name: $bundle_name"
echo "minimum macOS version: $bundle_macos_minimum"
echo "executable deployment target: 15.4"
echo "executable architectures: $(lipo -archs "$executable_path")"
echo "LDK architectures: $(lipo -archs "$ldk_path")"
if [[ -n "$signing_mode" ]]; then
  echo "signed LDK SHA-256: $actual_ldk_sha"
  echo "reviewed unsigned LDK SHA-256: $expected_ldk_sha (verified before signing)"
else
  echo "LDK SHA-256: $actual_ldk_sha"
fi
echo "embedded tag: $expected_tag"
echo "embedded source commit: $expected_source_commit"
if [[ "$signing_mode" == "--require-adhoc" || "$signing_mode" == "--require-acceptance-adhoc" ]]; then
  echo "signing identity: ad-hoc only (no publisher identity, notarization, or Gatekeeper trust)"
fi
if [[ "$signing_mode" == "--require-acceptance-adhoc" ]]; then
  echo "acceptance process-wrapper entitlements: absent"
fi
echo "bundle inspection passed"

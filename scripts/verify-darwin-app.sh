#!/bin/sh

set -eu

usage() {
  echo "Usage: $0 <path-to-app>" >&2
  exit 2
}

[ "$#" -eq 1 ] || usage

app_path=$1
info_plist="$app_path/Contents/Info.plist"

[ -d "$app_path" ] || { echo "ERROR: app bundle not found: $app_path" >&2; exit 1; }
[ -f "$info_plist" ] || { echo "ERROR: Info.plist not found" >&2; exit 1; }

executable_name=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleExecutable' "$info_plist")
bundle_id=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleIdentifier' "$info_plist")
executable_path="$app_path/Contents/MacOS/$executable_name"
embedded_ldk="$app_path/Contents/Frameworks/libldk_node.dylib"

[ -x "$executable_path" ] || { echo "ERROR: app executable is missing or not executable" >&2; exit 1; }
[ -f "$embedded_ldk" ] || { echo "ERROR: embedded libldk_node.dylib is missing" >&2; exit 1; }

codesign --verify --deep --strict --verbose=2 "$app_path"

if [ -n "${EXPECTED_BUNDLE_ID:-}" ] && [ "$bundle_id" != "$EXPECTED_BUNDLE_ID" ]; then
  echo "ERROR: bundle identifier does not match the intended data container" >&2
  echo "Expected: $EXPECTED_BUNDLE_ID" >&2
  echo "Actual:   $bundle_id" >&2
  exit 1
fi

signature_details=$(codesign -dv --verbose=4 "$app_path" 2>&1)
if [ "${REQUIRE_NONADHOC_SIGNATURE:-}" = '1' ]; then
  printf '%s\n' "$signature_details" | grep -q '^Signature=adhoc$' && {
    echo "ERROR: public candidate is ad-hoc signed" >&2
    exit 1
  }
  team_id=$(printf '%s\n' "$signature_details" | awk -F= '$1 == "TeamIdentifier" { print $2 }')
  [ -n "$team_id" ] && [ "$team_id" != 'not set' ] || {
    echo "ERROR: public candidate has no signing TeamIdentifier" >&2
    exit 1
  }
fi

entitlements_file=$(mktemp)
trap 'rm -f "$entitlements_file"' EXIT HUP INT TERM
codesign -d --entitlements :- "$app_path" >"$entitlements_file" 2>/dev/null

for entitlement in \
  com.apple.security.app-sandbox \
  com.apple.security.network.client \
  com.apple.security.network.server \
  com.apple.security.files.downloads.read-write \
  com.apple.security.files.user-selected.read-write
do
  value=$(/usr/libexec/PlistBuddy -c "Print :$entitlement" "$entitlements_file" 2>/dev/null || true)
  [ "$value" = "true" ] || {
    echo "ERROR: required entitlement is missing or false: $entitlement" >&2
    exit 1
  }
done

rpaths=$(otool -l "$executable_path" | awk '$1 == "path" { print $2 }')
[ -n "$rpaths" ] || { echo "ERROR: executable has no LC_RPATH" >&2; exit 1; }

unique_rpaths=$(printf '%s\n' "$rpaths" | sort -u)
[ "$unique_rpaths" = '@executable_path/../Frameworks' ] || {
  echo "ERROR: executable has a non-embedded or unexpected LC_RPATH:" >&2
  printf '%s\n' "$unique_rpaths" >&2
  exit 1
}

otool -L "$executable_path" | awk 'NR > 1 { print $1 }' | grep -qx '@rpath/libldk_node.dylib' || {
  echo "ERROR: executable is not linked to @rpath/libldk_node.dylib" >&2
  exit 1
}

executable_arches=$(lipo -archs "$executable_path" | tr ' ' '\n' | sort | tr '\n' ' ')
ldk_arches=$(lipo -archs "$embedded_ldk" | tr ' ' '\n' | sort | tr '\n' ' ')
[ "$executable_arches" = "$ldk_arches" ] || {
  echo "ERROR: executable and embedded LDK architectures differ" >&2
  echo "Executable: $executable_arches" >&2
  echo "LDK: $ldk_arches" >&2
  exit 1
}

echo "Verified macOS app invariants:"
echo "- deep code signature"
echo "- bundle identifier: $bundle_id"
echo "- sandbox, network, download, and user-selected-file entitlements"
echo "- embedded-only LDK runtime path"
echo "- matching executable and LDK architectures: $executable_arches"

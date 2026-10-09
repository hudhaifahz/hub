#!/bin/sh

set -eu

usage() {
  echo "Usage: PACKAGE_CLASS=<owner-test|developer-id-candidate> EXPECTED_BUNDLE_ID='<id>' CODESIGN_IDENTITY='<identity>' $0 <path-to-app> <output-zip>" >&2
  echo "owner-test also requires CODESIGN_IDENTITY=- and ALLOW_ADHOC=1." >&2
  echo "A developer-id-candidate is not a public release until it is notarized and Gatekeeper-verified." >&2
  exit 2
}

[ "$#" -eq 2 ] || usage

app_path=$1
output_zip=$2
identity=${CODESIGN_IDENTITY:-}
package_class=${PACKAGE_CLASS:-}
expected_bundle_id=${EXPECTED_BUNDLE_ID:-}

[ -n "$identity" ] || usage
[ -n "$expected_bundle_id" ] || usage
[ "$package_class" = 'owner-test' ] || [ "$package_class" = 'developer-id-candidate' ] || usage
[ "${output_zip##*.}" = 'zip' ] || { echo "ERROR: output package must end in .zip" >&2; exit 1; }

[ -d "$app_path" ] || { echo "ERROR: app bundle not found: $app_path" >&2; exit 1; }

case "$package_class" in
  owner-test)
    [ "$identity" = '-' ] || { echo "ERROR: owner-test packages must use ad-hoc signing" >&2; exit 1; }
    [ "${ALLOW_ADHOC:-}" = '1' ] || { echo "ERROR: set ALLOW_ADHOC=1 explicitly for an owner-test package" >&2; exit 1; }
    case "$(basename -- "$output_zip")" in
      *owner-test*.zip) ;;
      *) echo "ERROR: owner-test output filename must contain owner-test" >&2; exit 1 ;;
    esac
    ;;
  developer-id-candidate)
    [ "$identity" != '-' ] || { echo "ERROR: developer-id-candidate packages cannot be ad-hoc signed" >&2; exit 1; }
    case "$(basename -- "$output_zip")" in
      *candidate*.zip) ;;
      *) echo "ERROR: developer-id-candidate output filename must contain candidate" >&2; exit 1 ;;
    esac
    ;;
esac

for output_path in "$output_zip" "$output_zip.sha256" "$output_zip.manifest"; do
  [ ! -e "$output_path" ] || { echo "ERROR: refusing to overwrite existing package output: $output_path" >&2; exit 1; }
done

info_plist="$app_path/Contents/Info.plist"
executable_name=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleExecutable' "$info_plist")
executable_path="$app_path/Contents/MacOS/$executable_name"
entitlements_path="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)/build/darwin/entitlements.plist"
embedded_rpath='@executable_path/../Frameworks'

otool -l "$executable_path" | awk '$1 == "path" { print $2 }' | sort -u | while IFS= read -r rpath; do
  if [ "$rpath" != "$embedded_rpath" ]; then
    install_name_tool -delete_rpath "$rpath" "$executable_path"
  fi
done

if ! otool -l "$executable_path" | awk '$1 == "path" { print $2 }' | grep -qx "$embedded_rpath"; then
  install_name_tool -add_rpath "$embedded_rpath" "$executable_path"
fi

if [ "$package_class" = 'owner-test' ]; then
  codesign --force --deep --sign "$identity" --entitlements "$entitlements_path" "$app_path"
else
  codesign --force --deep --timestamp --options runtime --sign "$identity" \
    --entitlements "$entitlements_path" "$app_path"
fi
if [ "$package_class" = 'developer-id-candidate' ]; then
  REQUIRE_NONADHOC_SIGNATURE=1 EXPECTED_BUNDLE_ID="$expected_bundle_id" \
    "$(dirname -- "$0")/verify-darwin-app.sh" "$app_path"
else
  EXPECTED_BUNDLE_ID="$expected_bundle_id" "$(dirname -- "$0")/verify-darwin-app.sh" "$app_path"
fi

ditto -c -k --sequesterRsrc --keepParent "$app_path" "$output_zip"
shasum -a 256 "$output_zip" >"$output_zip.sha256"
{
  echo "package_class=$package_class"
  echo "bundle_id=$expected_bundle_id"
  echo "executable_architectures=$(lipo -archs "$executable_path")"
  echo "application_executable_sha256=$(shasum -a 256 "$executable_path" | awk '{ print $1 }')"
  echo "embedded_ldk_sha256=$(shasum -a 256 "$app_path/Contents/Frameworks/libldk_node.dylib" | awk '{ print $1 }')"
  echo "archive_size_bytes=$(stat -f '%z' "$output_zip")"
  echo "archive_sha256=$(awk '{ print $1 }' "$output_zip.sha256")"
} >"$output_zip.manifest"

echo "Created: $output_zip"
echo "SHA-256 manifest: $output_zip.sha256"
echo "Build manifest: $output_zip.manifest"

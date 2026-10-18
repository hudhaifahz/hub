#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <exact-source-commit>" >&2
  exit 64
fi

expected_commit="$1"
repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ ! "$expected_commit" =~ ^[a-f0-9]{40}$ ]]; then
  echo "acceptance build requires a full 40-character lowercase commit" >&2
  exit 1
fi
if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "acceptance build requires a clean checkout" >&2
  exit 1
fi

source_commit="$(git rev-parse HEAD)"
if [[ "$source_commit" != "$expected_commit" ]]; then
  echo "acceptance build commit does not match HEAD" >&2
  exit 1
fi
if [[ "$(node --version)" != v22.* ]]; then
  echo "Node.js 22.x is required for an acceptance build" >&2
  exit 1
fi
if ! command -v wails >/dev/null 2>&1 || [[ "$(wails version | head -n 1)" != "v2.16.0" ]]; then
  echo "Wails CLI v2.16.0 is required for an acceptance build" >&2
  exit 1
fi

ldk_library_sha256="879ea591a5f9636425ce56097ca0473df3fc05e62ef94135428a78a45087adcf"
version_label="owner-acceptance-${source_commit:0:12}"
source_url="https://github.com/hudhaifahz/hub/commit/${source_commit}"
export VITE_GRYNVAULT_SOURCE_URL="$source_url"

export MACOSX_DEPLOYMENT_TARGET=15.4
export CGO_CFLAGS="${CGO_CFLAGS:-} -mmacosx-version-min=15.4"
export CGO_CXXFLAGS="${CGO_CXXFLAGS:-} -mmacosx-version-min=15.4"
export CGO_LDFLAGS="${CGO_LDFLAGS:-} -framework UniformTypeIdentifiers -mmacosx-version-min=15.4"

mkdir -p frontend/dist
touch frontend/dist/tmp
wails generate module -tags "wails" -nocolour

(
  cd frontend
  yarn install --frozen-lockfile
  yarn prepare:wails
  yarn lint:js
  yarn tsc:compile
  yarn format
  yarn prepare:http
)

if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "acceptance checks changed the exact-commit checkout" >&2
  exit 1
fi

mkdir -p build
cp appicon.png build/appicon.png

wails build \
  -clean \
  -m \
  -nosyncgomod \
  -trimpath \
  --platform darwin/universal \
  -webview2 embed \
  -o "Grynvault Rebalance" \
  -tags "wails" \
  -ldflags "-X 'github.com/getAlby/hub/version.Tag=${version_label}' -X 'github.com/getAlby/hub/version.SourceCommit=${source_commit}' -X 'github.com/getAlby/hub/version.LDKLibrarySHA256=${ldk_library_sha256}' -extldflags '-Wl,-headerpad_max_install_names -framework UniformTypeIdentifiers'"

(
  cd frontend
  yarn prepare:http
)

if [[ -n "$(git status --porcelain=v1 --untracked-files=no)" ]]; then
  echo "acceptance build changed tracked files in the exact-commit checkout" >&2
  exit 1
fi

app_path="build/bin/Grynvault Rebalance.app"
frameworks_path="${app_path}/Contents/Frameworks"
executable_path="${app_path}/Contents/MacOS/Grynvault Rebalance"
ldk_module_dir="$(go list -m -f '{{.Dir}}' github.com/getAlby/ldk-node-go)"

mkdir -p "$frameworks_path"
if [[ -e "$frameworks_path/libldk_node.dylib" ]]; then
  chmod u+w "$frameworks_path/libldk_node.dylib"
fi
cp "${ldk_module_dir}/ldk_node/universal-macos/libldk_node.dylib" "$frameworks_path/"
chmod u+w "$frameworks_path/libldk_node.dylib"

while IFS= read -r existing_rpath; do
  [[ -z "$existing_rpath" ]] || install_name_tool -delete_rpath "$existing_rpath" "$executable_path"
done < <(otool -l "$executable_path" | awk '$1 == "cmd" && $2 == "LC_RPATH" { getline; getline; print $2 }' | sort -u)
install_name_tool -add_rpath @executable_path/../Frameworks "$executable_path"
chmod +x "$executable_path"

# Wails applies a temporary ad-hoc signature before the release-only RPATH and
# embedded-library adjustments above. Re-seal both modified components so this
# internal candidate can be integrity-checked without claiming a publisher
# identity or Apple notarization.
codesign --force --sign - "$frameworks_path/libldk_node.dylib"
codesign --force --sign - "$app_path"
./scripts/release/inspect-macos.sh \
  "$app_path" \
  "$version_label" \
  "$source_commit" \
  --require-acceptance-adhoc

echo "ad-hoc sealed, untagged acceptance candidate: ${app_path}"
echo "source commit: ${source_commit}"
echo "version label: ${version_label}"
echo "this candidate is not a release artifact and must not be uploaded or published"

#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <immutable-tag>" >&2
  exit 64
fi

release_tag="$1"
repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ ! "$release_tag" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
  echo "release tag contains unsupported characters" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "release build requires a clean checkout" >&2
  exit 1
fi

source_commit="$(git rev-parse HEAD)"
ldk_library_sha256="be8ee3aa741b9ef7f017dda80e90e523b4fd7ae6cc7b4d18a7c2314495c7ef9a"
tag_ref="refs/tags/${release_tag}"
tag_commit="$(git rev-parse --verify "${tag_ref}^{commit}" 2>/dev/null || true)"
if [[ -z "$tag_commit" || "$tag_commit" != "$source_commit" ]]; then
  echo "release tag must exist locally and resolve to HEAD" >&2
  exit 1
fi

if [[ "$(node --version)" != v22.* ]]; then
  echo "Node.js 22.x is required for a release build" >&2
  exit 1
fi
if ! command -v wails >/dev/null 2>&1 || [[ "$(wails version | head -n 1)" != "v2.16.0" ]]; then
  echo "Wails CLI v2.16.0 is required for a release build" >&2
  exit 1
fi

source_url="https://github.com/hudhaifahz/hub/commit/${source_commit}"
export VITE_GRYNVAULT_SOURCE_URL="$source_url"

# Bark's shipped macOS static libraries declare a 15.4 deployment target. Keep
# the application and all CGO compilation/linking at that same conservative
# floor instead of allowing Wails to advertise an older, unverified target.
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
  # Platform preparation intentionally swaps tracked adapters. Restore the
  # repository's HTTP baseline before evaluating source cleanliness; the Wails
  # build below prepares and embeds the desktop adapters again.
  yarn prepare:http
)

if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "release checks changed the tagged checkout" >&2
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
  -ldflags "-X 'github.com/getAlby/hub/version.Tag=${release_tag}' -X 'github.com/getAlby/hub/version.SourceCommit=${source_commit}' -X 'github.com/getAlby/hub/version.LDKLibrarySHA256=${ldk_library_sha256}' -extldflags '-Wl,-headerpad_max_install_names -framework UniformTypeIdentifiers'"

# Wails' frontend build leaves the desktop adapters in the tracked source
# locations. Restore the committed HTTP baseline before the final clean-tree
# assertion without changing the already embedded application assets.
(
  cd frontend
  yarn prepare:http
)

if [[ -n "$(git status --porcelain=v1 --untracked-files=no)" ]]; then
  echo "release build changed tracked files in the tagged checkout" >&2
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

echo "unsigned internal candidate: ${app_path}"
echo "source commit: ${source_commit}"
echo "source URL: ${source_url}"
echo "the owner-selected unsigned experimental path requires explicit packaging, PGP manifest signing, publication approval, and fresh-download verification"

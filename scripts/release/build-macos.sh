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
tag_commit="$(git rev-parse "${release_tag}^{commit}" 2>/dev/null || true)"
if [[ -z "$tag_commit" || "$tag_commit" != "$source_commit" ]]; then
  echo "release tag must exist locally and resolve to HEAD" >&2
  exit 1
fi

if [[ "$(node --version)" != v22.* ]]; then
  echo "Node.js 22.x is required for a release build" >&2
  exit 1
fi

source_url="https://github.com/hudhaifahz/hub/commit/${source_commit}"
export VITE_GRYNVAULT_SOURCE_URL="$source_url"

(
  cd frontend
  yarn install --frozen-lockfile
  yarn lint:js
  yarn tsc:compile
  yarn format
)

if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "release checks changed the tagged checkout" >&2
  exit 1
fi

mkdir -p build
cp appicon.png build/appicon.png

go run github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 build \
  -clean \
  -m \
  -nosyncgomod \
  -trimpath \
  --platform darwin/universal \
  -webview2 embed \
  -o "Grynvault Rebalance" \
  -tags "wails" \
  -ldflags "-X 'github.com/getAlby/hub/version.Tag=${release_tag}' -extldflags '-Wl,-headerpad_max_install_names'"

if [[ -n "$(git status --porcelain=v1 --untracked-files=no)" ]]; then
  echo "release build changed tracked files in the tagged checkout" >&2
  exit 1
fi

app_path="build/bin/Grynvault Rebalance.app"
frameworks_path="${app_path}/Contents/Frameworks"
executable_path="${app_path}/Contents/MacOS/Grynvault Rebalance"
ldk_module_dir="$(go list -m -f '{{.Dir}}' github.com/getAlby/ldk-node-go)"

mkdir -p "$frameworks_path"
cp "${ldk_module_dir}/ldk_node/universal-macos/libldk_node.dylib" "$frameworks_path/"

while IFS= read -r existing_rpath; do
  [[ -z "$existing_rpath" ]] || install_name_tool -delete_rpath "$existing_rpath" "$executable_path"
done < <(otool -l "$executable_path" | awk '$1 == "cmd" && $2 == "LC_RPATH" { getline; getline; print $2 }' | sort -u)
install_name_tool -add_rpath @executable_path/../Frameworks "$executable_path"
chmod +x "$executable_path"

echo "unsigned internal candidate: ${app_path}"
echo "source commit: ${source_commit}"
echo "source URL: ${source_url}"
echo "public distribution remains blocked on Developer ID signing, notarization, stapling, manifest signing, and fresh-download verification"

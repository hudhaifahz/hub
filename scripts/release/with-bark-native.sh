#!/usr/bin/env bash
# Build-only correction for Go 1.27.4's rejection of embedded LLVM IR.
# The original module cache stays unchanged. No compiler check is disabled.
set -euo pipefail
[[ $# -gt 0 ]] || { echo "usage: $0 <command> [arguments...]" >&2; exit 64; }
repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$repo_root"
[[ "$(uname -s)" == Darwin ]] || { exec "$@"; }
[[ -z "${GOFLAGS:-}" ]] || { echo "Run without GOFLAGS; this wrapper owns its isolated build workspace" >&2; exit 1; }
objcopy="${LLVM_OBJCOPY:-$(rustc --print sysroot)/lib/rustlib/$(rustc -vV | awk '/^host:/ {print $2}')/bin/llvm-objcopy}"
[[ -x "$objcopy" ]] || { echo "Install Rust llvm-tools-preview, or set LLVM_OBJCOPY" >&2; exit 1; }
work="$repo_root/.release-native"
mkdir -p "$work"
module_dir="$(go mod download -json gitlab.com/ark-bitcoin/bark-ffi-bindings/golang@v0.26.1 |
  python3 -c 'import json,sys; print(json.load(sys.stdin)["Dir"])')"
# Copy only generated Go code, license and the two supported desktop libraries.
mkdir -p "$work/bark"
if [[ ! -d "$work/bark/bark" ]]; then
  cp "$module_dir/go.mod" "$module_dir/LICENSE" "$work/bark/"
  cp -R "$module_dir/bark" "$work/bark/"
fi
cmp "$module_dir/go.mod" "$work/bark/go.mod"
cmp "$module_dir/LICENSE" "$work/bark/LICENSE"
diff -qr "$module_dir/bark" "$work/bark/bark" >/dev/null || {
  echo "Staged Bark bindings differ from the verified upstream module" >&2; exit 1;
}
for arch in arm64 amd64; do
  case "$arch" in
    arm64) before=bf160eef454928553732ce827e1eb048412ce46d8d7e51583c4aa07f787cebef; after=f5890d920462de7a9e63a50a4dca429a2223386ade41ed2d0b89a13910f95fcb ;;
    amd64) before=cca4ec846b21968363025e5efcd3d20d8d1c4f52eb7c9e1a76b9ee88a329c87d; after=286f0ca3105e8a1be9c9e7973c3f3b2ad6e8ecc89728152aa7bbdf8acae95e4b ;;
  esac
  original="$module_dir/lib/darwin_$arch/libbark_ffi_go.a"
  destination="$work/bark/lib/darwin_$arch/libbark_ffi_go.a"
  [[ "$(shasum -a 256 "$original" | awk '{print $1}')" == "$before" ]] || { echo "Unexpected upstream Bark archive" >&2; exit 1; }
  mkdir -p "$(dirname "$destination")"
  if [[ ! -f "$destination" ]] || [[ "$(shasum -a 256 "$destination" | awk '{print $1}')" != "$after" ]]; then
    rm -f "$destination"
    "$objcopy" --remove-section=__LLVM,__bitcode --remove-section=__LLVM,__cmdline "$original" "$destination"
  fi
  [[ "$(shasum -a 256 "$destination" | awk '{print $1}')" == "$after" ]] || { echo "Unexpected transformed Bark archive; use Rust 1.89.0 llvm-tools-preview" >&2; exit 1; }
  python3 scripts/release/verify-native-sections.py "$original" "$destination"
done
temporary_workspace="$(mktemp "$work/.go.work.XXXXXX")"
trap 'rm -f "$temporary_workspace"' EXIT
printf 'go 1.27.4\n' > "$temporary_workspace"
GOWORK=off go work edit -use="$repo_root" -use="$work/bark" "$temporary_workspace"
mv "$temporary_workspace" "$work/go.work"
trap - EXIT
export GOWORK="$work/go.work"
export MACOSX_DEPLOYMENT_TARGET=15.4
export CGO_LDFLAGS="${CGO_LDFLAGS:-} -framework UniformTypeIdentifiers -mmacosx-version-min=15.4"
exec "$@"

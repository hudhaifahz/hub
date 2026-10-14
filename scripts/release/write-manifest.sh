#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 4 ]]; then
  echo "usage: $0 <immutable-tag> <pgp-fingerprint> <output-manifest> <artifact>..." >&2
  exit 64
fi

release_tag="$1"
pgp_fingerprint="${2//[[:space:]]/}"
output_manifest="$3"
shift 3

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ ! "$release_tag" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
  echo "release tag contains unsupported characters" >&2
  exit 1
fi
if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "manifest generation requires a clean checkout" >&2
  exit 1
fi

source_commit="$(git rev-parse HEAD)"
tag_commit="$(git rev-list -n 1 "$release_tag" 2>/dev/null || true)"
if [[ -z "$tag_commit" || "$tag_commit" != "$source_commit" ]]; then
  echo "release tag must exist locally and resolve to HEAD" >&2
  exit 1
fi
if [[ ! "$pgp_fingerprint" =~ ^[A-Fa-f0-9]{40}$ ]]; then
  echo "PGP fingerprint must contain exactly 40 hexadecimal characters" >&2
  exit 1
fi
pgp_fingerprint_upper="$(printf '%s' "$pgp_fingerprint" | tr '[:lower:]' '[:upper:]')"

manifest_directory="$(dirname "$output_manifest")"
mkdir -p "$manifest_directory"
temporary_manifest="$(mktemp "${manifest_directory}/.manifest.XXXXXX")"
trap 'rm -f "$temporary_manifest"' EXIT

{
  echo "release_tag=${release_tag}"
  echo "source_commit=${source_commit}"
  echo "upstream_hub_base=b7d0b0be829ba387d6c2f95c54a1449b71f6cd0e"
  echo "ldk_node_commit=2dc742a8b2057519e22d757f510e5cc880731b2c"
  echo "ldk_node_go_commit=a381cacb18562dff960f24a12f43232633f8bc25"
  echo "build_workflow=.github/workflows/wails.yml@${source_commit}"
  echo "platform=macos-universal"
  echo "bundle_identifier=com.grynvault.rebalance"
  echo "pgp_fingerprint=${pgp_fingerprint_upper}"
  for artifact_path in "$@"; do
    if [[ ! -f "$artifact_path" ]]; then
      echo "artifact does not exist: $artifact_path" >&2
      exit 1
    fi
    artifact_name="$(basename "$artifact_path")"
    artifact_size="$(stat -f '%z' "$artifact_path")"
    artifact_sha="$(shasum -a 256 "$artifact_path" | awk '{print $1}')"
    printf 'artifact\t%s\t%s\t%s\n' "$artifact_name" "$artifact_size" "$artifact_sha"
  done
} >"$temporary_manifest"

mv "$temporary_manifest" "$output_manifest"
trap - EXIT
echo "manifest written to $output_manifest"

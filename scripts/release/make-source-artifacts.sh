#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <immutable-tag> <output-directory>" >&2
  exit 64
fi

release_tag="$1"
output_directory="$2"
repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ ! "$release_tag" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
  echo "release tag contains unsupported characters" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "source artifacts require a clean checkout" >&2
  exit 1
fi

source_commit="$(git rev-parse HEAD)"
tag_ref="refs/tags/${release_tag}"
tag_commit="$(git rev-parse --verify "${tag_ref}^{commit}" 2>/dev/null || true)"
if [[ -z "$tag_commit" || "$tag_commit" != "$source_commit" ]]; then
  echo "release tag must exist locally and resolve to HEAD" >&2
  exit 1
fi

mkdir -p "$output_directory"
safe_tag="${release_tag//\//-}"
prefix="grynvault-rebalance-${safe_tag}/"

git archive --format=tar.gz --prefix="$prefix" \
  -o "${output_directory}/grynvault-rebalance-${safe_tag}-source.tar.gz" \
  "$tag_ref"

echo "source artifacts written to $output_directory"

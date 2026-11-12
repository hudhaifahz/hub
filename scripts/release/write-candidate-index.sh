#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 4 ]]; then
  echo "usage: $0 <immutable-tag> <expected-source-commit> <output-index> <artifact>..." >&2
  exit 64
fi

release_tag="$1"
expected_source_commit="$2"
output_index="$3"
shift 3

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
source_commit="$(git rev-parse HEAD)"
tag_ref="refs/tags/${release_tag}"
tag_commit="$(git rev-parse --verify "${tag_ref}^{commit}" 2>/dev/null || true)"
if [[ -z "$tag_commit" || "$tag_commit" != "$source_commit" || "$source_commit" != "$expected_source_commit" ]]; then
  echo "release tag must exist locally and resolve to HEAD; expected source commit must also match" >&2
  exit 1
fi

index_directory="$(dirname "$output_index")"
mkdir -p "$index_directory"
temporary_index="$(mktemp "${index_directory}/.candidate-index.XXXXXX")"
trap 'rm -f "$temporary_index"' EXIT

{
  echo "status=UNSIGNED_PREPARATION_ONLY"
  echo "release_tag=${release_tag}"
  echo "source_commit=${source_commit}"
  echo "distribution=unsigned-experimental"
  echo "apple_signing=absent"
  echo "apple_notarization=absent"
  echo "pgp_signature=absent"
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
} >"$temporary_index"

mv "$temporary_index" "$output_index"
trap - EXIT
echo "unsigned preparation index written to $output_index"

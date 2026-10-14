#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "release script tests require a clean checkout" >&2
  exit 1
fi

test_tag="grynvault-release-script-test-$$"
branch_only_ref="grynvault-release-script-branch-test-$$"
test_directory="$(mktemp -d "${TMPDIR:-/tmp}/grynvault-release-script-test.XXXXXX")"
test_artifact="${test_directory}/artifact.bin"
test_manifest="${test_directory}/manifest.txt"
test_candidate_index="${test_directory}/candidate-index.txt"
test_fingerprint="0123456789abcdef0123456789abcdef01234567"
expected_fingerprint="0123456789ABCDEF0123456789ABCDEF01234567"

cleanup() {
  git tag -d "$test_tag" >/dev/null 2>&1 || true
  git branch -D "$branch_only_ref" >/dev/null 2>&1 || true
  rm -rf "$test_directory"
}
trap cleanup EXIT

git tag "$test_tag" HEAD
printf 'release-script-test\n' >"$test_artifact"

/bin/bash ./scripts/release/write-manifest.sh \
  "$test_tag" \
  "$test_fingerprint" \
  "$test_manifest" \
  "$test_artifact"

grep -Fqx "release_tag=${test_tag}" "$test_manifest"
grep -Fqx "source_commit=$(git rev-parse HEAD)" "$test_manifest"
grep -Fqx "pgp_fingerprint=${expected_fingerprint}" "$test_manifest"
grep -Fqx 'distribution=unsigned-experimental' "$test_manifest"
grep -Fqx 'apple_signing=absent' "$test_manifest"
grep -Fqx 'apple_notarization=absent' "$test_manifest"
grep -Eq '^artifact[[:space:]]+artifact\.bin[[:space:]]+[0-9]+[[:space:]]+[a-f0-9]{64}$' "$test_manifest"

/bin/bash ./scripts/release/write-candidate-index.sh \
  "$test_tag" \
  "$(git rev-parse HEAD)" \
  "$test_candidate_index" \
  "$test_artifact"
grep -Fqx 'status=UNSIGNED_PREPARATION_ONLY' "$test_candidate_index"
grep -Fqx 'distribution=unsigned-experimental' "$test_candidate_index"
grep -Fqx 'apple_signing=absent' "$test_candidate_index"
grep -Fqx 'apple_notarization=absent' "$test_candidate_index"
grep -Fqx 'pgp_signature=absent' "$test_candidate_index"
grep -Eq '^artifact[[:space:]]+artifact\.bin[[:space:]]+[0-9]+[[:space:]]+[a-f0-9]{64}$' "$test_candidate_index"

git branch "$branch_only_ref" HEAD
assert_rejects_branch_ref() {
  if "$@" >"$test_directory/branch-ref.log" 2>&1; then
    echo "release script accepted a branch in place of an immutable tag: $*" >&2
    exit 1
  fi
  grep -Fq 'release tag must exist locally and resolve to HEAD' "$test_directory/branch-ref.log"
}

assert_rejects_branch_ref ./scripts/release/build-macos.sh "$branch_only_ref"
assert_rejects_branch_ref ./scripts/release/make-source-artifacts.sh "$branch_only_ref" "$test_directory/source"
assert_rejects_branch_ref ./scripts/release/write-manifest.sh \
  "$branch_only_ref" \
  "$test_fingerprint" \
  "$test_directory/branch-manifest.txt" \
  "$test_artifact"
assert_rejects_branch_ref ./scripts/release/write-candidate-index.sh \
  "$branch_only_ref" \
  "$(git rev-parse HEAD)" \
  "$test_directory/branch-candidate-index.txt" \
  "$test_artifact"

echo "release script tests passed under /bin/bash"

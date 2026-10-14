#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

if [[ -n "$(git status --porcelain=v1)" ]]; then
  echo "release script tests require a clean checkout" >&2
  exit 1
fi

test_tag="grynvault-release-script-test-$$"
test_directory="$(mktemp -d "${TMPDIR:-/tmp}/grynvault-release-script-test.XXXXXX")"
test_artifact="${test_directory}/artifact.bin"
test_manifest="${test_directory}/manifest.txt"
test_fingerprint="0123456789abcdef0123456789abcdef01234567"
expected_fingerprint="0123456789ABCDEF0123456789ABCDEF01234567"

cleanup() {
  git tag -d "$test_tag" >/dev/null 2>&1 || true
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
grep -Eq '^artifact[[:space:]]+artifact\.bin[[:space:]]+[0-9]+[[:space:]]+[a-f0-9]{64}$' "$test_manifest"

echo "release script tests passed under /bin/bash"

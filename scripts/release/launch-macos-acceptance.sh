#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <Grynvault-Rebalance.app> <acceptance-work-dir>" >&2
  exit 64
fi

app_path="$1"
requested_workdir="$2"

if [[ ! -d "$app_path" || ! -x "$app_path/Contents/MacOS/Grynvault Rebalance" ]]; then
  echo "candidate application is missing its executable" >&2
  exit 1
fi
if lsof -nP -iTCP:21421 -sTCP:LISTEN >/dev/null 2>&1 || \
  lsof -nP -iTCP:21420 -sTCP:LISTEN >/dev/null 2>&1; then
  echo "Alby Hub or Grynvault Rebalance is still running; stop it normally before acceptance" >&2
  exit 1
fi

mkdir -p "$requested_workdir"
canonical_workdir="$(cd "$requested_workdir" && pwd -P)"
case "$canonical_workdir" in
  */grynvault-rebalance-acceptance/*) ;;
  *)
    echo "acceptance work directory must be inside a grynvault-rebalance-acceptance folder" >&2
    exit 1
    ;;
esac
if [[ -n "$(find "$canonical_workdir" -mindepth 1 -maxdepth 1 -print -quit)" ]]; then
  echo "acceptance work directory must be empty before its first guarded launch" >&2
  exit 1
fi
chmod 0700 "$canonical_workdir"

authority_path="${canonical_workdir}.grynvault-release-acceptance"
umask 077
printf 'grynvault-release-acceptance-v1\n%s\n' "$canonical_workdir" >"$authority_path"
chmod 0600 "$authority_path"

/usr/bin/open -n -F "$app_path" --args \
  "--grynvault-release-acceptance-work-dir=${canonical_workdir}" \
  --grynvault-release-acceptance-network-dark

echo "started network-dark release acceptance against ${canonical_workdir}"

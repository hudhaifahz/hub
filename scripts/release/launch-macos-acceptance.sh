#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <Grynvault-Rebalance.app> <acceptance-work-dir>" >&2
  exit 64
fi

app_path="$1"
requested_workdir="$2"
repo_root="$(git rev-parse --show-toplevel)"

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
log_path="${canonical_workdir}.grynvault-release-acceptance.log"
umask 077
printf 'grynvault-release-acceptance-v1\n%s\n' "$canonical_workdir" >"$authority_path"
chmod 0600 "$authority_path"

app_executable="$app_path/Contents/MacOS/Grynvault Rebalance"
profile_path="$repo_root/build/darwin/release-acceptance.sb"
GRYNVAULT_RELEASE_ACCEPTANCE_ACTIVE=1 \
  GRYNVAULT_RELEASE_ACCEPTANCE_WORK_DIR="$canonical_workdir" \
  /usr/bin/nohup /usr/bin/sandbox-exec \
  -f "$profile_path" \
  "$app_executable" \
  "--grynvault-release-acceptance-work-dir=${canonical_workdir}" \
  --grynvault-release-acceptance-network-dark \
  >"$log_path" 2>&1 &
candidate_pid=$!

sleep 1
if ! kill -0 "$candidate_pid" 2>/dev/null; then
  echo "guarded candidate exited during startup; inspect $log_path" >&2
  exit 1
fi

echo "started network-dark release acceptance as PID ${candidate_pid} against ${canonical_workdir}"
echo "startup log: ${log_path}"

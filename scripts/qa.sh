#!/usr/bin/env bash
#
# qa.sh — orchestrator auto-spawn QA Engineer (ADR-012).
#
# Alur: launch-qa (gate reviewing)
#   → claude -p di worktree (agent qa-engineer, quality_gates + acceptance)
#   → collect-qa (handoff → merge-ready / defect-found → running)
#
# Spawn memakai claude CLI dengan agent qa-engineer (permissionMode default).

set -euo pipefail

: "${M2S_ROOT:?M2S_ROOT tidak di-set — akar control repository.}"
M2S_BIN="${M2S_BIN:-$M2S_ROOT/bin/m2s}"

usage() {
  echo "Pemakaian: $0 --task <id> [--control <path>]" >&2
  exit 1
}

task=""
control="$M2S_ROOT"
while [ $# -gt 0 ]; do
  case "$1" in
    --task) task="$2"; shift 2 ;;
    --control) control="$2"; shift 2 ;;
    *) usage ;;
  esac
done
[ -n "$task" ] || usage

"$M2S_BIN" launch-qa --task "$task" --control "$control"

# Spawn QA Engineer di worktree.
echo "spawn qa-engineer untuk $task — agent dijalankan terpisah (manual/CI)"

#!/usr/bin/env bash
#
# review.sh — orchestrator auto-spawn Code Reviewer (ADR-012).
#
# Runner siapkan (gate + reservasi implementer), orchestrator spawn agent.
# m2s TIDAK spawn agent sendiri — pola sama launch-task (runner tipis).
#
# Alur: launch-review (gate implementation-complete)
#   → claude -p di worktree (agent code-reviewer, plan mode read-only)
#   → collect-review (handoff → reviewing / changes-requested)
#
# Spawn memakai claude CLI dengan agent code-reviewer (permissionMode plan).
# Worktree sudah disiapkan runner; agent cwd = worktree.

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

"$M2S_BIN" launch-review --task "$task" --control "$control"

# Spawn Code Reviewer di worktree. Runner mencetak instruksi dengan repo/branch/
# worktree; untuk otomasi penuh, worktree diambil dari reservasi.
# NOTE: claude -p adalah titik di mana agent benar-benar dijalankan.
# Digantikan oleh orchestrator project bila alur berbeda (mis. GitHub Actions).
echo "spawn code-reviewer untuk $task — agent dijalankan terpisah (manual/CI)"

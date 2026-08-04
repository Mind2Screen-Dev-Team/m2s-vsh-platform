#!/usr/bin/env bash
#
# Memasang ruleset agent-push-restriction (bypass m2s-approver) + ruleset
# agent-worker-restriction (kunci develop/staging dari push langsung worker)
# + main-agent-block (kunci main dari head agent/*, H-01 Phase 8) pada 3 repo
# org. ADR-008 §Langkah ADR-001 #5, V-10.
#
# Jalankan sebagai MANUSIA (admin org). Membutuhkan token gh dengan scope
# `repo`; ruleset POST adalah aksi yang tak boleh dilakukan agent (R-20).
#
# Env:
#   M2S_APPROVER_ID   app_id App m2s-approver (wajib).
#
# Pemakaian:
#   M2S_APPROVER_ID=<app_id> tools/apply-rulesets.sh

set -euo pipefail

: "${M2S_APPROVER_ID:?M2S_APPROVER_ID tidak di-set — app_id App m2s-approver.}"

REPOS=(
  Mind2Screen-Dev-Team/m2s-vsh-platform
  Mind2Screen-Dev-Team/m2s-vsh-project-backend
  Mind2Screen-Dev-Team/m2s-vsh-project-frontend
)
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APPROVER_RS="$ROOT/templates/github/rulesets/agent-push-restriction-approver.json"
WORKER_RS="$ROOT/templates/github/rulesets/agent-worker-restriction.json"

for repo in "${REPOS[@]}"; do
  echo "== $repo =="
  # agent-push-restriction: approver App (Integration) bypass `always`.
  # V-10: `always` (bukan pull_request) dibutuhkan agar API merge berhasil.
  sed "s/\"actor_id\": 0/\"actor_id\": $M2S_APPROVER_ID/" "$APPROVER_RS" \
    | gh api -X POST "repos/$repo/rulesets" --input - >/dev/null
  echo "  + agent-push-restriction (approver = always, bypass)"
  # agent-worker-restriction: kunci develop/staging dari push langsung worker,
  # tapi approver jg di-bypass di ruleset ini (semula cuma OrganizationAdmin,
  # sehingga menahan approver — V-10). Worker sendiri tidak di-bypass.
  sed "s/\"actor_id\": 0/\"actor_id\": $M2S_APPROVER_ID/" "$WORKER_RS" \
    | gh api -X POST "repos/$repo/rulesets" --input - >/dev/null
  echo "  + agent-worker-restriction (approver = always, bypass)"
  # main-agent-block (Phase 8, H-01): kunci main dari PR head agent/*.
  # Workflow path-enforcement trigger-nya [develop, staging], sehingga PR
  # agent/* → main TIDAK memicu CI sama sekali — tanpa check, tanpa tolak.
  # Ruleset ini menutup celah: rule update (non-fast-forward) menolak merge
  # branch agent/* ke main di level GitHub, tak tergantung workflow (Phase 8
  # A7 menemukan celah ini; test dummy PR #14 backend ter-bukti mergeable).
  sed "s/\"actor_id\": 0/\"actor_id\": $M2S_APPROVER_ID/" "$ROOT/templates/github/rulesets/main-agent-block.json" \
    | gh api -X POST "repos/$repo/rulesets" --input - >/dev/null
  echo "  + main-agent-block (main × head agent/*, H-01)"
done

echo "Selesai. Verifikasi: gh api repos/<repo>/rulesets --jq '.[] | {name, enforcement, bypass_actors}'"
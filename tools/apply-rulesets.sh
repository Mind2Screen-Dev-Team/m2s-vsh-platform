#!/usr/bin/env bash
#
# Memasang ruleset agent-push-restriction (bypass m2s-approver) + ruleset
# agent-worker-restriction (kunci develop/staging dari push langsung worker)
# pada 3 repo org. ADR-008 §Langkah ADR-001 #5.
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
  # agent-push-restriction: bypass m2s-approver (Integration) + manusia admin
  # (OrganizationAdmin, actor_id null). Merge lewat PR hanya untuk approver.
  sed "s/\"actor_id\": 0/\"actor_id\": $M2S_APPROVER_ID/" "$APPROVER_RS" \
    | gh api -X POST "repos/$repo/rulesets" --input - >/dev/null
  echo "  + agent-push-restriction (bypass m2s-approver)"
  # agent-worker-restriction: tanpa bypass App — worker tak boleh ubah
  # develop/staging; hanya manusia admin (OrganizationAdmin) lolos.
  gh api -X POST "repos/$repo/rulesets" --input "$WORKER_RS" >/dev/null
  echo "  + agent-worker-restriction"
done

echo "Selesai. Verifikasi: gh api repos/<repo>/rulesets --jq '.[] | {name, enforcement, bypass_actors}'"

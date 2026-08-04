#!/usr/bin/env bash
#
# Memasang ruleset agent-push-restriction (bypass m2s-approver) + ruleset
# agent-worker-restriction (kunci develop/staging dari push langsung worker)
# pada 3 repo org. ADR-008 §Langkah ADR-001 #5, V-10.
#
# CATATAN (Phase 8 A7): ruleset butuh plan GitHub Team. Org ini plan Free,
# sehingga POST ruleset apa pun -> 500. Script ini sengaja dipertahankan untuk
# saat plan dinaikkan, tetapi SAAT INI ia tidak dapat menjalankan apa pun.
# main-agent-block (H-01 tutup celah main) diarsipkan di berkas ruleset itu
# sendiri; pengganti aktif adalah tools/apply-branch-protection.sh yang menutup
# main via classic branch protection (tersedia di plan Free).
#
# Jalankan sebagai MANUSIA (admin org). Membutuhkan token gh dengan scope
# `repo`; ruleset POST adalah aksi yang tak boleh dilakukan agent (R-20).
#
# Idempotent: ruleset yang sudah ada di-UPDATE (PUT ke id-nya), yang belum
# di-CREATE (POST). Menjalankan ulang tidak membuat duplikat — diperbaiki
# setelah 422 pertama (Phase 4 mem-POST nama yang sudah ada).
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

# upsert_ruleset <repo> <nama> <berkas-template>
#   PUT bila ruleset bernama itu sudah ada di repo, POST bila belum.
#
# Payload ditulis ke berkas temp, lalu `--input <file>` — BUKAN `--input -`.
# Beberapa rilis gh mengirim body kosong lewat stdin ("unexpected end of input"
# JSON) ketika di-pipe; berkas temp menghindari itu sepenuhnya.
upsert_ruleset() {
  local repo="$1" name="$2" file="$3" id="" tmp
  id=$(gh api "repos/$repo/rulesets" --jq ".[] | select(.name==\"$name\") | .id" 2>/dev/null | head -1)
  tmp=$(mktemp)
  sed "s/\"actor_id\": 0/\"actor_id\": $M2S_APPROVER_ID/" "$file" > "$tmp"
  trap 'rm -f "$tmp"' RETURN
  if [ -n "$id" ]; then
    gh api -X PUT "repos/$repo/rulesets/$id" --input "$tmp" >/dev/null
    echo "  ~ $name (update id=$id)"
  else
    gh api -X POST "repos/$repo/rulesets" --input "$tmp" >/dev/null
    echo "  + $name (create)"
  fi
}

for repo in "${REPOS[@]}"; do
  echo "== $repo =="
  # agent-push-restriction: approver App (Integration) bypass `always`.
  # V-10: `always` (bukan pull_request) dibutuhkan agar API merge berhasil.
  upsert_ruleset "$repo" agent-push-restriction \
    "$ROOT/templates/github/rulesets/agent-push-restriction-approver.json"
  # agent-worker-restriction: kunci develop/staging dari push langsung worker,
  # tapi approver jg di-bypass di ruleset ini (semula cuma OrganizationAdmin,
  # sehingga menahan approver — V-10). Worker sendiri tidak di-bypass.
  upsert_ruleset "$repo" agent-worker-restriction \
    "$ROOT/templates/github/rulesets/agent-worker-restriction.json"
  done

echo "Selesai. Verifikasi: gh api repos/<repo>/rulesets --jq '.[] | {name, rules, bypass_actors}'"

#!/usr/bin/env bash
#
# Memasang branch protection pada `main` di 3 repo org (Phase 8, H-01/H-08).
#
# LATAR BELAKANG: ruleset (templates/github/rulesets/*.json) butuh plan GitHub
# Team — org ini plan Free, sehingga `gh api .../rulesets` selalu 500
# ("Upgrade to GitHub Team to enable this feature"). Pengganti yang tersedia
# di plan Free adalah classic branch protection, yang API-nya beda dan terbukti
# bekerja (200).
#
# Yang dipasang pada main:
#   - required_status_checks: path-enforcement (strict). PR agent/* → main TIDAK
#     memicu workflow (trigger [develop, staging]), sehingga check ini tak pernah
#     hijau → PR tak bisa merge. Ini menutup celah H-01 di main (Phase 8 A7:
#     PR dummy backend terbukti BLOCKED setelah protection aktif).
#   - allow_force_pushes: false, allow_deletions: false — main dilindungi.
#
# Jalankan sebagai MANUSIA dengan token admin: `gh auth refresh -s admin:org,repo`
# atau PAT yang punya scope admin. Aksi ini tak boleh dilakukan agent (R-20).
#
# Pemakaian:
#   GH_TOKEN=<token-admin> tools/apply-branch-protection.sh
#
# Idempotent: PUT branch protection meng-overwrite state; menjalankan ulang
# menghasilkan konfigurasi yang sama.

set -euo pipefail

: "${GH_TOKEN:?GH_TOKEN tidak di-set — token admin (scope admin:org, repo).}"

REPOS=(
  Mind2Screen-Dev-Team/m2s-vsh-platform
  Mind2Screen-Dev-Team/m2s-vsh-project-backend
  Mind2Screen-Dev-Team/m2s-vsh-project-frontend
)

PAYLOAD='{
  "required_status_checks": {"strict": true, "contexts": ["path-enforcement"]},
  "required_pull_request_reviews": {"required_approving_review_count": 0},
  "enforce_admins": false,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "restrictions": null
}'

for repo in "${REPOS[@]}"; do
  echo "== $repo =="
  code=$(curl -sS -o /dev/null -w "%{http_code}" \
    -X PUT \
    -H "Authorization: Bearer $GH_TOKEN" \
    -H "Accept: application/vnd.github+json" \
    "https://api.github.com/repos/$repo/branches/main/protection" \
    -d "$PAYLOAD")
  echo "  main protection -> HTTP $code"
done

echo "Selesai. Verifikasi: gh api repos/<repo>/branches/main/protection"

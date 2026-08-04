# Phase 8 — Branch Protection pada main (pengganti ruleset, plan Free)

**Status:** aktif, 2026-08-04
**Sumber:** Phase 8 A7 — penutupan celah H-01 di `main`

## Temuan A7: ruleset tidak tersedia di plan Free

Saat menguji H-01 secara live (PR dummy `agent/* → main` di backend), ditemukan
celah: workflow `path-enforcement` trigger-nya `[develop, staging]`, sehingga PR
ke `main` **tidak memicu CI sama sekali** — `gh pr checks` kosong,
`mergeStateStatus: CLEAN`, PR lolos tanpa guard.

Upaya menutup via ruleset gagal. `gh api .../rulesets` (POST/PUT) mengembalikan
**HTTP 500** untuk aturan apa pun. Dua ruleset Phase 4
(`agent-push-restriction`, `agent-worker-restriction`) juga kosong (`rules: null`)
sejak awal — tak pernah berhasil menulis isi. Diagnosis (dengan token admin):

```
GET /orgs/Mind2Screen-Dev-Team/rulesets
> 403 "Upgrade to GitHub Team to enable this feature."
```

**Ruleset (repo maupun org) butuh plan GitHub Team. Organisasi ini plan Free.**
Bukan soal token/script/payload.

## Pengganti yang aktif: classic branch protection

Tersedia di plan Free. Dipasang pada `main` di 3 repo:

- **required_status_checks:** `validate-changed-paths` (strict) — nama JOB
  workflow = nama check yang dituntut (check-github-artifacts.sh aturan #3:
  "nama job validate-changed-paths ADALAH nama required check"). Jangan tulis
  `path-enforcement` di sini; context yang salah membuat PR tetap BLOCKED
  walau check lulus (nama tak cocok).
- **allow_force_pushes:** false
- **allow_deletions:** false
- **enforce_admins:** false

Karena PR `agent/* → main` tidak memicu workflow, check `path-enforcement`
tidak pernah hijau untuk target main → **PR tak bisa merge**. Ini menutup celah
H-01 secara mekanis, tanpa butuh plan berbayar.

**Bukti rela, PR #14 backend (`agent/BE-999-dummy → main`):**

| | Sebelum protection | Setelah protection |
|---|---|---|
| `mergeStateStatus` | `CLEAN` (lolos) | **`BLOCKED`** |

## Skrip

`tools/apply-branch-protection.sh` — idempotent, PUT over state:

```
GH_TOKEN=<token-admin> tools/apply-branch-protection.sh
```

Token wajib scope `admin:org` + `repo` (`gh auth refresh -s admin:org,repo`
atau PAT classic). Agent dilarang (R-20); manusia yang menjalankan.

Sudah diterapkan ke 3 repo (control, backend, frontend) pada 2026-08-04, HTTP
200 pada ketiganya.

## Arsip ruleset

`templates/github/rulesets/main-agent-block.json` — diarsipkan (`enforcement:
disabled` + komentar `_status: archived`). Dirancang utk menutup celah yang sama
(target main x head agent/*, rule update) tetapi tak dapat dipasang di plan
Free. Aktifkan kembali bila org dinaikkan ke GitHub Team.

`tools/apply-rulesets.sh` dipertahankan (untuk saat plan dinaikkan) tetapi tidak
dapat menjalankan apa pun saat ini — tidak memproses `main-agent-block` lagi.

## Upgrade ke GitHub Team? (keputusan terbuka)

Ruleset lebih unggul dari branch protection (head-ref filter per branch,
bypass actor, granular). Bila org dinaikkan ke GitHub Team, aktifkan
`main-agent-block` + jalan `apply-rulesets.sh`. Biaya vs manfaat jadi keputusan
Project Manager; sampai saat itu branch protection sudah menutup celah yang
sama.
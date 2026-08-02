# Migrasi Org — Runbook Operator

**Fase:** transisi setelah Phase 4 (§60)
**Tujuan:** pindahkan 3 repo `fajarcandraaa/*` → `Mind2Screen-Dev-Team/*`, lalu
tegakkan ADR-001 #5 (dua GitHub App) yang selama ini mustahil karena V-08.
**Rujukan:** ADR-008 (keputusan), ADR-007 (latar), D-02/D-03, V-06..V-08.

Dokumen ini untuk **operator manusia**. Transfer, branch protection, dan aturan
adalah aksi yang tak boleh dilakukan agent (ADR-001 #4, R-20). Perintah `gh` di
bawah siap-tempel; jalankan dengan `!` di prompt Claude Code agar output masuk ke
sesi.

---

## Ringkas alur

```
0. Prasyarat: kunci semua sesi agent — transfer mengubah URL mid-session
1. Ganti SEMUA ref fajarcandraaa/ -> Mind2Screen-Dev-Team/  (patch #1)
2. Verifikasi build + make verify hijau
3. Transfer 3 repo ke org
4. Verifikasi pasca-transfer (proteksi, ruleset, CI, GitGuardian)
5. Buat 2 GitHub App + pasang ruleset bypass (ADR-001 #5)
```

⚠️ **Urutan wajib.** Transfer SEBELUM ref diganti = CI di repo aplikasi menunjuk
control repo lama yang sudah redirect — rapuh, dan window-nya tak perlu. Ganti ref
dulu, verifikasi, baru transfer.

---

## Patch #1 — ganti ref `fajarcandraaa/` → `Mind2Screen-Dev-Team/`

### Kenapa ini SATU patch, bukan beberapa PR

Agent tidak bisa menyentuh `cmd/m2s/**` (deny, human-only write). Tiga import di
sana menyebut module path; `go.mod` tak bisa diganti lebih dulu tanpa memecah
build. Karena itu seluruh ref harus berganti **bersamaan** dalam satu langkah.

Ada **10 referensi fungsional** (belum termasuk ~55 teks dokumen yang auto-redirect
dan boleh dibiarkan):

| File | Baris | Referensi |
|---|---|---|
| `go.mod` | 1 | `module github.com/fajarcandraaa/m2s-vsh-platform` |
| `cmd/m2s/commands.go` | 12, 13 | import contract, registry |
| `cmd/m2s/pathcheck.go` | 21 | import pathmatch |
| `internal/registry/registry.go` | 25, 26 | import contract, pathmatch |
| `internal/registry/registry_test.go` | 12, 171, 231 | import contract; pr_url fixture ×2 |
| `schemas/examples/handoff-BE-101.valid.yaml` | 56 | pr_url |
| `templates/github/workflows/path-enforcement.yml` | 96 | `repository:` checkout control repo |

Jalankan dari root control repo:

```bash
# Modul path (go.mod + 5 import). Path import adalah string persis, jadi
# penggantian per-baris aman. command di bawah memakai perl -pi agar string
# literal, bukan regex.
perl -pi -e 's{github\.com/fajarcandraaa/m2s-vsh-platform}{github.com/Mind2Screen-Dev-Team/m2s-vsh-platform}g' \
  go.mod \
  cmd/m2s/commands.go cmd/m2s/pathcheck.go \
  internal/registry/registry.go internal/registry/registry_test.go

# pr_url fixture (backend) — repositori backend pindah juga.
perl -pi -e 's{github\.com/fajarcandraaa/m2s-vsh-project-backend}{github.com/Mind2Screen-Dev-Team/m2s-vsh-project-backend}g' \
  internal/registry/registry_test.go schemas/examples/handoff-BE-101.valid.yaml

# Workflow template: checkout control repo.
sed -i '' 's/repository: fajarcandraaa\/m2s-vsh-platform/repository: Mind2Screen-Dev-Team\/m2s-vsh-platform/' \
  templates/github/workflows/path-enforcement.yml
```

Verifikasi **semua** lenyap:

```bash
grep -rn "fajarcandraaa" --include='*.go' --include='*.yml' --include='*.yaml' \
  cmd internal schemas templates .github 2>/dev/null
# Harus kosong. (tests/negative/github-workflow.test.sh sengaja DIKECUALIKAN:
# @fajarcandraaa di sana adalah fixture test negatif, bukan ref. JANGAN diubah.)
```

Lalu gate:

```bash
make build && make verify
```

Commit + merge ke `main` control repo. Lalu turunkan template ke 6 salinan repo
aplikasi (ulangi pola distribusi Phase 4: satu PR per base branch, `repository:`
line saja).

---

## Patch #2 (human-only) — dua hal Phase 4 yang belum diterapkan

Sambil di sana, terapkan sisa `docs/operator/phase-4-human-only-patches.md`:
1. Target Makefile `verify-github`
2. Daftarkan `worktree-lifecycle.sh` (T-04)

Keduanya tidak terkait migrasi, tapi repo sedang dibuka untuk human edit — sekali
jadi.

---

## Langkah transfer

Pindahkan tiga repo. Jalankan satu per satu, dalam urutan apa pun (independen):

```bash
# Control repo.
gh api -X POST repos/fajarcandraaa/m2s-vsh-platform/transfer \
  -f new_owner=Mind2Screen-Dev-Team

gh api -X POST repos/fajarcandraaa/m2s-vsh-project-backend/transfer \
  -f new_owner=Mind2Screen-Dev-Team

gh api -X POST repos/fajarcandraaa/m2s-vsh-project-frontend/transfer \
  -f new_owner=Mind2Screen-Dev-Team
```

⚠️ **Sebelum transfer:** kunci semua sesi agent. URL berubah mid-session; worktree
yang masih menunjuk `fajarcandraaa/` akan gagal push. Update remote lokal:

```bash
cd <clone>/m2s-vsh-platform
git remote set-url origin git@github.com:Mind2Screen-Dev-Team/m2s-vsh-platform.git
# ulangi untuk backend, frontend.
```

GitHub meng-redirect URL lama otomatis, jadi push/pull lama masih jalan untuk
sementara — tapi jangan diandalkan.

---

## Verifikasi pasca-transfer

Wajib, bukan opsional. Bandingkan dengan baseline sebelum transfer
(ADR-008 mencatat "belum diverifikasi" apakah proteksi ikut).

```bash
# 1. Branch protection: required check ikut?
for r in m2s-vsh-project-backend m2s-vsh-project-frontend; do
  for b in develop staging; do
    printf '%-30s %-8s ' "$r" "$b"
    gh api "repos/Mind2Screen-Dev-Team/$r/branches/$b/protection" \
      --jq '.required_status_checks.contexts // ["HILANG — daftar ulang"]' 2>&1 | head -1
  done
done
# Expected: validate-changed-paths di 4 branch. Kalau HILANG, jalankan ulang
# perintah di docs/operator/branch-protection.md §Langkah 4 (PUT protection).

# 2. Ruleset ikut? (baseline = 0)
for r in m2s-vsh-platform m2s-vsh-project-backend m2s-vsh-project-frontend; do
  printf '%-30s rulesets=%s\n' "$r" "$(gh api repos/Mind2Screen-Dev-Team/$r/rulesets --jq 'length')"
done

# 3. CI jalan? buka satu PR kecil, amati run.
gh api "repos/Mind2Screen-Dev-Team/m2s-vsh-platform/actions/runs" --jq '.total_count'

# 4. GitGuardian ikut? (app_id 46505) — cek check-runs PR baru. Kalau hilang,
#    reinstall dari pengaturan GitHub → Installations → GitGuardian → org.
```

Kalau branch protection **hilang** saat transfer, perintah PUT di
`branch-protection.md` §Langkah 4 mengembalikannya. Jangan aktifkan required check
sebelum satu run hijau di lokasi baru.

---

## Langkah ADR-001 #5 — dua GitHub App (setelah org aktif)

**Ini keputusan yang memakai `admin:org`.** Token agent tidak punya scope itu.
**Membuat App tidak punya REST endpoint** — dilakukan lewat UI oleh manusia,
sehingga langkah 1 dan 3 di bawah adalah aksi manusia, bukan agent.

1. Buat App `m2s-worker` dan `m2s-approver` di:
   `https://github.com/organizations/Mind2Screen-Dev-Team/settings/apps/new`
   - Hak minimal: `Contents: Read & Write`, `Pull requests: Read & Write`
   - Izinkan hanya 3 repo yang relevan (control, backend, frontend)
   - Unduh **private key** masing-masing App saat dibuat; simpan di luar repo
     (mis. `~/.claude/secrets/`). Kunci `m2s-approver` adalah aset bernilai
     tinggi (ADR-001) — jangan commit ke mana pun.
   - Catat `app_id` tiap App. Dapat dibaca ulang via API setelah lahir:
     `gh api apps/<slug> --jq .id` (slug `m2s-worker` / `m2s-approver`).
2. Pasang ruleset per App. Bypass `actor_type: Integration` kini **legal** di org
   (V-08 hanya berlaku repo personal). Payload kanonik ada di
   `templates/github/rulesets/`, pemasang di `tools/apply-rulesets.sh`:
   - `agent-push-restriction-approver.json` — bypass `Integration` App
     `m2s-approver` (`bypass_mode: pull_request`), kunci develop/staging;
     hanya App approver yang boleh merge lewat PR.
   - `agent-worker-restriction.json` — bypass hanya manusia `OrganizationAdmin`;
     worker tak boleh mengubah develop/staging, hanya push `agent/*` + buka PR.

```bash
# <APP_ID> = id App m2s-approver.
M2S_APPROVER_ID=<APP_ID> tools/apply-rulesets.sh
```

3. **Verifikasi §66 #9** (Implementer tidak merge PR sendiri): worker App buka PR,
   coba merge → harus ditolak. Ini momen pertama acceptance itu teruji.

---

## Rollback

ADR-008 §Rollback. Transfer-balik ke `fajarcandraaa`, pulihkan ref (patch #1
terbalik), verifikasi CI. Harga: satu siklus penuh PR + merge. Verifikasi
pasca-transfer (#4) adalah penjaga agar ini tak perlu.

---

## Setelah migrasi bersih

- Update remote lokal semua clone.
- Update `docs/operator/branch-protection.md`: ganti `fajarcandraaa/` → org di
  seluruh contoh.
- Hapus repo probe `m2s-vsh-rules-probe` bila selesai.
- Phase 5 (§61 Tool Pilot) bisa mulai — tidak lagi terblokir oleh kepemilikan org.

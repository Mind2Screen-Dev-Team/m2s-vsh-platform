# Phase 8 Hardening — Enforcement Gaps Post-Phase-7

**Tanggal:** 2026-08-03
**Status:** draft
**Tujuan:** Menutup tiga enforcement gap yang terekspos selama Phase 7 multi-repo pilot.

## Latar Belakang

Phase 7 (multi-repo pilot) selesai dan berjalan, tetapi mengekspos tiga kesalahan
eksekusi yang tidak tertahan enforcement layer:

1. **PR agent → main langsung**, bukan develop → staging → main (§44 dilanggar).
2. **Branch planning pakai pola `agent/*`** padahal bukan task.
3. **Contract terlalu ketat** — forbid file yang scaffolding wajib hasilkan
   (`go.mod`, `src/app/layout.tsx`), menyebabkan loop CI + remediasi manual.

Semua tertulis di arsitektur, tapi tidak ada guard yang menahan penyimpangan
pada level eksekusi. Hardening ini menambah guard — enforcement, bukan sekadar
dokumentasi.

---

## H-01: CI Reject PR agent → main langsung

**Masalah:** Workflow `path-enforcement.yml` hanya validasi path, bukan base branch.
Agent bisa push PR ke `main` dan lolos CI.

**Guard:** workflow memeriksa `github.base_ref`. PR dengan head `agent/*` yang
target `main` (pada repo aplikasi) → fail-closed.

```yaml
- name: Validasi base branch
  if: steps.scope.outputs.mode == 'agent'
  run: |
    case "${{ github.base_ref }}" in
      develop) exit 0 ;;
      staging) exit 0 ;;
      *)
        echo "::error::branch agent/* wajib target develop atau staging (§44), bukan ${{ github.base_ref }}"
        exit 1
        ;;
    esac
```

**Catatan:** control repo main-only (konvensi), jadi guard ini hanya aktif di
repo aplikasi (backend/frontend). Flag `APPLICATION_REPO=true` di env.

---

## H-02: Branch Naming Lint — `agent/*` hanya untuk task

**Masalah:** Branch planning (`agent/phase-7-plan-specs`) lolos ke pattern `agent/*`,
lalu CI malformed karena slug bukan task-id. Deteksi terjadi setelah push, rework.

**Guard:** lint branch name terhadap konvensi — `agent/<TASK-ID>-<slug>` OR
`worktree-<context>`. Tidak boleh campur.

```bash
case "$HEAD_REF" in
  agent/*)
    task_id=$(printf '%s' "$HEAD_REF" | sed -n 's#^agent/\([A-Z][A-Z0-9]*-[0-9]\{1,\}\)-.*#\1#p')
    if [ -z "$task_id" ]; then
      echo "::error::branch agent/* wajib pola agent/<task-id>-<slug>; planning/docs pakai worktree-*"
      exit 1
    fi
    ;;
  worktree-*) ;;
  *) exit 0 ;;  # non-agent, bukan task
esac
```

Ini sudah ada di `path-enforcement.yml` step "Tentukan lingkup pemeriksaan",
tetapi H-02 memperjelas: planning artifacts WAJIB `worktree-*`, tidak boleh
`agent/*` tanpa task-id.

---

## H-03: Contract Realism Check — scaffolding files wajib diizinkan

**Masalah:** Contract BE-102 forbid `go.mod`, FE-102 forbid `src/app/layout.tsx`.
Keduanya wajib dibuat scaffolding. Agent menaati contract → CI fail → remediasi.

**Guard:** `validate-task.sh` (atau `m2s validate-task`) memeriksa contract
terhadap daftar "scaffold-mandatory files" per stack:

```bash
# scripts/validate-task.sh — tambahan check
scaffold_forbidden() {
  local contract="$1"
  local stack="$2"   # go | nextjs
  local allowed
  allowed=$(yq '.paths.allowed[]' "$contract")

  case "$stack" in
    go)
      # Repo Go baru selalu hasilkan go.mod (go mod init)
      if ! echo "$allowed" | grep -q 'go.mod'; then
        echo "::error::contract forbid/omit go.mod — scaffolding go mod init wajib buat file ini (§29.6 shared file)"
        exit 1
      fi
      ;;
    nextjs)
      # create-next-app selalu hasilkan layout.tsx + globals.css
      for f in "src/app/layout.tsx" "src/app/globals.css"; do
        if ! echo "$allowed" | grep -q "$f"; then
          echo "::error::contract omit $f — create-next-app wajib hasilkan file ini"
          exit 1
        fi
      done
      ;;
  esac
}
```

**Waktu check:** pre-task (validate-task sebelum launch), bukan post-CI. Contract
yang tidak realistis ditolak SEBELUM agent mulai — tidak ada loop remediasi.

---

## Daftar implementasi

| # | Guard | Lokasi | Fail mode |
|---|---|---|---|
| H-01 | Reject agent/* → main | `path-enforcement.yml` (repo aplikasi) | fail-closed exit 1 |
| H-02 | Branch naming lint | `path-enforcement.yml` step scope | fail-closed exit 1 |
| H-03 | Scaffold realism check | `scripts/validate-task.sh` | fail-closed exit 2 |

Semua fail-closed. Tidak ada tambahan tool/dependency.

---

## Verifikasi

1. **H-01:** buat dummy PR agent/* → main di repo aplikasi → CI reject.
2. **H-02:** push branch `agent/planning-xyz` → CI reject (bukan task-id).
3. **H-03:** `validate-task` contract tanpa go.mod/layout.tsx → exit 2.
4. **Regression:** `make verify` masih hijau — guard baru tidak melanggar
   existing test (§68 negative tests).

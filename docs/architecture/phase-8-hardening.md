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

---

## H-04: Architecture Constraints di Role Definition

**Masalah:** Agent mulai eksekusi tanpa membaca arsitektur. Kesalahan Phase 7
(§44 branch flow, §29.6 shared file) semuanya karena section arsitektur tidak
dimuat ke konteks sebelum bekerja.

**Guard:** `.claude/agents/*.md` memuat blok wajib `Architecture Constraints`
yang mendaftar section arsitektur yang HARUS dibaca sebelum mulai:

```markdown
## Architecture Constraints (wajib baca sebelum kerja)

Sebelum mengerjakan task apa pun, baca section berikut dan ikuti:

- §44 Branch Strategy — base branch task, target merge (develop, bukan main)
- §29.6 Shared File Ownership — file apa yang boleh/kewajiban kamu tulis
- §16 Universal Rules — larangan yang berlaku mutlak
- contract yang ditunjuk task (CONTRACT-*)

Pelanggaran terhadap section ini adalah bug, bukan pilihan.
```

**Verifikasi:** `TestAgentFrontmatterFieldsAreSupported` (sudah ada) ditambah
test yang memeriksa tiap writer-role template memuat `Architecture Constraints`.

---

## H-05: Task Spec sebagai Otoritas — validate pre-launch

**Masalah:** Task spec (YAML) adalah source of truth, tapi tidak divalidasi
lengkap sebelum launch. Contract BE-102 forbid go.mod → agent ikut → CI fail
setelah kerja habis.

**Guard:** `scripts/validate-task.sh` diperkuat — selain H-03 scaffold realism,
validasi spec penuh SEBELUM agent mulai:

```bash
# validate-task.sh — pre-launch check (bukan post-CI)
check_base_branch() {
  # ownership.base_branch harus ada di repo target
  git -C "$repo" show-ref --verify "refs/heads/$base" || {
    echo "::error::base_branch $base tidak ada di repo $repo"
    exit 2
  }
}

check_allowed_nonempty() {
  # allowed minimal 1 path; forbidden wajib .claude/** + .task/**
  # (sesuai task.schema.json paths.forbidden allOf)
}
```

**Waktu:** pre-task, di runner `launch-task.sh` sebelum spawn agent. Contract
yang tidak realistis → tolak launch, bukan mulai task yang sudah pasti gagal.

---

## H-06: Pre-flight Checklist di Runner

**Masalah:** Runner `launch-task.sh` menerima task yang sudah tidak valid
(contract missing, branch naming salah) → agent mulai, lalu CI fail.

**Guard:** `scripts/launch-task.sh` tambah pre-flight sebelum spawn worktree:

```bash
preflight() {
  local spec="$1"
  local task_id repo base

  task_id=$(yq '.task.id' "$spec")
  repo=$(yq '.ownership.repository' "$spec")
  base=$(yq '.ownership.base_branch' "$spec")

  # 1. Contract referenced exists
  for c in $(yq '.task.contract_ids[]' "$spec"); do
    [ -f "contracts/$c.yaml" ] || deny "contract $c tidak ada"
  done

  # 2. Base branch ada di repo target
  git -C "../$repo" show-ref --verify "refs/heads/$base" || deny "base branch $base tidak ada"

  # 3. Branch naming: agent/<task-id>-<slug> valid
  [[ "$(yq '.ownership.branch' "$spec")" =~ ^agent/${task_id//-/-}-.*$ ]] \
    || deny "branch name tidak mengikuti pola agent/<task-id>-<slug>"

  # 4. Scaffold realism (H-03)
  validate_scaffold "$spec" || deny "contract omit scaffold-mandatory file"

  echo "preflight lulus — task $task_id siap"
}
```

**Fail mode:** preflight gagal → exit 2, task tidak pernah launch. Agent tidak
pernah mulai pada spec yang sudah invalid.

---

## H-07: TL/SA Gate — contract review sebelum eksekusi

**Masalah:** Phase 7 contract ditulis dan langsung dieksekusi tanpa review
TL/SA dulu. Review jadi tanggung jawab manusia di akhir, bukan gate di awal.

**Guard:** flow wajib dua-gate sebelum implementasi:

```
CONTRACT ditulis (TL/SA)
    → TL/SA review (validasi: realistic, scaffolding-aware, base branch benar)
    → approve (sign-off di control/tasks/status/)
    → baru BE + FE launch
```

**Enforcement:** hook `validate-handoff.sh` (sudah ada, fail-closed) diperluas —
membaca task spec, memeriksa status contract sudah `approved` sebelum subagent
implementation diizinkan berhenti sukses.

```bash
# validate-handoff.sh — tambahan
approved_status() {
  local task_id="$1"
  grep -q "status: approved" "control/tasks/specifications/$task_id.yaml" \
    || deny "contract $task_id belum approved — implementasi tidak boleh mulai"
}
```

**Catatan:** status field ada di task.schema.json (`task.status`). Nilai
`approved` sebagai sign-off TL/SA didokumentasikan di sini; eksekusi
`launch-task` memeriksa status ini.

---

## Daftar implementasi

| # | Guard | Lokasi | Fail mode |
|---|---|---|---|
| H-01 | Reject agent/* → main | `path-enforcement.yml` (repo aplikasi) | fail-closed exit 1 |
| H-02 | Branch naming lint | `path-enforcement.yml` step scope | fail-closed exit 1 |
| H-03 | Scaffold realism check | `scripts/validate-task.sh` | fail-closed exit 2 |
| H-04 | Architecture Constraints di agent template | `.claude/agents/*.md` + test | test fail |
| H-05 | Task spec validate pre-launch | `scripts/validate-task.sh` | fail-closed exit 2 |
| H-06 | Pre-flight checklist runner | `scripts/launch-task.sh` | fail-closed exit 2 |
| H-07 | TL/SA gate — contract approved | `validate-handoff.sh` | fail-closed |

Semua fail-closed. Tidak ada tambahan tool/dependency.

---

## Verifikasi

1. **H-01:** buat dummy PR agent/* → main di repo aplikasi → CI reject.
2. **H-02:** push branch `agent/planning-xyz` → CI reject (bukan task-id).
3. **H-03:** `validate-task` contract tanpa go.mod/layout.tsx → exit 2.
4. **H-04:** test agent template — writer-role tanpa `Architecture Constraints` → test fail.
5. **H-05:** `validate-task` spec dengan base_branch tidak ada → exit 2.
6. **H-06:** `launch-task` spec tanpa contract referenced → preflight deny exit 2, agent tidak start.
7. **H-07:** task spec status bukan `approved` → subagent implementation diblokir.
8. **Regression:** `make verify` masih hijau — guard baru tidak melanggar
   existing test (§68 negative tests).

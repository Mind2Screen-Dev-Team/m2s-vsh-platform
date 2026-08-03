# Prompt untuk Session Baru — Implementasi Phase 8 Hardening

Salin blok di bawah ini ke session Claude Code baru untuk melanjutkan pekerjaan
Phase 8 (implementasi hardening H-01..08).

---

## Posisi awal

M2S-VSH control repo. Phase 5 (§61) dan Phase 7 (§63) selesai dan ter-merge.
Phase 7 menyelesaikan multi-repo pilot (backend + frontend paralel, 9 PR).
Phase 8 blueprint sudah di-merge sebagai dokumen (PR #21).

## Yang sudah ada di main control

| Komponen | Status |
|---|---|
| Phase 7 pipeline | ✅ 9 PR merged, endpoint + CORS live |
| `docs/architecture/phase-8-hardening.md` | ✅ blueprint H-01..H-08 (merged, PR #21) |
| Task specs | CONTRACT-102 approved, BE/BE-fix/FE/QA completed |
| Backlog PILOT-1 | done |
| README roadmap | Phase 7 ✅, Phase 6 ⬜ (dilewati D-P7-2), Phase 8 ➡️ berikutnya |

Repos dalam scope:
- `Mind2Screen-Dev-Team/m2s-vsh-platform` (control)
- `Mind2Screen-Dev-Team/m2s-vsh-project-backend` (Go — punya endpoint, CORS)
- `Mind2Screen-Dev-Team/m2s-vsh-project-frontend` (Next.js — punya StatusCard)

## Fase: Implementasi Hardening H-01..08

Dokumen sumber: `docs/architecture/phase-8-hardening.md` (sudah di main).

### Delapan guard yang harus diimplementasi ke kode nyata

| # | Guard | Target file |
|---|---|---|
| H-01 | `agent/*` wajib target `develop` (staging/main ditolak) | `templates/github/workflows/path-enforcement.yml` + turunan 3 repo |
| H-02 | Branch naming lint — `agent/*` = task, `worktree-*` = planning | `path-enforcement.yml` step scope |
| H-03 | Scaffold realism — contract forbid go.mod/layout.tsx → reject pre-task | `scripts/validate-task.sh` |
| H-04 | Architecture Constraints wajib di agent template | `.claude/agents/*.md` + test `internal/contract/agents_test.go` |
| H-05 | Task spec validate pre-launch (base branch, allowed) | `scripts/validate-task.sh` |
| H-06 | Pre-flight checklist runner | `scripts/launch-task.sh` (atau `internal/.../runner`) |
| H-07 | TL/SA gate — contract `approved` sebelum implementasi | `.claude/hooks/validate-handoff.sh` |
| H-08 | Branch promotion naik satu level (develop→main tanpa staging ditolak) | `path-enforcement.yml` |

### Prinsip implementasi

1. **Ikuti dokumen blueprint persis** — kode contoh sudah ada di `phase-8-hardening.md`,
   terjemahkan ke file nyata.
2. **Fail-closed semua** — exit 2 untuk guard path/task, exit 1 untuk CI.
3. **Tanpa tool baru** — hanya shell, Go, workflow yaml. Jangan tambah dependency.
4. **`make verify` tetap hijau** — guard baru tidak boleh melanggar existing test (§68).
5. **Worktree pattern `worktree-*` untuk branch non-task** — planning/docs/implementasi
   hardening pakai `worktree-phase-8-*`, bukan `agent/*`.
6. **Sync template → deploy** — `path-enforcement.yml` template kanonik ada di
   `templates/github/workflows/`, turunkan ke `main` control + 2 repo aplikasi.
   `make verify-github` memeriksa keselarasan.
7. **PR per guard atau per grup logis** — jangan satu PR raksasa; tiap PR CI pass
   dan mergeable.

### Urutan yang disarankan

1. **H-02** dulu (branch naming lint) — murah, mencegah kesalahan naming di task berikutnya.
2. **H-01 + H-08** (branch flow) — satu paket di path-enforcement.yml.
3. **H-03 + H-05** (validate-task) — contract realism + pre-launch.
4. **H-06** (launch-task pre-flight).
5. **H-07** (validate-handoff gate).
6. **H-04** (agent template + test) — paling invasive, terakhir.
7. `make verify` full — regresi hijau.

### Acceptance criteria Phase 8

- H-01..08 terimplementasi ke file nyata (bukan hanya dokumen).
- `make verify` hijau penuh (format, vet, test, wrapper, schema, agents, hook, §68).
- Template workflow di 3 repo sinkron (`make verify-github` pass).
- Tiap guard punya verifikasi: CI reject / exit 2 / test fail.
- Catat pengukuran baseline biaya token per role (Phase 8 tujuan §64).

## Caveat

- Branch app repo tetap hanya `main` saat ini — implementasi H-01/H-08 akan
  membuat guard yang **menolak agent→main**. Pastikan ada `develop` branch di
  repo backend + frontend SEBELUM guard aktif, atau guard akan blokir semua
  PR task. **Buat `develop` dari `main` dulu** (sinkron), lalu aktifkan guard.
- `git worktree`, `git checkout`, `git reset --hard`, `rm -rf` denied untuk agent.
  Branch switch manual (user) atau symbolic-ref workaround untuk kerja di repo lain.
- `.claude/**`, `Makefile`, `cmd/m2s/**` human-only write — agent tidak boleh edit.
  Perubahan settings/agents/templates via PR ke main, merge oleh manusia.
- Merge flow harus `agent/* → develop → staging → main` (H-01/H-08). PR ke main
  langsung ditolak. Merge oleh manusia di main.

## Dokumen rujukan

- `docs/architecture/phase-8-hardening.md` — blueprint H-01..08 (otoritas)
- `docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md` — §44 branch, §63, §64
- `docs/operator/phase-5-cleanup-and-phase-7-prep.md` — pola fase sebelumnya
- `docs/decisions/component-inventory.md` — ownership (path-enforcement.yml owner DevOps)
- `templates/github/workflows/path-enforcement.yml` — template kanonik workflow
- `internal/contract/agents_test.go` — test agent boundary
- `schemas/task.schema.json` — task contract format

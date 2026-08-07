# Handoff: Implementasi Otomasi Status Task + Auto-spawn Review/QA (ADR-011 + ADR-012)

**Dibuat:** 2026-08-07
**Sumber:** ADR-011-task-status-automation.md + ADR-012-autospawn-review-qa.md (draft PR #52)
**Status:** menunggu persetujuan ADR-011 + ADR-012 sebelum implementasi
**PR draft:** https://github.com/Mind2Screen-Dev-Team/m2s-vsh-platform/pull/52

> **Baca dulu:** ADR-011 (`docs/adr/ADR-011-task-status-automation.md`) dan ADR-012
> (`docs/adr/ADR-012-autospawn-review-qa.md`). ADR-012 **bergantung** pada ADR-011
> (prasyarat status hidup). Kalau salah satu belum approved, jangan implementasi —
> tunggu persetujuan dulu.

---

## Konteks

Status task §33 (23 nilai `taskStatus`) tidak pernah ditulis sebagai file
`control/tasks/status/<id>.yaml`. Audit menemukan:

- `grep "control/tasks/status\|task-status"` di `scripts/` dan `cmd/` = nol hasil.
  CLI `m2s` cuma MEMBACA status (gate H-07 cek `technical-ready`), tidak pernah menulis.
- `control/tasks/status/` (writable path PM §17.7) tak pernah terisi.
- Transisi hidup task (`running → implementation-complete → reviewing → qa-testing →
  merge-ready → merged → released`) tak punya penulis.
- Yang otomatis hanya reservasi (`internal/registry/`): `active → reserved-pending-merge → released`.

Yang SUDAH ada dan bisa disambung:
- `handoffStatus` enum di `schemas/common.schema.json`: `implementation-complete`,
  `changes-requested`, `defect-found`, `blocked`, `failed`. Agent sudah melaporkan ini di handoff.
- `writerRole` enum: subset role yang boleh memegang reservasi.
- Reservasi sudah otomatis via `m2s` CLI.

---

## Keputusan ADR-011 (ringkas)

**Hybrid A+B**: status ditulis oleh mekanisme yang menyelesaikan tahap, bukan PM mencatat.

1. **Runner (A)** tulis status deterministic di titik `m2s` yang sudah ada:
   - `reserve-paths` sukses → `reserved`
   - `launch-task` sukses → `running`
   - `collect-result` dengan PR → `reviewing`
   - `release-reservation` → `released`
   - merge event → `merged`
2. **Agent (B)** tulis status judgement via subcommand baru `m2s update-status`.
3. **Satu owner per status** (prinsip #4). Tabel owner di ADR-011.
4. **Sinkronkan reservasi ↔ §33**: `active→reserved`, `reserved-pending-merge→merge-ready`,
   `released→released`, `cancelled→cancelled`.

---

## Scope pekerjaan

**Dua ADR, dua fase implementasi (urutan wajib):**
- **Fase 1 = ADR-011**: status YAML hidup + sinkron reservasi. (Bagian 1-5 bawah)
- **Fase 2 = ADR-012**: auto-spawn review/QA, **bergantung** Fase 1. (Bagian 6-7 bawah)

### 1. Subcommand `m2s update-status` (utama, baru)

Di `cmd/m2s/commands.go` + `cmd/m2s/main.go`:

```
m2s update-status --task <id> --status <taskStatus> [--reason <teks>]
```

Harus:
- Validasi `<status>` terhadap enum `taskStatus` (`schemas/common.schema.json`).
- Validasi transisi dari status saat ini terhadap state machine §33 (dari contract
  `control/tasks/specifications/<id>.yaml`).
- Validasi **owner**: role agent yang menulis (dari identitas App, pola
  `scripts/gh-app-token.sh`) harus diizinkan menulis status itu (tabel owner).
- Menulis `control/tasks/status/<id>.yaml`:
  ```yaml
  schema_version: "1.0"
  task_id: <id>
  status: <taskStatus>
  updated_at: <RFC3339+zone>
  by: <role>
  reason: <teks>       # opsional
  ```
- Mengikuti format `docs/operator/task-status.md` yang sudah ada.

Contoh pemakaian (dari dokumentasi ADR-011):
```bash
m2s update-status --task BE-301 --status defect-found --reason "test gagal: timeout 5s"
```

### 2. Runner menulis status deterministic

Di `cmd/m2s/commands.go`, pada titik yang sudah ada:
- `cmdReservePaths` sukses → tulis status `reserved`.
- `cmdLaunchTask` sukses → tulis status `running`.
- `cmdCollectResult` dengan `prURL` → tulis status `reviewing`.
- `cmdReleaseReservation` → tulis status `released` / `cancelled`.

Tulis file yang sama dengan format di atas. Ini tidak mengubah perilaku agent — runner
sudah ada di titik-titik itu.

### 3. Sinkronkan reservasi ↔ §33

Di `internal/registry/registry.go`, mapping saat `Transition`:

| Reservasi | taskStatus |
|---|---|
| `active` | `reserved` |
| `reserved-pending-merge` | `merge-ready` |
| `released` | `released` |
| `cancelled` | `cancelled` |

Dokumentasikan mapping ini di ADR-011 + doc arsitektur.

### 4. Izin agent panggil binary

- `.claude/agents/*.md` — tambah izin agent panggil `m2s update-status` (per role,
  sesuai tabel owner).
- `settings.json`/permissions — pastikan binary `m2s` bisa dipanggil agent.

### 5. Test

- Unit test untuk `update-status`: status valid/invalid, transisi diizinkan/ditolak,
  owner benar/salah.
- Test untuk penulisan status di `reserve-paths`/`launch-task`/`collect-result`/
  `release-reservation`.
- `make test` + `make verify` hijau.

---

## File yang terdampak

| File | Perubahan |
|---|---|
| `cmd/m2s/commands.go` | Subcommand baru (`update-status`, `launch-review`, `launch-qa`) + tulis status deterministic |
| `cmd/m2s/main.go` | Daftarkan `update-status`, `launch-review`, `launch-qa` |
| `internal/registry/registry.go` | Mapping reservasi ↔ §33; reservasi role review/QA |
| `schemas/common.schema.json` | Mungkin tambah enum/schema status file (validasi) |
| `.claude/agents/*.md` | Izin agent panggil `m2s` (per role, tabel owner) |
| `docs/operator/task-status.md` | Update format + owner + subcommand |
| `cmd/m2s/*_test.go` | Test baru |

### 6. Runner baru `m2s launch-review` (ADR-012)

Di `cmd/m2s/commands.go` + `cmd/m2s/main.go`:

```
m2s launch-review --task <id> [--repo <path>]
```

Harus:
- **Gate**: status task = `implementation-complete` (dari `control/tasks/status/<id>.yaml`).
  Jangan spawn kalau status beda.
- Reservasi role `code-reviewer` (read-only, tanpa writable path — A-03).
- Worktree dari branch PR → review diff terhadap `develop`.
- Spawn code-reviewer; contract review di-ekstrak dari handoff implementer.
- Baca handoff reviewer (`decision`, `findings`) → **runner tulis status**:
  - `approve` → `reviewing`
  - `request-changes` → `changes-requested`

### 7. Runner baru `m2s launch-qa` (ADR-012)

```
m2s launch-qa --task <id> [--repo <path>]
```

Harus:
- **Gate**: status = `reviewing`.
- Reservasi role `qa-engineer` (writable, `quality_gates` di contract).
- Worktree dari branch PR → jalankan quality_gates + acceptance criteria + regresi.
- Baca handoff QA (`findings`) → **runner tulis status**:
  - pass → `qa-testing` → `ci-passed` → `merge-ready`
  - defect → `defect-found`

### 8. Trigger otomatis (ADR-012)

`implementation-complete` → auto `launch-review` → `reviewing` → auto `launch-qa` →
`merge-ready`. Trigger = runner orchestration (bukan GitHub Actions), sesuai ADR-012
keputusan #3. Fix (`changes-requested`/`defect-found`) → status `running` → implementer
lanjut di worktree sama (§32 langkah 16).

---

## Kendala / catatan

- `cmd/m2s/**` dan `Makefile` di-deny Edit/Write di `.claude/settings.json`.
  Sesi ini (atau user) harus izinkan edit ke path itu, atau kerja lewat task contract
  dengan role yang berhak.
- Deny `git checkout`/`switch`/`worktree` di settings.json **dipertahankan** (keputusan
  user 2026-08-07). Agent tak boleh pindah branch — isolasi worktree (prinsip #3).
  Kerja branch lewat jalur yang diizinkan (worktree runner / `!` manual), bukan melemahkan guard.
- ADR-012 menuntut ADR-011 approved dulu — implementasi dua fase, jangan digabung.
- Merge ke `main` tetap milik manusia.

---

## Definition of Done

**Fase 1 (ADR-011):**
- [ ] `m2s update-status` jalan, validasi status + transisi + owner.
- [ ] Status `reserved`/`running`/`reviewing`/`released` ditulis otomatis runner.
- [ ] Reservasi ↔ §33 sinkron (mapping didokumentasikan).
- [ ] `make test` + `make verify` hijau.
- [ ] `control/tasks/status/<id>.yaml` terisi untuk task yang sudah jalan.

**Fase 2 (ADR-012):**
- [ ] `m2s launch-review` jalan, gate `implementation-complete`, tulis `reviewing`/`changes-requested`.
- [ ] `m2s launch-qa` jalan, gate `reviewing`, tulis `merge-ready`/`defect-found`.
- [ ] Trigger otomatis `implementation-complete → review → QA` (runner orchestration).
- [ ] Fix findings → `running` → implementer lanjut worktree sama.
- [ ] `make test` + `make verify` hijau.

**Semua:**
- [ ] ADR-011 + ADR-012 approved, PR #52 merged.

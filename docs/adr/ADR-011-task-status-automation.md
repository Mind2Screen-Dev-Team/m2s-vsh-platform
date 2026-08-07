# ADR-011: Otomasi status task §33 (hybrid runner + agent)

**Tanggal:** 2026-08-07
**Decider:** owner arsitektur (Mindtoscreen)
**Status:** proposed (draft untuk review)

## Context

Status task §33 (23 nilai taskStatus) didefinisikan penuh di
`schemas/common.schema.json` dan `docs/architecture/...-v0.1.0-Architecture.md`,
tetapi **tidak pernah ditulis sebagai file** `control/tasks/status/<id>.yaml`.
Audit 2026-08-07 menunjukkan:

1. `grep "control/tasks/status\|task-status"` di `scripts/` dan `cmd/` → **nol hasil**.
   `m2s` CLI cuma **membaca** status (gate H-07 cek `technical-ready`), tidak pernah menulis.
2. `control/tasks/status/` (writable path PM, §17.7) tak pernah terisi — riwayat task
   Phase 7 hilang, `docs/operator/task-status.md` mengakui ini.
3. Transisi hidup task (`running → implementation-complete → reviewing → qa-testing →
   merge-ready → merged → released`) **tidak punya penulis**. Semua konseptual, jalan di
   kepala PM + git history.
4. Yang benar-benar berjalan otomatis hanya **reservasi** (`internal/registry/`):
   `active → reserved-pending-merge → released`. Tapi state ini memakai istilah sendiri
   (`reserved-pending-merge`) yang **tidak ada** di enum 23 §33 — dua bahasa state parallel
   yang tak sinkron.

Sementara itu, **dua struktur yang dibutuhkan otomasi sudah ada**:
- `handoffStatus` enum di schema: `implementation-complete`, `changes-requested`,
  `defect-found`, `blocked`, `failed`. Agent **sudah** melaporkan status lewat handoff
  (Q9, §35) — runner tinggal menulisnya.
- `writerRole` enum: subset role yang boleh memegang reservasi.

## Decision

Terapkan **hybrid A+B**: status ditulis oleh mekanisme yang menyelesaikan tahap,
bukan PM yang mencatat. Runner menulis transisi deterministic; agent menulis transisi
judgement lewat subcommand baru `m2s update-status`.

### 1. Runner menulis status deterministic (Opsi A)

`m2s` CLI menulis file status otomatis di titik yang sudah deterministic:

| Titik | Status yang ditulis |
|---|---|
| `reserve-paths` sukses | `reserved` |
| `launch-task` sukses | `running` |
| CI hijau (bila runner diintegrasikan CI) | `ci-passed` |
| `collect-result` dengan PR | `reviewing` |
| merge event | `merged` |
| `release-reservation` | `released` |

Ini tidak butuh ubah perilaku agent — runner sudah ada di titik-titik ini.

### 2. Agent menulis status judgement lewat `m2s update-status` (Opsi B)

Subcommand baru:

```
m2s update-status --task <id> --status <taskStatus> [--reason <teks>]
```

- Memvalidasi `<status>` terhadap enum taskStatus + transisi yang diizinkan dari status
  saat ini (§33 state machine).
- Memvalidasi **siapa** yang menulis: role agent di `.task/` menentukan status mana yang
  boleh dia tulis (tabel owner di bawah).
- Menulis `control/tasks/status/<id>.yaml` dengan `updated_at`, `by`, `reason`.
- Agent panggil lewat runner dengan identitas App-nya (pola sama `gh-app-token.sh`),
  bukan langsung — konsisten ADR-004 dan blocklist Bash.

### 3. Owner per status — satu penulis, prinsip #4

| Status | Penulis |
|---|---|
| `draft`, `technical-ready` | **TL/SA** |
| `needs-business-clarification`, `analysis-ready`, `technical-analysis`, `needs-technical-clarification` | **PM** |
| `reserved`, `running`, `ci-passed`, `merged`, `released` | **Runner** (otomatis) |
| `implementation-complete` | **Worker/implementer** |
| `reviewing` | **Runner** (event PR) |
| `changes-requested` | **Code Reviewer** |
| `qa-testing` | **Runner/QA agent** |
| `defect-found` | **QA Engineer** |
| `merge-ready` | **QA/PM** |
| `documented` | **Technical Writer** |
| `staging-verified` | **DevOps** |
| `cancelled`, `failed`, `superseded`, `blocked` | **PM** + Runner |

Ini menegakkan prinsip #6 (implementer tidak menulis `qa-testing`/`merged`) dan §30
(worker tidak melepas reservasi).

### 4. Sinkronkan reservasi dengan §33

Mapping reservasi → taskStatus, ditulis otomatis oleh runner di titik yang sama:

| Reservasi | taskStatus |
|---|---|
| `active` (setelah reserve-paths) | `reserved` |
| (worktree dibuat) | `running` |
| `reserved-pending-merge` | `merge-ready` |
| `released` | `released` |
| `cancelled` | `cancelled` |

Ini menutup celah dua-bahasa-state: reservasi jadi mekanisme, §33 jadi bahasa.

### 5. Backlog → Issue optional (belakangan)

Otomasi status YAML adalah prasyarat untuk integrasi GitHub Projects (ADR-010, pending).
Setelah YAML status hidup, `sync-project` tinggal membaca file → Projects. Tidak ada
keputusan baru di sini; ADR-010 tetap valid.

## Rationale

### Mengapa runner + agent, bukan hanya runner

Runner hanya tahu titik deterministic (launch, merge). Transisi judgement —
`changes-requested`, `defect-found`, `documented`, `staging-verified` — hanya agent
yang tahu. Agent sudah melaporkannya di handoff; `update-status` tinggal memindahkan
laporan itu ke file status yang bisa diverifikasi.

### Mengapa `m2s update-status`, bukan agent tulis file langsung

Prinsip #4 satu penulis. Kalau agent tulis YAML langsung, siapa yang memvalidasi
transisi + owner? Bisa dilanggar. `m2s update-status` menjadi satu gerbang: validate
status valid + transisi diizinkan + role berhak. Agent tidak pernah pegang format file.

### Mengapa handoff sebagai input

Agent sudah mengisi `status: implementation-complete` di handoff (Q9). Runner
`collect-result` sudah memvalidasinya. Menulis status YAML dari handoff = hampir gratis,
data sudah ada. Tidak perlu agent tambahan.

### Alternatif yang ditolak

| Opsi | Alasan ditolak |
|---|---|
| PM menulis semua status manual | Telah terbukti gagal (Phase 7) — tak ada yang menulis |
| Agent tulis YAML langsung | Melanggar prinsip #4; tak ada validasi transisi/owner |
| GitHub Actions sync saja (tanpa runner) | Tak menyelesaikan akar — YAML status tetap kosong |
| Satu status global, tanpa owner | Melanggar prinsip #6 dan §30 |

## Consequences

- **Positif**: status task ter-persist dan dapat diaudit; transisi hidup punya penulis;
  reservasi ↔ §33 sinkron; fondasi untuk ADR-010 (Projects).
- **Biaya**: subcommand baru `update-status` + validasi transisi + owner di runner;
  izin agent panggil binary; test baru.
- **Risiko**: `update-status` bisa disalahgunakan agent (role → status di luar owner).
  Mitigasi: validasi owner di runner, bukan di prompt — enforce lewat kode.
- **Yang tidak diputuskan**: apakah CI `ci-passed` diintegrasikan runner atau tetap
  event-driven; apakah `documented`/`staging-verified` butuh runner terpisah atau cukup
  agent; kapan ADR-010 (Projects) aktif.

## Backward compatibility / rollback

- Tidak mengubah perilaku agent yang ada. `reserve-paths`/`launch-task`/`collect-result`
  tetap jalan; hanya bertambah menulis file status.
- Rollback: nonaktifkan penulisan status di runner + hapus subcommand. File YAML lama
  tidak rusak.
- Handoff schema tidak berubah.

**Menimpa:** tidak ada ADR sebelumnya yang menimpa. Mengisi gap writable path
`control/tasks/status/**` yang kosong. Relevan: ADR-004 (#3 reservasi), ADR-010
(integrasi Projects, pending).

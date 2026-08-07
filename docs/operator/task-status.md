# Task Status — format persist status task

**Owner:** Runner + role per tabel owner (ADR-011; path `control/tasks/status/**`)
**Otoritas nilai:** `schemas/common.schema.json` `$defs.taskStatus`
**Validasi:** `schemas/task-status.schema.json` (Kind validator `m2s`)
**Status:** aktif, 2026-08-07 (ADR-011)

## Mengapa ada

`control/tasks/status/` tak pernah terisi — empat task Phase 7 berjalan tanpa
satu pun jejak status tersimpan. Akibatnya riwayat task hilang dan tidak ada
yang bisa membedakan "task berstatus apa" setelah selesai. ADR-011 mengotomasi
penulisan: status ditulis oleh mekanisme yang menyelesaikan tahap, bukan PM yang
mencatat.

## Format

Satu berkas per task: `control/tasks/status/<task-id>.yaml`. `<task-id>` = nilai
`task.id` dari spec (`BE-101`, `FE-102`, dst). Berkas adalah **snapshot** status
terakhir, bukan riwayat — riwayat transisi dicatat schema terpisah
`task-state.schema.json` (dokumen referensi).

```yaml
# <task-id> — status task
schema_version: "1.0"
task_id: BE-101
status: merge-ready
updated_at: "2026-08-04T10:15:00+07:00"
by: backend-engineer
reason: "QA lulus, siap merge"   # opsional
```

| Field | Wajib | Isi |
|---|---|---|
| `schema_version` | ya | `"1.0"` |
| `task_id` | ya | sama dengan `<task-id>` pada nama berkas |
| `status` | ya | satu nilai `taskStatus` (daftar di bawah) |
| `updated_at` | ya | RFC 3339 dengan zona |
| `by` | ya | role penulis — selalu nilai enum `role`, tak pernah `runner` |

Properti tambahan diizinkan (mis. `pr_url`, `reason`) tetapi tidak dibakukan.

## Nilai status (taskStatus)

Dari `schemas/common.schema.json` `$defs.taskStatus`, 23 nilai:

`draft`, `needs-business-clarification`, `analysis-ready`, `technical-analysis`,
`needs-technical-clarification`, `technical-ready`, `reserved`, `running`,
`implementation-complete`, `reviewing`, `changes-requested`, `qa-testing`,
`defect-found`, `ci-passed`, `merge-ready`, `merged`, `documented`,
`staging-verified`, `released`, `cancelled`, `failed`, `superseded`, `blocked`

Transisi mengikuti state machine §33 arsitektur, ditegakkan `m2s update-status`.
Gate H-07 (Phase 8) memakai `technical-ready` sebagai tanda sign-off TL/SA
sebelum launch.

## Siapa menulis (tabel owner, ADR-011)

Satu penulis per status (prinsip #4). Runner menulis status deterministic pada
titik yang sudah ada; agent menulis status judgement lewat `m2s update-status`.

| Status | Penulis |
|---|---|
| `draft`, `technical-ready` | technical-lead-system-analyst |
| `needs-business-clarification`, `analysis-ready`, `technical-analysis`, `needs-technical-clarification` | project-manager |
| `reserved`, `running`, `ci-passed`, `merged`, `released` | runner (by = role pemilik task) |
| `implementation-complete` | role pemilik task (implementer) |
| `reviewing` | runner |
| `changes-requested` | code-reviewer (runner tulis atas nama) |
| `qa-testing` | qa-engineer |
| `defect-found` | qa-engineer |
| `merge-ready` | qa-engineer / project-manager |
| `documented` | technical-writer |
| `staging-verified` | devops-release |
| `cancelled`, `failed`, `superseded`, `blocked` | project-manager / runner |

Implementer tidak boleh menulis `qa-testing`/`merged` (prinsip #6). Status
runner-owned (`reserved`, `running`, `reviewing`, `ci-passed`, `merged`,
`released`) menolak semua agent — satu-satunya penulisnya runner.

## Subcommand `m2s update-status`

Agent menulis status judgement lewat subcommand (ADR-011 opsi B):

```bash
m2s update-status --task <id> --status <taskStatus> --by <role> [--reason <teks>]
```

Validasi tiga lapis, ditegakkan di kode bukan prompt:
1. `<status>` anggota enum `taskStatus` (schema).
2. Transisi sah dari status saat ini (state machine §33).
3. `<by>` berhak atas `<status>` (tabel owner di atas).

Runner menulis status deterministic tanpa subcommand — pada `reserve-paths`
(`reserved`), `launch-task` (`running`), `collect-result` (status dari handoff),
dan `release-reservation` (`released`/`cancelled`). Sinkronisasi reservasi ↔ §33:
`active→reserved`, `reserved-pending-merge→merge-ready`, `released→released`,
`cancelled→cancelled`.

## Contoh

Setelah task BE-101 di-launch (ditulis runner):

```yaml
schema_version: "1.0"
task_id: BE-101
status: running
updated_at: "2026-08-04T10:15:00+07:00"
by: backend-engineer
```

Implementer menyatakan selesai:

```bash
m2s update-status --task BE-101 --status implementation-complete --by backend-engineer
```

## Catatan

- Ini bukan `schemas/reservation.schema.json` — reservasi adalah hak path,
  status task adalah lifecycle. Keduanya terpisah.
- `schemas/task-state.schema.json` mencatat transisi (`from`/`to`/`at`) sebagai
  dokumen referensi; berkas status di sini adalah snapshot terkini. Keduanya
  tidak saling menggantikan.

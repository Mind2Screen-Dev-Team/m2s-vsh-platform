# Task Status — format persist status task

**Owner:** Project Manager (§17.7 writable path `control/tasks/status/**`)
**Otoritas nilai:** `schemas/common.schema.json` `$defs.taskStatus`
**Status:** aktif, 2026-08-04 (C1 backlog Phase 8)

## Mengapa ada

`control/tasks/status/` adalah writable path PM tapi tak pernah terisi — empat
task Phase 7 berjalan tanpa satu pun jejak status tersimpan. Akibatnya riwayat
task hilang dan tidak ada yang bisa membedakan "task berstatus apa" setelah
selesai. Dokumen ini membakukan formatnya supaya status task **ter-persist**.

## Format

Satu berkas per task: `control/tasks/status/<task-id>.yaml`. `<task-id>` = nilai
`task.id` dari spec (`BE-101`, `FE-102`, dst).

```yaml
# <task-id> — status task
schema_version: "1.0"
task_id: BE-101
status: merge-ready
updated_at: "2026-08-04T10:15:00+07:00"
by: project-manager
```

| Field | Wajib | Isi |
|---|---|---|
| `schema_version` | ya | `"1.0"` |
| `task_id` | ya | sama dengan `<task-id>` pada nama berkas |
| `status` | ya | satu nilai `taskStatus` (daftar di bawah) |
| `updated_at` | ya | RFC 3339 dengan zona |
| `by` | ya | role yang menulis (`project-manager`, `technical-lead-system-analyst`) |

Properti tambahan diizinkan (mis. `pr_url`, `reason`) tetapi tidak dibakukan.

## Nilai status (taskStatus)

Dari `schemas/common.schema.json` `$defs.taskStatus`, 23 nilai:

`draft`, `needs-business-clarification`, `analysis-ready`, `technical-analysis`,
`needs-technical-clarification`, `technical-ready`, `reserved`, `running`,
`implementation-complete`, `reviewing`, `changes-requested`, `qa-testing`,
`defect-found`, `ci-passed`, `merge-ready`, `merged`, `documented`,
`staging-verified`, `released`, `cancelled`, `failed`, `superseded`, `blocked`

Kapan nilai mana dipakai diikuti state machine §33 arsitektur. Gate H-07 (Phase 8)
memakai `technical-ready` sebagai tanda sign-off TL/SA sebelum launch.

## Siapa menulis

**Project Manager** menulis status di `control/tasks/status/`. PM punya tool
Write pada path itu (§17.7); tidak ada subcommand runner khusus — menulis
berkas langsung dengan format di atas.

TL/SA menyetujui (`technical-ready`), PM yang mencatat.

## Contoh

Setelah task BE-101 di-launch:

```yaml
schema_version: "1.0"
task_id: BE-101
status: running
updated_at: "2026-08-04T10:15:00+07:00"
by: project-manager
```

Setelah merge:

```yaml
schema_version: "1.0"
task_id: BE-101
status: merged
updated_at: "2026-08-04T16:00:00+07:00"
by: project-manager
pr_url: "https://github.com/Mind2Screen-Dev-Team/m2s-vsh-project-backend/pull/13"
```

## Catatan

- Ini bukan `schemas/reservation.schema.json` — reservasi adalah hak path,
  status task adalah lifecycle. Keduanya terpisah.
- Belum ada schema JSON utk status task. Bila mekanisme jadi subcommand runner
  di masa depan, tambahkan `schemas/task-status.schema.json` + validasi. Untuk
  C1, format doc cukup.

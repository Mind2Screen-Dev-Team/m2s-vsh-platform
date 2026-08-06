---
name: project-start
description: Titik masuk pengembangan baru. Ketika user memberikan deskripsi project, link design, atau requirement awal — skill ini mengorkestrasi PM agent untuk melakukan structured interview, menulis task contract YAML yang valid, dan memandu user ke m2s launch-task. Gunakan skill ini setiap kali user ingin memulai fitur/project baru tanpa perlu tahu urutan agent yang harus dipakai.
---

# Project Start

Skill ini adalah **titik masuk pengembangan** di M2S-VSH. Ia menerima input apapun dari
user (deskripsi ide, link design, requirement, atau brief singkat) lalu mengorkestrasi
alur menuju task contract yang siap dijalankan `m2s launch-task`.

Gunakan skill ini ketika user:
- Ingin memulai fitur atau project baru
- Memberikan link design dan minta dianalisa untuk pengembangan
- Memberikan requirement atau brief dan minta diturunkan ke task
- Bertanya "dari mana mulai?" atau "bagaimana cara mulai?"

## Filosofi

Skill ini **tidak menggantikan** dokumen pre-development (`scripts/project-kickoff.sh`).
Ia adalah jalur cepat ketika user sudah punya gambaran yang cukup dan ingin langsung
ke task contract. Jika project belum punya dokumen Discovery/BRD/PRD sama sekali,
rekomendasikan `project-kickoff.sh` dulu — tapi jangan blokir jika user mau lanjut.

## Alur Kerja

### Fase 1 — Intake & Klarifikasi

Lakukan structured interview singkat. Kumpulkan informasi berikut (boleh digabung,
tidak harus satu per satu jika user sudah memberi cukup konteks):

1. **Nama project / fitur** — apa yang akan dibangun
2. **Stack teknologi** — frontend, backend, mobile, fullstack, atau kombinasi
3. **Scope** — apa yang IN dan apa yang OUT untuk task ini
4. **Role yang dibutuhkan** — siapa yang kerjakan (BE engineer, FE engineer, TL/SA, dll)
5. **Repo target** — nama repo GitHub yang akan disentuh agent
6. **Path yang boleh disentuh** — file/direktori mana saja yang boleh diubah agent
7. **Acceptance criteria** — kapan task dianggap selesai (minimal 1 kriteria konkret)
8. **Quality gate** — perintah verifikasi (`make verify`, `npm test`, dll)

Jika user memberikan link design: analisa dulu designnya, ekstrak komponen, API yang
dibutuhkan, dan flow utama — gunakan itu sebagai bahan intake.

Jika user memberikan dokumen pre-dev (Discovery/BRD/PRD dari `control/pre-dev/`):
baca dokumen itu sebagai sumber intake, minimalkan pertanyaan ulang.

Jika informasi belum cukup untuk menulis contract yang valid: tanyakan poin yang kurang.
Jangan mengarang path atau acceptance criteria.

### Fase 2 — Tulis Task Contract

Setelah intake cukup, tulis task contract YAML ke
`control/tasks/specifications/<TASK-ID>.yaml`.

**Aturan TASK-ID:**
- Format: `^[A-Z][A-Z0-9]*-[0-9]+$` (contoh: `BE-301`, `FE-301`, `PM-101`)
- Prefix sesuai domain: `BE` backend, `FE` frontend, `PM` project-management,
  `CONTRACT` contract-change, `DESIGN` design, `QA` QA, `WIRE` wireframe/UI
- Nomor: lanjutkan dari ID tertinggi yang ada di `control/tasks/specifications/`

**Format contract (wajib valid terhadap `schemas/task.schema.json`):**

```yaml
schema_version: "1.0"

task:
  id: <TASK-ID>
  title: "<judul singkat, imperatif>"
  type: <lihat enum di bawah>
  project: <nama project>
  status: draft

ownership:
  role: <writer role>
  repository: <nama repo GitHub>
  base_branch: develop
  branch: agent/<TASK-ID>-<slug-lowercase>

execution:
  isolation: worktree
  background: true
  max_turns: 40
  timeout_minutes: 30

paths:
  allowed:
    - <path/glob yang boleh ditulis>
  forbidden:
    - .claude/**
    - .task/**
    - <path lain yang harus dilindungi>

inputs:
  - <file referensi yang harus dibaca agent>

acceptance_criteria:
  - "<kriteria konkret, verifiable>"

quality_gates:
  - <perintah gate, contoh: make verify>

outputs:
  - pull-request

stop_conditions:
  - "contract change required"
```

**Enum `task.type`:** `backend-implementation`, `frontend-implementation`,
`fullstack-implementation`, `mobile-implementation`, `integration`, `bugfix`,
`refactor`, `test-authoring`, `design`, `documentation`, `devops`, `contract-change`

**Enum `ownership.role` (writer role):** `technical-lead-system-analyst`,
`ui-ux-designer`, `backend-engineer`, `frontend-engineer`, `fullstack-engineer`,
`mobile-engineer`, `android-developer`, `ios-developer`, `devops-release`

**Catatan `base_branch`:** selalu `develop` untuk task agent (§44). `main` hanya
untuk PR yang sudah melalui develop → staging → main. Jangan set ke `main`.

### Fase 3 — Gate Manusia & Petunjuk Lanjut

Setelah contract ditulis:

1. Tampilkan ringkasan contract (task-id, title, role, repo, paths allowed, criteria).
2. Tanyakan apakah ada yang perlu diubah sebelum dilanjutkan.
3. Setelah user setuju, tampilkan perintah untuk menjalankan task:

```bash
# dari akar control repo
./scripts/launch-task.sh <TASK-ID>
```

4. Ingatkan: merge akhir ke `main` tetap dilakukan manusia, bukan agent.

## Aturan Khusus

- **Jangan tulis contract dengan `status: ready` langsung.** Mulai dari `draft`,
  biarkan user yang approve dan ubah ke `ready` sebelum `launch-task`.
- **Jangan set `base_branch: main`.** Task agent selalu target `develop`.
- **`paths.forbidden` wajib memuat `.claude/**` dan `.task/**`** — schema validator
  akan menolak contract tanpa keduanya.
- **Satu task = satu contract.** Jika scope besar, pecah jadi beberapa task (BE + FE
  terpisah, atau fitur A + fitur B terpisah).
- **Jangan spawn agent tanpa task contract.** Skill ini menghasilkan contract terlebih
  dahulu; eksekusi agent dimulai oleh user via `launch-task`, bukan oleh skill ini.

## Contoh Trigger

- "Mulai fitur login, stack Next.js + Go"
- "Ini design dashboard: [link]. Buat task untuk implementasi FE-nya"
- "Saya punya PRD di control/pre-dev/04-prd.md, buat task contract untuk fitur booking"
- "Dari mana saya mulai untuk fitur notifikasi?"
- "Buat contract untuk BE endpoint `/api/v1/users`"

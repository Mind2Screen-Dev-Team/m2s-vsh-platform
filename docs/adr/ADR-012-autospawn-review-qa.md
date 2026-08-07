# ADR-012: Auto-spawn Code Reviewer + QA setelah implementation-complete

**Tanggal:** 2026-08-07
**Decider:** owner arsitektur (Mindtoscreen)
**Status:** proposed (draft untuk review)

## Context

Arsitektur §32 langkah 14-15 menetapkan *"Code Reviewer melakukan read-only review"*
dan *"QA melakukan acceptance/regression test"* — keduanya bagian dari alur end-to-end.
Namun **tidak ada mekanisme yang memicunya**:

1. `grep "launch-qa\|launch-review\|spawn.*qa"` di `scripts/`, `cmd/`, `.claude/` → **nol
   hasil** untuk runner. Yang ada cuma `reserve-paths`, `launch-task`, `collect-result`,
   `release-reservation`.
2. `qa-engineer` dan `code-reviewer` ada sebagai definisi agent
   (`.claude/agents/*.md`, `cmb-agent-review`), tapi **tak dipanggil siapa pun** kecuali
   manual.
3. `collect-result` hanya memvalidasi handoff + memindahkan reservasi ke
   `reserved-pending-merge`. Tidak spawn apa pun.
4. Akibatnya review + QA **manual**, tidak ada yang memaksa, tidak ada jejak status
   (`reviewing`, `qa-testing` — konseptual, tak ditulis, lihat ADR-011).

**Prasyarat:** ADR-011 (proposed) membangun status YAML hidup + trigger transisi.
ADR-012 memakai status `implementation-complete` sebagai **sinyal** untuk memicu
review/QA. Tanpa ADR-011, auto-spawn tak punya pemicu yang bisa diverifikasi.

## Decision

Auto-spawn Code Reviewer dan QA setelah status `implementation-complete`, didorong
oleh runner. Dua runner baru, pola sama `launch-task`.

### 1. Runner baru `m2s launch-review`

```
m2s launch-review --task <id> [--repo <path>]
```

- Membaca status task dari `control/tasks/status/<id>.yaml`.
- **Gate**: hanya jalan bila status = `implementation-complete` (ADR-011 menulisnya).
- Reservasi baru untuk role `code-reviewer` (read-only, tanpa writable path — pola
  `writerRole` di schema mengecualikan code-reviewer karena read-only, A-03).
- Buat worktree dari branch PR (bukan base) → review diff terhadap `develop`.
- Spawn code-reviewer agent, contract review di-ekstrak dari handoff implementer.
- Handoff reviewer (`decision`, `findings`) → **runner tulis status**:
  - `decision: approve` → `reviewing` → menunggu QA
  - `decision: request-changes` → `changes-requested` → kembali ke implementer

### 2. Runner baru `m2s launch-qa`

```
m2s launch-qa --task <id> [--repo <path>]
```

- **Gate**: status = `reviewing` (review approve, belum QA).
- Reservasi untuk role `qa-engineer` (writable, quality_gates di contract).
- Buat worktree dari branch PR → jalankan `quality_gates` di contract + acceptance
  criteria + regresi.
- Handoff QA (`findings`) → **runner tulis status**:
  - pass → `qa-testing` → `ci-passed` → `merge-ready`
  - defect → `defect-found` → kembali ke implementer

### 3. Trigger otomatis (bukan manual call)

Orkestrasi:
```
status: implementation-complete (ditulis runner ADR-011)
  → m2s launch-review auto (spawn code-reviewer)
  → changes-requested? → kembali implementer
  → reviewing (approve)
  → m2s launch-qa auto (spawn qa-engineer)
  → defect-found? → kembali implementer
  → merge-ready
```

Di mana trigger ini hidup? Dua opsi, keputusan di bawah:

| Opsi | Trigger | Pro |
|---|---|---|
| **a. Runner orchestration** | `m2s` CLI yang chain (setelah `collect-result`, auto panggil launch-review) | Satu tempat, testable, tak butuh infra |
| **b. GitHub Actions event** | workflow `on: pull_request` → spawn | Ter-sentral, tak butuh polling |

**Keputusan: opsi (a)** — runner orchestration. Konsisten dengan ADR-004 (runner
tipis, logika di binary yang di-test), tak menambah infra CI, dan sinkron dengan
status YAML yang sudah ditulis runner. Opsi (b) bisa ditambahkan nanti bila butuh
spawn lintas-repo/remote.

### 4. Siapa yang memutuskan hasil review/QA

- **Code Reviewer** → `approve` / `request-changes` (lewat handoff, `decision` field).
- **QA** → pass / defect (lewat handoff, `findings`).
- **Runner** yang menulis status hasilnya ke `control/tasks/status/` — agent tak pegang
  file status (prinsip #4, konsisten ADR-011).

### 5. Fix findings → kembali implementer

`changes-requested` / `defect-found` → status balik ke `running` → implementer lanjut
di worktree sama (bukan spawn baru). Runner yang menulis transisi. Ini sesuai §32
langkah 16 *"Implementer memperbaiki findings melalui task yang sama"*.

## Rationale

### Mengapa runner baru, bukan agent yang spawn agent

Agent tidak boleh spawn agent lain — tidak ada mekanisme itu, dan akan memecah
prinsip #3 (satu task satu worktree satu proses). Runner adalah satu-satunya tempat
yang sudah men-spawn agent (launch-task). Menambah launch-review/launch-qa = perpanjang
pola yang sudah teruji.

### Mengapa gate status, bukan event PR

Status `implementation-complete` adalah sinyal domain yang ditulis runner (ADR-011).
Event PR (`pull_request`) bisa datang sebelum implementer selesai (PR dibuka saat
WIP). Gate status lebih tepat: hanya spawn ketika implementer benar-benar menyatakan
selesai lewat handoff.

### Mengapa fix di worktree sama

§32 langkah 16 eksplisit. Spawn baru untuk fix = buang worktree, kerja ulang.
Lanjut worktree sama = konteks agent terjaga, cost kecil.

### Alternatif yang ditolak

| Opsi | Alasan ditolak |
|---|---|
| Review/QA manual (sekarang) | Tak ada yang memaksa; status kosong; bergantung PM ingat |
| Agent spawn agent | Tak ada mekanisme; pecah prinsip #3 |
| Trigger via PR event saja | PR bisa WIP; sinyal kurang tepat dari status |
| GitHub Actions sebagai satu-satunya trigger | Tambah infra CI; runner sudah punya konteks |

## Consequences

- **Positif**: review + QA otomatis setelah implement; siklus fix tertutup
  (request-changes → running); status hidup dari ADR-011 jadi orkestrasi nyata;
  §32 langkah 14-15 akhirnya dieksekusi.
- **Biaya**: dua runner baru + reservasi review/QA + contract review/QA +
  quality_gates enforcement + test. Bergantung ADR-011 (status) approved dulu.
- **Risiko**: QA/review agent salah gate (spawn padahal status salah). Mitigasi: gate
  status di runner, bukan prompt. Reviewer read-only dipastikan plan mode (A-03).
- **Yang tidak diputuskan**: apakah code-reviewer butuh PR review formal di GitHub atau
  cukup handoff; apakah QA butuh environment/staging terpisah; kapan fix loop berhenti
  (batas iterasi).

## Dependencies

- **ADR-011** (proposed, PR #52): status YAML hidup + `implementation-complete`
  ditulis runner. **Prasyarat wajib** — tanpa ini launch-review tak punya gate.
- **ADR-010** (proposed, PR #52): integrasi GitHub Projects. Optional — auto-spawn
  jalan tanpa Projects; Projects cuma view.

## Backward compatibility / rollback

- Tidak mengubah perilaku agent implementer. Runner baru bersifat tambahan.
- Rollback: nonaktifkan launch-review/launch-qa; kembali ke manual. Status YAML
  tidak rusak.
- Handoff schema code-reviewer sudah ada (`decision` field, §23.7) — tak berubah.

**Menimpa:** tidak ada ADR sebelumnya yang menimpa; mengisi gap eksekusi §32 langkah
14-15 yang belum punya runner. Relevan: ADR-004 (pola runner), ADR-011 (status,
prasyarat), ADR-001 (merge authority — tetap manusia).

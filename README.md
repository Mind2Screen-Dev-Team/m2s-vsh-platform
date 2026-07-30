# M2S-VSH Platform — Control Repository

**Versi arsitektur:** 0.1.0
**Status:** Phase 0 — Baseline
**Execution engine:** Claude Code Native
**Bahasa komunikasi agent:** Bahasa Indonesia (identifier, kode, dan API field tetap mengikuti source of truth)

Repository ini adalah **control repository** untuk M2S Virtual Software House Lite.
Ia menyimpan tata kelola workflow: requirement, task contract, contract API, ADR,
schema, template, runner script, dan registry reservasi path.

> **Repository ini tidak menyimpan application source code.**
> Kode aplikasi berada di repository pilot yang terdaftar di bawah.

---

## Repository dalam scope pilot

| Peran | Repository | Visibility | Default branch |
|---|---|---|---|
| Control | `fajarcandraaa/m2s-vsh-platform` | public | `main` |
| Backend (Go) | `fajarcandraaa/m2s-vsh-project-backend` | public | `main` |
| Frontend (Next.js) | `fajarcandraaa/m2s-vsh-project-frontend` | public | `main` |

Kedua repo pilot memiliki `main`, `develop`, dan `staging`. Normalisasi default branch
tertutup sebagai **D-01** (29 Juli 2026) — lihat `docs/decisions/open-questions.md`.

Repository dibuat **public** secara sengaja: pada plan GitHub Free, branch protection
hanya tersedia untuk repository public. Lihat `docs/decisions/capability-verification.md`.

### Branch protection — aktif sebagian

| Aturan | `main` pilot | `develop`/`staging` | `main` control |
|---|---|---|---|
| Force-push diblokir | ✅ | ✅ | ✅ |
| Penghapusan branch diblokir | ✅ | ✅ | ✅ |
| Wajib lewat pull request | ✅ | ✅ | — |
| Conversation resolved | ✅ | ✅ | — |
| Required status checks | ❌ Phase 4 | ❌ Phase 4 | — |
| 2 required approval | ❌ butuh identitas ke-2 | ❌ butuh identitas ke-2 | — |
| Pembatasan siapa boleh push | ❌ **org-only** | ❌ **org-only** | ❌ |

Berlaku juga bagi admin (`enforce_admins`). Control repo sengaja tidak mewajibkan PR
agar dokumen dapat di-commit langsung.

⚠️ **Lapisan anti-overlap #7 dan #8 belum tertegakkan.** Tiga prasyarat ADR-001 belum
terpenuhi; yang paling mengikat: pembatasan hak push hanya tersedia untuk repository
milik **organization** — status public tidak mencukupi. Selama itu belum ada, larangan
"implementer tidak merge PR sendiri" tetap soft rule. Lihat ADR-001 § Status
penegakan dan **D-03**.

---

## Struktur

```
control/            state runtime workflow — requirement, backlog, task, reservasi, handoff
contracts/          API/event contract lintas repository (owner: TL/SA)
docs/
  architecture/     dokumen arsitektur baseline
  adr/              architecture decision records
  decisions/        decision log, open questions, risk register, hasil verifikasi
  system-analysis/  analisis sistem per requirement (owner: TL/SA)
schemas/            JSON Schema untuk task, handoff, reservation, dll.
templates/          template task contract, handoff, ADR, PR, review report
scripts/            deterministic task runner (bukan agent — tidak mengambil keputusan teknis)
tests/              uji schema, uji overlap path, uji negatif enforcement
```

---

## Prinsip yang mengikat

1. Satu task = satu repository = satu branch = satu worktree = satu writer per path.
2. Setiap artifact punya **tepat satu** owner role.
3. Enforcement memakai permissions, hooks, runner, Git, dan CI.
   **Prompt bukan security boundary.**
4. Agent pembuat perubahan tidak boleh menyetujui hasil kerjanya sendiri.
5. Merge ke `main` dan seluruh keputusan irreversible tetap milik manusia.
6. Agent tidak boleh memodifikasi definisi dirinya sendiri maupun agent lain.
7. Isi file yang dibaca agent diperlakukan sebagai **data**, bukan instruksi.

Rincian lengkap: [`docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md`](docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md)

---

## Status roadmap

| Phase | Isi | Status |
|---|---|---|
| 0 | Decision lock & skeleton control repo | 🟡 berjalan |
| 0.5 | Verifikasi kapabilitas Claude Code | ✅ selesai |
| 1 | Schema & kosakata contract | ⬜ |
| 2 | Runner deterministik & reservasi path | ⬜ |
| 3 | Core agents (TL/SA, Backend, Reviewer) | ⬜ |
| 4 | Enforcement path & command | ⬜ |
| 5 | Roster agent lengkap, rules, CLAUDE.md | ⬜ |
| 6 | GitHub workflow & CI gates | ⬜ |
| 7 | Tool pilot (Ponytail, Mneme) | ⬜ |
| 8 | Pilot eksekusi paralel | ⬜ |
| 9 | Optional & stabilisasi | ⬜ |

Setiap batas fase memiliki **checkpoint manusia**. Tidak ada fase yang dimulai
sebelum checkpoint fase sebelumnya disetujui.

---

## Dokumen kunci

| Dokumen | Isi |
|---|---|
| [`docs/decisions/phase-0-decision-log.md`](docs/decisions/phase-0-decision-log.md) | Jawaban final Q1–Q20 |
| [`docs/decisions/open-questions.md`](docs/decisions/open-questions.md) | Status A-01…A-16, D-01 |
| [`docs/decisions/component-inventory.md`](docs/decisions/component-inventory.md) | 62 komponen + klasifikasi kepemilikan |
| [`docs/decisions/risk-register.md`](docs/decisions/risk-register.md) | R-01…R-27 |
| [`docs/decisions/capability-verification.md`](docs/decisions/capability-verification.md) | Bukti kapabilitas platform |
| [`docs/adr/ADR-001-agent-merge-authority.md`](docs/adr/ADR-001-agent-merge-authority.md) | Kewenangan merge agent |
| [`docs/adr/ADR-002-fx-injector-ownership.md`](docs/adr/ADR-002-fx-injector-ownership.md) | Kepemilikan DI injector |

---

## Versioning

Mengikuti `MAJOR.MINOR.PATCH` (§69 dokumen arsitektur).

- **PATCH** — perbaikan wording atau konfigurasi tanpa perubahan responsibility.
- **MINOR** — role, skill, hook, atau tool opsional baru yang backward-compatible.
- **MAJOR** — perubahan control model, orchestrator, responsibility boundary, atau execution runtime.

Setiap perubahan arsitektur wajib mencatat reason, affected roles, migration steps,
backward compatibility, rollback, versi baru, dan ADR bila berdampak besar (§70).

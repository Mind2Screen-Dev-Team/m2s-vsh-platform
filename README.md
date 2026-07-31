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
| Required status checks | ❌ Phase 4 (§60) | ❌ Phase 4 (§60) | — |
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
templates/agents/   13 definisi role agent — sumber kanonik (human-only write)
.claude/agents/     subset yang aktif di control repo: PM dan TL/SA (Q10)
scripts/            wrapper tipis ke bin/m2s — pola scripts/<runner>.sh (Q11)
cmd/m2s/            source runner: 5 subcommand (human-only write)
internal/           pathmatch (deteksi overlap R-03), contract, registry
Makefile            build & quality gate (human-only write)
tests/              uji schema, uji overlap path, uji negatif enforcement
```

Binary `bin/m2s` **dibangun lokal dan tidak di-commit** (ADR-004 #5). Jalankan
`make build` lebih dulu; wrapper akan menolak dengan pesan jelas bila belum ada.

```bash
make build     # kompilasi bin/m2s
make check     # fmt + vet + test -race
make verify    # check + wrapper tipis + schema RE2 + 13 definisi agent
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

Roadmap mengikuti **§56–§64 dokumen arsitektur**. Tabel ini adalah cerminan, bukan
sumber — bila berbeda, dokumen arsitektur yang berlaku, kecuali pada bagian yang
ditimpa ADR.

| Phase | Isi | Sumber | Status |
|---|---|---|---|
| 0 | Baseline — control repo, pilot project, auth, protected branches | §56 | 🟡 berjalan |
| 1 | Task Contract dan Runner — schema, validasi, reservasi, launcher | §58 ⇄ | ✅ selesai |
| 2 | Core Agents — 13 project agent (ADR-005) | §57 ⇄ | ✅ selesai |
| 3 | Path Enforcement — PreToolUse hook, dangerous-command, CI path validation | §59 | ⬜ |
| 4 | GitHub Workflow — PR template, CODEOWNERS, required checks, merge queue | §60 | ⬜ |
| 5 | Tool Pilot — Ponytail & Mneme project-scoped | §61 | ⬜ |
| 6 | UI/UX Optional — Open Design pada workspace terisolasi | §62 | ⬜ |
| 7 | Multi-Repo Pilot — backend + frontend paralel | §63 | ⬜ |
| 8 | Stabilization — pengukuran token, review cycle, escaped defect | §64 | ⬜ |

⇄ **Phase 1 dan 2 ditukar** oleh [ADR-003](docs/adr/ADR-003-phase-order-contract-before-agents.md):
task contract dikerjakan sebelum core agents, karena definisi agent memuat batas path
yang bentuknya ditetapkan `task.schema.json`. Isi dan kriteria Done kedua fase tidak
berubah — hanya urutannya. Akibatnya nomor fase tidak lagi sejajar dengan nomor §.

Setiap batas fase memiliki **checkpoint manusia**. Tidak ada fase yang dimulai
sebelum checkpoint fase sebelumnya disetujui.

**Verifikasi kapabilitas Claude Code** (hasil: `docs/decisions/capability-verification.md`)
dikerjakan sebagai bagian Phase 0, memperluas §56 dengan bukti platform sebelum
agent dibangun. Ia **bukan fase tersendiri**.

> **Aturan perubahan roadmap.** Penomoran dan isi fase hanya boleh menyimpang dari
> §56–§64 melalui **ADR** yang menyatakan bagian mana yang ditimpa, sesuai §70.
> Perubahan langsung pada tabel ini tanpa ADR adalah cacat dokumentasi.
>
> Saat merujuk fase di dokumen lain, tulis nomor § berdampingan — `Phase 5 (§61)` —
> agar rujukan tetap sahih bila penomoran bergeser.

---

## Dokumen kunci

| Dokumen | Isi |
|---|---|
| [`docs/decisions/phase-0-decision-log.md`](docs/decisions/phase-0-decision-log.md) | Jawaban final Q1–Q20 |
| [`docs/decisions/open-questions.md`](docs/decisions/open-questions.md) | Status A-01…A-16, D-01…D-05, V-01…V-05, dan daftar bagian arsitektur yang ditimpa ADR |
| [`docs/decisions/component-inventory.md`](docs/decisions/component-inventory.md) | 62 komponen + klasifikasi kepemilikan |
| [`docs/decisions/risk-register.md`](docs/decisions/risk-register.md) | R-01…R-27 |
| [`docs/architecture/roles-extension-v0.1.0.md`](docs/architecture/roles-extension-v0.1.0.md) | §E1–§E4 — rincian 4 role tambahan (ADR-005) |
| [`docs/decisions/path-overlap-matrix.md`](docs/decisions/path-overlap-matrix.md) | Semantik glob & 24 kasus uji konflik reservasi (R-03) |
| [`docs/decisions/capability-verification.md`](docs/decisions/capability-verification.md) | Bukti kapabilitas platform |
| [`docs/adr/ADR-001-agent-merge-authority.md`](docs/adr/ADR-001-agent-merge-authority.md) | Kewenangan merge agent |
| [`docs/adr/ADR-002-fx-injector-ownership.md`](docs/adr/ADR-002-fx-injector-ownership.md) | Kepemilikan DI injector |
| [`docs/adr/ADR-003-phase-order-contract-before-agents.md`](docs/adr/ADR-003-phase-order-contract-before-agents.md) | Pertukaran urutan Phase 1 dan 2 |
| [`docs/adr/ADR-004-contract-format-and-runner-implementation.md`](docs/adr/ADR-004-contract-format-and-runner-implementation.md) | Format kontrak & bahasa runner |
| [`docs/adr/ADR-005-additional-engineering-roles.md`](docs/adr/ADR-005-additional-engineering-roles.md) | Empat role engineering tambahan |
| [`docs/adr/ADR-006-agent-definition-baseline.md`](docs/adr/ADR-006-agent-definition-baseline.md) | Baseline definisi 13 agent, `effort`, prasyarat platform, distribusi role |

---

## Versioning

Mengikuti `MAJOR.MINOR.PATCH` (§69 dokumen arsitektur).

- **PATCH** — perbaikan wording atau konfigurasi tanpa perubahan responsibility.
- **MINOR** — role, skill, hook, atau tool opsional baru yang backward-compatible.
- **MAJOR** — perubahan control model, orchestrator, responsibility boundary, atau execution runtime.

Setiap perubahan arsitektur wajib mencatat reason, affected roles, migration steps,
backward compatibility, rollback, versi baru, dan ADR bila berdampak besar (§70).

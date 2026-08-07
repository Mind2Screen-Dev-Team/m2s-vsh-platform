# Component Inventory & Ownership Mapping

**Tanggal:** 29 Juli 2026
**Versi arsitektur:** 0.1.0
**Cakupan:** seluruh komponen dokumen arsitektur yang perlu diterjemahkan menjadi
konfigurasi atau file executable.

**Total: 62 komponen**, diklasifikasikan ke dalam empat tingkat kepemilikan.

---

## 1. Klasifikasi kepemilikan

Aturan pemisah yang dipakai:

> Bila hilangnya sebuah file mengubah **perilaku** agent bagi orang lain, file itu
> **wajib** version-controlled. Bila hanya mengubah kenyamanan, ia local-only.

| Tingkat | Definisi | Lokasi |
|---|---|---|
| **Platform-global** | Berlaku untuk semua project M2S; tidak memuat nama project atau path repo | control repo |
| **Project-specific** | Terikat pada satu produk/pilot | control repo, per namespace project |
| **Repository-specific** | Terikat pada satu application repository | di dalam repo tersebut |
| **Personal / local-only** | Preferensi pribadi atau rahasia; tidak version-controlled | mesin masing-masing |

---

## 2. Platform-global (22 komponen)

Dimiliki **Human Workflow Maintainer**. Berada di control repository.

| # | Komponen | Bentuk | Fase |
|---|---|---|---|
| 1 | Dokumen arsitektur baseline | Markdown | ada |
| 2 | `VERSION` | teks | P0 ✅ |
| 3 | Decision log, open questions, risk register, capability verification | Markdown | P0 ✅ |
| 4–10 | 7 JSON Schema — task, handoff, reservation, failure, review-report, capability, task-state | JSON Schema | P1 ✅ |
| 11–17 | 7 template — task contract, handoff, failure report, review report, ADR, DESIGN.md, PR | Markdown/YAML | P1, P6, P9 ✅ |
| 18 | 5 runner script — validate-task, reserve-paths, launch-task, collect-result, release-reservation | shell | P2 ✅ |
| 19 | Library 6 hook | shell | P4 ✅ |
| 20 | 13 template agent definition | Markdown + frontmatter | P2 ✅ |
| 21 | Template rules generik — architecture, security, testing, universal-agent-rules, rule-precedence | Markdown | P5 ✅ |
| 22 | Template CI workflow + CODEOWNERS + checklist branch protection | YAML | P6 ✅ |

**Status backlog (5 Agustus 2026).** Komponen #4–#22 selesai:

- #4–10 — 8 schema di `schemas/` (`common` sebagai kosakata bersama + 7 dokumen).
  `failure`, `review-report`, `capability`, dan `task-state` adalah dokumen
  registry/referensi: didaftarkan `NewValidator` agar `$ref` ter-resolve, tetapi
  bukan `Kind` — belum ada subcommand runner yang memvalidasinya. Diuji
  `TestRegistrySchemasCompile` + `TestRegistryExamplesValidate`.
- #11–17 — `templates/artifacts/` (task contract, handoff, failure report,
  review report, ADR, design handoff) + `templates/github/PULL_REQUEST_TEMPLATE.md`.
  Template `DESIGN.md` diwujudkan sebagai `design-handoff.md`: design system
  itu sendiri sudah ada di `design/DESIGN.md` (Phase 5), yang belum ada adalah
  bentuk handoff-nya (Phase 6).
- #21 — `templates/rules/`. Deploy ke `.claude/rules/` adalah langkah manusia
  (`.claude/**` human-only-write) — lihat `docs/operator/rules-deployment.md`.

Ditambah kebijakan platform yang berbentuk dokumen: definisi task state machine (§33),
severity taxonomy (§48), retry rules (§49), konvensi branch `agent/<task-id>-<slug>`,
rule precedence (§15), larangan universal (§16), capability governance (§41),
PRPM approval flow (§9.4).

---

## 3. Project-specific (15 komponen)

Berada di control repository, di bawah namespace project.

| Komponen | Owner | Writable path |
|---|---|---|
| Project registry entry | PM | `control/projects/**` |
| Requirements + `REQ-*` | PM | `control/requirements/**` |
| Backlog + prioritas | PM | `control/backlog/**` |
| Task status | Runner + role (tabel owner ADR-011) | `control/tasks/status/**` |
| Release scope | PM | `control/releases/**` |
| Project report | PM | `control/reports/**` |
| System analysis | TL/SA | `docs/system-analysis/**` |
| Business rule catalogue `BR-*` | TL/SA | `docs/system-analysis/**` |
| API/Event contract `CONTRACT-*` | TL/SA | `contracts/**` |
| ADR | TL/SA | `docs/adr/**` |
| Task specification + Task DAG | TL/SA | `control/tasks/specifications/**` |
| Path ownership matrix | TL/SA | `control/tasks/specifications/**` |
| Reservation registry | **Runner** | `control/reservations/**` |
| Handoff records | Runner (isi dari agent) | `control/handoffs/**` |
| Traceability matrix | QA | `qa/**` di app repo |

`DESIGN.md`, design token, flow, wireframe, prototype, dan handoff dimiliki UI/UX
pada `design/**`. Phase 6 (§62) menetapkan lokasinya di **control repository**:
design workspace terisolasi dari application worktree, sehingga design tidak
pernah menulis repo aplikasi. Repo frontend menerima salinan `design/DESIGN.md` +
token lewat task sync terpisah. Lihat `docs/operator/phase-6-open-design.md`.

---

## 4. Repository-specific (16 komponen)

| Komponen | Owner | Catatan |
|---|---|---|
| `CLAUDE.md` | Workflow Maintainer | konteks + pointer, **bukan** rule keras |
| `.claude/agents/` subset | Workflow Maintainer | hanya role relevan bagi repo itu |
| `.claude/rules/` stack-specific | TL/SA (isi), Maintainer (integrasi) | ditulis manual, version-controlled |
| `.claude/skills/` | Workflow Maintainer | |
| `.claude/settings.json` | Workflow Maintainer | **tempat hard rules** |
| `.claude/hooks/` | Workflow Maintainer | |
| `.mneme/project_memory.json` | TL/SA | engineering dilarang mengedit |
| `.github/workflows/` | DevOps | |
| `CODEOWNERS` | Workflow Maintainer | human-only |
| Branch protection / rulesets | Manusia | setting GitHub, bukan file |
| Definisi quality gate command | TL/SA | per stack, per task |
| Backend module + unit test | Backend task owner | |
| Frontend feature + component test | Frontend task owner | |
| Integration/E2E test + QA artifact | QA | `tests/integration/**`, `tests/e2e/**`, `qa/**` |
| `design/DESIGN.md` + token | UI/UX | repo frontend |
| Dokumentasi pengguna/operator + `CHANGELOG.md` | Technical Writer | |

---

## 5. Personal / local-only (9 komponen)

**Tidak version-controlled.**

| Komponen | Alasan |
|---|---|
| `~/.claude/settings.json` | hanya preferensi pribadi yang tidak mengubah project behaviour (§38) |
| `.claude/settings.local.json` | override lokal per developer |
| Kredensial GitHub, token GitHub App | rahasia; milik manusia (§13.1) |
| `.env`, `*.pem`, `*.key`, `credentials*.json` | dilarang diakses agent (§42.3) |
| Absolute path mesin, `M2S_WORKTREE_ROOT` | mesin-spesifik |
| Isi fisik worktree | ephemeral, di luar repo |
| `.task/` snapshot contract | disuntikkan runtime oleh runner |
| Instalasi plugin pada mesin Maintainer | bukan runtime dependency repo |
| Session log, transcript, `.DS_Store` | ephemeral |

---

## 6. Pemetaan menurut tempat konfigurasi

### `CLAUDE.md` — konteks, bukan security

**Isi yang tepat:** identitas dan domain project; terminologi (prosa Bahasa
Indonesia, identifier mengikuti source of truth); pointer ke task contract sebagai
otoritas; ringkasan larangan universal §16; pointer ke `.claude/rules/`; catatan
bahwa CLAUDE.md berada di peringkat rendah pada rule precedence §15.

**Isi yang salah tempat:** daftar allowed/forbidden path (milik task contract),
pola secret (milik `permissions.deny`), daftar command destruktif (milik hook +
deny), kredensial apa pun.

---

### `.claude/agents/` — role definition

Frontmatter yang **terverifikasi didukung**: `name`, `description`, `tools`,
`disallowedTools`, `model`, `effort`, `permissionMode`, `background`, `isolation`,
`maxTurns`, `skills`, `hooks`, `mcpServers`, `color`.

Body memuat purpose, owns, responsibilities, allowed, prohibited, definition of done,
stop conditions.

**Tidak boleh ada di sini:** secret, absolute path mesin, allowed path spesifik task
(berubah per task — milik task contract).

**Catatan:** `effort` semula belum ditetapkan Appendix A dokumen arsitektur.
**Ditetapkan 31 Juli 2026** oleh ADR-006 #2: `high` bagi PM, TL/SA, dan Code
Reviewer; `low` bagi Technical Writer; `medium` bagi sepuluh role sisanya. Nilai
ini menjadi baseline pengukuran biaya token **Phase 8 (§64)**.

Sumber kanonik ketiga belas definisi adalah `templates/agents/<role>.md`.
`.claude/agents/` memuat subset yang aktif per repository (Q10); keduanya dijaga
identik oleh `TestDeployedAgentsMatchTemplates`.

---

### `.claude/rules/` — soft rules, path-scoped

`architecture.md`, `security.md`, `testing.md`, `<stack>.md`.

Setiap rule yang berasal dari referensi eksternal **wajib** memuat source note dan
tanggal review (§11.2). Ditulis langsung dan version-controlled; Rulix tidak dipakai
pada v0.1.0 (§38, §10).

---

### `.claude/skills/` — reusable procedure

Wajib: single responsibility, trigger jelas, I/O contract, tidak menulis di luar
role ownership, tidak menyembunyikan external command, tidak menginstal dependency,
memiliki version + owner + evaluation cases (§39).

**Bukan** tempat untuk permanent project facts.

---

### `.claude/settings.json` — **tempat hard rules**

| Blok | Isi | Menegakkan |
|---|---|---|
| `permissions.deny` | `.env*`, `**/secrets/**`, `**/*.pem`, `**/*.key`, `**/credentials*.json` | §42.3 |
| `permissions.deny` | `.claude/**`, `.mneme/project_memory.json`, `.task/**` | §16.1, R-12 |
| `permissions.deny` | Bash destruktif §42.2 | §42.2 |
| `permissions.deny` | Installer: `npm install`, `yarn add`, `pip install`, `go get`, `prpm install`, `claude plugin` | §16.5, acceptance #12 — **tidak dienumerasi §42.2, ditambahkan** |
| `permissions.deny` | `git -C`, `git checkout`, `git switch`, `git worktree` | §16.2, R-15 |
| `permissions.allow` | whitelist minimum per repo | least privilege |
| `hooks` | registrasi 6 hook | §42 |
| `env` | `PONYTAIL_DEFAULT_MODE`, `PONYTAIL_SUBAGENT_MATCHER` | §5.5 + Q17 |
| `defaultMode` | **bukan** `bypassPermissions` | §16.5, R-09 |

---

### Hooks — enforcement runtime

| Hook | Event | Fail mode |
|---|---|---|
| `validate-path-scope.sh` | PreToolUse | **fail-closed** (`exit 2`) |
| `block-dangerous-command.sh` | PreToolUse | **fail-closed** |
| `block-secret-paths.sh` | PreToolUse | **fail-closed** |
| `audit-tool-use.sh` | PostToolUse | fail-open (audit), wajib redaksi secret |
| `validate-handoff.sh` | SubagentStop | **fail-closed** |
| `worktree-lifecycle.sh` | WorktreeCreate/Remove | **fail-closed** — terdaftar di `settings.json` (T-04 ditutup; catatan "belum dipanggil runtime" per 2026-08-01 sudah tidak berlaku) |

**Wajib:** `exit 2`, bukan `exit 1` — `exit 1` tidak memblokir apa pun (T-01).
Setiap hook memeriksa dependensinya di awal dan memiliki self-test.

---

### Schemas, Templates, Scripts

- **Schemas** — dibaca runner, hook, dan CI. Perubahan bersifat breaking → butuh version bump.
- **Templates** — menjamin handoff tidak pernah hanya "selesai" (§35).
- **Scripts** — **bukan agent**; tidak membuat technical decision (§13.4); wajib idempotent.

---

## 7. Daftar human-only write

Tidak boleh ditulis agent mana pun, dalam keadaan apa pun:

```
.claude/agents/**
.claude/hooks/**
.claude/settings.json
.claude/settings.local.json
CLAUDE.md
.github/CODEOWNERS
control/reservations/**
.task/**
cmd/m2s/**                          (source runner — penegak batas path)
Makefile                            (target build runner)
governance/capability-registry.yaml
branch protection / rulesets        (setting GitHub)
seluruh secret, token, credential
```

`cmd/m2s/**` masuk daftar ini karena binary hasil kompilasinya adalah satu-satunya
otoritas atas registry reservasi dan validasi path. Agent yang dapat menulis source
runner dapat melonggarkan batas yang mengikatnya sendiri. `bin/**` tidak dicantumkan
karena tidak di-commit (ADR-004 #5) — ia gitignored dan dibangun lokal.

---

## 8. Matriks ownership artifact

| Artifact | Owner | Reviewer | Concurrent writer |
|---|---|---|---|
| Business requirements | PM | Product Owner | 1 |
| System analysis | TL/SA | PM | 1 |
| Business rule catalogue | TL/SA | PM | 1 |
| API/Event contract | TL/SA | Backend, Frontend | 1 |
| ADR | TL/SA | Manusia untuk high risk | 1 |
| Mneme memory | TL/SA | Manusia/Reviewer | 1 |
| `DESIGN.md` | UI/UX | PM, TL/SA | 1 |
| Backend module | Backend task owner | Code Reviewer | 1 per path |
| Frontend feature | Frontend task owner | Code Reviewer | 1 per path |
| Unit/component test | Implementer | Code Reviewer | 1 per path |
| Integration/E2E test | QA | TL/SA, Reviewer | 1 per path |
| Review report | Code Reviewer *(via runner)* | — | 1 per review |
| CI/infra | DevOps | Reviewer, Manusia | 1 per path |
| Dokumentasi & `CHANGELOG.md` | Technical Writer | PM, TL/SA | 1 per path |
| Reservation registry | **Runner** | — | 1 |
| **Shared file** — `go.mod`, `go.sum`, lockfile, `app/injector/**`, migration registry, global enum | **owner yang ditunjuk per task** | Reviewer | 1 |
| Agent/rules/hooks config | Workflow Maintainer | Technical Approver | 1 |

---

## 9. Batas kepemilikan yang mudah tertukar

Tiga role berpotensi menyentuh narasi release. Batasnya ditetapkan eksplisit:

| Role | Menulis |
|---|---|
| Technical Writer | `CHANGELOG.md`, `release-notes/**` |
| PM | `control/releases/**` |
| DevOps | release manifest di `ops/**` |

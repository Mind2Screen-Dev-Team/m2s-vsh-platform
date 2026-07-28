# M2S Virtual Software House Lite v0.1.0

## Claude Code Native Architecture

**Nama singkat:** M2S-VSH Lite  
**Versi:** 0.1.0  
**Status:** Proposed Architecture / Pilot Baseline  
**Tanggal baseline:** 28 Juli 2026  
**Pemilik arsitektur:** Mindtoscreen  
**Execution engine:** Claude Code Native  
**Bahasa komunikasi agent:** Bahasa Indonesia  

---

## 0. Ringkasan Eksekutif

M2S-VSH Lite v0.1.0 adalah arsitektur automated software-development workflow yang menggunakan Claude Code sebagai execution engine utama tanpa Hermes Agent atau control plane agent tambahan.

Arsitektur ini sengaja dibuat lebih ringan daripada rancangan M2S-VSH v1.0.0. Fokus versi ini bukan membangun platform multi-agent yang sepenuhnya otonom, melainkan membuktikan bahwa beberapa role AI dapat bekerja secara terstruktur, paralel, dapat diaudit, dan tidak saling menimpa pekerjaan.

Arsitektur dibangun dengan prinsip berikut:

1. Project Manager Agent mengelola objective, scope, task state, prioritas, dan handoff.
2. Technical Lead & System Analyst Agent menjaga konsistensi system behaviour, technical design, contract, dan pembagian technical work.
3. Setiap agent implementasi hanya menangani satu task, satu repository, satu branch, dan satu worktree.
4. Setiap writable path hanya memiliki satu active writer pada satu waktu.
5. Backend dan Frontend boleh bekerja paralel hanya setelah contract bersama disetujui.
6. Agent pembuat perubahan tidak boleh menyetujui hasil kerjanya sendiri.
7. Rules, hooks, permissions, CI, dan Git protections digunakan sebagai enforcement; prompt saja bukan security boundary.
8. Tool eksternal tidak boleh dipasang atau diperbarui secara otomatis oleh agent pelaksana.
9. Shared files, generated files, architecture decisions, dan design system mempunyai designated owner.
10. Production deployment dan keputusan irreversible tetap memerlukan human approval.

### Keputusan utama versi 0.1.0

- **Hermes Agent tidak digunakan.**
- **Claude Code custom subagents digunakan sebagai role agents.**
- **Write-capable subagents wajib memakai `isolation: worktree`.**
- **Satu task lintas beberapa repository dilarang; task harus dipecah per repository.**
- **Ponytail diadopsi secara selektif untuk engineering dan review.**
- **Mneme dipilotkan per repository dalam mode `warn` sebelum `strict`.**
- **Konsep `DESIGN.md` diadopsi, tetapi design system harus dibuat khusus untuk project, bukan menyalin identitas brand lain.**
- **Open Design bersifat opsional dan hanya digunakan dalam design workspace terisolasi.**
- **PRPM tidak digunakan untuk autonomous package installation.**
- **Rulix ditunda sampai M2S memakai lebih dari satu AI coding tool atau rule duplication benar-benar menjadi masalah.**
- **Awesome Cursor Rules dan versi Chinese digunakan sebagai reference library, bukan runtime dependency.**

---

# Bagian I — Tujuan, Scope, dan Non-Goals

## 1. Tujuan Arsitektur

M2S-VSH Lite bertujuan untuk:

- mengotomasi workflow pengembangan dari requirement sampai pull request;
- membagi tanggung jawab agent secara eksplisit;
- mencegah duplikasi tugas dan file overlap;
- mendukung pekerjaan Backend, Frontend, QA, dan dokumentasi secara paralel;
- menjaga konsistensi lintas repository;
- menyediakan audit trail atas task, keputusan, perubahan file, review, dan hasil test;
- mempertahankan human control pada keputusan bisnis, perubahan berisiko, dan release production;
- membangun fondasi yang dapat dinaikkan ke arsitektur lebih besar apabila kebutuhan multi-project dan durable orchestration muncul.

## 2. Scope v0.1.0

Versi ini mencakup:

- project discovery dan requirement management;
- system analysis dan technical design;
- UI/UX design handoff;
- backend implementation;
- frontend implementation;
- unit, integration, dan end-to-end testing;
- code review;
- CI validation;
- documentation update;
- pull request dan merge workflow;
- staging deployment;
- production release dengan human gate;
- worktree-based parallel execution;
- deterministic path validation;
- architecture decision governance per repository.

## 3. Non-Goals v0.1.0

Versi ini tidak mencoba menyediakan:

- autonomous company operations 24/7;
- persistent task queue lintas mesin dan restart;
- multi-provider model orchestration;
- Telegram, Slack, atau WhatsApp agent gateway;
- autonomous production deployment;
- auto-install dan auto-update plugin dari public registry;
- autonomous rule learning;
- automatic architecture decision generation tanpa approval;
- guaranteed zero merge conflict dalam semua kondisi;
- replacement penuh untuk Product Owner, security engineer, atau human technical approver.

Istilah “tidak overlap” pada dokumen ini berarti arsitektur mencegah concurrent writer, duplicate responsibility, dan uncoordinated contract changes. Arsitektur tetap tidak dapat menjamin bahwa dua perubahan berbeda tidak mempunyai semantic impact yang sama apabila task decomposition salah. Karena itu, TL/SA review dan contract-first workflow tetap wajib.

---

# Bagian II — Analisis Repository dan Tool

## 4. Matriks Keputusan

| Repository / Tool | Fungsi Utama | Keputusan v0.1.0 | Role Utama | Cara Penggunaan | Catatan Risiko |
|---|---|---|---|---|---|
| `DietrichGebert/ponytail` | Heuristik anti-overengineering dan minimal-diff engineering | **Adopt Selectively** | Backend, Frontend, Code Reviewer | Project-scoped plugin atau explicit skill | Jangan diterapkan ke PM, TL/SA, QA, UI/UX; jangan memakai mode ultra sebagai default |
| `MnemeHQ/mneme` | Deterministic architecture-decision governance sebelum edit | **Pilot / Adopt Partially** | TL/SA sebagai owner; semua writer sebagai consumer | Per-repository, project-scoped, warn-first | Saat ini local-repo dan single-developer; retrieval memiliki gap dan hook fail-open |
| `VoltAgent/awesome-design-md` | Collection `DESIGN.md` dari berbagai design language | **Reference and Adapt** | UI/UX Designer | Referensi struktur dan token; buat `DESIGN.md` sendiri | Jangan menyalin visual identity brand secara mentah |
| `nexu-io/open-design` | Local-first AI design workspace dan artifact generator | **Optional Tool** | UI/UX Designer | Dedicated design workspace atau scoped MCP | Jangan biarkan Open Design men-spawn nested coding agent pada production worktree |
| `pr-pm/prpm` | Registry/package manager untuk prompts, rules, skills, agents, dan hooks | **Defer for Runtime; Controlled Discovery Only** | Human Workflow Maintainer dan TL/SA | Search/info/manual audit; internal distribution di masa depan | Public prompt/plugin supply-chain risk; agent tidak boleh auto-install |
| `danielcinome/rulix` | Single source of truth dan compiler untuk multi-tool rules | **Defer to v0.2.0** | Workflow Maintainer dan TL/SA | Gunakan ketika Cursor/AGENTS.md/Copilot juga menjadi target | Claude-only belum membutuhkan rule conversion; menambah Node 22 dan generated-rule workflow |
| `PatrickJS/awesome-cursorrules` | Curated collection modern Cursor rules | **Reference Only** | TL/SA, Backend, Frontend, QA | Ambil ide, verifikasi, lalu tulis ulang menjadi M2S rules | Kualitas dan freshness setiap rule dapat berbeda; format utamanya Cursor |
| `LessUp/awesome-cursorrules-zh` | Chinese-localized Cursor rules collection | **Reference Only / Not Runtime** | TL/SA bila membutuhkan coverage tambahan | Referensi sekunder | Bahasa dan format tidak sesuai runtime Claude-first; berpotensi redundant dengan sumber utama |

---

## 5. Ponytail

### 5.1 Temuan

Ponytail mengarahkan coding agent untuk menggunakan solution ladder berikut:

1. Pastikan fitur memang perlu dibangun.
2. Reuse sesuatu yang sudah ada di codebase.
3. Gunakan standard library.
4. Gunakan native platform capability.
5. Gunakan dependency yang sudah terpasang.
6. Pilih implementasi paling sederhana.
7. Baru tulis minimum code yang benar-benar diperlukan.

Ponytail tetap mengecualikan security, trust-boundary validation, data-loss handling, accessibility, dan requirement eksplisit dari proses simplification.

### 5.2 Yang diadopsi

- YAGNI sebelum menambah abstraction.
- Existing-pattern-first.
- Standard-library-first.
- No dependency tanpa alasan terukur.
- Minimum correct diff.
- Root-cause fix, bukan symptom patch.
- Delete before add ketika aman.
- Satu runnable check untuk non-trivial logic.

### 5.3 Yang tidak diadopsi sebagai aturan universal

Ponytail tidak boleh diaktifkan ke seluruh agent karena:

- PM perlu mengeksplorasi scope dan risiko, bukan sekadar minimum output.
- TL/SA perlu mempertimbangkan extensibility dan system boundaries.
- UI/UX tidak dapat dinilai hanya berdasarkan minimum code.
- QA perlu memperluas edge cases, bukan mengecilkan coverage.
- Technical Writer membutuhkan completeness, bukan shortest text.

### 5.4 Role pengguna

- Backend Engineer Agent.
- Frontend Engineer Agent.
- Code Reviewer Agent.
- DevOps Agent hanya untuk script atau configuration change yang sederhana.

### 5.5 Konfigurasi yang direkomendasikan

```bash
PONYTAIL_DEFAULT_MODE=full
PONYTAIL_SUBAGENT_MATCHER='^(backend-engineer|frontend-engineer|code-reviewer|devops-release)$'
```

Mode `ultra` tidak digunakan secara default.

### 5.6 Ownership

Ponytail installation dan update dimiliki oleh Human Workflow Maintainer. Engineering agents tidak boleh mengubah mode atau versi plugin sendiri.

---

## 6. Mneme

### 6.1 Temuan

Mneme adalah governance layer per repository untuk mencatat architecture decisions, melakukan deterministic retrieval, dan memblokir edit yang melanggar recorded decisions sebelum edit ditulis ke disk.

Current phase Mneme masih berfokus pada:

- local repository;
- single developer;
- project-scoped governance;
- deterministic keyword-based retrieval;
- validation dan enforcement sebelum generation/edit.

Mneme bukan:

- workflow orchestrator;
- generalized agent memory;
- multi-repo governance system;
- deployment governance;
- auto-fix system;
- security sandbox.

### 6.2 Yang diadopsi

- `.mneme/project_memory.json` per repository.
- ADR-to-governance workflow.
- Project-scoped Claude Code hook.
- `warn` mode untuk pilot.
- `/mneme-context` sebelum perubahan non-trivial.
- `/mneme-review` pada akhir implementation task.
- CI governance check.

### 6.3 Yang tidak diandalkan

Mneme tidak digunakan sebagai:

- path lock;
- task reservation service;
- cross-repository source of truth;
- permission boundary;
- replacement code review;
- satu-satunya compliance check.

Alasannya:

- file-name keyword retrieval dapat melewatkan keputusan yang relevan;
- generic file name dapat menyebabkan low retrieval;
- hook fail-open ketika executable tidak tersedia, timeout, atau OS error;
- multi-repository/team governance masih deferred pada proyek Mneme.

### 6.4 Role pengguna

**Owner:** Technical Lead & System Analyst Agent.  
**Consumers:** Backend, Frontend, QA, DevOps, Code Reviewer.  
**Approver:** Human Technical Approver untuk keputusan high risk.

### 6.5 Rule ownership

- Hanya TL/SA yang boleh membuat atau mengubah recorded architecture decision.
- Engineering agent boleh mengajukan decision-change request.
- Engineering agent dilarang mengedit `.mneme/project_memory.json`.
- Mneme mode hanya boleh diubah oleh Workflow Maintainer.

### 6.6 Rollout

1. Minggu pertama: `warn` mode.
2. Perbaiki decision scope dan false positive.
3. Aktifkan `strict` hanya untuk keputusan stabil dan high-confidence.
4. Tetap jalankan post-change review dan CI.

---

## 7. Awesome DESIGN.md

### 7.1 Temuan

Repository ini mengumpulkan `DESIGN.md` yang menjelaskan palette, typography, layout, component language, responsive behaviour, dan prompt guidance agar coding agent dapat menghasilkan UI yang lebih konsisten.

### 7.2 Yang diadopsi

M2S menggunakan `DESIGN.md` sebagai source of truth visual pada setiap project frontend.

`DESIGN.md` minimal memuat:

- product personality;
- colour tokens;
- typography;
- spacing scale;
- grid dan layout;
- component principles;
- interaction states;
- motion rules;
- responsive behaviour;
- accessibility requirements;
- do/don't examples;
- implementation notes.

### 7.3 Yang tidak diadopsi

- Tidak menyalin mentah design system Linear, Vercel, Stripe, atau brand lain.
- Tidak menggunakan public visual identity sebagai final brand direction.
- Tidak membiarkan Frontend Agent mengubah `DESIGN.md` sendiri.

### 7.4 Role pengguna

**Owner:** UI/UX Designer Agent.  
**Reviewer:** Project Manager untuk product fit; TL/SA untuk system behaviour dan feasibility.  
**Consumer:** Frontend Engineer Agent dan QA Agent.

---

## 8. Open Design

### 8.1 Temuan

Open Design adalah local-first desktop/web design environment yang dapat menggunakan Claude Code dan runtime agent lain untuk menghasilkan prototype, dashboard, deck, image, video, dan file artifact. Open Design mengombinasikan base prompt, `DESIGN.md`, dan `SKILL.md`.

### 8.2 Keputusan

Open Design tidak menjadi komponen wajib M2S-VSH Lite.

Open Design dapat digunakan apabila project membutuhkan:

- rapid prototype;
- UI concept exploration;
- visual design system preview;
- presentation or media artifact;
- HTML/CSS design handoff.

### 8.3 Batas penggunaan

- Open Design harus berjalan pada design workspace atau design repository terpisah.
- Output tidak boleh langsung ditulis ke implementation worktree.
- UI/UX Agent mengekspor approved design handoff.
- Frontend Agent mengimplementasikan handoff pada task terpisah.
- Jangan menjalankan Open Design local runtime yang men-spawn Claude Code di dalam worker Claude Code yang sedang mengedit application repository.
- Bila MCP digunakan, hanya UI/UX Agent yang mendapat akses.
- MCP write path dibatasi ke `design/**`, `prototypes/**`, dan `artifacts/**`.

### 8.4 Role pengguna

- UI/UX Designer Agent: primary.
- Project Manager: review output.
- Frontend Agent: consumer only.

---

## 9. PRPM

### 9.1 Temuan

PRPM menyediakan registry dan package manager untuk prompts, rules, agents, skills, hooks, dan collections lintas AI coding tools.

### 9.2 Keputusan

PRPM tidak menjadi runtime dependency v0.1.0.

PRPM hanya digunakan secara controlled untuk:

- mencari referensi capability;
- melihat package metadata;
- mengevaluasi package sebelum di-vendor;
- potensi distribusi internal M2S package pada versi berikutnya.

### 9.3 Larangan

- Agent tidak boleh menjalankan `prpm install` secara otomatis.
- Agent tidak boleh menggunakan self-improvement pattern yang memasang package saat task berlangsung.
- Package public tidak boleh masuk project tanpa source review.
- Collection besar tidak boleh dipasang sekaligus tanpa inventory file.
- Version floating dilarang untuk package yang sudah disetujui.

### 9.4 Approval flow

```text
Agent mengajukan capability request
        ↓
TL/SA menilai relevance dan overlap
        ↓
Workflow Maintainer melakukan source review
        ↓
License dan security review
        ↓
Install pada sandbox project
        ↓
Evaluation
        ↓
Pin version dan commit inventory
```

### 9.5 Role pengguna

- Human Workflow Maintainer: install/update.
- TL/SA: approve usefulness dan architectural fit.
- Engineering agents: request only.

---

## 10. Rulix

### 10.1 Temuan

Rulix dapat menjadikan `.rulix/rules/` sebagai source of truth lalu menghasilkan rules untuk Claude Code, Cursor, dan AGENTS.md. Rulix juga memvalidasi duplicate, vague rules, conflict, dan token budget.

### 10.2 Keputusan

Rulix ditunda sampai salah satu kondisi berikut terpenuhi:

- M2S menggunakan Cursor dan Claude Code secara bersamaan;
- M2S perlu menghasilkan AGENTS.md lintas tool;
- jumlah project rules menjadi sulit diaudit;
- rule drift antar tool muncul;
- rule token budget menjadi bottleneck nyata.

### 10.3 Alasan tidak digunakan sekarang

- v0.1.0 adalah Claude Code Native.
- Claude Code sudah mempunyai `.claude/rules/` path-scoped.
- Menambah Rulix menciptakan generated-file layer yang belum dibutuhkan.
- Node.js 22 menjadi dependency tambahan hanya untuk rule conversion.

### 10.4 Future ownership

Bila diaktifkan pada v0.2.0:

- `.rulix/rules/` menjadi canonical source.
- `.claude/rules/` menjadi generated output.
- Generated output tidak boleh diedit manual.
- TL/SA memiliki isi architecture rules.
- Workflow Maintainer memiliki config dan sync process.

---

## 11. Awesome Cursor Rules dan Chinese Localization

### 11.1 Keputusan

Kedua repository digunakan hanya sebagai reference library.

### 11.2 Cara penggunaan yang diperbolehkan

1. Cari rule yang relevan dengan stack.
2. Verifikasi terhadap dokumentasi resmi framework.
3. Hapus rule yang obsolete, terlalu umum, atau kontradiktif.
4. Tulis ulang menjadi project-specific Claude rule.
5. Tambahkan source note dan tanggal review.
6. Jalankan rule conflict review.

### 11.3 Cara penggunaan yang dilarang

- Menyalin seluruh repository ke `.claude/rules/`.
- Mengaktifkan semua rules sekaligus.
- Menganggap curated rule sebagai official framework guidance.
- Menggunakan Chinese-localized rule tanpa technical review.
- Mengizinkan agent mengunduh atau mengganti rule saat implementation task.

### 11.4 Role pengguna

- TL/SA: architecture dan workflow references.
- Backend/Frontend: mengajukan candidate rule.
- QA: testing-rule references.
- Workflow Maintainer: final integration.

---

# Bagian III — Arsitektur Sistem

## 12. High-Level Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│ HUMAN GOVERNANCE                                            │
│ Product Owner / Founder / Technical Approver                │
└────────────────────────────┬────────────────────────────────┘
                             │ objective, approval, release
                             ▼
┌─────────────────────────────────────────────────────────────┐
│ PROJECT MANAGER AGENT — CONTROL SESSION                     │
│ scope, priority, task state, handoff, progress              │
│ tidak mengedit application source code                      │
└────────────────────┬───────────────────┬────────────────────┘
                     │                   │
                     ▼                   ▼
┌────────────────────────────┐  ┌─────────────────────────────┐
│ TL / SYSTEM ANALYST AGENT  │  │ UI/UX DESIGNER AGENT       │
│ system flow, contract, ADR │  │ DESIGN.md, prototype       │
│ task technical readiness   │  │ design handoff             │
└──────────────┬─────────────┘  └──────────────┬──────────────┘
               │ approved work orders          │ approved design
               └───────────────┬────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ DETERMINISTIC TASK RUNNER                                   │
│ membaca task contract, memvalidasi reservation,             │
│ menjalankan Claude role di dedicated process/worktree       │
└──────────────┬────────────────┬────────────────┬─────────────┘
               ▼                ▼                ▼
     ┌────────────────┐ ┌────────────────┐ ┌────────────────┐
     │ Backend Agent  │ │ Frontend Agent │ │ DevOps Agent   │
     │ isolated WT    │ │ isolated WT    │ │ isolated WT    │
     └───────┬────────┘ └───────┬────────┘ └───────┬────────┘
             │                  │                  │
             └──────────┬───────┴──────────┬───────┘
                        ▼                  ▼
               ┌────────────────┐  ┌────────────────┐
               │ QA Agent       │  │ Code Reviewer  │
               │ test ownership │  │ read-only      │
               └────────┬───────┘  └────────┬───────┘
                        └──────────┬─────────┘
                                   ▼
┌─────────────────────────────────────────────────────────────┐
│ GITHUB                                                      │
│ branch → PR → CI → review → merge queue → staging           │
└────────────────────────────┬────────────────────────────────┘
                             ▼
                    Human Production Approval
```

---

## 13. Komponen Arsitektur

### 13.1 Human Governance

Human tetap memiliki:

- business objective;
- final scope decision;
- budget dan timeline trade-off;
- high-risk architecture approval;
- production access;
- secret ownership;
- agent/tool installation approval;
- final release approval.

### 13.2 Project Manager Control Session

Project Manager Agent menjadi coordinator utama, tetapi bukan coding orchestrator yang bebas mengedit repository.

PM Agent:

- menerima objective;
- melakukan requirement interview;
- mengelola backlog dan dependency status;
- meminta TL/SA analysis;
- mengeksekusi approved Task DAG;
- memonitor agent completion;
- mengelola handoff dan escalation;
- menyusun project report.

PM Agent hanya menjalankan task runner deterministik. PM tidak membuat branch/worktree secara improvisasi melalui rangkaian `cd` pada shared terminal.

### 13.3 Technical Governance

TL/SA menghasilkan:

- requirement analysis;
- use cases;
- system flow;
- business rules;
- data requirements;
- API/event contract;
- architecture decision;
- technical task breakdown;
- repository assignment;
- path scope;
- technical acceptance criteria.

### 13.4 Deterministic Task Runner

Task runner bukan agent. Task runner adalah shell script atau CLI ringan yang:

- membaca approved task contract;
- memastikan task status `ready`;
- memastikan dependency selesai;
- memeriksa path reservation;
- memilih repository;
- membuat dedicated branch/worktree atau meminta native worktree isolation;
- menjalankan Claude Code dengan role yang ditentukan;
- menyediakan prompt/task contract;
- menyimpan session ID dan log;
- mengembalikan result status;
- membersihkan reservation setelah task selesai atau gagal.

Task runner tidak membuat technical decision.

### 13.5 Claude Code Role Agents

Role agents disimpan pada project `.claude/agents/`, bukan hanya plugin, karena project agents dapat mempunyai role-specific tools, hooks, MCP, permission mode, dan worktree isolation.

Write-capable role wajib menggunakan:

```yaml
isolation: worktree
background: true
```

Read-only reviewer tidak wajib membuat worktree.

### 13.6 GitHub Collaboration

GitHub menjadi collaboration dan audit layer:

- protected branches;
- one branch per task;
- pull request per implementation task atau coherent task group;
- CODEOWNERS;
- CI checks;
- review requirements;
- merge queue;
- release tags;
- audit trail.

---

# Bagian IV — General Rules

## 14. Rule Classification

Rules dibagi menjadi dua jenis.

### 14.1 Hard Rules

Hard rules harus ditegakkan dengan permissions, hooks, runner, Git, atau CI.

Contoh:

- allowed paths;
- denied paths;
- no direct push;
- no branch switching;
- no secret access;
- no destructive command;
- no production deployment;
- one active writer per path;
- required tests;
- PR approval requirement.

### 14.2 Soft Rules

Soft rules berupa guidance yang dimuat melalui `CLAUDE.md`, `.claude/rules`, agents, atau skills.

Contoh:

- coding style;
- design principles;
- naming preference;
- architectural heuristics;
- communication style;
- Ponytail simplification ladder.

Soft rule tidak boleh digunakan sebagai satu-satunya enforcement untuk security atau file ownership.

---

## 15. Rule Precedence

Urutan prioritas tertinggi ke terendah:

1. Human safety, legal, dan production governance.
2. Managed security settings dan repository protections.
3. Task contract: repository, branch, allowed paths, forbidden paths.
4. Approved ADR dan Mneme project decisions.
5. Project-specific architecture and domain rules.
6. Role-agent definition.
7. Skill instructions.
8. User prompt untuk task tersebut.
9. External reference rules.

Bila terjadi konflik:

- agent harus berhenti;
- tulis conflict report;
- jangan memilih rule sendiri;
- eskalasi ke TL/SA atau PM sesuai jenis keputusan.

---

## 16. Rules yang Wajib Dipatuhi Semua Agent

### 16.1 Scope dan Ownership

Semua agent wajib:

- bekerja hanya pada task ID yang diberikan;
- membaca task contract sebelum bekerja;
- memastikan repository dan branch benar;
- hanya menulis pada `allowed_paths`;
- memperlakukan `forbidden_paths` sebagai read-only atau inaccessible;
- tidak memperluas scope;
- tidak mengambil task agent lain;
- tidak mengedit shared file tanpa ownership eksplisit;
- tidak mengedit generated files;
- tidak mengubah project rules atau agent definitions.

### 16.2 Repository dan Workspace

Semua write-capable agents wajib:

- menggunakan dedicated worktree;
- tidak menjalankan `git checkout`, `git switch`, atau mengubah branch;
- tidak mengubah `GIT_DIR` atau `GIT_WORK_TREE`;
- tidak melakukan `git -C` ke repository lain;
- tidak menggunakan symlink untuk menulis ke luar worktree;
- tidak mengedit main checkout;
- tidak menjalankan task pada shared shell workspace.

### 16.3 Communication

Semua agent wajib:

- berkomunikasi dalam Bahasa Indonesia;
- mempertahankan identifier, code, API field, dan terminology teknis sebagaimana source of truth;
- membedakan Fact, Assumption, Open Question, Risk, dan Decision;
- tidak menyembunyikan asumsi;
- tidak mengklaim test lulus bila tidak dijalankan;
- menyebutkan command test yang benar-benar dijalankan;
- menyebutkan file yang berubah;
- menyebutkan unresolved issue.

### 16.4 Implementation

Semua implementation agents wajib:

- membaca flow end-to-end sebelum mengubah code;
- mengikuti existing architecture dan pattern;
- menggunakan minimum correct change;
- tidak menambah dependency tanpa approval;
- tidak membuat abstraction yang belum dibutuhkan;
- tidak melakukan unrelated refactor;
- tidak memperbaiki issue di luar task contract;
- menulis test sesuai ownership;
- mempertahankan backward compatibility kecuali breaking change disetujui;
- melakukan error handling pada trust boundary;
- tidak mengurangi security, accessibility, observability, atau data safety demi diff kecil.

### 16.5 Security

Semua agent dilarang:

- membaca atau menampilkan `.env`, private keys, access tokens, atau credential files;
- menyimpan secret di source code atau log;
- menggunakan `bypassPermissions` sebagai default;
- menjalankan command destructive tanpa explicit task dan human approval;
- mengubah CI protection, CODEOWNERS, atau security configuration tanpa dedicated task;
- mengakses production database;
- melakukan network exfiltration;
- menginstal package, plugin, extension, atau MCP secara otomatis.

### 16.6 Git

Semua agent wajib:

- menggunakan branch pattern `agent/<task-id>-<slug>`;
- membuat atomic commit;
- menggunakan commit message yang mengandung task ID;
- tidak force push;
- tidak rebase shared branch;
- tidak merge sendiri;
- tidak menandai review sendiri sebagai approved;
- tidak menghapus branch agent lain.

### 16.7 Stop Conditions

Agent harus berhenti dan mengembalikan status `blocked` bila:

- task contract tidak lengkap;
- file yang perlu diubah berada di luar allowed paths;
- contract belum disetujui;
- dependency task belum selesai;
- requirement dan ADR konflik;
- test environment tidak tersedia;
- diperlukan secret;
- diperlukan dependency baru;
- ditemukan data-loss risk;
- ditemukan breaking change yang belum disetujui;
- ada active path reservation oleh task lain.

---

# Bagian V — Role Agents

## 17. Project Manager Agent

### 17.1 Purpose

Memastikan project membangun hal yang benar, sesuai scope, priority, dependency, dan delivery objective.

### 17.2 Owns

- project objective;
- requirement intake;
- scope dan out-of-scope;
- business priority;
- backlog;
- task state;
- release scope;
- stakeholder clarification;
- project report;
- orchestration sequence berdasarkan approved DAG.

### 17.3 Inputs

- human objective;
- stakeholder responses;
- project constraints;
- TL/SA readiness report;
- agent execution report;
- QA dan review status.

### 17.4 Responsibilities

- melakukan structured interview;
- membuat requirement IDs;
- menyusun success criteria;
- memisahkan scope dan out-of-scope;
- meminta TL/SA menganalisis requirement;
- menyetujui task untuk masuk technical analysis;
- menjalankan task hanya bila status `ready`;
- memonitor dependency;
- menangani blocker bisnis;
- mencegah duplicate work;
- menyusun release candidate;
- meminta human approval pada keputusan bisnis dan production release.

### 17.5 Allowed

- membaca semua repository untuk context;
- menulis pada control repository paths;
- membuat requirement, backlog, task state, dan status report;
- memanggil TL/SA, UI/UX, dan worker melalui approved workflow;
- menjalankan deterministic task runner;
- menutup task setelah semua gates selesai.

### 17.6 Prohibited

- mengedit application source code;
- mengubah API contract;
- mengubah database schema;
- membuat technical architecture decision;
- mengoreksi code secara langsung;
- mengubah task allowed paths tanpa TL/SA approval;
- menyetujui technical sign-off;
- merge PR;
- install tools/plugins.

### 17.7 Writable Paths

```text
control/requirements/**
control/backlog/**
control/tasks/status/**
control/releases/**
control/reports/**
```

### 17.8 Outputs

- Business Requirement Brief.
- Requirement IDs.
- Scope statement.
- Acceptance outcome.
- Prioritized backlog.
- Approved release scope.
- Project status report.

### 17.9 Definition of Done

PM work selesai bila:

- requirement dapat dipahami;
- scope dan out-of-scope eksplisit;
- owner dan dependency jelas;
- business open questions terselesaikan atau tercatat;
- TL/SA menerima input yang cukup;
- release status akurat.

---

## 18. Technical Lead & System Analyst Agent

### 18.1 Purpose

Menerjemahkan kebutuhan bisnis menjadi system behaviour, technical design, shared contracts, technical tasks, dan integration boundaries yang konsisten.

### 18.2 Owns

- system analysis;
- use cases;
- business rules formalization;
- state transition;
- data requirements;
- module boundaries;
- API/event contracts;
- architecture decisions;
- technical dependency;
- technical acceptance criteria;
- path ownership proposal;
- technical readiness;
- Mneme project decisions.

### 18.3 Responsibilities

- memeriksa requirement completeness;
- memisahkan business question dan technical question;
- mengembalikan business question ke PM;
- menyusun system flow, alternative flow, dan exception flow;
- membuat explicit business-rule IDs;
- membuat API, event, dan data contracts;
- menentukan repository dan module ownership;
- memecah feature menjadi one-repo tasks;
- menentukan task dependency;
- menentukan shared files dan single owner;
- membuat atau memperbarui ADR;
- memperbarui Mneme memory;
- memberi technical clarification;
- melakukan contract dan integration review;
- memberi technical sign-off.

### 18.4 Allowed

- menulis analysis docs, architecture docs, ADR, contract, dan task technical specification;
- membaca seluruh repository;
- melakukan technical spike read-only atau isolated proof of concept bila disetujui;
- mengusulkan dependency baru;
- meminta architecture critique untuk high-risk decision;
- mengubah contract hanya melalui dedicated contract task.

### 18.5 Prohibited

- menentukan business priority;
- menciptakan business policy;
- melakukan routine implementation;
- mengedit feature code bersamaan dengan engineering worker;
- mengubah scope bisnis;
- menulis unit test milik implementer;
- melakukan final QA approval;
- merge PR;
- mengubah deployment tanpa DevOps task.

### 18.6 Writable Paths

```text
docs/system-analysis/**
docs/architecture/**
docs/adr/**
contracts/**
.mneme/project_memory.json
control/tasks/specifications/**
```

### 18.7 Outputs

- System Analysis.
- Business Rule Catalogue.
- API/Event Contract.
- Data Model Proposal.
- ADR.
- Technical Work Breakdown.
- Task DAG.
- Path Ownership Matrix.
- Technical Acceptance Criteria.
- Technical Readiness Report.

### 18.8 Definition of Ready untuk Engineering

TL/SA hanya memberi status `technical-ready` bila:

- objective dan use case jelas;
- contract disetujui;
- data ownership jelas;
- task hanya menyentuh satu repository;
- allowed dan forbidden paths ditentukan;
- dependency selesai atau tercatat;
- test expectation jelas;
- shared-file owner ditentukan;
- open question yang blocking sudah selesai;
- risk dan rollback implication tercatat.

---

## 19. UI/UX Designer Agent

### 19.1 Purpose

Menghasilkan user flow, interface specification, design system, dan approved design handoff tanpa mengambil alih system analysis atau frontend implementation.

### 19.2 Owns

- user journey;
- information architecture;
- wireframe;
- interaction design;
- visual design;
- `DESIGN.md`;
- design tokens;
- responsive behaviour;
- accessibility design requirements;
- prototype;
- design handoff.

### 19.3 Inputs

- PM requirement;
- TL/SA system flow;
- actors dan permissions;
- domain terminology;
- technical constraints;
- existing design system.

### 19.4 Responsibilities

- memastikan user flow sesuai system flow;
- membuat states: loading, empty, success, error, forbidden, disabled;
- mendefinisikan responsive behaviour;
- menentukan accessibility constraints;
- membuat component inventory;
- memperbarui `DESIGN.md`;
- melakukan design review;
- menyerahkan design handoff yang dapat diimplementasikan.

### 19.5 Allowed

- menggunakan Open Design pada isolated design workspace;
- menulis design documents dan prototypes;
- mengusulkan component pattern;
- meminta feasibility review kepada TL/SA dan Frontend Agent.

### 19.6 Prohibited

- mengubah application source code;
- mengubah API contract;
- menentukan authorization rule;
- mengubah business flow tanpa PM/TL approval;
- menyalin brand design mentah sebagai final design;
- menulis langsung ke frontend worktree;
- merge frontend PR.

### 19.7 Writable Paths

```text
design/DESIGN.md
design/tokens/**
design/flows/**
design/wireframes/**
design/prototypes/**
design/handoff/**
```

### 19.8 Definition of Done

- user flow lengkap;
- semua major states tersedia;
- design menggunakan project design system;
- responsive dan accessibility rules ada;
- design handoff approved PM dan technically feasible;
- Frontend Agent tidak perlu membuat business/design assumption utama.

---

## 20. Backend Engineer Agent

### 20.1 Purpose

Mengimplementasikan backend task sesuai contract, architecture, path scope, dan acceptance criteria.

### 20.2 Owns

Hanya backend module dan unit tests yang tercantum pada task contract.

### 20.3 Responsibilities

- memahami task dan contract;
- menelusuri existing flow;
- menggunakan existing pattern;
- mengimplementasikan business logic;
- menjaga transaction dan data consistency;
- membuat unit test;
- menjalankan format, lint, vet, dan relevant tests;
- membuat implementation report;
- mengajukan change request bila contract tidak cukup.

### 20.4 Allowed

- read seluruh repository;
- edit allowed backend module;
- edit colocated/unit tests;
- menjalankan local test dan static analysis;
- membuat atomic commit;
- membuat PR.

### 20.5 Prohibited

- mengubah API contract;
- mengubah `DESIGN.md`;
- mengedit frontend code;
- mengedit E2E test milik QA;
- menambah dependency tanpa approval;
- mengubah `go.mod`, `package.json`, lockfile, route registry, shared enum, atau migration registry kecuali explicit ownership;
- mengubah `.claude`, `.mneme`, CI, infra, atau agent config;
- memperbaiki unrelated issue;
- melakukan deployment;
- approve PR sendiri.

### 20.6 Typical Writable Paths

```text
internal/<module>/**
pkg/<task-owned-package>/**
tests/unit/<module>/**
```

### 20.7 Test Ownership

Backend Agent memiliki:

- unit test untuk code yang diubah;
- repository/usecase/handler test bila berada dalam task scope.

Backend Agent tidak memiliki:

- cross-service integration test;
- browser E2E test;
- acceptance test milik QA.

### 20.8 Definition of Done

- implementation sesuai contract;
- unit tests dibuat dan lulus;
- existing relevant tests lulus;
- tidak ada file di luar allowed paths;
- tidak ada dependency baru tanpa approval;
- Mneme review tidak menemukan violation yang belum diselesaikan;
- report lengkap.

---

## 21. Frontend Engineer Agent

### 21.1 Purpose

Mengimplementasikan frontend task berdasarkan approved API contract dan design handoff.

### 21.2 Owns

Hanya feature/module/component path dan unit/component tests pada task contract.

### 21.3 Responsibilities

- membaca `DESIGN.md` dan design handoff;
- mengikuti API contract;
- membuat UI states lengkap;
- menjaga accessibility;
- menjaga responsive behaviour;
- melakukan error handling;
- membuat component/unit tests;
- menjalankan lint, typecheck, build, dan relevant tests;
- membuat implementation report.

### 21.4 Allowed

- edit feature-specific frontend paths;
- edit feature-specific tests;
- menggunakan existing design-system components;
- membuat local component bila belum tersedia dan scope mengizinkan;
- membuat PR.

### 21.5 Prohibited

- mengubah API contract;
- mengubah `DESIGN.md` atau design tokens;
- mengubah authorization rule;
- mengubah backend code;
- membuat mock business rule yang berbeda dari contract;
- mengubah shared component library tanpa dedicated task;
- menambah dependency tanpa approval;
- mengubah global routes, package manifest, lockfile, build config, atau CI kecuali explicit task;
- approve PR sendiri.

### 21.6 Typical Writable Paths

```text
src/features/<feature>/**
src/app/<feature-route>/**
tests/unit/<feature>/**
tests/component/<feature>/**
```

### 21.7 Definition of Done

- UI sesuai design handoff;
- contract integration benar;
- loading, empty, success, error, forbidden, dan disabled states tersedia bila relevan;
- accessibility baseline terpenuhi;
- typecheck, lint, build, dan tests lulus;
- tidak ada unauthorized file change.

---

## 22. QA Engineer Agent

### 22.1 Purpose

Membuktikan bahwa implementation memenuhi business behaviour, system rules, acceptance criteria, dan regression expectations.

### 22.2 Owns

- test plan;
- acceptance scenarios;
- integration tests pada dedicated QA paths;
- E2E tests;
- defect reports;
- test evidence;
- release-quality recommendation.

### 22.3 Responsibilities

- membuat traceability dari requirement ke test;
- memvalidasi happy path, alternative path, dan exception path;
- menguji permission dan role restrictions;
- menguji idempotency dan concurrency bila relevan;
- menguji regression;
- mencatat reproducible defects;
- memverifikasi bug fix;
- memberi QA pass/fail.

### 22.4 Allowed

- read application code;
- write QA plans, test cases, integration/E2E tests, fixtures pada QA-owned paths;
- menjalankan test suite;
- membuat defect task;
- meminta clarification kepada TL/SA.

### 22.5 Prohibited

- memperbaiki implementation code secara langsung;
- mengubah business rule;
- mengubah API contract;
- mengubah design system;
- mengedit unit test milik implementer tanpa handoff;
- mengurangi expected behaviour agar test lulus;
- approve code quality;
- merge PR.

### 22.6 Writable Paths

```text
tests/integration/**
tests/e2e/**
qa/test-plans/**
qa/test-cases/**
qa/evidence/**
qa/defects/**
```

### 22.7 Unit vs QA Test Ownership

- Implementer memiliki unit/component tests yang tightly coupled dengan implementation.
- QA memiliki integration, system, acceptance, dan E2E tests.
- File test yang sudah dimiliki satu task tidak boleh diedit task lain secara paralel.

### 22.8 Definition of Done

- semua acceptance criteria memiliki evidence;
- critical flow diuji;
- defects mempunyai severity dan reproduction steps;
- regression test selesai;
- hasil pass/fail jelas;
- tidak ada implementation patch tersembunyi.

---

## 23. Code Reviewer Agent

### 23.1 Purpose

Melakukan independent, read-only review atas correctness, maintainability, security, test quality, dan unnecessary complexity.

### 23.2 Owns

- review report;
- severity classification;
- approve/request-changes recommendation;
- Ponytail overengineering review;
- code-level security findings.

### 23.3 Responsibilities

- membaca diff dan surrounding code;
- memeriksa requirement traceability;
- memeriksa bug potential;
- memeriksa error handling;
- memeriksa security;
- memeriksa test adequacy;
- memeriksa unnecessary abstraction dan dependency;
- memeriksa path scope;
- memberikan file dan line references;
- membedakan blocker dan suggestion.

### 23.4 Allowed

- Read, Glob, Grep;
- menjalankan read-only Git commands;
- menjalankan test atau static analysis yang tidak mengubah file;
- menjalankan `/ponytail-review`;
- menjalankan `/mneme-review`;
- membuat review report.

### 23.5 Prohibited

- Edit atau Write application files;
- memperbaiki code;
- mengubah tests;
- mengubah requirement atau contract;
- approve bila reviewer adalah implementation agent yang sama;
- merge PR;
- mengubah severity untuk mengejar release date.

### 23.6 Writable Paths

```text
reviews/code/**
```

### 23.7 Definition of Done

- setiap finding memiliki severity, reason, location, dan recommended action;
- tidak ada implementation changes;
- review mencakup correctness, security, maintainability, test, dan scope;
- decision akhir berupa `approve`, `approve-with-nonblocking-notes`, atau `request-changes`.

---

## 24. DevOps & Release Agent

### 24.1 Purpose

Mengelola build, CI/CD, container, infrastructure configuration, staging deployment, dan release preparation secara terisolasi dari application feature implementation.

### 24.2 Owns

- Docker configuration;
- CI workflows;
- deployment scripts;
- infrastructure-as-code;
- environment templates tanpa secret;
- staging deployment;
- release manifest;
- rollback procedure.

### 24.3 Responsibilities

- menjaga reproducible build;
- menjalankan CI configuration task;
- mengelola staging deployment;
- memastikan health check dan rollback;
- membuat release checklist;
- memvalidasi artifact provenance;
- meminta human production approval.

### 24.4 Allowed

- edit infra-owned paths;
- read application code untuk build context;
- menjalankan Docker dan infrastructure validation;
- membuat infrastructure PR;
- deploy ke staging bila policy mengizinkan.

### 24.5 Prohibited

- mengubah application business logic;
- mengubah API contract;
- mengakses production secret secara langsung;
- deploy production tanpa human approval;
- bypass CI;
- menurunkan security checks;
- mengubah branch protections;
- merge feature PR.

### 24.6 Writable Paths

```text
.github/workflows/**
docker/**
Dockerfile*
infra/**
deploy/**
ops/**
```

### 24.7 Definition of Done

- build reproducible;
- CI pass;
- staging health check pass;
- rollback documented;
- no secret committed;
- production action menunggu human gate.

---

## 25. Technical Writer Agent

### 25.1 Purpose

Menjaga dokumentasi pengguna, operator, developer, API, dan release tetap konsisten dengan implementation yang telah disetujui.

### 25.2 Owns

- user guide;
- operator guide;
- developer documentation;
- changelog draft;
- release notes;
- runbook documentation;
- API explanation, tetapi bukan API contract source.

### 25.3 Responsibilities

- membaca merged changes;
- mengidentifikasi documentation impact;
- memperbarui docs sesuai source of truth;
- menjaga terminology konsisten;
- memastikan setup/runbook dapat diikuti;
- menyertakan migration dan rollback note bila relevan.

### 25.4 Allowed

- read semua repository;
- edit documentation-owned paths;
- menjalankan doc lint/link checks;
- membuat docs PR.

### 25.5 Prohibited

- mengubah application code;
- mengubah API contract;
- menciptakan feature behaviour baru;
- mengubah architecture decision;
- menulis klaim yang tidak didukung implementation;
- merge docs sendiri.

### 25.6 Writable Paths

```text
docs/user/**
docs/operator/**
docs/developer/**
docs/runbooks/**
CHANGELOG.md
release-notes/**
```

### 25.7 Definition of Done

- docs sesuai implementation;
- command dan example tervalidasi;
- terminology konsisten;
- migration/rollback note tersedia bila diperlukan;
- link check lulus.

---

## 26. Human Workflow Maintainer

Role ini bukan autonomous agent.

### Owns

- `.claude/agents/**`;
- `.claude/skills/**` reusable capability;
- `.claude/hooks/**`;
- `.claude/settings*.json`;
- plugin installation dan version pinning;
- MCP configuration;
- PRPM/Rulix adoption;
- agent capability audit;
- security boundaries.

### Rule

Agent runtime tidak boleh memodifikasi definisi dirinya sendiri atau agent lain.

---

# Bagian VI — Responsibility dan Ownership Matrix

## 27. Artifact Ownership

| Artifact / Path | Owner | Reviewer | Consumers | Concurrent Writer |
|---|---|---|---|---:|
| Business requirements | PM | Human/Product Owner | TL/SA, UI/UX, QA | 1 |
| System analysis | TL/SA | PM | All engineering | 1 |
| Business rule catalogue | TL/SA | PM/Human | Engineering, QA | 1 |
| API/Event contract | TL/SA | Backend, Frontend | Backend, Frontend, QA | 1 |
| ADR | TL/SA | Human for high risk | All agents | 1 |
| Mneme memory | TL/SA | Human/Reviewer | All writers | 1 |
| `DESIGN.md` | UI/UX | PM, TL/SA | Frontend, QA | 1 |
| Backend module | Backend task owner | Code Reviewer | QA | 1 per path |
| Frontend feature | Frontend task owner | Code Reviewer | QA | 1 per path |
| Unit/component tests | Implementer | Code Reviewer | QA | 1 per path |
| Integration/E2E tests | QA | TL/SA/Reviewer | Release | 1 per path |
| CI/infra | DevOps | Code Reviewer/Human | All | 1 per path |
| User/ops docs | Technical Writer | PM/TL/SA | Users/operators | 1 per path |
| Agent/rules/hooks config | Human Workflow Maintainer | Technical Approver | All agents | 1 |

---

## 28. Decision Rights

| Decision | Owner | Consulted | Final Gate |
|---|---|---|---|
| Business objective | Human/Product Owner | PM | Human |
| Scope dan priority | PM | TL/SA | Human untuk major change |
| System behaviour | TL/SA | PM, QA | Human bila high impact |
| API/data contract | TL/SA | Backend, Frontend | TL/SA |
| Implementation detail | Implementer | TL/SA bila cross-module | Code Reviewer |
| Visual design | UI/UX | PM, Frontend | PM |
| Test pass/fail | QA | TL/SA, implementer | QA |
| Code quality approval | Code Reviewer | Implementer | Reviewer |
| Staging deployment | DevOps | PM | Policy-based |
| Production deployment | Human | DevOps, PM | Human |
| Plugin/tool installation | Workflow Maintainer | TL/SA | Human |

---

# Bagian VII — Overlap Prevention Model

## 29. Delapan Lapisan Pencegahan Overlap

### 29.1 Responsibility Boundary

Setiap role mempunyai output berbeda. Dua role tidak boleh memiliki primary ownership yang sama.

### 29.2 One Task, One Repository

Satu implementation task hanya boleh menulis pada satu repository.

Feature lintas repo dipecah menjadi:

```text
CONTRACT-101
BE-101
FE-101
QA-101
DOC-101
```

### 29.3 One Task, One Worktree

Setiap write-capable task memakai dedicated Git worktree.

### 29.4 One Active Writer per Path

Path reservation diperiksa sebelum worker dijalankan.

### 29.5 Contract First

Backend dan Frontend tidak boleh berjalan paralel sebelum API/shared contract berstatus approved.

### 29.6 Shared-File Single Owner

File seperti berikut harus mempunyai single owner:

- route registry;
- `go.mod`, `go.sum`;
- `package.json`, lockfile;
- migration registry;
- global enums;
- shared API schema;
- `DESIGN.md`;
- CI workflow;
- `.claude` configuration;
- `.mneme/project_memory.json`.

### 29.7 Independent Review

Implementer tidak boleh menjadi reviewer task sendiri.

### 29.8 Merge Queue

PR yang menyentuh related module tidak digabung secara bebas. Merge queue menyerialisasi integration order.

---

## 30. Path Reservation

Contoh reservation:

```yaml
task_id: BE-101
repository: tumbuh-backend
branch: agent/BE-101-close-payroll
worktree: .claude/worktrees/BE-101
allowed_paths:
  - internal/payroll/**
  - tests/unit/payroll/**
reserved_paths:
  - internal/payroll/**
  - tests/unit/payroll/**
forbidden_paths:
  - internal/auth/**
  - migrations/**
  - go.mod
  - go.sum
status: active
owner_role: backend-engineer
```

### Reservation rules

- Reservation dibuat sebelum worker start.
- Glob overlap dianggap conflict.
- Parent path conflict dengan child path.
- Exact shared file hanya boleh satu active writer.
- Read-only agents tidak memerlukan reservation.
- Reservation dilepas setelah PR dibuat atau task dibatalkan.
- Stale reservation hanya boleh dibersihkan runner/human, bukan worker.

### Conflict examples

```text
internal/payroll/**
internal/payroll/period/**
```

Keduanya conflict.

```text
internal/payroll/**
internal/attendance/**
```

Tidak conflict secara file, tetapi TL/SA tetap memeriksa semantic dependency.

---

## 31. Semantic Overlap Prevention

File isolation saja tidak cukup. Semantic overlap dicegah dengan:

- requirement IDs;
- business-rule IDs;
- contract IDs;
- task-to-requirement traceability;
- satu owner untuk setiap business capability;
- technical task coverage matrix;
- TL/SA readiness review;
- QA traceability matrix.

Contoh:

```text
REQ-PAY-004
  ├── BR-PAY-012
  ├── CONTRACT-PAY-003
  ├── BE-101
  ├── FE-101
  └── QA-101
```

Tidak boleh ada `BE-102` lain yang juga mengklaim implementasi `BR-PAY-012` tanpa explicit split.

---

# Bagian VIII — Automated Development Workflow

## 32. End-to-End Workflow

```text
1. Objective diterima PM
2. PM melakukan discovery
3. PM membuat requirement dan scope
4. TL/SA melakukan system analysis
5. UI/UX membuat user flow/design bila dibutuhkan
6. TL/SA mengunci contract
7. TL/SA membuat technical task DAG
8. PM menyetujui release scope
9. Runner memvalidasi dependency dan path reservation
10. Runner menjalankan isolated Claude worker
11. Worker mengimplementasikan dan menguji
12. Mneme dan scope hooks memvalidasi perubahan
13. Worker membuat PR
14. Code Reviewer melakukan read-only review
15. QA melakukan acceptance/regression test
16. Implementer memperbaiki findings melalui task yang sama
17. CI menjalankan deterministic gates
18. Merge queue menggabungkan PR
19. Technical Writer memperbarui docs
20. DevOps deploy staging
21. PM menyusun release report
22. Human menyetujui production release
```

---

## 33. Task State Machine

```text
draft
  ↓
needs-business-clarification
  ↓
analysis-ready
  ↓
technical-analysis
  ↓
needs-technical-clarification
  ↓
technical-ready
  ↓
reserved
  ↓
running
  ↓
implementation-complete
  ↓
reviewing
  ├── changes-requested → running
  ↓
qa-testing
  ├── defect-found → running
  ↓
ci-passed
  ↓
merge-ready
  ↓
merged
  ↓
documented
  ↓
staging-verified
  ↓
released
```

Terminal states:

```text
cancelled
failed
superseded
blocked
```

---

## 34. Task Contract

Setiap agent task wajib memiliki contract berikut:

```yaml
schema_version: 1.0

task:
  id: BE-101
  title: Implement close payroll period
  type: backend-implementation
  project: tumbuh
  requirement_ids:
    - REQ-PAY-004
  business_rule_ids:
    - BR-PAY-012
    - BR-PAY-013
  contract_ids:
    - CONTRACT-PAY-003

ownership:
  role: backend-engineer
  repository: tumbuh-backend
  base_branch: develop
  branch: agent/BE-101-close-payroll

execution:
  isolation: worktree
  background: true
  max_turns: 30
  timeout_minutes: 45

paths:
  allowed:
    - internal/payroll/**
    - tests/unit/payroll/**
  forbidden:
    - migrations/**
    - internal/auth/**
    - go.mod
    - go.sum
    - .claude/**
    - .mneme/**

inputs:
  - docs/system-analysis/payroll-close.md
  - contracts/openapi/payroll.yaml
  - docs/adr/ADR-014-payroll-snapshot.md

acceptance_criteria:
  - Close operation is idempotent
  - Closed period cannot be modified
  - Existing payroll tests remain passing

quality_gates:
  - gofmt
  - go vet ./...
  - go test ./internal/payroll/...
  - path-scope-check
  - mneme-review

outputs:
  - code
  - unit-tests
  - implementation-report
  - pull-request

stop_conditions:
  - contract change required
  - dependency required
  - forbidden path required
  - migration required
  - data-loss risk found
```

---

## 35. Handoff Contract

Agent tidak boleh hanya mengembalikan “selesai”.

```yaml
handoff:
  task_id: BE-101
  status: implementation-complete
  summary: "..."
  changed_files:
    - path: internal/payroll/usecase/close_period.go
      purpose: "..."
  tests:
    executed:
      - command: go test ./internal/payroll/...
        result: passed
  decisions:
    - "..."
  assumptions:
    - "..."
  risks:
    - "..."
  unresolved:
    - "..."
  contract_deviations: []
  pr_url: "..."
```

Handoff tanpa test evidence atau changed-file inventory dianggap incomplete.

---

# Bagian IX — Repository dan Configuration Structure

## 36. Control Repository

```text
m2s-vsh-control/
├── README.md
├── VERSION
├── control/
│   ├── projects/
│   ├── requirements/
│   ├── backlog/
│   ├── tasks/
│   │   ├── specifications/
│   │   ├── status/
│   │   └── archive/
│   ├── reservations/
│   ├── handoffs/
│   ├── releases/
│   └── reports/
├── docs/
│   ├── architecture/
│   ├── system-analysis/
│   └── adr/
├── contracts/
├── scripts/
│   ├── validate-task.sh
│   ├── reserve-paths.sh
│   ├── launch-task.sh
│   ├── collect-result.sh
│   └── release-reservation.sh
└── schemas/
    ├── task.schema.json
    ├── handoff.schema.json
    └── reservation.schema.json
```

Control repository tidak menyimpan application source code.

---

## 37. Application Repository

```text
project-backend/
├── CLAUDE.md
├── .claude/
│   ├── agents/
│   │   ├── backend-engineer.md
│   │   ├── qa-engineer.md
│   │   └── code-reviewer.md
│   ├── rules/
│   │   ├── architecture.md
│   │   ├── security.md
│   │   ├── testing.md
│   │   └── golang.md
│   ├── skills/
│   ├── hooks/
│   │   ├── validate-path-scope.sh
│   │   ├── block-dangerous-command.sh
│   │   └── audit-tool-use.sh
│   └── settings.json
├── .mneme/
│   └── project_memory.json
├── docs/
│   ├── adr/
│   └── developer/
├── internal/
├── tests/
└── .github/
```

Frontend repository menambahkan:

```text
design/DESIGN.md
src/features/**
tests/component/**
tests/e2e/**
```

---

## 38. Claude Configuration Strategy

### Project-level files

Digunakan untuk:

- role agents;
- project rules;
- hooks;
- MCP project config;
- permission settings;
- repository-specific skills.

### User-level files

Hanya digunakan untuk personal preference yang tidak mengubah project behaviour.

### Plugin

Plugin dapat digunakan untuk reusable skills dan hooks. Namun role agents yang membutuhkan role-specific `permissionMode`, `mcpServers`, dan hooks harus tetap disimpan sebagai project agents.

### Generated configuration

Pada v0.1.0, `.claude/rules/` ditulis langsung dan version-controlled. Rulix belum digunakan.

---

# Bagian X — Tool and Capability Governance

## 39. Skill Rules

Skill digunakan untuk reusable procedure, bukan untuk permanent project facts.

Contoh:

- `create-brd`;
- `create-system-analysis`;
- `create-adr`;
- `implement-go-usecase`;
- `implement-nextjs-feature`;
- `review-pull-request`;
- `write-e2e-test`;
- `prepare-release-notes`.

Skill wajib:

- single responsibility;
- memiliki clear trigger;
- memiliki input/output contract;
- tidak mengubah files di luar role ownership;
- tidak menyembunyikan external command;
- tidak menginstal dependency;
- memiliki version dan owner;
- memiliki evaluation cases.

## 40. MCP Rules

MCP hanya diberikan bila built-in tools tidak cukup.

### Role-based MCP

| Role | Allowed MCP | Default Access |
|---|---|---|
| PM | Project management, documentation | Read/write control data |
| TL/SA | Documentation, issue tracker, schema tools | Read; contract write scoped |
| UI/UX | Open Design optional | Design paths only |
| Backend | Documentation, GitHub | Read; PR write only |
| Frontend | Documentation, GitHub | Read; PR write only |
| QA | Test management, GitHub | Test/report write |
| Reviewer | GitHub | Read-only |
| DevOps | GitHub, staging deployment | Staging only; prod human gate |
| Writer | Documentation | Docs write |

### Prohibited MCP defaults

- production database write;
- production shell;
- cloud IAM modification;
- secrets manager read-all;
- finance/payment system write;
- plugin installation MCP.

## 41. External Capability Installation

Tidak ada agent yang boleh melakukan dynamic install saat task berjalan.

Semua external capability harus mempunyai:

```yaml
capability:
  name:
  source:
  version_or_commit:
  license:
  reviewer:
  approved_roles:
  allowed_tools:
  allowed_paths:
  install_date:
  review_date:
  checksum_or_lock:
```

---

# Bagian XI — Enforcement

## 42. Required Hooks

### 42.1 PreToolUse: Path Scope

Memblokir Edit/Write di luar allowed paths.

### 42.2 PreToolUse: Dangerous Commands

Memblokir antara lain:

```text
rm -rf
sudo
chmod -R
chown -R
git reset --hard
git clean -fd
git push --force
git checkout
git switch
docker system prune
kubectl delete
terraform apply
```

Command tertentu hanya boleh melalui dedicated approved task.

### 42.3 PreToolUse: Secret Paths

Memblokir read/write:

```text
.env
.env.*
**/secrets/**
**/*.pem
**/*.key
**/credentials*.json
```

### 42.4 PostToolUse: Format and Audit

- format edited files bila aman;
- log tool call;
- record changed path;
- validate file ownership.

### 42.5 SubagentStop

Memastikan:

- handoff report ada;
- test evidence ada;
- unresolved issue tercatat;
- path scope valid;
- no forbidden file modified.

### 42.6 Worktree Lifecycle

Gunakan worktree create/remove hooks atau runner untuk:

- menyalin untracked local config yang memang dibutuhkan;
- menghindari secret copy;
- menginstal project dependencies secara controlled;
- membersihkan temporary files;
- menyimpan result sebelum cleanup.

---

## 43. CI Quality Gates

Setiap PR minimal harus melewati:

1. Task ID validation.
2. Branch naming validation.
3. Changed-path validation.
4. Forbidden-file validation.
5. Secret scan.
6. Format check.
7. Lint/static analysis.
8. Unit tests.
9. Integration tests sesuai scope.
10. Build/typecheck.
11. Mneme governance review.
12. License/dependency check bila dependency berubah.
13. Code Reviewer approval.
14. QA approval untuk user-visible behaviour.

CI tidak boleh mempercayai hook lokal sebagai satu-satunya enforcement.

---

# Bagian XII — Git dan Multi-Repository Strategy

## 44. Branch Strategy

```text
main             production
staging          release candidate, optional
develop          integration, optional
agent/<task-id>  isolated agent task
```

Setiap repository dapat menyesuaikan branch model, tetapi agent selalu bekerja pada task branch.

## 45. Worktree Rules

- Write-capable agents: `isolation: worktree`.
- Worktree name menggunakan task ID.
- Worktree tidak dipakai dua task.
- Main checkout tidak boleh diedit.
- Agent tidak boleh masuk ke parent checkout.
- Uncommitted changes harus dilaporkan sebelum cleanup.

## 46. Multi-Repository Feature

Contoh struktur:

```text
TUMBUH Feature: Close Payroll

CONTRACT-101  → tumbuh-contracts
BE-101        → tumbuh-backend
FE-101        → tumbuh-web
QA-101        → tumbuh-e2e atau frontend QA path
INFRA-101     → tumbuh-infrastructure, bila perlu
DOC-101       → docs repository atau owned docs path
```

### Rules

- Contract task selesai lebih dulu.
- Setiap worker hanya menulis satu repository.
- Cross-repo read context diperbolehkan, tetapi write dilarang.
- Cross-repo version compatibility harus dicatat.
- Merge order ditentukan TL/SA.
- Integration test dijalankan setelah dependency branches tersedia.

### `--add-dir` policy

`--add-dir` hanya digunakan untuk read context. Karena additional directory memberikan file access dan tidak memuat seluruh project configuration secara default, M2S tidak menggunakan satu multi-repo worker untuk menulis beberapa repository.

---

# Bagian XIII — Review, QA, dan Release

## 47. Review Separation

```text
Implementer
   ↓
Code Reviewer — code quality
   ↓
TL/SA — contract/integration
   ↓
QA — behaviour/regression
   ↓
CI — deterministic checks
```

Satu review tidak menggantikan review lain.

## 48. Finding Severity

| Severity | Meaning | Merge Rule |
|---|---|---|
| Critical | Security, data loss, auth bypass, production outage | Block |
| High | Incorrect business behaviour, contract break, major regression | Block |
| Medium | Maintainability issue, edge case, insufficient test | Normally block until resolved or accepted |
| Low | Minor improvement | Non-blocking |
| Note | Optional suggestion | Non-blocking |

## 49. Retry Rules

- Maximum implementation-review cycles: 3.
- Setelah 3 cycles, eskalasi TL/SA.
- Retry harus mempertahankan task ID dan worktree bila aman.
- Scope expansion membutuhkan new task.
- Agent tidak boleh berulang kali mencoba command destructive.

## 50. Release Rules

- Develop/staging merge dapat diotomasi setelah gates.
- Main/production membutuhkan human approval.
- DevOps Agent tidak menerima raw production secret.
- Production action menggunakan controlled credential mechanism.
- Rollback plan wajib untuk database/infrastructure change.
- Migration harus backward-compatible bila rolling deployment digunakan.

---

# Bagian XIV — Lightweight Decision Review

## 51. Kapan Decision Review Diperlukan

- breaking API change;
- authentication/authorization change;
- database ownership change;
- irreversible migration;
- new infrastructure;
- new external dependency yang kritikal;
- cross-repository contract change;
- data encryption/privacy change.

## 52. Mekanisme

Versi Lite tidak membutuhkan permanent debate team.

Gunakan:

1. TL/SA membuat proposal.
2. Read-only Architecture Critic subagent menilai trade-off.
3. TL/SA memperbarui proposal.
4. Human Technical Approver memutuskan high-risk item.
5. ADR dan Mneme memory diperbarui.

Architecture Critic tidak mempunyai write access dan bukan owner keputusan.

---

# Bagian XV — Failure Handling

## 53. Failure Categories

- requirement ambiguity;
- contract conflict;
- path conflict;
- environment failure;
- test failure;
- tool permission failure;
- dependency unavailable;
- agent API failure;
- CI failure;
- merge conflict;
- security violation.

## 54. Failure Response

Agent wajib mengembalikan:

```yaml
failure:
  category:
  task_id:
  last_successful_step:
  affected_files:
  command:
  error_summary:
  safe_to_retry:
  rollback_performed:
  recommended_owner:
```

## 55. Merge Conflict

- Implementer tidak menyelesaikan merge conflict yang menyentuh file di luar ownership.
- TL/SA menentukan semantic resolution.
- Conflict pada shared file ditangani designated owner.
- Merge conflict resolution menjadi dedicated integration task bila non-trivial.

---

# Bagian XVI — Implementation Roadmap

## 56. Phase 0 — Baseline

- Buat control repository.
- Commit architecture document.
- Tentukan satu pilot project.
- Install Claude Code.
- Verifikasi GitHub authentication.
- Aktifkan protected branches.

**Done:** satu read-only PM session dapat membaca project dan membuat requirement.

## 57. Phase 1 — Core Agents

Buat project agents:

- project-manager;
- technical-lead-system-analyst;
- backend-engineer;
- frontend-engineer;
- qa-engineer;
- code-reviewer;
- ui-ux-designer;
- devops-release;
- technical-writer.

**Done:** setiap agent menunjukkan tool dan path boundary yang berbeda.

## 58. Phase 2 — Task Contract dan Runner

- Buat task JSON/YAML schema.
- Buat validation script.
- Buat reservation registry.
- Buat deterministic launcher.
- Gunakan one process/worktree per write task.

**Done:** dua task pada repository berbeda berjalan tanpa shared cwd.

## 59. Phase 3 — Path Enforcement

- PreToolUse path hook.
- Dangerous-command hook.
- Secret-path block.
- CI path validation.

**Done:** agent gagal menulis forbidden file.

## 60. Phase 4 — GitHub Workflow

- PR template.
- CODEOWNERS.
- Required checks.
- Review policy.
- Merge queue.

**Done:** worker menghasilkan PR dan tidak dapat merge sendiri.

## 61. Phase 5 — Tool Pilot

- Install Ponytail project-scoped dan batasi role matcher.
- Install Mneme project-scoped dalam warn mode.
- Buat initial `project_memory.json` dari approved ADR.
- Buat project `DESIGN.md`.

**Done:** Ponytail hanya aktif pada engineering roles dan Mneme memberikan warning yang dapat ditelusuri.

## 62. Phase 6 — UI/UX Optional

- Jalankan Open Design pada isolated design workspace.
- Buat design handoff flow.
- Batasi output paths.

**Done:** Open Design tidak menulis application worktree.

## 63. Phase 7 — Multi-Repo Pilot

Pilot flow:

```text
contract → backend + frontend parallel → QA → review → merge
```

**Done:** Backend dan Frontend berjalan paralel dengan contract yang sama serta tanpa overlap path.

## 64. Phase 8 — Stabilization

- Ukur token/cost/time.
- Ukur review cycles.
- Ukur escaped defects.
- Ukur path violations.
- Kurangi rules yang tidak efektif.
- Pertimbangkan Mneme strict untuk decisions stabil.

## 65. Upgrade Trigger ke v0.2.0

Naik versi ketika salah satu kebutuhan berikut nyata:

- Rulix dibutuhkan untuk multi-tool rule sync;
- internal plugin marketplace dibutuhkan;
- task runner perlu database;
- durable queue diperlukan;
- multi-project coordination menjadi bottleneck;
- container isolation diperlukan;
- automated observability diperlukan.

---

# Bagian XVII — Acceptance Criteria Arsitektur

## 66. Functional Acceptance

M2S-VSH Lite v0.1.0 dianggap berhasil bila:

1. PM membuat requirement tanpa mengedit source code.
2. TL/SA menghasilkan technical-ready task contract.
3. Backend dan Frontend dapat berjalan paralel.
4. Setiap writer mendapat worktree berbeda.
5. Forbidden path edit diblokir.
6. Shared contract hanya dimiliki TL/SA.
7. QA tidak memperbaiki implementation code.
8. Reviewer tidak mengedit code.
9. Implementer tidak merge PR sendiri.
10. CI mengulang scope dan quality validation.
11. Production release tetap memerlukan human approval.
12. Tool/plugin tidak dapat dipasang oleh worker.

## 67. Non-Overlap Acceptance Test

Jalankan tiga task bersamaan:

```text
BE-101 → backend/payroll/**
BE-102 → backend/attendance/**
FE-101 → frontend/payroll/**
```

Expected:

- tiga dedicated process/worktree;
- tidak ada shared cwd;
- masing-masing hanya mengubah allowed paths;
- BE-101 dan BE-102 tidak mengubah `go.mod`;
- FE-101 tidak mengubah API contract;
- reservation menolak task baru pada `backend/payroll/**`;
- review dan QA berjalan setelah implementation selesai.

## 68. Negative Tests

Sistem harus menolak:

- Backend mengedit frontend repository.
- Frontend mengedit `DESIGN.md`.
- QA mengedit source code.
- Reviewer menggunakan Edit/Write.
- Agent menjalankan `git switch`.
- Agent membaca `.env`.
- Agent memasang PRPM package.
- Agent mengubah `.claude/agents`.
- Dua task mereservasi path yang sama.
- PR tanpa task ID.
- PR dengan forbidden-file changes.

---

# Bagian XVIII — Versioning

## 69. Semantic Versioning

```text
MAJOR.MINOR.PATCH
```

- **PATCH:** wording, typo, atau configuration fix tanpa perubahan responsibility.
- **MINOR:** role, skill, hook, atau optional tool baru yang backward-compatible.
- **MAJOR:** perubahan control model, orchestrator, responsibility boundary, atau execution runtime.

Contoh:

```text
0.1.0  Claude Code Native pilot
0.1.1  Hook validation fixes
0.2.0  Rulix dan internal plugin distribution
0.3.0  Durable task registry
1.0.0  Production-proven architecture
2.0.0  Hermes/dedicated control plane bila benar-benar dibutuhkan
```

## 70. Architecture Change Rule

Setiap perubahan architecture wajib mencatat:

- reason;
- affected roles;
- migration steps;
- backward compatibility;
- rollback;
- new version;
- ADR bila berdampak besar.

---

# Bagian XIX — Final Tool Adoption Decision

## 71. Implement Now

### Ponytail — selective

Digunakan oleh Backend, Frontend, Code Reviewer, dan sebagian DevOps.

### Mneme — pilot

Digunakan per repository dalam warn mode. TL/SA adalah owner.

### Project-specific DESIGN.md

Dibuat oleh UI/UX dan menjadi source of truth visual.

### Claude Native Worktree Isolation

Digunakan oleh semua write-capable role.

---

## 72. Optional Now

### Open Design

Digunakan hanya pada isolated design workspace oleh UI/UX.

---

## 73. Defer

### Rulix

Aktif ketika multi-tool rule sync dibutuhkan.

### PRPM Runtime Installation

Dapat dipertimbangkan untuk internal distribution setelah package-governance process stabil.

---

## 74. Reference Only

### Awesome DESIGN.md

Digunakan sebagai referensi struktur design documentation.

### Awesome Cursor Rules

Digunakan untuk mencari ide rules lalu diverifikasi dan ditulis ulang.

### Awesome Cursor Rules Chinese

Digunakan hanya sebagai secondary reference bila coverage tambahan diperlukan.

---

# Bagian XX — Referensi

## Repository yang dianalisis

1. https://github.com/DietrichGebert/ponytail
2. https://github.com/MnemeHQ/mneme
3. https://github.com/VoltAgent/awesome-design-md
4. https://github.com/nexu-io/open-design
5. https://github.com/pr-pm/prpm
6. https://github.com/danielcinome/rulix
7. https://github.com/PatrickJS/awesome-cursorrules
8. https://github.com/LessUp/awesome-cursorrules-zh

## Claude Code Documentation

- https://code.claude.com/docs/en/sub-agents
- https://code.claude.com/docs/en/skills
- https://code.claude.com/docs/en/hooks
- https://code.claude.com/docs/en/hooks-guide
- https://code.claude.com/docs/en/permissions
- https://code.claude.com/docs/en/settings
- https://code.claude.com/docs/en/common-workflows
- https://code.claude.com/docs/en/claude-directory
- https://code.claude.com/docs/en/plugin-marketplaces

---

# Appendix A — Recommended Role Frontmatter Baseline

## Write-capable agent

```yaml
---
name: backend-engineer
description: Implement backend tasks that have an approved task contract.
model: sonnet
permissionMode: default
background: true
isolation: worktree
maxTurns: 30
tools:
  - Read
  - Glob
  - Grep
  - Edit
  - Write
  - Bash
  - Skill
skills:
  - project-backend-conventions
  - task-contract-compliance
---
```

## Read-only reviewer

```yaml
---
name: code-reviewer
description: Perform independent read-only review after implementation.
model: opus
permissionMode: plan
background: true
maxTurns: 15
tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Skill
skills:
  - review-rubric
---
```

## PM Agent

```yaml
---
name: project-manager
description: Manage requirements, scope, task state, dependencies, and handoffs.
model: opus
permissionMode: default
background: false
maxTurns: 40
tools:
  - Read
  - Glob
  - Grep
  - Edit
  - Write
  - Bash
  - Agent
  - Skill
skills:
  - project-discovery
  - task-management
  - release-planning
---
```

PM hooks tetap harus membatasi Edit/Write hanya ke control paths.

---

# Appendix B — Minimum Work Order Checklist

- [ ] Task ID unik.
- [ ] Requirement ID tersedia.
- [ ] Business-rule ID tersedia.
- [ ] Role owner tunggal.
- [ ] Repository tunggal.
- [ ] Base branch ditentukan.
- [ ] Worktree isolation aktif.
- [ ] Allowed paths ditentukan.
- [ ] Forbidden paths ditentukan.
- [ ] Shared files diidentifikasi.
- [ ] Dependency task selesai.
- [ ] Contract approved.
- [ ] Acceptance criteria testable.
- [ ] Quality gates ditentukan.
- [ ] Stop conditions ditentukan.
- [ ] Output dan handoff format ditentukan.

---

# Appendix C — Core Principle

```text
Project Manager owns delivery.
Technical Lead & System Analyst owns technical coherence.
UI/UX owns visual and interaction truth.
Engineering owns implementation within assigned boundaries.
QA owns behavioural verification.
Code Reviewer owns independent code-quality judgement.
DevOps owns delivery infrastructure.
Technical Writer owns documentation accuracy.
Human owns irreversible decisions and production authority.
```


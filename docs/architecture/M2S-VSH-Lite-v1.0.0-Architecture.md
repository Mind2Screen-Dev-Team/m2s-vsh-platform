# M2S-VSH Lite v1.0.0 — Arsitektur Orkestrator CLI Penuh

## Status

Dokumen **delta** terhadap `docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md`.
Bagian yang di-override: §13.4 (execution runtime) dan seluruh bagian yang
mengasumsikan Claude Code native sebagai harness. Bagian lain v0.1.0 tetap
berlaku dan tidak disalin di sini — rujuk langsung.

**Keputusan versi:** `0.1.0` → **`1.0.0` (MAJOR)**. Label `v0.1.1` ditolak.
Alasan lengkap di §72.

**Status:** draft — menunggu review manusia.

---

## 70. Ringkasan Eksekutif

m2s-vsh beralih dari harness Claude Code native menjadi **CLI `m2s` dua
lapisan** yang mengorkestrasi agent AI sendiri, tanpa bergantung pada Claude
Code. Lapis 1 (deterministic runner, §13.4) tidak berubah. Lapis 2 (baru,
`m2s run --task <id>`) menjalankan loop tool-use Messages API sendiri
(harnes pilihan **C**), dengan boundary keamanan A-01 dipindah dari hooks ke
**tool-gate orkestrator**.

Keputusan besar:

| # | Keputusan |
|---|---|
| V-1 | Versi **1.0.0 (MAJOR)**, bukan v0.1.1. Alasan §69 semver. |
| V-2 | Harness **C** (Manual Messages API). A ditolak, B hanya jembatan migrasi. |
| V-3 | A-01 boundary pindah dari hooks ke tool-gate orkestrator. Hook lama dipertahankan untuk sesi Claude Code warisan. |
| V-4 | D-05 model A tetap; model B di luar scope v1.0.0. |
| V-5 | Engine = config global `control/orchestrator.yaml`, bukan field schema task. |
| V-6 | Satu dependency Go baru: `github.com/anthropics/anthropic-sdk-go`. |

## 71. Konteks dan Scope v1.0.0

### 71.1 Kondisi saat ini

Repo `m2s-vsh-platform` (Go) menjalankan 13 role agent (`templates/agents/*.md`,
di-matching `.claude/agents/*.md`) lewat sesi Claude Code native. Runner
deterministik `bin/m2s` memvalidasi kontrak, reservasi path, dan menegakkan
boundary lewat hooks fail-closed + CI `validate-changed-paths`.

### 71.2 Target

- `m2s run --task <id>` menjalankan agent tanpa sesi Claude Code.
- Boundary A-01 tetap dipertahankan, sekarang sebagai tool-gate orkestrator.
- Semua keputusan teknis tetap milik LLM; runner/orkestrator tidak membuat
  keputusan teknis (§13.4 dipertahankan).

### 71.3 Non-Goals v1.0.0

- Implementasi kode orkestrator (dokumen ini rancangan; eksekusi per fase §78).
- Migrasi data reservasi / perubahan schema task.
- Rulix, plugin marketplace, merge queue, durable queue (kandidat v0.2.0 / v1.1.0).
- Model B multi-project (D-05) — §77.
- Managed Agents / Claude Agent SDK sebagai engine utama.

## 72. Keputusan Versi — v0.1.1 ditolak, MAJOR 1.0.0

### 72.1 Aturan semver §69

```text
MAJOR.MINOR.PATCH
PATCH  = wording, typo, atau configuration fix tanpa perubahan responsibility.
MINOR  = role, skill, hook, atau optional tool baru yang backward-compatible.
MAJOR  = perubahan control model, orchestrator, responsibility boundary,
         atau execution runtime.
```

Contoh yang dicantumkan §69: `0.1.1  Hook validation fixes` dan
`0.2.0  Rulix dan internal plugin distribution`.

### 72.2 Verifikasi label v0.1.1

Lapis 2 orkestrator memenuhi definisi **MAJOR** §69 dua kali:

1. **Orchestrator** — struktur baru: Lapis 2 orkestrator di atas runner.
2. **Execution runtime** — runtime eksekusi berubah dari "sesi Claude Code
   native" menjadi "loop Messages API yang dijalankan binary m2s".

`v0.1.1` adalah kelas "hook validation fixes" — contoh §69 menaruhnya di sana.
Menamai perubahan orchestrator + runtime dengan PATCH menipu semver.

### 72.3 Mengapa bukan 0.2.0

Trigger §65 (Rulix, plugin marketplace, database, durable queue, multi-project
bottleneck, container isolation, automated observability) **tidak terpicu**:
- Durable queue tidak dibutuhkan — eksekusi tetap synchronous, satu task satu
  sesi.
- Database tidak dibutuhkan — registry file-based + state YAML cukup.

Kenaikan ke `0.2.0` adalah kelas Rulix/plugin distribution, bukan perubahan
orchestrator. Kelas perubahan ini adalah MAJOR → `1.0.0`.

### 72.4 Mengapa 1.0.0 tidak menunggu "production-proven"

Contoh §69 `1.0.0 Production-proven` adalah deskripsi state, bukan gate semver.
Gate semantiknya: MAJOR bila orchestrator/execution runtime berubah — sudah
berubah. Production-proofing adalah pekerjaan fase hardening (F6), bukan alasan
menahan angka versi.

### 72.5 Efek tertulis

- `README.md` badge judul: `v0.1.0` → `1.0.0`.
- Dokumen ini: `M2S-VSH-Lite-v1.0.0-Architecture.md` (menggantikan peran
  v0.1.0 sebagai referensi execution).
- `docs/decisions/open-questions.md`: entri "Keputusan Versi 1.0.0".
- ADR-011 mencatat simpangan label v0.1.1.

## 73. Evaluasi Harness — A/B/C

### 73.1 Kriteria

| Kriteria | A: Managed Agents | B: Claude Agent SDK | C: Manual Messages API |
|---|---|---|---|
| Independen dari Claude Code | Ya (API baru) | **Tidak** — SDK adalah Claude Code sebagai library | **Ya** — loop dibangun sendiri |
| `ANTHROPIC_BASE_URL` proxy (9Router) | ❌ **Blocking** — Anthropic-hosted; proxy tak berlaku | ✅ (Claude Code honor proxy) | ✅ (base_url override) |
| Worktree git lokal runner | ❌ sandbox container remote; repo via resource `github_repository` | ✅ cwd = worktree lokal | ✅ cwd = worktree lokal |
| A-01 enforcement | ❌ lemah — toolset server-executed, tak ada interposisi write/edit | ✅ hooks native | ✅ **terkuat** — gate tiap tool_use termasuk bash |
| Hook stack existing (5 hooks) | ❌ tak berjalan; port ulang | ✅ berjalan | ⚠️ port logika hook ke guard Go |
| Kontrak snapshot `.task/contract.json` 0444 + info/exclude | ❌ tak berlaku di container | ✅ | ✅ |
| Bahasa integrasi | SDK Go ada (beta) | Node/TS → bridge proses + runtime Node baru | Go-native, stdlib + SDK |
| Approval/permission | `always_ask` + tool_confirmation | permission system | guard + `--await-approval` |
| Subagents/threads | multiagent thread (25 max) | native | sesi per-role; DAG dijalankan orkestrator |
| Beta/durability | beta header; rate limit 60 env-op; data residency | stable | stable (Messages API) |
| Biaya migrasi | tinggi — rewrite sandbox + enforcement | rendah | sedang |

### 73.2 Keputusan: C

**C dipilih.** Alasan:

1. **Tujuan eksplisit "tanpa bergantung pada Claude Code".** A dan B gagal
   syarat ini dengan cara berbeda: A mengganti Claude Code dengan platform
   Anthropic-hosted yang tetap di luar kendali kita; B adalah Claude Code
   itu sendiri sebagai library.
2. **Proxy 9Router.** Seluruh pilot berjalan lewat routing model lokal
   (`ANTHROPIC_BASE_URL`). A menghilangkan jalur ini (sandbox Anthropic-hosted).
   B dan C menghormatinya.
3. **Enforcement A-01.** A memindahkan sandbox ke container Anthropic dan
   menutup jalur interposisi write/edit — melemahkan invariant A-01. C
   memperkuat: bash sendiri di-guard (R-08 hilang, lihat §76).
4. **Kesesuaian arsitektur.** C mempertahankan model "runner mengendalikan
   worktree git lokal, agent bekerja di dalamnya" yang sudah dibuktikan pilot.

Harga C: implementasi loop tool-use — itulah deliverable Lapis 2.

### 73.3 Peran B (opsional, jembatan migrasi)

Selama migrasi bertahap, role yang belum diport tetap berjalan sebagai sesi
Claude Code (`engine: claude-code` di `control/orchestrator.yaml`). Kontrak
tidak berubah; engine adalah keputusan config (V-5), bukan schema. B
menghilang setelah seluruh role diport.

### 73.4 Peran A (ditolak sebagai engine; catatan masa depan)

Managed Agents layak dievaluasi ulang bila: (1) klien menuntut sandbox
managed penuh, atau (2) proxy 9Router berhenti digunakan. Tidak ada kerja
sekarang.

## 74. Arsitektur Lapis 2 — Orkestrator

### 74.1 Diagram

```text
Lapis 1 (tetap): deterministic runner
  validate-task, reserve-paths, launch-task, collect-result,
  release-reservation, check-path, validate-changed-paths
  + internal/contract, internal/pathmatch, internal/registry
  + scripts/<sub>.sh wrapper tipis (ADR-004 #2)
  TIDAK BERUBAH. Exit code 0/1/2 dipertahankan.

Lapis 2 (baru): orkestrator — m2s run --task <id>
  ├── resolve spec        control/tasks/specifications/<id>.yaml
  ├── validate-task       (komposisi subcommand yang ada)
  ├── reserve-paths       (komposisi)
  ├── launch-task         (komposisi; worktree + .task/contract.json)
  ├── engine select       control/orchestrator.yaml → engine: messages
  ├── agent loop          internal/orchestrator
  │     ├── prompt assembly  contract.json + role system prompt + governance
  │     ├── transport        internal/harness (anthropic-sdk-go,
  │     │                     base_url dari ANTHROPIC_BASE_URL,
  │     │                     model dari frontmatter role template)
  │     ├── tools            bash, read, write, edit, glob, grep,
  │     │                     m2s_check_path, m2s_submit_handoff,
  │     │                     m2s_submit_failure
  │     ├── guard            tiap tool_use → guard sebelum eksekusi (§76)
  │     ├── audit            .task/audit.jsonl (ditulis orkestrator)
  │     └── loop bounds       execution.max_turns, execution.timeout_minutes
  ├── quality_gates       jalankan tiap perintah di worktree (Q4, per-task)
  ├── git/PR              commit → push branch → gh pr create --base develop
  ├── handoff             tool m2s_submit_handoff → validasi
  │                        contract.KindHandoff → tulis control/handoffs/<id>.yaml
  └── collect-result -pr  → reservasi → reserved-pending-merge (Q12)
```

Prinsip §13.4 dipertahankan: **orkestrator bukan agent**. Orkestrator tidak
membuat keputusan teknis; ia mengeksekusi kontrak dan memverifikasi boundary.
Seluruh keputusan teknis = LLM dalam loop, dibatasi guard.

### 74.2 Modul Go baru

```text
internal/orchestrator/
  orchestrator.go   state machine task (draft→…→reserved-pending-merge)
  loop.go           agent loop: request→tool_use→guard→exec→tool_result→…
  guard.go          boundary A-01 (reuse internal/pathmatch + isInside)
  tools.go          definisi tool (nama, desc, JSON schema)
  prompt.go         assembly system prompt dari contract + role template
  handoff.go        validasi handoff/failure
internal/harness/
  client.go         anthropic-sdk-go wrapper; base_url/model dari env + config
  sse.go            (SDK handle; fallback raw HTTP SSE bila proxy tak cocok)
cmd/m2s/run.go      subcommand m2s run (register di main.go)
scripts/run.sh      wrapper tipis (pola Q11 / ADR-004 #2)
control/orchestrator.yaml   config engine, model routing, timeout global
```

Dependency baru: hanya `github.com/anthropics/anthropic-sdk-go`. Justifikasi:
transport inti — SDK menangani streaming, retry, typed error, base_url
override. Proxy 9Router melayani format Anthropic di `ANTHROPIC_BASE_URL`
(dibuktikan Claude Code berjalan di atasnya). Risiko kompatibilitas proxy
tool-use → ADR-010 + verifikasi V-11 (F2).

`internal/pathmatch` tetap satu-satunya otoritas overlap. `isInside` /
`resolveEventual` (saat ini di `cmd/m2s/main.go`) di-export ke
`internal/pathmatch` agar `check-path` dan guard orkestrator berbagi satu
implementasi — bukan duplikasi. (Ponytail: satu implementasi, dua pemanggil.)

### 74.3 State reservasi dan handoff

Tidak berubah untuk reservasi: tetap repo control, CLI update setelah session.
Baru: `control/handoffs/<id>.yaml` sebagai hasil handoff orkestrator, divalidasi
`handoff.schema.json`.

## 75. Matriks Pemetaan Fitur — Claude Code → Pengganti

Lampiran A berisi matriks lengkap per fitur. Inti:

| Fitur Claude Code | Penggunaan sekarang | Pengganti (engine C) |
|---|---|---|
| Subagents `.claude/agents/*.md` | 13 role; isolasi worktree | Orkestrator spawn sesi Messages per role; system prompt = role template + contract + governance; model dari frontmatter role |
| Permission system (settings.json) | deny secret/.claude/.task/cmd-m2s/Makefile | `guard.go`: deny list di kode + `m2s check-path` tiap path tool |
| Hooks PreToolUse/PostToolUse | path scope, danger cmd, secret, audit | Interceptor orkestrator sebelum/sesudah tiap tool_use |
| Hooks SubagentStop | handoff validation | Tool `m2s_submit_handoff`; input schema = handoff.schema.json |
| Worktree isolation | §29.3, `$HOME/.m2s/worktrees` | Tidak berubah — runner `launch-task` yang buat; orkestrator cwd = worktree |
| CLAUDE.md + `.claude/rules/*` | governance + soft rules | Assembly system prompt: architecture constraints, `rules/`, DESIGN.md bila role frontend |
| `--add-dir` read context (A-11) | read-only | Tool `read` di-scope worktree + `inputs` contract di-inject sebagai read context |
| CLI `claude -p` / spawn | sesi Claude Code | Loop streaming Messages API |
| Skills/plugins | Ponytail, dll | Tidak dibutuhkan — instruksi skill-like di-inject dari `templates/` per role |
| Permission prompts / human approval | A-03 reviewer plan mode | `--await-approval`: guard menahan aksi berisiko; reviewer tetap read-only |
| Mneme | warn mode, CI gate | **Tidak berubah** — tool eksternal di CI/hook, bukan bagian loop |
| Secret scan (GitGuardian) | CI | Tidak berubah — CI |
| Hooks worktree bawaan | R-26 | Tidak berubah untuk sesi warisan; untuk engine messages, secret-scan worktree milik runner |

## 76. Invariant A-01 — Tool-Gate Orkestrator dan Uji Konkrit

### 76.1 Boundary baru

Setiap `tool_use` dari LLM melewati `guard.go` **sebelum eksekusi**. Urutan
guard identik dengan `check-path` hari ini:

1. Tool name dalam allowlist role (writer: bash/read/write/edit/glob/grep/m2s_*;
   reviewer: read/glob/grep/m2s_submit_handoff tanpa write).
2. Path tool → resolve `file_path` terhadap worktree (`isInside` +
   `resolveEventual`, symlink-aware) → `pathmatch.IsForbidden` dulu →
   `pathmatch.IsAllowed`. Forbidden menang (matriks §4.8).
3. Bash → cwd dipaksa worktree; port blocklist `block-dangerous-command.sh`
   ke Go (rm -rf, sudo, chmod -R, git reset --hard, git switch, git -C, dll);
   env di-scrub (tidak ada secret).
4. Gagal guard → `tool_result` `is_error: true` + audit `.task/audit.jsonl` +
   hitung pelanggaran. Pelanggaran pertama tidak menghentikan loop (LLM boleh
   koreksi); pelanggaran berulang (default N=3) → status `failed`, handoff
   `failure` kategori `path-violation`.
5. Net kedua tetap ada: CI `validate-changed-paths` + branch protection. Guard
   lokal bukan satu-satunya enforcement (konsisten §43).

**Invariant A-01 dipertahankan:** forbidden path ditolak pada semua harness.
Untuk sesi Claude Code warisan, hooks tetap aktif; untuk sesi messages, guard
menggantikan posisi hook sebagai primary. ADR-012 mencatat pemindahan boundary.

### 76.2 Uji konkrit (bukan narasi)

- **Unit:** `internal/pathmatch` (24 kasus R-03, sudah ada) — tidak berubah.
- **Unit:** `cmd/m2s/run_test.go` — **fake model**: `httptest.Server` melayani
  `POST /v1/messages` dengan transcript skrip. Orkestrator diarahkan via
  base-url override.
  - Positif: LLM emit `tool_use write` `file_path: internal/payroll/x.go`
    (allowed) → guard lolos → file benar ada di disk.
  - Negatif 1: `file_path: internal/auth/token.go` (di luar allowed) → deny,
    file tidak pernah dibuat.
  - Negatif 2: `file_path: go.mod` (forbidden) → deny.
  - Negatif 3: `file_path: ../../etc/passwd` (keluar worktree) → deny.
  - Bash: `command: git switch main` → deny; `command: go test ./...` →
    izinkan, cwd worktree.
- **Integrasi:** `tests/orchestrator/path-guard.test.sh` + target Makefile
  `verify-orchestrator` (dibangun F3).
- **Regression:** `make verify` tetap hijau — guard tidak mengubah
  `check-path`, hooks, atau CI.

## 77. D-05 Multi-Project — status

**Di luar scope v1.0.0.** Model A (satu control repo per project) dipertahankan
apa adanya. Orkestrator TIDAK menambah kunci `project` ke reservation
(menghindari migrasi berkas yang D-05 tandai mahal). Empat titik kerja model B
di D-05 tetap berdiri untuk v1.1.0. Tidak ada perubahan schema.

Ini keputusan sadar, bukan lupa — pemicu §65 (multi-project bottleneck) belum
terpenuhi.

## 78. Rencana Fase — 1 PR per fase

| Fase | Isi | Kriteria done terukur | Verifikasi |
|---|---|---|---|
| **F0: versi + dokumen** | VERSION→1.0.0 (README); `M2S-VSH-Lite-v1.0.0-Architecture.md`; ADR-010/011/012; open-questions (Keputusan Versi + D-05) | `grep v0.1.0 README.md` kosong; dokumen + ADR ada; status open-questions akurat | `make verify` hijau |
| **F1: `m2s run` skeleton** | `internal/orchestrator` + `cmd/m2s/run.go`; komposisi validate→reserve→launch; `--dry-run` tanpa model; exit code 0/1/2; `scripts/run.sh`; `control/orchestrator.yaml` | `m2s run --task BE-101 --dry-run` menghasilkan reservasi + worktree + `.task/contract.json`, idempotent | `go test ./cmd/m2s/ -run TestRunDryRun`; `make verify` |
| **F2: transport + loop** | `internal/harness` (SDK, base_url env, model dari role template); loop tool_use; tool `bash` + `m2s_submit_handoff` (schema handoff); audit log; bounds max_turns/timeout | Fake-model test: LLM emit bash + handoff → loop selesai, handoff tervalidasi | `go test ./cmd/m2s/ -run TestRunFakeModel` |
| **F3: A-01 guard** | `internal/orchestrator/guard.go`; isInside/pathmatch di-share ke `internal/pathmatch`; port danger-cmd + secret-path; negatif test | Tabel uji §76.2 lulus semua; write forbidden tak pernah ke disk | `tests/orchestrator/path-guard.test.sh`; `make verify-orchestrator` |
| **F4: prompt + gates** | Prompt assembly (role + contract + governance); quality_gates di worktree; handoff ke `control/handoffs/`; failure tool; state transitions | Happy path fake-model penuh: write allowed → gates lulus → handoff valid → reservasi ke reserved-pending-merge | `go test ./cmd/m2s/ -run TestRunEndToEnd` |
| **F5: git/PR + review** | commit atomic → push → `gh pr create --base develop` (H-01 base guard tetap CI); engine selector `claude-code` sebagai jembatan; approval gate `--await-approval` | `m2s run` end-to-end di scratch repo: PR muncul, CI `validate-changed-paths` hijau, worker tak bisa merge | `make verify`; check CI scratch repo |
| **F6: hardening** | retry/backoff, resume dari audit log, token/cost report, secret guard hardening, chaos test, docs operator | Chaos: kill loop mid-run → resume → state konsisten; docs/operator/run.md | test chaos + `make verify` + `verify-orchestrator` |

Dependensi: F1→F2→F3 berurutan (loop butuh transport, guard butuh loop). F4
paralel sebagian dengan F3 (prompt tak bergantung guard). F5 butuh F4. F6 semua.

## 79. Risiko

| ID | Risiko | Mitigasi |
|---|---|---|
| V-11 (baru) | Kompatibilitas proxy 9Router dengan tool-use loop Messages API | Verifikasi di F2 dengan fake + real proxy; fallback raw HTTP SSE di `internal/harness` bila SDK tak cocok (dokumen ADR-010) |
| — | Model via proxy (mis. `gratisan`/`deepseek-v4-flash-free`) tool-use berkualitas rendah | Contract `model:` per role tetap; orkestrator tidak mengubah id; verifikasi ke proxy, jangan asumsi (aturan execution-parallel §5) |
| — | Role diport ke messages → batas A-01 tidak lagi ganda (hook+guard) | Guard jadi primary; CI tetap net kedua. ADR-012 mencatat. |
| — | Migrasi bertahap (engine campuran) selama F5 | Engine selector `claude-code` + `messages`; kontrak tidak berubah |

## 80. Acceptance Criteria v1.0.0

1. `m2s run --task <id>` menyelesaikan satu task end-to-end di scratch repo
   tanpa sesi Claude Code: reservasi → worktree → agent loop → quality gates →
   handoff → PR → `reserved-pending-merge`.
2. Invariant A-01 diverifikasi lewat tabel uji §76.2 — termasuk negatif path,
   bukan narasi.
3. `make verify` + `make verify-orchestrator` hijau.
4. Seluruh 13 role punya `engine: messages` di config, atau yang masih
   `claude-code` tercatat sebagai jembatan migrasi aktif.
5. Tidak ada perubahan `task.schema.json`; engine adalah config, bukan kontrak.
6. D-05 model A dipertahankan; model B di luar scope tercatat di open-questions.

## 81. Lampiran A — Matriks Pemetaan Lengkap

| # | Fitur Claude Code | Penggunaan sekarang | Pengganti (engine C) | Catatan |
|---|---|---|---|---|
| 1 | Subagents `.claude/agents/*.md` | 13 role, isolasi worktree | Orkestrator spawn sesi Messages per role | System prompt = role template + contract + governance |
| 2 | Permission system (settings.json) | deny secret/.claude/.task/cmd-m2s/Makefile | `guard.go` deny list + `m2s check-path` | Setara deny list, di Go |
| 3 | PreToolUse/PostToolUse hooks | path scope, danger cmd, secret, audit | Interceptor orkestrator | Logika diport dari `.claude/hooks/*.sh` |
| 4 | SubagentStop hook | handoff validation | Tool `m2s_submit_handoff` | Schema = handoff.schema.json |
| 5 | WorktreeLifecycle hook | worktree init/cleanup | Runner `launch-task`/`release-reservation` | Tidak berubah |
| 6 | Worktree isolation `isolation: worktree` | §29.3 | Runner yang buat; orkestrator cwd = worktree | Tidak berubah |
| 7 | CLAUDE.md + `.claude/rules/*` | governance + soft rules | Assembly system prompt | Termasuk architecture constraints (§H-04) |
| 8 | `--add-dir` read context (A-11) | read-only | Tool `read` di-scope worktree + `inputs` contract | Tetap read-only |
| 9 | CLI `claude -p` / spawn | sesi Claude Code | Loop streaming Messages API | transport `internal/harness` |
| 10 | Skills/plugins | Ponytail, dll | Inject dari `templates/` per role | Tidak dibutuhkan skill runtime |
| 11 | Permission prompts | A-03 reviewer plan mode | `--await-approval` | Reviewer tetap read-only |
| 12 | Mneme | warn mode, CI gate | Tidak berubah | Tool eksternal di CI/hook |
| 13 | Secret scan (GitGuardian) | CI | Tidak berubah | CI |
| 14 | `.task/contract.json` snapshot | 0444 + info/exclude | Dipertahankan | Runner tulis, agent tak pernah |
| 15 | `.claude/agents/*` model frontmatter | routing model per role | Dibaca orkestrator | dari role template |
| 16 | Bash tool | shell di worktree | `guard.go` bash blocklist + cwd paksa | Menutup R-08 |

Tidak ada sel kosong — setiap fitur yang digunakan sekarang punya pengganti.
Fitur yang "tidak berubah" berarti dipegang runner/CI, bukan sesi Claude Code.

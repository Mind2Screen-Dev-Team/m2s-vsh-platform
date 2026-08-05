# Setup Klien — Implementasi M2S-VSH di GitHub Org Baru

**Versi:** v0.1.0 · **Bahasa:** Bahasa Indonesia (konvensi repo)
**Audience:** manusia (operator/owner GitHub) yang mengimplementasikan project
ini di org klien — bukan Mind2Screen-Dev-Team.

Dokumen ini memandu setup lengkap dari **control repo sampai repo aplikasi**
(1 atau lebih, stack bebas) di GitHub org yang berbeda. Setelah mengikuti
dokumen ini, klien punya control repo + repo aplikasi berjalan dengan pipeline
agent, branch protection, dan required check `validate-changed-paths` yang aktif.

---

## Bagian 1 — Ikhtisar

### Struktur project

| Repo | Isi | Peran |
|---|---|---|
| `m2s-vsh-platform` | repositori pengatur — arsitektur, runner `m2s`, schema task, templates, governance, docs | sumber aturan & konfigurasi |
| **Repo aplikasi (1+)** | tiap repo = satu jenis (BE, FE, mobile, fullstack, dst) | repo tempat kode produk dibuat |

**Repo aplikasi bisa apa saja — tidak harus backend+frontend.** Arsitektur
mendukung backend, frontend, mobile, fullstack (backend+frontend dalam satu
repo), android/ios, atau campuran. Pilih peran pengerja sesuai jenis repo
(`backend-engineer`, `frontend-engineer`, `fullstack-engineer`,
`mobile-engineer`, `android-developer`, `ios-developer`). Minimal 1 repo
aplikasi; jumlah dan kombinasinya bebas (ADR-005, §46 Multi-Repository Feature).

**Project pilot saat ini (real):** 2 repo aplikasi — `m2s-vsh-project-backend`
(Go) + `m2s-vsh-project-frontend` (Next.js). Ini **keputusan pilot**, bukan
batasan arsitektur. Klien bebas memilih formasi sendiri.

### Prasyarat

- GitHub **org** (plan **Team** disarankan — mendukung ruleset; Free plan pakai
  fallback classic branch protection, lihat Bagian 2 §5)
- Akun admin org (`admin:org` + `repo` scope)
- `gh` CLI (authenticated), `git`, `make`, Go ≥ 1.26
- `age` (enkripsi key GitHub App)
- Akses model: Anthropic API langsung ATAU proxy 9Router (setup klien sendiri —
  **jangan** menyalin `ANTHROPIC_BASE_URL` dari org asal, lihat Bagian 2 langkah 10)

### Konvensi yang berlaku

- Branch: `develop` / `staging` / `main`. Agent `agent/*` → `develop` saja
  (H-01); `main` merge milik manusia (ADR-001 #2).
- Human-only paths (off-limits utk agent): `cmd/m2s/**`, `Makefile`,
  `.claude/**`, `.github/**`, `.mneme/**`, `governance/**`, `.task/**`, secrets.
  Daftar lengkap: `docs/operator/phase-8-human-only-patches.md`.
- Template kanonik di `templates/` → salin ke lokasi aktif (human-run), bukan
  sunting salinan aktif.

---

## Bagian 2 — Langkah setup berurutan

> **Pembagian peran:** langkah 1–3 dan 10 (clone, build, template sync,
> konfigurasi model) dapat dijalankan klien — referensi yang disebut di sana
> client-safe. Langkah 4–6 dan 9 (GitHub App, branch protection, workflow,
> verifikasi penegakan) menuntut **dokumen internal** dan identitas penegakan —
> dijalankan **tim Mind2Screen**, bukan diserahkan sbg dokumen ke klien.
> Langkah 7–8 (integrasi) menjelaskan cara repo saling terhubung — dibaca utk
> pemahaman, bukan langkah eksekusi mandiri.

### 1. Clone + rename ke org klien

```bash
# ganti <KLIEN-ORG>, <REPO-APLIKASI-1..N> sesuai formasi klien
git clone https://github.com/Mind2Screen-Dev-Team/m2s-vsh-platform.git
# pilot: clone 2 repo aplikasi (backend + frontend)
git clone https://github.com/Mind2Screen-Dev-Team/m2s-vsh-project-backend.git
git clone https://github.com/Mind2Screen-Dev-Team/m2s-vsh-project-frontend.git
# formasi lain: clone/seed repo aplikasi sesuai stack (mobile, fullstack, dst)
#   agent role sesuai stack — lihat Bagian 1 Struktur project
```

Buat repo baru di org klien, lalu ganti semua referensi org/repo lama ke org
klien. **Repo aplikasi sebanyak formasi klien (1+); untuk tiap repo aplikasi
terapkan langkah rename yang sama di bawah.**

**Control repo (`m2s-vsh-platform`):**
- `go.mod` — `module github.com/Mind2Screen-Dev-Team/m2s-vsh-platform` → `github.com/<KLIEN-ORG>/m2s-vsh-platform`
- `cmd/m2s/commands.go`, `cmd/m2s/pathcheck.go` — import path
- `internal/registry/registry.go`, `registry_test.go` — import path
- `templates/github/workflows/path-enforcement.yml` — `repository:` line →
  `<KLIEN-ORG>/m2s-vsh-platform`
- `tools/apply-rulesets.sh`, `tools/apply-branch-protection.sh` — variabel
  `REPOS=( Mind2Screen-Dev-Team/... )` → repo klien
- `templates/github/CODEOWNERS`, `.github/CODEOWNERS` — `@fajarcandraaa` → `@<KLIEN-OWNER>`
- `README.md` + `docs/operator/*` — referensi org lama di perintah `gh`

**Tiap repo aplikasi** (contoh di bawah pakai stack pilot BE Go + FE Next.js;
sesuaikan dgn stack repo klien):
- `go.mod` (stack Go) — module path → org klien
- `cmd/server/main.go` (stack Go) — import `github.com/Mind2Screen-Dev-Team/m2s-vsh-project-backend/...` → org klien
- `.github/workflows/path-enforcement.yml` — `repository:` line (checkout control repo) → `<KLIEN-ORG>/m2s-vsh-platform`
- `.github/CODEOWNERS` — `@fajarcandraaa` → `@<KLIEN-OWNER>`
- `README.md` — tautan control repo (stale `fajarcandraaa`)

> **Catatan:** `tests/negative/github-workflow.test.sh` memuat `@fajarcandraaa`
> sbg fixture negatif sengaja — jangan diubah.

### 2. Bangun + verify control repo

```bash
cd <control-repo>
make build      # bangun bin/m2s
make verify     # format, vet, test, schema, agents, hooks, test negatif
```

`make verify` **harus hijau penuh** sebelum lanjut. `TestDeployedAgentsMatchTemplates`
menuntut `.claude/agents/` identik dgn `templates/agents/` — kalau kamu ubah
template, sync dulu (langkah 3).

### 3. Human-only template sync

Template kanonik → salin ke lokasi aktif (manual, bukan agent):

```bash
# di control repo
for r in frontend-engineer project-manager technical-lead-system-analyst technical-writer; do
  cp "templates/agents/$r.md" ".claude/agents/$r.md"
done
cp templates/rules/*.md .claude/rules/          # ref rules-deployment.md
cp templates/governance/capability-registry.yaml governance/   # ref capability-registry-deploy.md
cp templates/github/CODEOWNERS .github/CODEOWNERS
cp templates/github/PULL_REQUEST_TEMPLATE.md .github/PULL_REQUEST_TEMPLATE.md
cp templates/github/workflows/path-enforcement.yml .github/workflows/path-enforcement.yml
# ulangi CODEOWNERS + PR template + workflow ke .github/ backend & frontend
```

Docs rujukan: `docs/operator/rules-deployment.md`, `capability-registry-deploy.md`.

### 4. Buat dua GitHub App (bagian tersulit)

Dua identitas: **`m2s-worker`** (buat PR) dan **`m2s-approver`** (review +
merge). Prosedur lengkap + teruji: `docs/operator/org-migration.md` §Langkah
ADR-001 #5. Ringkasan:

1. Buat App di `https://github.com/organizations/<KLIEN-ORG>/settings/apps/new`
   untuk `m2s-worker` dan `m2s-approver`:
   - Permission: **Contents: Read & Write**, **Pull requests: Read & Write**
   - Restrict ke control repo + semua repo aplikasi klien
2. Download private key tiap App. Enkripsi:
   ```bash
   age -p -o ~/.claude/secrets/m2s-worker.pem.age <worker-key>.pem
   age -p -o ~/.claude/secrets/m2s-approver.pem.age <approver-key>.pem
   rm <worker-key>.pem <approver-key>.pem   # plaintext dihapus
   ```
   Passphrase (beda per App) di password manager. Kunci `m2s-approver` aset
   bernilai tinggi — jangan commit.
3. Catat `app_id` tiap App:
   ```bash
   gh api apps/m2s-worker --jq .id
   gh api apps/m2s-approver --jq .id
   ```
4. **Env var yang dipakai `scripts/gh-app-token.sh`** (wajib diset utk runner):

   | Env | Nilai |
   |---|---|
   | `M2S_APP_ID` | app_id `m2s-worker` |
   | `M2S_APP_KEY_PATH` | path `~/.claude/secrets/m2s-worker.pem.age` |
   | `M2S_APP_KEY_PASS` | passphrase worker |
   | `M2S_APP_INSTALL_ID` | installation id worker |
   | `M2S_APPROVER_ID` | app_id `m2s-approver` (utk `tools/apply-rulesets.sh`) |
   | `GH_TOKEN` | token admin (utk install ruleset/branch protection) |

   `scripts/gh-app-token.sh` mendukung `.pem` dan `.pem.age`. Nilai-nilai ini
   TIDAK ada di `.env.example` — dokumentasikan sendiri di config klien.

> **Penting (V-08):** `actor_type: Integration` (GitHub App) sbg bypass actor
> HANYA valid di repo **organization**, bukan akun personal. Klien WAJIB pakai
> org. Lihat `docs/decisions/open-questions.md` V-08.

### 5. Branch protection + rulesets

**Plan Team (org)** — ruleset + required check:

```bash
# install ruleset agent-push-restriction + agent-worker-restriction
M2S_APPROVER_ID=<app_id-approver> tools/apply-rulesets.sh
```

**Free plan fallback** — classic branch protection di `main` (guard `validate-changed-paths`,
no force-push, no deletions):

```bash
GH_TOKEN=<admin-token> tools/apply-branch-protection.sh
```

Rasional Free plan: `docs/operator/phase-8-branch-protection.md`. Panduan
lengkap 4-langkah (prove CI → required check → protection → troubleshoot):
`docs/operator/branch-protection.md`.

### 6. Workflow CI

Pastikan di **control repo + tiap repo aplikasi**:
`.github/workflows/path-enforcement.yml` ada dgn `repository:` menunjuk control
repo klien, trigger `branches: [develop, staging]`, job `validate-changed-paths`
(nama ini = nama required check). Kontrol repo sendiri punya salinan dgn trigger
`[main]`.

Keselarasan template workflow diperiksa lewat `make verify` (test negatif
`tests/negative/github-workflow.test.sh`, penegak `tests/lib/check-github-artifacts.sh`).

### 7. Integrasi Control repo ↔ Repo aplikasi

Bagian ini menjelaskan **bagaimana repositori pengatur (control) terhubung ke
tiap repositori aplikasi**. Integrasi lewat **5 titik sambung** berikut:

| # | Titik sambung | File / lokasi | Isi yang disesuaikan klien |
|---|---|---|---|
| 1 | **Checkout control di CI app** | `.github/workflows/path-enforcement.yml` di tiap repo aplikasi | baris `repository:` → `<KLIEN-ORG>/m2s-vsh-platform` |
| 2 | **CODEOWNERS** | `.github/CODEOWNERS` di control + tiap app | `@fajarcandraaa` → `@<KLIEN-OWNER>` |
| 3 | **Daftar repo (rulesets)** | `tools/apply-rulesets.sh`, `tools/apply-branch-protection.sh` | variabel `REPOS=( ... )` → repo klien |
| 4 | **Cakupan GitHub App** | pengaturan install App di UI GitHub | restrict ke control + semua repo aplikasi |
| 5 | **Daftar repo runner** | `README.md` (tabel repo pilot) + field `ownership.repository` di tiap task contract | nama repo → nama repo klien |

**Alur CI (control ↔ app):** saat PR agent dibuka di repo aplikasi, CI app:
1. checkout repo aplikasi (isi PR)
2. checkout **control repo** ke folder `.control` (lewat baris `repository:`)
3. bangun binary `m2s` dari `.control/go.mod`
4. baca **kontrak tugas** dari `.control/control/tasks/specifications/${TASK_ID}.yaml`
5. jalankan `validate-changed-paths` — cek path yang berubah sesuai kontrak

Jadi repo aplikasi **mengambil** kontrak + runner dari control repo setiap kali
CI jalan. Tanpa titik sambung #1, CI app tak bisa menemukan kontrak tugasnya.

**Daftar repo runner:** tidak ada "registrasi repo" terpusat. Runner tahu repo
aplikasi dari: (a) tabel repo pilot di `README.md`, (b) field
`ownership.repository` di tiap task contract (`schemas/common.schema.json`).
Klien cukup menamai repo konsisten di ketiga tempat.

> **Runtime config (belum otomatis):** saat ini CORS backend hardcoded `*`
> (`internal/handler/status.go`) dan URL API frontend hardcoded
> `http://localhost:8080` (`StatusDashboard.tsx`). Klien dengan host/port
> berbeda harus **edit source** kedua file — belum ada file konfigurasi.

### 8. Integrasi Antar Repo Aplikasi

Repo aplikasi (backend, frontend, mobile, dst) **independen — tidak terhubung
langsung satu sama lain**. Ini by design (arsitektur: "app repo bicara ke
control; app tak bicara langsung"). Ada dua bentuk keterhubungan:

**A. Ikatan lewat kontrak di control repo (desain):**
- TL/SA menulis **kontrak API** di control repo (mis. `contracts/CONTRACT-201.yaml`
  → shape `GET /api/v1/status`)
- **Backend** implementasi endpoint sesuai kontrak
- **Frontend** baca kontrak utk tahu shape respon, tulis kode konsumsi
- Konsistensi dijamin oleh kontrak bersama — bukan kode yang saling referensi

```
        control repo (kontrak)
       /                      \
  backend (implement)     frontend (konsumsi)
       \                      /
          runtime: HTTP + CORS
```

**B. Ikatan runtime (HTTP):**
- Frontend panggil backend lewat URL (`http://localhost:8080/api/v1/status`)
- Backend izinkan cross-origin via CORS `*`
- Klien dengan setup beda: sesuaikan URL di source frontend + pastikan CORS
  backend mengizinkan origin frontend

**Jika klien butuh repo aplikasi saling terhubung langsung** (mis. backend
memanggil service lain): arsitektur saat ini mendukung lewat **kontrak** di
control repo, bukan referensi kode lintas repo. Tambahkan kontrak baru di
`contracts/`, lalu tiap repo implement/konsumsi via kontrak itu.

### 9. Verifikasi end-to-end

```bash
# control
make verify    # hijau

# bukti agent pipeline: PR agent/* → develop, CI validate-changed-paths jalan
# bukti main protected: push agent/* → main DITOLAK (405 rule violations)
# bukti merge flow: worker PR → approver APPROVE → approver merge
```

Referensi status final + checklist: `docs/operator/status-adr001-five-complete.md`.

### 10. Konfigurasi model (Anthropic / 9Router)

`~/.claude/settings.json` org asal memakai `ANTHROPIC_BASE_URL:
http://localhost:20128/v1` (proxy 9Router local) — itu **khusus Mind2Screen**,
JANGAN disalin. Klien pakai:

- Anthropic API langsung (default, tanpa `ANTHROPIC_BASE_URL`), ATAU
- Proxy/9Router sendiri, set `ANTHROPIC_BASE_URL` + `ANTHROPIC_AUTH_TOKEN` +
  `ANTHROPIC_MODEL` sesuai setup klien.

Agent template memakai `model: gratisan` (nama combo 9Router). Kalau klien pakai
Anthropic langsung, ganti `model:` di `templates/agents/*.md` ke model Anthropic
yang valid (`sonnet` / `opus` / `haiku`), lalu sync ke `.claude/agents/`
(langkah 3). Detail model routing: `docs/operator/execution-parallel.md`.

---

## Bagian 3 — Klasifikasi dokumen: INTERNAL vs CLIENT-SAFE

Klasifikasi seluruh dokumen repo berdasarkan **keamanan informasi**, bukan
sekadar kegunaan setup. Sebagian besar dokumen operator/arsitektur memuat
**detail penegakan internal** (cara guard bekerja, lapis mana yang dapat
dilewati, app_id, mekanika bypass actor, temuan risk register). Dokumen semacam
itu **tidak dikirim ke klien** — klien yang membacanya belajar cara melewati
penegakan.

> **Prinsip:** klien menerima **prosedur yang perlu dijalankan**, bukan
> **mekanika penegakan** yang melindunginya.

### ⛔ INTERNAL — JANGAN dikirim ke klien

| Dokumen | Alasan internal |
|---|---|
| `docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md` | §XI enforcement (hooks, pola secret path), prinsip "prompt bukan security boundary", rule precedence, temuan R-07/R-08, rollback merge-authority. Klien belajar cara guard bekerja → cara melewatinya. **Ganti: ringkasan konsep client-safe** (lihat bawah) |
| `docs/architecture/phase-8-hardening.md` | Mendokumentasikan **celah** penegakan H-01..08 — apa yang dulu lemah, bagaimana PR agent bisa lompati staging/main |
| `docs/architecture/roles-extension-v0.1.0.md` | Desain role/agent internal + permission + larangan |
| `docs/operator/branch-protection.md` | app_id `4461216`/`4461262`, nama org, temuan V-08 bypass actor, data probe empiris, mekanika `bypass_mode` |
| `docs/operator/org-migration.md` | Runbook migrasi org: model dua-identitas, penanganan private key, bypass `actor_type: Integration`, nama org/repo |
| `docs/operator/hook-enforcement.md` | Tabel defense-in-depth memuat kolom "dapat dilewati?" — membuka R-07 (Bash arbitrer) dan R-08 (elakan string) + semantik fail-open |
| `docs/operator/phase-8-human-only-patches.md` | Mekanika guard (H-03/05/06/07 fail-closed), alasan H-08 dimatikan |
| `docs/operator/execution-parallel.md` | Proses eksekusi tim internal + celah model routing |
| `docs/operator/status-adr001-five-complete.md` | Tabel app_id, penanganan private key, "kunci approver aset bernilai tinggi", batas V-10 |
| `docs/operator/phase-8-branch-protection.md` | Runbook penegakan: scope `GH_TOKEN` admin, nama repo, bukti HTTP |
| `docs/adr/ADR-001-agent-merge-authority.md` | Rasional membuka prinsip prompt-bukan-boundary + aturan keras dua-identitas |
| `docs/adr/ADR-007-github-workflow-enforcement.md` | Proses keputusan internal, menimpa arsitektur §60 |
| `docs/adr/ADR-008-repo-ownership-migration.md` | Strategi plan klien (D-02: "repo klien private butuh plan Team") — membocorkan strategi harga/klien |
| `docs/decisions/open-questions.md` | Temuan V-08 (bypass actor), V-10 (batas merge approver), klaster D-02/D-03, kontradiksi fail-open |
| `docs/decisions/risk-register.md` | Seluruh daftar risiko R-XX + mitigasi + apa yang masih terbuka |
| `scripts/gh-app-token.sh` | Mencetak token instalasi App — mekanika identitas penegakan (JWT, `.pem.age`) |
| `tools/apply-rulesets.sh` | Install ruleset dgn bypass actor approver; membocorkan batas plan Free |
| `tools/apply-branch-protection.sh` | Install penegakan dgn scope `GH_TOKEN`, nama org/repo |
| `templates/github/rulesets/*.json` | Payload ruleset — bypass actor + mekanika penegakan |
| Semua `docs/operator/phase-*` lain (metrics, token-baseline, token-tracking, prompt, patches Phase 4/5/6) | Rekam fase internal + pengukuran biaya |
| `docs/operator/mneme-warn-patch.md`, `worktree-cleanup.md`, `task-status.md` | Governance + hygiene internal |
| `docs/phase-7-pilot-plan.md`, `docs/decisions/D-P7-*`, `phase-0-decision-log.md` | Rekam keputusan fase internal |
| `m2s-vsh-platform-phase-7-plan-specs/` (dir sibling) | Snapshot frozen — abaikan, jangan dikirim |

### ✅ CLIENT-SAFE — boleh dikirim ke klien

| Dokumen / artefak | Alasan aman |
|---|---|
| `docs/operator/client-setup.md` (dokumen ini) | Prosedur setup — ditulis untuk klien; tanpa mekanika bypass/risk |
| `docs/operator/rules-deployment.md` | Prosedur `cp` rules generik ke `.claude/rules/`. Tanpa nilai org, tanpa secret |
| `docs/operator/capability-registry-deploy.md` | Prosedur `mkdir`+`cp` capability registry. Bersih |
| `templates/rules/*.md` | Aturan agent generik (arsitektur, keamanan, testing, precedence) — konten kanonik, bukan mekanika penegakan |
| `templates/governance/capability-registry.yaml` | Template registry kapabilitas — data, bukan penegakan |
| `templates/agents/*.md` | Definisi role agent — deskripsi tugas/batas, dipakai klien untuk menjalankan pipeline |
| `templates/github/CODEOWNERS`, `PULL_REQUEST_TEMPLATE.md` | Artefak GitHub generik |
| `templates/github/workflows/path-enforcement.yml` | Workflow CI yang klien perlu pasang (validasi path). Isinya kode CI, bukan penjelasan cara melewatinya |
| Kode repo aplikasi (seed) | Contoh implementasi — endpoint, komponen UI, test |

### Ringkasan konsep client-safe (pengganti dokumen arsitektur)

Klien **tidak** menerima dokumen arsitektur kanonik. Sebagai gantinya, sampaikan
ringkasan konsep berikut (aman, tanpa mekanika penegakan):

- **Pipeline berbasis kontrak** — API/shared contract disetujui dulu, baru
  implementasi paralel per repo aplikasi.
- **Role terpisah** — PM (backlog/scope), TL/SA (contract + technical design),
  UI/UX (design handoff), engineering per stack (BE/FE/mobile/fullstack), QA
  (acceptance), technical writer (dokumentasi).
- **Task contract** — tiap pekerjaan agent punya spesifikasi: repo, branch,
  path yang boleh disentuh, acceptance criteria, quality gate.
- **Eksekusi terisolasi** — tiap task jalan di worktree sendiri; tidak ada dua
  task menulis path yang sama.
- **Branch flow** — `agent/*` → `develop` → `staging` → `main`; merge ke `main`
  milik manusia.
- **Gate manusia** — contract approve, design approve, dan merge akhir dipegang
  manusia, bukan agent.

Yang **tidak** disebutkan dalam ringkasan: cara hook bekerja, lapis mana yang
dapat dilewati, daftar risiko, app_id, mekanika ruleset/bypass.

### Ringkasan eksekutif

> **Kirim ke klien:** dokumen ini + `rules-deployment.md` +
> `capability-registry-deploy.md` + `templates/` (kecuali `rulesets/*.json`) +
> ringkasan konsep di atas.
> **JANGAN kirim:** dokumen arsitektur kanonik, seluruh `docs/adr/*`,
> `docs/decisions/*`, `docs/operator/*` lain (branch-protection, org-migration,
> hook-enforcement, status-adr001, phase-*), `scripts/gh-app-token.sh`,
> `tools/*.sh`, `templates/github/rulesets/*.json`.
>
> **Konsekuensi praktis:** langkah setup yang menuntut dokumen internal
> (GitHub App, ruleset, branch protection) **dijalankan oleh tim Mind2Screen**
> — bukan diserahkan ke klien sbg dokumen. Klien menerima hasil setup + dokumen
> ini + template.

---

## Bagian 4 — Catatan

- **Bahasa dokumen:** Bahasa Indonesia (konvensi repo — pertahankan di doc klien).
- **Versi baseline:** v0.1.0. Perubahan MAJOR (model kontrol, orchestrator,
  boundary) perlu ADR baru.
- **Model:** `templates/agents/*.md` memakai `model: gratisan` (combo 9Router
  org asal). Klien yang pakai Anthropic langsung harus ganti ke `sonnet`/`opus`
  dan sync `.claude/agents/`. Risiko: `gratisan` = `deepseek-v4-flash-free`,
  kualitas rendah utk role berat (pm/tl-sa/code-reviewer).
- **Merge queue** ditunda ke v0.2.0 (org Free, ADR-007 #3) — tak ada di setup.
- **Multi-project** (field `project` runner) model B ke v0.2.0 (D-05).

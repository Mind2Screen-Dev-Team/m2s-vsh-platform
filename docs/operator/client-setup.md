# Setup Klien — Implementasi M2S-VSH di GitHub Org Baru

**Versi:** v0.1.0 · **Bahasa:** Bahasa Indonesia (konvensi repo)
**Audience:** manusia (operator/owner GitHub) yang mengimplementasikan project
ini di org klien — bukan Mind2Screen-Dev-Team.

Dokumen ini memandu setup lengkap dari **control repo sampai repo pilot**
(backend + frontend) di GitHub org yang berbeda. Setelah mengikuti dokumen ini,
klien punya 3 repo berjalan dengan pipeline agent, branch protection, dan
required check `validate-changed-paths` yang aktif.

---

## Bagian 1 — Ikhtisar

### Struktur project

| Repo | Isi | Peran |
|---|---|---|
| `m2s-vsh-platform` | control — arsitektur, runner `m2s`, schema task, templates, governance, docs | otoritas konfigurasi |
| `m2s-vsh-project-backend` | Go server — `GET /api/v1/status` multi-component | repo aplikasi (BE) |
| `m2s-vsh-project-frontend` | Next.js — dashboard `/status` multi-status, polling 30s | repo aplikasi (FE) |

### Prasyarat

- GitHub **org** (plan **Team** disarankan — mendukung ruleset; Free plan pakai
  fallback classic branch protection, lihat Bagian 2 §5)
- Akun admin org (`admin:org` + `repo` scope)
- `gh` CLI (authenticated), `git`, `make`, Go ≥ 1.26
- `age` (enkripsi key GitHub App)
- Akses model: Anthropic API langsung ATAU proxy 9Router (setup klien sendiri —
  **jangan** menyalin `ANTHROPIC_BASE_URL` dari org asal, lihat Bagian 2 §8)

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

### 1. Clone + rename ke org klien

```bash
# ganti <KLIEN-ORG> dan nama repo sesuai
git clone https://github.com/Mind2Screen-Dev-Team/m2s-vsh-platform.git
git clone https://github.com/Mind2Screen-Dev-Team/m2s-vsh-project-backend.git
git clone https://github.com/Mind2Screen-Dev-Team/m2s-vsh-project-frontend.git
```

Buat repo baru di org klien, lalu ganti semua referensi org/repo lama ke org
klien:

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

**Repo aplikasi (backend + frontend):**
- `go.mod` (BE) — module path → org klien
- `cmd/server/main.go` (BE) — import `github.com/Mind2Screen-Dev-Team/m2s-vsh-project-backend/...` → org klien
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
   - Restrict ke 3 repo project
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

Pastikan di **3 repo**: `.github/workflows/path-enforcement.yml` ada dgn
`repository:` menunjuk control repo klien, trigger `branches: [develop, staging]`,
job `validate-changed-paths` (nama ini = nama required check). Kontrol repo
sendiri punya salinan dgn trigger `[main]`.

Keselarasan template workflow diperiksa lewat `make verify` (test negatif
`tests/negative/github-workflow.test.sh`, penegak `tests/lib/check-github-artifacts.sh`).

### 7. Verifikasi end-to-end

```bash
# control
make verify    # hijau

# bukti agent pipeline: PR agent/* → develop, CI validate-changed-paths jalan
# bukti main protected: push agent/* → main DITOLAK (405 rule violations)
# bukti merge flow: worker PR → approver APPROVE → approver merge
```

Referensi status final + checklist: `docs/operator/status-adr001-five-complete.md`.

### 8. Konfigurasi model (Anthropic / 9Router)

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

## Bagian 3 — Daftar dokumen: DIBUTUHKAN vs TIDAK utk setup klien

Klasifikasi seluruh `docs/operator/*.md`, `docs/architecture/*.md`, dan file
lain di repo — mana yang klien baca saat setup, mana yang bisa diabaikan.

### ✅ DIBUTUHKAN utk setup klien

| Dokumen | Alasan |
|---|---|
| `docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md` | **Kanonik §0–70** — rules, roles, ownership, workflow. Sumber kebenaran |
| `docs/operator/branch-protection.md` | **Setup branch protection** 4-langkah + ruleset + troubleshooting. Paling penting |
| `docs/operator/org-migration.md` | §Langkah ADR-001 #5 = prosedur buat 2 GitHub App + age-enkripsi + ruleset (aktif) |
| `docs/operator/rules-deployment.md` | Copy `templates/rules/*` → `.claude/rules/` |
| `docs/operator/capability-registry-deploy.md` | Copy `templates/governance/*` → `governance/` |
| `docs/operator/hook-enforcement.md` | 6 hooks + settings + `validate-changed-paths` — cara kerja + test |
| `docs/operator/phase-8-human-only-patches.md` | Daftar human-only paths + alasan |
| `docs/operator/execution-parallel.md` | Rule eksekusi paralel BE+FE + model routing (baru) |
| `docs/operator/status-adr001-five-complete.md` | Model dua-identitas worker/approver final + checklist verifikasi |
| `docs/operator/phase-8-branch-protection.md` | Free plan classic protection rationale |
| `docs/adr/ADR-001-agent-merge-authority.md` | Keputusan kewenangan merge (main manusia, agent → develop) |
| `docs/adr/ADR-007-github-workflow-enforcement.md` | Keputusan required check + workflow enforcement |
| `docs/adr/ADR-008-repo-ownership-migration.md` | Keputusan org ownership (mengapa org wajib) |
| `docs/decisions/open-questions.md` | V-08 (org mandatory utk Integration bypass), V-10 (merge flow) |
| `scripts/gh-app-token.sh` | Generator token GitHub App (env `M2S_APP_*`) |
| `tools/apply-rulesets.sh` | Install ruleset (org plan) |
| `tools/apply-branch-protection.sh` | Install classic protection (Free plan) |
| `templates/github/*` (CODEOWNERS, PR template, workflow, rulesets) | Kanonik → salin ke `.github/` 3 repo |
| `templates/rules/*`, `templates/governance/*`, `templates/agents/*` | Kanonik → salin ke `.claude/`/`governance/` |

### ⛔ TIDAK diperlukan utk setup klien

| Dokumen | Alasan tidak diperlukan |
|---|---|
| `docs/operator/phase-4-human-only-patches.md` | Patch Phase 4 sudah applied — historical |
| `docs/operator/phase-5-cleanup-and-phase-7-prep.md` | Patch Phase 5 sudah applied — historical |
| `docs/operator/phase-6-open-design.md` | Rekam design workspace Phase 6 — bukan setup |
| `docs/operator/phase-6-planning-prompt.md` | Session prompt Phase 6 — mati (dead) |
| `docs/operator/phase-8-implementation-prompt.md` | Session prompt Phase 8 — nyebut 3 repo org asal |
| `docs/operator/phase-8-metrics.md` | Pengukuran token §64 — rekam |
| `docs/operator/phase-8-token-baseline.md` | Baseline token per role §64 — rekam |
| `docs/operator/phase-8-token-tracking.md` | Proposal C2 tak diadopsi |
| `docs/operator/mneme-warn-patch.md` | Patch governance internal (Mneme warn mode) |
| `docs/operator/worktree-cleanup.md` | Hygiene cleanup — org-specific, bukan setup |
| `docs/operator/task-status.md` | Konvensi format status task (borderline — opsional) |
| `docs/architecture/roles-extension-v0.1.0.md` | Spesifikasi ADR-005 — historical |
| `docs/architecture/phase-8-hardening.md` | Blueprint H-01..08 — guard sudah shipped di workflow |
| `docs/phase-7-pilot-plan.md` | Rekam plan Phase 7 |
| `docs/decisions/*` (D-P7-*, phase-0-decision-log, dsb) | Keputusan fase — sudah diterapkan |
| `m2s-vsh-platform-phase-7-plan-specs/` (dir sibling) | Snapshot frozen Phase 7 — abaikan |

### Ringkasan eksekutif

> **Baca saat setup:** architecture kanonik + branch-protection + org-migration
> (§App) + rules-deployment + capability-registry-deploy + hook-enforcement.
> **Abaikan:** semua `phase-*-prompt`/`phase-*-metrics`/`phase-*-patches` yang
> sudah applied, snapshot dir, rekam keputusan fase.

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

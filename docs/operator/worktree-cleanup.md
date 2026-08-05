# Patch human-only — cleanup worktree stale dan branch merged

`git worktree`, `git checkout`, `git switch`, `rm -rf`, dan `git reset --hard`
ditolak `deny` di `.claude/settings.json` (§42.2, §16.2, R-15). Agent karena itu
tidak dapat menghapus worktree maupun direktorinya. Ini batas yang benar: agent
yang dapat menghapus worktree dapat menghapus pekerjaan task lain.

Cleanup di bawah **dijalankan manusia**. Diaudit 5 Agustus 2026 dari akar control
repository.

## Keadaan yang diaudit

19 worktree terdaftar pada `.claude/worktrees/` plus satu di luar
(`m2s-vsh-platform-phase-7-plan-specs`). Seluruhnya `git status` bersih —
tidak ada uncommitted work yang akan hilang.

### Aman dihapus — isi sudah ada di `main` (0 commit unik)

| Worktree | Branch |
|---|---|
| `agent-ac5ad36989e60a75d` | `agent/agent-ac5ad36989e60a75d` |
| `agent-af1ab6b1ed080d43d` | `agent/agent-af1ab6b1ed080d43d` |
| `hook-fix` | `agent/hook-fix` |
| `phase-7-final-test` | `agent/phase-7-final-test` |
| `settings-allowlist-fix` | `agent/settings-allowlist-fix` |
| `worktree-phase-8-hardening` | `agent/worktree-phase-8-hardening` |
| `org-migration` | `worktree-org-migration` |
| `phase-4-github-workflow` | `worktree-phase-4-github-workflow` |
| `phase-5-cleanup` | `worktree-phase-5-cleanup` |
| `phase-5-docs-update` | `worktree-phase-5-docs-update` |
| `phase-5-tool-pilot` | `worktree-phase-5-tool-pilot` |
| `fix-be102-contract` | `worktree-phase-7-contract-fix` |
| `fix-be102fix-contract` | `worktree-phase-7-contract-fix-2` |
| `phase-7-docs-update` | `worktree-phase-7-docs-update` |
| `fix-fe102-contract` | `worktree-phase-7-fe-contract-fix` |
| `qa-102-report` | `worktree-phase-7-qa-report` |
| `phase-8-implementation-prompt` | `worktree-phase-8-implementation-prompt` |
| `../m2s-vsh-platform-phase-7-plan-specs` | `worktree-phase-7-planning` |

### JANGAN hapus tanpa memeriksa — punya commit yang tidak ada di `main`

| Branch | Commit unik | Catatan |
|---|---|---|
| `agent/be-102-cors-fix` | 1 | worktree `be-102-cors-fix`; periksa apakah perbaikan CORS sudah masuk lewat jalur lain |
| `worktree-phase-8-hardening` | 3 | worktree `phase-8-hardening-draft`; Phase 8 di-merge lewat PR #24–#29, tetapi tiga commit ini tidak ter-squash — periksa isinya lebih dulu |
| `worktree-phase-6-design-infra` | 2 | PR #31 di-**squash**-merge, sehingga commit aslinya tidak ada di `main` meski isinya sudah masuk (merge commit `53586d8`). Aman dihapus setelah dikonfirmasi. |

Squash merge selalu meninggalkan branch sumber "unmerged" menurut
`git rev-list branch ^main`. Itu bukan tanda pekerjaan hilang — periksa PR-nya.

## Perintah

### 1. Hapus worktree yang isinya sudah di `main`

```bash
cd <akar-control-repo>
for w in \
  agent-ac5ad36989e60a75d agent-af1ab6b1ed080d43d hook-fix phase-7-final-test \
  settings-allowlist-fix worktree-phase-8-hardening org-migration \
  phase-4-github-workflow phase-5-cleanup phase-5-docs-update phase-5-tool-pilot \
  fix-be102-contract fix-be102fix-contract phase-7-docs-update \
  fix-fe102-contract qa-102-report phase-8-implementation-prompt; do
  git worktree remove ".claude/worktrees/$w"
done
git worktree remove ../m2s-vsh-platform-phase-7-plan-specs
git worktree prune -v
```

`git worktree remove` menolak bila worktree punya perubahan uncommitted — itu
pengaman, jangan tambahkan `--force` tanpa memeriksa `git status` di dalamnya.

### 2. Hapus branch yang isinya sudah di `main`

```bash
for b in \
  agent/agent-ac5ad36989e60a75d agent/agent-af1ab6b1ed080d43d agent/hook-fix \
  agent/phase-7-final-test agent/settings-allowlist-fix \
  agent/worktree-phase-8-hardening worktree-org-migration \
  worktree-phase-4-github-workflow worktree-phase-5-cleanup \
  worktree-phase-5-docs-update worktree-phase-5-tool-pilot \
  worktree-phase-7-contract-fix worktree-phase-7-contract-fix-2 \
  worktree-phase-7-docs-update worktree-phase-7-fe-contract-fix \
  worktree-phase-7-planning worktree-phase-7-qa-report \
  worktree-phase-8-implementation-prompt; do
  git branch -d "$b"
done
```

`git branch -d` (bukan `-D`) menolak branch yang belum merged — biarkan begitu.
Branch yang menolak berarti ada commit yang perlu diperiksa lebih dulu.

### 3. Periksa tiga branch sisa

```bash
git log --oneline agent/be-102-cors-fix ^main
git log --oneline worktree-phase-8-hardening ^main
git log --oneline worktree-phase-6-design-infra ^main
gh pr list --state merged --limit 20   # cocokkan dengan PR yang sudah masuk
```

Setelah dipastikan isinya sudah masuk lewat squash merge:

```bash
git worktree remove .claude/worktrees/be-102-cors-fix
git worktree remove .claude/worktrees/phase-8-hardening-draft
git worktree remove .claude/worktrees/worktree-phase-6-design-infra
git branch -D agent/be-102-cors-fix worktree-phase-8-hardening worktree-phase-6-design-infra
```

`-D` dipakai di sini secara sadar: squash merge membuat `-d` selalu menolak.
Jangan memakainya sebelum langkah 3 dijalankan.

### 4. Verifikasi

```bash
git worktree list          # hanya checkout utama + worktree yang sedang dipakai
git branch                 # hanya main + branch kerja aktif
git status                 # bersih
make verify                # tetap hijau
```

## Kenapa ini menumpuk

`EnterWorktree`/`ExitWorktree` menghapus direktori worktree tetapi tidak selalu
menghapus branch-nya, dan worktree yang dibuat runner (`m2s launch-task`)
memang tidak membersihkan dirinya — §30 menyatakan stale reservation hanya boleh
dibersihkan runner atau manusia, bukan worker. Cleanup periodik karena itu
memang pekerjaan manusia, bukan bug.

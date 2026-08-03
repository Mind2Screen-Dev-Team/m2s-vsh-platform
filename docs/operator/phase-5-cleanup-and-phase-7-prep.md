# Phase 6 Prep — Cleanup + Phase 7 Preparation

**PR #12 (docs update) harus di-merge lebih dulu sebelum PR ini.**

## Patch 1: Ponytail warn → full

File: `.claude/settings.json`

```diff
-    "PONYTAIL_DEFAULT_MODE": "warn",
+    "PONYTAIL_DEFAULT_MODE": "full",
```

Alasan: §5.5 arsitektur merekomendasikan `full`. Mode `warn` untuk pilot; pilot
Phase 5 berhasil, 3 skill baru tidak konflik dengan Ponytail.

## Patch 2: Register worktree-lifecycle.sh

File: `.claude/settings.json`

Tambahkan blok hook setelah `SubagentStop`:

```json
    "WorktreeCreate": [
      {
        "matcher": "*",
        "hooks": [
          { "type": "command", "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/worktree-lifecycle.sh" }
        ]
      }
    ],
    "WorktreeRemove": [
      {
        "matcher": "*",
        "hooks": [
          { "type": "command", "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/worktree-lifecycle.sh" }
        ]
      }
    ]
```

Alasan: T-04 dari Phase 4 — `worktree-lifecycle.sh` ada, self-test lulus,
tetapi tidak pernah dipanggil runtime. Hook ini menyediakan mitigasi R-26
(salin untracked config tanpa secret, instal dependency controlled, bersihkan
temp files, simpan result sebelum cleanup).

Kedua patch di atas ditulis manual oleh manusia karena `.claude/settings.json`
ada di `permissions.deny`.

---

## Persiapan Phase 7 (§63 Multi-Repo Pilot)

### Tujuan

Pilot end-to-end: contract approval → backend + frontend paralel → QA →
code review → merge ke develop.

### Prasyarat

- [ ] PR #12 merged (docs Phase 5)
- [ ] Ponytail `full` mode
- [ ] worktree-lifecycle hook terdaftar
- [ ] Repo backend (`m2s-vsh-project-backend`) + frontend (`m2s-vsh-project-frontend`) ter-clone lokal
- [ ] Clone lokal up-to-date (saat ini `behind 2`)
- [ ] Runner `bin/m2s` ter-build (`make build`)

### Fitur pilot

Pilih satu fitur kecil yang bisa dipecah menjadi:

```
CONTRACT-102 → create endpoint + response schema
BE-102       → implement backend endpoint
FE-102       → consume endpoint, render UI
QA-102       → acceptance + integration test
```

### Komponen yang perlu dipersiapkan

1. **Task contract factory** — `scripts/launch-task.sh` perlu mendukung
   reservation cross-repo. Saat ini reservation per-path di satu repo;
   Multi-Repo butuh PATH_OWNERSHIP_MATRIX lintas dua repo.

2. **Parallel launch** — Dua agent (BE + FE) dijalankan bersamaan setelah
   contract approved. Runner harus bisa spawn dua worktree paralel tanpa
   shared cwd.

3. **QA gate** — QA Agent membaca approved contract + BE/PR + FE/PR →
   menjalankan acceptance test → menghasilkan QA report.

4. **Merge order** — BE → FE → QA (sequential, bukan paralel). TL/SA
   menentukan merge order setelah semua PR siap.

### Yang sudah tersedia

- Path reservation (pathmatch, 24 kasus R-03)
- Contract schema (`task.schema.json`, `contracts/` path TL/SA)
- GitHub Workflow (CI gate, required checks, review requirement)
- PR pipeline (worker → reviewer → approver → merge)
- Mneme governance (warn-first, ADR memory)

### Decision yang perlu dibuat di Phase 7

| # | Decision | Owner |
|---|---|---|
| D-P7-1 | Fitur pilot (pilih satu yang cukup kecil) | PM + TL/SA |
| D-P7-2 | Apakah perlu Open Design (Phase 6) sebelum Phase 7? | UI/UX |
| D-P7-3 | Merge order: BE → FE → QA, atau BE + FE paralel merge? | TL/SA |
| D-P7-4 | `contracts/` path di repo control atau dedicated? | TL/SA |

### Acceptance criteria Phase 7 (§63)

```
contract → backend + frontend parallel → QA → review → merge
```

**Done:** Backend dan Frontend berjalan paralel dengan contract yang sama
serta tanpa overlap path.

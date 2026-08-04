# Prompt untuk Session Baru — Planning & Implementasi Phase 6 (§62 UI/UX Optional)

Salin blok di bawah ini ke session Claude Code baru untuk merencanakan dan
mengerjakan Phase 6. Fokus prompt ini adalah **planning dulu** — ikuti alur
rencana-sebelum-eksekusi (konvensi repo: [[m2s-vsh-cara-kerja]]).

---

## Posisi awal

M2S-VSH control repo. **Phase 8 selesai total** (2026-08-04): guard H-01..08,
branch protection 3 repo, arsip contract Phase 7, backward $64 metrics. PR
#24-#29 merged. `make verify` hijau penuh.

## Yang sudah ada (jangan dikerjakan ulang)

| Komponen | Status |
|---|---|
| `design/DESIGN.md` | ✅ design system 167 baris — colour tokens, typography, spacing, motion, aksesibilitas (control repo) |
| Skill `ui-ux-pro-max`, `emilkowalski`, `stop-slop` | ✅ terpasang (Phase 5) |
| Agent `ui-ux-designer`, `frontend-engineer` | ✅ template + test (Phase 5/8) |
| Worktree pattern | ✅ `m2s launch-task` runner — agent tak `git checkout` (Phase 8) |
| Phase 6 ditandai ⬜ dilewati di README | D-P7-2: boleh dijalankan bila fitur berikut butuh design |

## Dokumen otoritas

- `docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md` §62 (Phase 6)
- `docs/architecture/phase-8-hardening.md` — guard yang berlaku (wajib dipatuhi)
- `docs/decisions/D-P7-2-phase-6-skip.md` — kenapa dulu dilewati
- `design/DESIGN.md` — design system existing (sumber)
- `docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md` §44 (branch),
  §16 (universal rules), §29.6 (shared file)

## Apa itu Phase 6 (§62)

> Jalankan **Open Design pada isolated design workspace**. Buat design handoff
> flow. Batasi output paths.
> **Done:** Open Design tidak menulis application worktree.

## Batasan Phase 8 yang WAJIB dipegang (jangan dilanggar)

1. `.claude/**`, `cmd/m2s/**`, `Makefile`, `.github/**` human-only-write.
   Agent tak bisa sunting, bahkan dari worktree. Perubahan di sana → commit di
   branch + draft PR, manusia merge.
2. Merge flow `agent/* → develop → staging → main` (H-01/H-08). PR agent ke main
   ditolak. Manusia merge di main.
3. Branch non-task pakai `worktree-*`, bukan `agent/*` (H-02). Planning pakai
   `worktree-phase-6-*`.
4. Worktree pattern: `m2s launch-task` (runner) buat worktree; agent tak
   `git checkout`/`git worktree`/`git reset --hard`/`rm -rf` (deny).
5. `make verify` tetap hijau — guard baru tak boleh melanggar existing test (§68).

## Pertanyaan yang harus dijawab PLANNING dulu (sebelum implementasi)

1. **Scope fitur**: Phase 6 butuh "fitur berikut yang memerlukan design".
   D-P7-2 dilewati karena StatusCard statis. Fitur apa sekarang yang butuh
   design komprehensif? Kalau belum ada — Phase 6 dijalankan di atas fitur apa?
   (Belum ditentukan — ini keputusan Project Manager / UI/UX dulu.)
2. **Open Design workspace**: di mana? Isolated design workspace (tidak menyentuh
   application worktree). Pola worktree `design/`? Atau repo terpisah?
   DESIGNGrant: design workspace harus ter-isolasi dari app worktree.
3. **Design handoff flow**: bagaimana UI/UX menyerahkan ke frontend-engineer?
   Format (DESIGN.md? file per component? token)? Alur contract → design →
   implementation?
4. **Output paths**: batasi output design (contoh `design/**`, `docs/design/**`)?
   Contract path allowed utk ui-ux-designer harus terdefinisi.

## Deliverables Phase 6 (per §62, diturunkan jadi konkret)

- Rencana Open Design workspace (lokasi + isolasi)
- Design handoff flow (formal, bukan ad hoc)
- Batasan output paths design
- Bukti: design tak menulis application worktree (kriteria Done)

## Urutan yang disarankan

1. **Decision gate**: tentukan fitur target design (tanpa fitur, Phase 6 tak
   punya objek — sebaguna D-P7-2 dulu dilewati karena tak ada fitur).
2. **Planning**: tulis rencana (worktree pattern `worktree-phase-6-*`), tentukan
   struktur design workspace, handoff flow, output paths. Komit + draft PR.
3. **ADR** bila simpang jauh dari §62 — konvensi repo: ADR wajib utk simpangan.
4. Implementasi, `make verify` hijau, PR per grup logis, manusia merge.

## Caveat

- Phase 6 **opsional** — §62 sendiri bilang "Open Design bersifat opsional dan
  hanya digunakan dalam design workspace terisolasi". Kalau belum ada fitur yang
  butuh design, keputusan sah: pertahankan skip (perbarui D-P7-2 yang sudah ada),
  atau tentukan fitur yang akan dikerjakan dulu.
- Jangan buat design system baru — `design/DESIGN.md` sudah ada (167 baris).
  Phase 6 = Open Design flow + handoff, bukan redesign system.
- `git worktree`, `git checkout`, `git reset --hard`, `rm -rf` denied utk agent.

## Dokumen rujukan lain

- `docs/operator/phase-8-branch-protection.md` — kenapa main dilindungi
- `docs/operator/phase-8-human-only-patches.md` — pola patch utk path denied
- `docs/operator/task-status.md` — format status task (kalau ada task design)
- `docs/operator/phase-8-metrics.md` — §64 hasil, konteks stabilisasi
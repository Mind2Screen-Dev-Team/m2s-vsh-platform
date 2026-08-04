# Phase 6 — Open Design Workspace

**Status:** ✅ selesai (2026-08-04) — redeskripsi `design/` + handoff flow + isolasi.
**Sumber:** §62 dokumen arsitektur, §8 (Open Design), R-17, D-P7-2.

## Tujuan

Mendefinisikan design workspace terisolasi + alur handoff design → frontend,
sesuai §62. Dulu dilewati (D-P7-2) karena fitur pilot StatusCard statis tak butuh
design; kini dijalankan sebagai **design infra** — struktur artifact, handoff
flow formal, dan bukti isolasi design dari application worktree.

## Lokasi workspace

Design workspace = **worktree control repo**, bukan repo/workspace lain.

```
branch:  worktree-design-*  (dari main control repo)
cwd sesi UI/UX: design worktree → .claude/worktrees/<name>/
artifact: design/**  (owner ui-ux-designer, §19.7)
```

Alasan:
- Open Design wajib berjalan pada design workspace terisolasi (§8.3, R-17).
- Design worktree control repo **fisik terpisah** dari repo frontend — sesi yang
  jalan di sana tidak punya akses menulis ke application worktree.
- Tidak perlu repo ke-4: kontrol repo kecil, design artifact memang miliknya.

## Struktur design/ (control repo)

```
design/
├── DESIGN.md                  # design system (colour, type, spacing, motion, a11y)
├── tokens/                    # token tambahan: spacing, radius, border, status map
├── flows/                     # user flow (mermaid)
├── wireframes/                # layout ASCII per screen/state
├── prototypes/                # artifact HTML/CSS statis (output Open Design)
└── handoff/
    └── DESIGN-<N>/            # per task: handoff.md + contract.yaml
```

## Design handoff flow

```
PM requirement + TL/SA system flow
  → UI/UX (design worktree control) tulis design/handoff/DESIGN-<N>/*
  → PM + TL/SA review → status approved (mirip gate H-07)
  → frontend-engineer task contract sebut handoff sbg input
  → FE baca handoff (Read) + implement src/** (forbid design/** tetap)
  → QA verifikasi vs handoff
```

Format handoff (`design/handoff/DESIGN-<N>/`):
- `handoff.md` — ringkasan, input, spesifikasi komponen, states, responsive,
  aksesibilitas, DoD frontend.
- `contract.yaml` — ref silang ke task FE, daftar state, breakpoint, token.

Batas peran:
- **FE tak menulis `design/**`** — wilayah UI/UX (§21.5 forbid). Handoff adalah
  input Read-only bagi FE.
- **UI/UX tak menulis `src/**`** — wilayah FE (§19.6 forbid).

## Output path design

Wilayah sah UI/UX (`ui-ux-designer` writable):

```text
design/DESIGN.md
design/tokens/**
design/flows/**
design/wireframes/**
design/prototypes/**
design/handoff/**
```

Forbidden:
```text
src/**  .claude/**  .mneme/**  .task/**
```

## Bukti isolasi (kriteria Done §62)

Test negatif `tests/negative/design-isolation.test.sh` (dijalankan `make verify`):

| Kasus | Hasil |
|-------|-------|
| Edit/Write ke `../m2s-vsh-project-frontend/src/**` | ditolak exit 2 |
| Edit/Write ke `../m2s-vsh-project-frontend/design/**` | ditolak exit 2 |
| Edit/Write ke `src/**` di dalam worktree | ditolak exit 2 |
| Edit/Write ke `.claude/**` | ditolak exit 2 |
| `git -C ../frontend commit` | ditolak exit 2 |
| Edit/Write ke `design/**` (wilayah sah) | diizinkan exit 0 |

## Cara menjalankan design task

1. Task design (mis. `DESIGN-1`, role `ui-ux-designer`, repo control):
   `scripts/launch-task.sh` buat worktree dari `main`, materialize
   `.task/contract.json` (allowed `design/**`), spawn agent UI/UX di sana.
2. Task sync design system ke frontend (mis. `FE-DESIGN-1`, role
   `frontend-engineer`, repo frontend): salin `design/DESIGN.md` + `design/tokens/**`
   dari control ke `design/` di frontend. Contract allowed path eksplisit utk
   menimpa template FE forbid `design/**` (task contract > template, §15).
3. Merge flow: `agent/* → develop → staging → main` (H-01/H-08). PR design ke
   control repo main via manusia.

## Referensi

- §62 (Phase 6), §8 (Open Design), §19.7 (UI/UX path) — dokumen arsitektur
- D-P7-2 (histori skip; superseded)
- R-17 (risiko isolasi design)
- `tests/negative/design-isolation.test.sh` — guard isolasi
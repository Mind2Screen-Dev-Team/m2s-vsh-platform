# Phase 7 Pilot Plan — Multi-Repo Pilot

**Tanggal:** 2026-08-03
**Arsitektur:** §63 M2S-VSH Lite v0.1.0
**Status:** ✅ **selesai** (3 Agustus 2026)

## Ringkasan

Pilot end-to-end pertama: contract approval → backend + frontend paralel → QA →
code review → merge.

Repositori yang terlibat:
- `Mind2Screen-Dev-Team/m2s-vsh-platform` (control)
- `Mind2Screen-Dev-Team/m2s-vsh-project-backend` (Go, seed-only)
- `Mind2Screen-Dev-Team/m2s-vsh-project-frontend` (Next.js, seed-only)

## Fitur Pilot: System Status

`GET /api/v1/status` — endpoint health sederhana dengan response `{status, version, uptime_seconds}`.

Dipilih karena:
- Kedua repo seed-only — tidak ada kode yang perlu dimodifikasi.
- Contract jelas — satu endpoint, response sederhana.
- Path terpisah sempurna — BE di `internal/handler/`, FE di `src/components/`.
- Zero shared-file contention — tidak ada `go.mod` atau `package.json` yang perlu disentuh.

## Keputusan

| ID | Decision | Keputusan |
|---|---|---|
| D-P7-1 | Fitur pilot | System Status: GET /api/v1/status + StatusCard |
| D-P7-2 | Phase 6 sebelum Phase 7? | Skip — pilot terlalu kecil, no design handoff |
| D-P7-3 | Merge order | Sequential: BE merge → FE merge → QA → review |
| D-P7-4 | Contracts path | Control repo (`contracts/`) |

Lihat `docs/decisions/D-P7-*.md` untuk masing-masing ADR.

## Task DAG

```
CONTRACT-102 (TL/SA) → BE-102 + FE-102 (parallel)
    → QA-102 (integration)
    → Code Reviewer
    → merge develop
```

## Path Reservation

| Task | Repo | Allowed Paths |
|---|---|---|
| CONTRACT-102 | control | `contracts/CONTRACT-102.yaml` |
| BE-102 | backend | `cmd/server/**`, `internal/handler/**`, `internal/handler/status_test.go` |
| FE-102 | frontend | `src/components/StatusCard/**`, `src/app/status/**` |
| QA-102 | control | `qa/CONTRACT-102/**` |

Non-overlap: beda repositori = beda filesystem. Masing-masing repo hanya satu writer.

## Prasyarat Eksekusi

- [x] Backend repo up-to-date
- [x] Frontend repo up-to-date
- [x] `make build` di control repo (runner binary)
- [x] `make verify` green
- [x] Runner `bin/m2s` ada dan tervalidasi

## Hasil (3 Agustus 2026)

| Repo | PR | Isi | Merged |
|---|---|---|---|
| control | #16 | planning (contract + 4 specs + 4 ADR) | ✅ |
| control | #17 | QA-102 report (PASS 5/5) | ✅ |
| control | #18 | BE-102 contract fix (go.mod shared file) | ✅ |
| control | #19 | BE-102-fix spec (CORS task) | ✅ |
| control | #20 | FE-102 contract fix (scaffold files) | ✅ |
| backend | #9 | status endpoint | ✅ |
| backend | #11 | CORS header fix | ✅ |
| frontend | #5 | StatusCard + /status | ✅ |

**Verifikasi live:** `GET /api/v1/status` → HTTP 200, CORS header `*`, `go test -race` PASS.

### Acceptance criteria §63 — terpenuhi
- Backend + Frontend berjalan paralel dengan contract yang sama ✅
- Tanpa overlap path (BE `internal/handler/`, FE `src/components/`) ✅
- QA acceptance + integration PASS ✅

### Catatan proses (learning Phase 7)
- Contract awal terlalu ketat (forbid `go.mod`/`layout.tsx` yang scaffolding wajib) → 3 PR fix contract
- Branch planning awalnya `agent/*` → harus `worktree-*` (2 PR rename)
- Semua PR ke main langsung, bukan `develop→staging→main` — menjadi dasar hardening H-01/H-08
- Hardening blueprint: `docs/architecture/phase-8-hardening.md`

# Phase 7 Pilot Plan — Multi-Repo Pilot

**Tanggal:** 2026-08-03
**Arsitektur:** §63 M2S-VSH Lite v0.1.0
**Status:** planning

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

- [ ] Backend repo up-to-date
- [ ] Frontend repo up-to-date
- [ ] `make build` di control repo (runner binary)
- [ ] `make verify` green
- [ ] Runner `bin/m2s` ada dan tervalidasi

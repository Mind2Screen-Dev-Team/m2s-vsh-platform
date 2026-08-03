# ADR D-P7-1: Fitur Pilot Phase 7 — System Status

**Tanggal:** 2026-08-03
**Decider:** PM + TL/SA
**Status:** approved

## Context

Phase 7 (§63) membutuhkan satu fitur yang cukup kecil untuk pilot multi-repo
pertama. Kedua repositori aplikasi (backend Go, frontend Next.js) masih seed-only.

## Decision

**Fitur pilot: System Status — `GET /api/v1/status` + StatusCard component.**

Contract:

```yaml
GET /api/v1/status
200:
  status: "ok" | "degraded"   # string
  version: string              # semver
  uptime_seconds: number       # float64
```

## Rationale

1. **Seed-only repos** — tidak ada kode existing, tidak perlu modifikasi.
2. **Contract sederhana** — satu endpoint, response flat, zero nested object.
3. **Path terpisah sempurna** — BE `internal/handler/`, FE `src/components/StatusCard/`, tidak overlap.
4. **Zero shared-file contention** — tidak perlu edit `go.mod` atau `package.json`.
5. **Verifikasi gampang** — `go test ./...` + `npm run build` langsung hijau.
6. **Mudah rollback** — satu file handler + satu component, mudah revert.

## Consequences

- Pilot mencakup full pipeline tanpa kompleksitas domain logic.
- Tidak menguji multi-file coordination (tidak perlu untuk pilot minimal).
- Meniadakan kebutuhan Open Design (Phase 6 skipped).
# Design Handoff — DESIGN-2: Dashboard Multi-Status

**Status:** draft — menunggu review PM + TL/SA (approved after review, H-07 gate)
**Design workspace:** `design/handoff/DESIGN-2/`
**Owner:** ui-ux-designer · **Consumer:** frontend-engineer

## Ringkasan

Perluas halaman `/status` jadi dashboard multi-service. Satu halaman + satu
component `StatusDashboard` yang mengonsumsi `GET /api/v1/status` (diperluas,
CONTRACT-201: field lama + `services` array 3 komponen). `StatusCard` jadi
presentational — menampilkan satu komponen, tanpa fetch sendiri.

## Input

| Sumber | Path |
|--------|------|
| Design system | `design/DESIGN.md` |
| Token status | `design/tokens/tokens.md` |
| API contract | `contracts/CONTRACT-201.yaml` → `{status, version, uptime_seconds, services[]}` |

## Layout halaman

```
┌────────────────────────────────────────────┐
│ System Status                    [↻ Refresh] │
│  auto-refresh tiap 30 detik                  │
├────────────────────────────────────────────┤
│ ┌─────────┐ ┌─────────┐ ┌─────────┐        │
│ │backend- │ │database │ │  auth   │        │
│ │  api    │ │         │ │         │        │
│ └─────────┘ └─────────┘ └─────────┘        │
│   (grid-cols-1 → md:2 → lg:3, gap-4)        │
└────────────────────────────────────────────┘
```

## Komponen

### StatusDashboard (baru)

| Properti | Nilai |
|----------|-------|
| Container | section, `aria-live="polite"` (region status dinamis) |
| Grid | `grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4` |
| Fetch | `GET /api/v1/status`, `setInterval` 30.000 ms, clear on unmount |
| Refresh manual | button `aria-label="Perbarui status"`, reset interval |
| Header | h1 "System Status" + hint "auto-refresh tiap 30 detik" |
| Warna | `--accent-success/warning/danger`, `--text-muted` utk unknown |

### StatusCard (refactor jadi presentational)

| Properti | Nilai |
|----------|-------|
| Props | `component: {id, status, version, uptime, last_checked, latency_ms, message}` |
| Container | card: `bg-surface`, border `--border-default`, radius `--radius-md`, shadow `--shadow-card`, padding `--space-5` |
| Header | dot indikator (12px round) + teks status capital, gap `--space-3` |
| Body | baris: Versi, Uptime, Latensi, Terakhir dicek; font 0.875rem, `--text-secondary` |
| Message | baris opsional bila ada, font 0.75rem |
| Hapus | self-fetch, useState/useEffect, max-w-md |

### Dot indikator status

| Status | Warna |
|--------|-------|
| `ok` | `--accent-success` |
| `degraded` | `--accent-warning` |
| `error` | `--accent-danger` |
| `unknown` | `--text-muted` |

## States

| State | Perilaku | UI |
|-------|----------|-----|
| Loading | fetch pending | skeleton grid 3 kartu pulse (`animate-pulse`) |
| Empty | services kosong/0 | teks "Tidak ada komponen status" + refresh |
| Success | 200 + services ≥ 1 | grid 3 kartu + header |
| Error | fetch gagal | teks "Gagal: <msg>" `--accent-danger` + tombol refresh retry |

## Responsive

- `<640px`: grid 1 kolom, padding `--space-4`.
- `640–1024px`: grid 2 kolom.
- `>1024px`: grid 3 kolom.
- Page: `min-h-screen`, padding `--space-8`, `bg-primary`.

## Accessibility

- Kontras text ≥ 4.5:1 (token DESIGN.md sudah AA).
- `prefers-reduced-motion`: matikan pulse animation (loading skeleton statis).
- Region status `aria-live="polite"` — perubahan status diumumkan screen reader.
- Button refresh fokusable, `aria-label` deskriptif.
- Dot murni dekoratif → `aria-hidden`; teks status tetap dibaca.

## Token referensi

Semua gaya dari `design/DESIGN.md`; tambahan status mapping di
`design/tokens/tokens.md`. Prototype `design/prototypes/status-screen.html`
(DESIGN-1) adalah preview visual kartu tunggal — dashboard pakai token yang sama,
bukan kode yang disalin.

## Definition of Done (frontend)

- `src/app/status/page.tsx` + `src/components/StatusDashboard/StatusDashboard.tsx` + `src/components/StatusCard/StatusCard.tsx`
- Semua state tersedia: loading, empty, success, error.
- Polling 30s + manual refresh.
- Contract integration benar (CONTRACT-201).
- Accessibility baseline + responsive per tabel di atas.
- Vitest: loading→success, error+refresh, empty.

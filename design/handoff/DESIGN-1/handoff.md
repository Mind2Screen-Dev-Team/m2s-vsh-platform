# Design Handoff — DESIGN-1: System Status Screen

**Status:** draft — menunggu review PM + TL/SA (approved after review, H-07 gate)
**Design workspace:** `design/handoff/DESIGN-1/`
**Owner:** ui-ux-designer · **Consumer:** frontend-engineer

## Ringkasan

Handoff tampilan System Status di frontend: satu halaman `/status` + satu
component `StatusCard` yang mengkonsumsi `GET /api/v1/status`.

## Input

| Sumber | Path |
|--------|------|
| Design system | `design/DESIGN.md` |
| Token status | `design/tokens/tokens.md` |
| Flow | `design/flows/system-status-flow.md` |
| Wireframe | `design/wireframes/status-screen-wireframe.md` |
| Prototype | `design/prototypes/status-screen.html` |
| API contract | `GET /api/v1/status` → `{status, version, uptime_seconds}` |

## Spesifikasi komponen

### StatusCard

| Properti | Nilai |
|----------|-------|
| Container | card: `bg-surface`, border `--border-default`, radius `--radius-md`, shadow `--shadow-card`, padding `--space-5`, `max-w-md` |
| Header | dot indikator (12px, round) + teks status capital, gap `--space-3`, font-semibold |
| Body | dua baris: Versi, Uptime; font 0.875rem, warna `--text-secondary`, label bold |
| Card Bukan interactive | tanpa focus management, tanpa hover state |

### Dot indikator status

| Status | Warna |
|--------|-------|
| `ok` | `--accent-success` |
| `degraded` | `--accent-warning` |
| error/unknown | `--accent-danger` / `--text-muted` |

## States

| State | Perilaku | UI |
|-------|----------|-----|
| Loading | fetch pending | teks "Memuat status sistem..." + pulse (`animate-pulse`) |
| Success | 200 `status: ok` | dot hijau + `ok` + detail |
| Degraded | 200 `status: degraded` | dot kuning + `degraded` + detail |
| Error | fetch gagal | teks "Gagal: <msg>" warna `--accent-danger` |
| Unknown | tanpa data | dot `--text-muted`, label "tidak diketahui" |

## Responsive

- `<640px`: card full-width, padding `--space-4`.
- `640px+`: card `max-w-md`, centered.
- Page: `min-h-screen`, flex center, `bg-primary`.

## Accessibility

- Kontras text ≥ 4.5:1 (token DESIGN.md sudah AA).
- `prefers-reduced-motion`: matikan pulse animation.
- Status perubahan (loading → data) sebaiknya `aria-live="polite"` bila
  dinamis; card non-interactive tak butuh keyboard handler.
- Icon/gambar: dot murni dekoratif → `aria-hidden` atau text label sudah ada.

## Token referensi

Semua gaya dari `design/DESIGN.md`; tambahan status mapping di
`design/tokens/tokens.md`. Prototype `design/prototypes/status-screen.html`
adalah preview visual (bukan kode yang disalin).

## Definition of Done (frontend)

- `src/app/status/page.tsx` + `src/components/StatusCard/StatusCard.tsx`
- Semua state tersedia: loading, success, degraded, error.
- Contract integration benar (`GET /api/v1/status`).
- Accessibility baseline + responsive per tabel di atas.
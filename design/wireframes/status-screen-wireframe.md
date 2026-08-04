# Wireframe — System Status Screen

**Design workspace:** `design/wireframes/`
**Referensi:** `design/handoff/DESIGN-1/`

## Layout (desktop, > 1024px)

```
┌────────────────────────────────────────────────────────┐
│  System Status                                         │
│  ┌────────────────────────────────────────────────┐    │
│  │  StatusCard                                    │    │
│  │  ● ok                           [max-w-md]     │    │
│  │                                                │    │
│  │  Versi: v0.1.0                                 │    │
│  │  Uptime: 3j 12m                                │    │
│  └────────────────────────────────────────────────┘    │
│                                        (kiri tengah)   │
└────────────────────────────────────────────────────────┘
```

- Halaman: `min-h-screen`, `bg-primary`, centering via flex.
- Card: `max-w-md`, `bg-surface`, border `--border-default`, radius `--radius-md`,
  shadow `--shadow-card`, padding `--space-5`.
- Header card: dot (12px) + status texto, gap `--space-3`, font-semibold.

## States

### Loading
```
┌──────────────────────────────┐
│  Memuat status sistem...     │   <- animate-pulse
└──────────────────────────────┘
```

### Error
```
┌──────────────────────────────┐
│  Gagal: <pesan>              │   <- text-danger
└──────────────────────────────┘
```

### Success / Degraded (detail card)
```
┌──────────────────────────────┐
│  ● ok          ● degraded    │   <- dot 12px round
│  Versi:  v0.1.0              │
│  Uptime: 3j 12m              │
└──────────────────────────────┘
```

## Responsive

- `<640px`: single column, card full-width, padding `--space-4`.
- `640-1024px`: card tetap max-w-md, horizontal center.
- `>1024px`: default.

## Catatan implementasi

Dot indikator mapping status → warna ada di `design/tokens/tokens.md`.
Keyboard: card tidak interactive — hanya info render, no focus management.
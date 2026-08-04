# Design Tokens — M2S-VSH

Referensi token tambahan di luar `DESIGN.md`. Sumber pertama tetap
`design/DESIGN.md` (colour, typography, spacing). File ini mencatat token yang
tidak ada di DESIGN.md tapi dipakai prototype / handoff.

## Spacing (ringkas dari DESIGN.md)

| Token | Value |
|-------|-------|
| `--space-1` | 0.25rem |
| `--space-2` | 0.5rem |
| `--space-3` | 0.75rem |
| `--space-4` | 1rem |
| `--space-6` | 1.5rem |
| `--space-8` | 2rem |

## Radius

| Token | Value | Usage |
|-------|-------|-------|
| `--radius-sm` | 0.375rem (6px) | button, input |
| `--radius-md` | 0.5rem (8px) | card, panel |

## Border

| Token | Value |
|-------|-------|
| `--border-default` | 1px `--border-default` (DESIGN.md) |
| `--border-focus` | 2px `--border-focus` ring, offset 2px |

## Shadow

| Token | Light | Dark |
|-------|-------|------|
| `--shadow-card` | `0 1px 3px rgba(0,0,0,0.1)` | `0 1px 3px rgba(0,0,0,0.3)` |

## Status indicator

StatusCard memakai dot indikator. Mapping status → warna:

| Status | Token |
|--------|-------|
| ok | `--accent-success` |
| degraded | `--accent-warning` |
| error | `--accent-danger` |
| unknown | `--text-muted` |

Dipakai oleh prototype `design/prototypes/status-screen.html` dan handoff
`design/handoff/DESIGN-1/`.

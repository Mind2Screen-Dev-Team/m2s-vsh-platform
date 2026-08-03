# M2S-VSH Design System

## Product Personality

M2S-VSH is a developer tooling platform. The interface should feel precise, trustworthy, and unobtrusive. It gets out of the way so engineers can focus on building.

- **Tone:** Technical, direct, calm under pressure
- **Density:** Information-dense; whitespace is earned, not default
- **Motion:** Functional, not decorative. Animations serve state indication and spatial consistency.
- **Accessibility:** WCAG 2.1 AA minimum; keyboard-first for all interactive elements

## Colour Tokens

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `--bg-primary` | `#F8FAFC` | `#0F172A` | Page background |
| `--bg-surface` | `#FFFFFF` | `#1E293B` | Cards, panels, modals |
| `--bg-elevated` | `#FFFFFF` | `#334155` | Floating elements, dropdowns |
| `--text-primary` | `#1E293B` | `#F8FAFC` | Headings, primary text |
| `--text-secondary` | `#64748B` | `#94A3B8` | Labels, metadata, captions |
| `--text-muted` | `#94A3B8` | `#475569` | Disabled, placeholder |
| `--border-default` | `#E2E8F0` | `#334155` | Dividers, borders |
| `--border-focus` | `#2563EB` | `#60A5FA` | Focus rings |
| `--accent-primary` | `#2563EB` | `#3B82F6` | Primary actions, links |
| `--accent-primary-hover` | `#1D4ED8` | `#2563EB` | Primary hover |
| `--accent-success` | `#059669` | `#10B981` | Success states, passing checks |
| `--accent-warning` | `#D97706` | `#F59E0B` | Warnings, pending states |
| `--accent-danger` | `#DC2626` | `#EF4444` | Errors, failures, destructive actions |
| `--accent-info` | `#0891B2` | `#22D3EE` | Info, tips, neutral highlights |

## Typography

| Role | Font | Size | Weight | Line Height | Letter Spacing |
|------|------|------|--------|-------------|----------------|
| Display | system-ui | 2rem (32px) | 700 | 1.1 | -0.02em |
| H1 | system-ui | 1.5rem (24px) | 600 | 1.2 | -0.01em |
| H2 | system-ui | 1.25rem (20px) | 600 | 1.3 | 0 |
| H3 | system-ui | 1.125rem (18px) | 500 | 1.4 | 0 |
| Body | system-ui | 1rem (16px) | 400 | 1.5 | 0 |
| Small | system-ui | 0.875rem (14px) | 400 | 1.5 | 0 |
| Caption | system-ui | 0.75rem (12px) | 400 | 1.4 | 0.01em |
| Mono | ui-monospace, SFMono-Regular, Menlo, monospace | 0.875rem | 400 | 1.5 | 0 |

## Spacing Scale

| Token | Value |
|-------|-------|
| `--space-1` | 0.25rem (4px) |
| `--space-2` | 0.5rem (8px) |
| `--space-3` | 0.75rem (12px) |
| `--space-4` | 1rem (16px) |
| `--space-5` | 1.25rem (20px) |
| `--space-6` | 1.5rem (24px) |
| `--space-8` | 2rem (32px) |
| `--space-10` | 2.5rem (40px) |
| `--space-12` | 3rem (48px) |
| `--space-16` | 4rem (64px) |

## Grid and Layout

- **Container max-width:** 1280px
- **Sidebar width:** 240px (collapsible to 64px)
- **Content padding:** `--space-4` to `--space-6`
- **Card padding:** `--space-4` to `--space-5`
- **Gap between cards:** `--space-4`
- **Form field gap:** `--space-3`

## Component Principles

### Buttons

| Variant | Background | Text | Border | Hover |
|---------|----------|------|--------|-------|
| Primary | `--accent-primary` | white | none | `--accent-primary-hover` |
| Secondary | `--bg-surface` | `--text-primary` | `--border-default` | `--bg-primary` |
| Ghost | transparent | `--text-secondary` | none | `--bg-primary` |
| Danger | `--accent-danger` | white | none | darken 10% |

- **Border radius:** 0.375rem (6px)
- **Padding:** `--space-2` `--space-4`
- **Height:** 2.25rem (36px) standard, 2rem (32px) compact
- **Active state:** `transform: scale(0.97)` — instant feedback on press

### Inputs

- **Border:** 1px `--border-default`
- **Border radius:** 0.375rem
- **Focus:** 2px `--border-focus` ring, offset 2px
- **Error:** border `--accent-danger`, icon + message below
- **Disabled:** `--text-muted` text, `--bg-primary` background

### Cards

- **Background:** `--bg-surface`
- **Border:** 1px `--border-default`
- **Border radius:** 0.5rem (8px)
- **Shadow:** `0 1px 3px rgba(0,0,0,0.1)` (light), `0 1px 3px rgba(0,0,0,0.3)` (dark)
- **Hover (interactive):** shadow increases, no border change

### Tables

- **Header:** `--bg-primary` background, `--text-secondary` text, uppercase caption style
- **Row hover:** `--bg-primary` tint
- **Selected row:** `--accent-primary` at 10% opacity
- **Cell padding:** `--space-3` vertical, `--space-4` horizontal

## Interaction States

| State | Visual Treatment |
|-------|-----------------|
| Default | As defined above |
| Hover | Subtle background shift or shadow increase |
| Active/Pressed | `transform: scale(0.97)` for buttons; instant feedback |
| Focus | 2px ring `--border-focus`, offset 2px |
| Disabled | `--text-muted`, no hover effects, cursor not-allowed |
| Loading | Spinner replaces content, maintains dimensions |
| Empty | Centered icon + text, `--text-secondary` |
| Error | `--accent-danger` border/icon, error message below |
| Success | `--accent-success` icon, brief checkmark animation |

## Motion Rules

- **Default duration:** 150-200ms for UI elements
- **Easing:** `cubic-bezier(0.23, 1, 0.32, 1)` (ease-out) for enters; `cubic-bezier(0.77, 0, 0.175, 1)` (ease-in-out) for moves
- **No animation on:** keyboard shortcuts, command palette toggle, rapid-repeated actions
- **Springs:** Only for drag, swipe, gesture-driven interactions
- **Reduced motion:** `@media (prefers-reduced-motion: reduce)` — cross-fade opacity, no transform motion

## Responsive Behaviour

| Breakpoint | Behaviour |
|------------|-----------|
| < 640px | Single column, sidebar becomes drawer, tables scroll horizontally |
| 640-1024px | Two columns where applicable, sidebar collapsible |
| > 1024px | Full layout, sidebar expanded by default |

## Accessibility Requirements

- All interactive elements keyboard accessible (Tab order logical)
- Focus visible on all focusable elements
- Color contrast minimum 4.5:1 for text, 3:1 for UI components
- Screen reader labels for icon-only buttons
- `aria-live` regions for dynamic status updates
- Skip link for main content

## Do / Don't

### Do
- Use system font stack for performance and familiarity
- Animate only `transform` and `opacity` for 60fps
- Provide loading states for all async operations
- Use exact property names in transitions (`transition: transform 200ms ease-out`)
- Respect `prefers-reduced-motion`

### Don't
- Use `ease-in` for UI animations — feels sluggish
- Animate from `scale(0)` — elements should appear from `scale(0.95)` with opacity
- Use `transition: all` — specify exact properties
- Animate layout properties (width, height, margin, padding)
- Use decorative animations on high-frequency actions

## Implementation Notes

- CSS custom properties (variables) for all tokens
- Dark mode via `data-theme="dark"` on root or `@media (prefers-color-scheme: dark)`
- Component library: prefer existing design-system components; create local only when scope allows
- Tailwind CSS utility-first approach recommended if project uses Tailwind
- `clsx` for conditional className strings; `cva` for variant-driven components
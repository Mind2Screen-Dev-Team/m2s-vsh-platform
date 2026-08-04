# User Flow — System Status Screen

**Design workspace:** `design/flows/`
**Referensi:** `design/handoff/DESIGN-1/`, backend `GET /api/v1/status`

## Flow

```mermaid
flowchart TD
    A[Masuk halaman /status] --> B{Data status?}
    B -->|loading| C[Tampilkan loading state]
    C --> D{Response?}
    D -->|200 ok| E[Tampilkan StatusCard - status ok]
    D -->|200 degraded| F[Tampilkan StatusCard - status degraded]
    D -->|error| G[Tampilkan error state]
    G --> H[Tombol coba lagi]
    H --> B
```

## State

| State | Trigger | UI |
|-------|---------|-----|
| Loading | fetch pending | teks "Memuat status sistem..." + pulse |
| Success | 200, `status: ok` | dot hijau + `ok` + version + uptime |
| Degraded | 200, `status: degraded` | dot kuning + `degraded` |
| Error | fetch gagal | teks "Gagal: <pesan>" merah |

## Skenario

- User buka `/status` langsung → auto fetch.
- Backend turun → error state + retry.
- Backend balas `degraded` → warning dot, bukan merah (masih tersedia sebagian).

Flow minimal — status screen pilot tidak punya multi-step navigation.

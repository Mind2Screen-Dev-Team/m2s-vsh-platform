# ADR D-P7-2: Phase 6 (Open Design) sebelum Phase 7

**Tanggal:** 2026-08-03
**Decider:** UI/UX
**Status:** superseded (4 Agustus 2026) — Phase 6 dijalankan sebagai design infra

> **Superseded.** Keputusan skip di bawah berlaku untuk konteks Phase 7: fitur
> pilot StatusCard statis sehingga tidak butuh design handoff. Pada 4 Agustus
> 2026 Phase 6 dijalankan dengan objek berbeda — bukan design untuk sebuah fitur,
> melainkan **design workflow infra**: struktur `design/`, handoff flow formal,
> dan bukti isolasi design dari application worktree (§62 kriteria Done).
> Lihat [`docs/operator/phase-6-open-design.md`](../operator/phase-6-open-design.md).
> Alasan skip di bawah tetap sahih sebagai catatan historis.

## Context

Phase 6 (§62) adalah Open Design workspace. Phase 7 (§63) butuh design handoff
bila ada UI komponen baru. Fitur pilot System Status menggunakan komponen sederhana.

## Decision

**Skip Phase 6. Tidak ada Open Design sebelum Phase 7.**

## Rationale

1. **Phase 6 opsional** — §62: "Open Design bersifat opsional dan hanya digunakan
   dalam design workspace terisolasi."
2. **Fitur terlalu kecil** — StatusCard komponen statis, tidak ada variabel layout
   atau interaksi kompleks.
3. **Existing design system cukup** — `design/DESIGN.md` di repo frontend sudah
   punya typography + spacing token.
4. **Avoid ceremony** — Open Design workspace menambah langkah tanpa menambah
   nilai untuk pilot minimal.

## Consequences

- UI/UX tidak perlu jalankan Open Design workspace.
- Frontend agent pakai existing `design/DESIGN.md` tanpa handoff baru.
- Phase 6 tetap bisa dijalankan nanti bila fitur berikutnya butuh design lebih
  komprehensif.
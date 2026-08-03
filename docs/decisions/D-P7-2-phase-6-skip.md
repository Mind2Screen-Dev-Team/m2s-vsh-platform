# ADR D-P7-2: Phase 6 (Open Design) sebelum Phase 7

**Tanggal:** 2026-08-03
**Decider:** UI/UX
**Status:** approved

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
# ADR-012: Invariant A-01 — boundary pindah dari hooks ke tool-gate orkestrator

**Tanggal:** 2026-08-06
**Decider:** technical-lead-system-analyst
**Status:** draft

## Context

Pada v0.1.0, invariant A-01 (forbidden path ditolak) ditegakkan lewat hooks
fail-closed (PreToolUse/PostToolUse) + CI `validate-changed-paths`. Orkestrator
baru (ADR-010, engine C) menjalankan agent tanpa sesi Claude Code — hooks tidak
berjalan. Boundary harus tetap dipertahankan pada harness baru, dengan cara uji
konkrit, bukan sekadar narasi.

## Decision

**Boundary A-01 pindah dari hooks ke tool-gate orkestrator** (`guard.go`),
dipanggil sebelum setiap `tool_use` dieksekusi.

1. Guard sequence identik dengan `check-path` hari ini (allowlist role →
   resolve path → `pathmatch.IsForbidden` → `IsAllowed`; bash blocklist
   `block-dangerous-command.sh` diport ke Go; env di-scrub; cwd dipaksa
   worktree). Detail: §76.1 doc v1.0.0.
2. `isInside`/`resolveEventual` di-export ke `internal/pathmatch` — satu
   implementasi untuk `check-path` dan guard (no duplikasi).
3. Hook lama **dipertahankan** untuk sesi Claude Code warisan (jembatan migrasi,
   ADR-010). Untuk sesi messages, guard menggantikan posisi hook sebagai primary.
4. Net kedua tidak berubah: CI `validate-changed-paths` + branch protection.
5. Pengujian konkrit wajib (bukan narasi): fake model `httptest.Server` +
   tabel uji negatif/positif (§76.2), integrasi `tests/orchestrator/path-guard.test.sh`
   + target Makefile `verify-orchestrator`.

## Rationale

Batas keamanan yang "diceritakan" tidak dapat diverifikasi. Guard sebagai
interposer di tiap tool_use — termasuk bash — menutup celah R-08 (blocklist
bash dapat dilewati saat bash bebas). Ini memperkuat, bukan sekadar
menerjemahkan, enforcement v0.1.0.

## Consequences

- Guard jadi primary boundary untuk sesi messages; batas tidak lagi ganda
  (hook + guard) untuk sesi itu. Net kedua (CI) tetap ada — konsisten §43.
- Pelanggaran berulang (default N=3) menghentikan loop → status `failed` +
  handoff `failure` kategori `path-violation`.
- Test regresi `make verify` harus tetap hijau: guard tidak mengubah
  `check-path`, hooks, atau CI.

## Backward compatibility / rollback

- Hook lama tetap aktif untuk engine `claude-code`; rollback = kembalikan
  seluruh role ke engine itu tanpa perubahan kontrak.
- Test guard dapat dibatalkan tanpa efek runtime, tapi melemahkan kepastian
  A-01 pada harness baru.

## Menimpa

- Asumsi v0.1.0 bahwa A-01 dijamin hooks; status "sebagian" A-09/R-08 untuk
  celah bash menjadi tertutup pada engine messages.

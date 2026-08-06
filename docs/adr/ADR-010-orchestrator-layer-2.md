# ADR-010: Orkestrator Lapis 2 — Manual Messages API (harness C)

**Tanggal:** 2026-08-06
**Decider:** technical-lead-system-analyst + project-manager
**Status:** draft

## Context

m2s-vsh beralih dari harness Claude Code native menjadi CLI dua lapisan yang
mengorkestrasi agent AI sendiri, tanpa bergantung pada Claude Code. Lapis 2
(`m2s run --task <id>`) butuh engine yang menjalankan loop agent. Tiga
kandidat dievaluasi: A (Managed Agents), B (Claude Agent SDK), C (Manual
Messages API). Keputusan ini menetapkan pilihan dan mencatat penolakan.

## Decision

**Harness C — Manual Messages API** (`client.messages.*` + loop tool-use
sendiri, dibangun di atas `github.com/anthropics/anthropic-sdk-go`) adalah
engine orkestrator.

1. A (Managed Agents) **ditolak** sebagai engine.
2. B (Claude Agent SDK) **ditolak** sebagai engine utama; dipertahankan hanya
   sebagai jembatan migrasi opsional (`engine: claude-code` di config).
3. Loop, guard, tools, prompt assembly diimplementasikan di
   `internal/orchestrator/` (§74.2 doc v1.0.0).
4. Engine dipilih lewat config `control/orchestrator.yaml` (V-5), bukan field
   schema task.

## Rationale

### 1. Syarat keras "tanpa bergantung pada Claude Code"

A menggantikan Claude Code dengan platform Anthropic-hosted — tetap di luar
kendali kita. B adalah Claude Code itu sendiri sebagai library. Hanya C yang
memenuhi kemandirian penuh.

### 2. Proxy 9Router

Seluruh pilot berjalan lewat `ANTHROPIC_BASE_URL` (routing model lokal).
A menghilangkan jalur ini — sandbox Anthropic-hosted tidak menghormati proxy.
B dan C menghormatinya; C menambah kontrol base_url di harness.

### 3. Invariant A-01

A memindahkan sandbox ke container Anthropic dan menutup jalur interposisi
write/edit — melemahkan enforcement. C memperkuat: guard di tiap tool_use
termasuk bash, menutup celah R-08 (blocklist bash).

### 4. Biaya

C menuntut implementasi loop tool-use — itu justru deliverable Lapis 2. B
lebih murah tapi menolak tujuan; A lebih mahal (rewrite seluruh model
sandbox + enforcement) sekaligus lemah pada dua syarat keras.

## Consequences

- Dependency baru: `github.com/anthropics/anthropic-sdk-go`.
- Port logika hooks (danger-command, secret-path, path-scope) ke guard Go.
- Risiko kompatibilitas proxy tool-use → V-11, diverifikasi di F2.
- Role yang belum diport tetap `engine: claude-code` selama F5 (jembatan).

## Backward compatibility / rollback

- Runner Lapis 1 tidak berubah; rollback = kembalikan `engine` ke
  `claude-code` untuk seluruh role tanpa perubahan kontrak.
- Dokumen yang ditimpa: §13.4 v0.1.0 untuk bagian execution runtime;
  §73-§74 doc v1.0.0 menggantikannya.

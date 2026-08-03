# ADR D-P7-3: Merge Order — Sequential, bukan Parallel

**Tanggal:** 2026-08-03
**Decider:** TL/SA
**Status:** approved

## Context

BE-102 dan FE-102 berjalan paralel. Merge order menentukan kapan masing-masing
masuk develop dan bagaimana QA dilakukan.

## Decision

**Sequential merge: BE → FE → QA → Code Review → merge develop.**

Implementasi tetap paralel. Merge ke develop sequence.

## Rationale

1. **Avoid cascade conflicts** — BE dan FE bekerja independen; merge
   simultan tanpa koordinasi bisa bikin conflict di CI level.
2. **Contract verification upstream** — BE merger duluan, QA bisa verifikasi
   contract compliance sebelum FE masuk.
3. **Easier rollback** — bila BE perlu revert, FE tidak sudah merged.
4. **Single-threaded review bottleneck** — manusia (Code Reviewer + m2s-approver)
   lebih natural satu per satu.

## Consequences

- Parallel implementation, sequential merge.
- BE merge → TL/SA verify contract compliance → FE merge.
- QA-102 dijalankan setelah kedua PR exist.
- Total wall-clock: max(BE-time, FE-time) + sequential merge + QA.
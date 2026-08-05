# ADR-009: Enforcement untuk repo klien private + status item tertunda

**Tanggal:** 2026-08-05
**Decider:** owner arsitektur (Mindtoscreen)
**Status:** approved

## Context

Empat item tertunda dari v0.1.0 perlu keputusan final agar register status tidak
menggantung: enforcement untuk repo klien private (D-02), pembatasan hak
push/merge (D-03), lima contract Phase 7 yang invalid terhadap schema, dan merge
queue.

## Decision

1. **D-02 — enforcement klien private: diterima sementara tanpa enforcement.**
   Upgrade ke GitHub Team (~$4/seat × 14) ditangguhkan sampai project klien
   private pertama masuk. Migrasi ke GitLab **ditolak**.
2. **D-03 — pembatasan hak push/merge: tertutup** lewat migrasi org (ADR-008)
   + ruleset `agent-push-restriction`/`agent-worker-restriction`. Alur
   worker→PR→approver→merge teruji (§66 #9).
3. **Contract Phase 7 yang invalid: dipertahankan sebagai arsip**, tidak
   diperbaiki. Mengubah `role`/`status`/`base_branch` akan memalsukan riwayat.
4. **Merge queue: ditunda** dari v0.1.0. Tidak ada trigger nyata (single
   project, task jarang bentrok); gate anti-overlap #7/#8 sudah ditegakkan
   lewat ruleset + review. Re-evaluasi bila task paralel sering bentrok atau
   klien private butuh enforcement penuh.

## Rationale

### 1. D-02 — GitHub Team vs GitLab

| Opsi | Biaya | Protected branch private | Approval wajib | Merge queue |
|---|---|---|---|---|
| GitHub Free (sekarang) | 0 | ❌ | ❌ | ❌ |
| GitHub Team | ~$4/seat × 14 | ✅ | ✅ | ✅ |
| GitLab Free | 0 | ✅ | ❌ (Premium) | ❌ (Premium) |
| GitLab Premium | ~$19/seat | ✅ | ✅ | ✅ |

GitLab Free menutup gap protected-branch-private yang GitHub Free tidak punya,
tetapi tidak menutup approval wajib dan merge queue — keduanya Premium.
GitLab Premium lebih mahal dari GitHub Team **dan** menuntut port seluruh
enforcement stack (GitHub Actions, App dua-identitas, ruleset, runner m2s) yang
ditulis untuk GitHub. GitLab hanya layak bila klien menuntut self-host.

### 2. D-03

Migrasi org (ADR-008) sudah menutup ini. Ruleset membatasi identitas App:
worker tak bisa merge, approver yang merge. Tidak ada kerja tersisa selain
pembaruan status register.

### 3. Contract Phase 7 invalid

Kelima file ditulis sebelum schema distandarkan dan melanggar
`task.schema.json` (`role: backend-developer` vs `backend-engineer`,
`status: completed/approved` vs nilai `taskStatus`, `base_branch: main` ditolak
runner ADR-001 #2). Dipindah ke `control/tasks/archive/` sebagai catatan
sejarah (PR #30). Perbaikan akan memalsukan riwayat Phase 7; penghapusan
menghilangkan jejak audit.

### 4. Merge queue

Ditunda dari v0.1.0 (Q16 + ADR-007 #3). Tidak ada trigger §65 yang nyata;
gate anti-overlap #7/#8 tertegakkan lewat ruleset + required review. Merge
queue hanya menambah biaya (plan berbayar) tanpa nilai sekarang.

## Consequences

- **D-02**: klien private pertama menuntut keputusan upgrade Team + persetujuan
  owner org (peran arsitektur adalah `member`). Risiko enforcement soft-rule
  pada klien private diterima eksplisit.
- **D-03**: register status akurat; tidak ada lagi item "menunggu".
- **Contract Phase 7**: arsip tetap ada sbg catatan; spec aktif memakai pola
  valid `schemas/examples/task-BE-101.valid.yaml`.
- **Merge queue**: ditutup sbg keputusan, bukan sisa pekerjaan; dibuka lagi
  bila trigger muncul.

## Backward compatibility / rollback

- D-02: tetap di Free = status quo; upgrade = keputusan terpisah.
- D-03: sudah tereksekusi; rollback = batalkan migrasi org (tidak disarankan).
- Contract Phase 7: arsip tidak mengubah perilaku apa pun.
- Merge queue: tidak ada perubahan runtime.

**Menimpa:** status "terbuka/menunggu" pada D-02 dan D-03 di
`docs/decisions/open-questions.md`; pembacaan ADR-007 #8 "Done tercapai
sebagian" menjadi penuh utk enforcement (sisa hanya merge queue, ditunda).

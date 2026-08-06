# ADR-011: Bump Versi MAJOR ke 1.0.0 — label v0.1.1 ditolak

**Tanggal:** 2026-08-06
**Decider:** technical-lead-system-analyst + project-manager
**Status:** draft

## Context

Permintaan awal menulis versi baru sebagai **v0.1.1**. §69 semver menyatakan
perubahan orchestrator / execution runtime / responsibility boundary = MAJOR,
dan §65 mendaftar "container isolation" + "automated observability" sebagai
trigger kenaikan versi. Label v0.1.1 (PATCH) harus diverifikasi terhadap aturan
itu — tidak boleh diasumsikan.

## Decision

**Versi = `1.0.0` (MAJOR). Label `v0.1.1` DITOLAK.**

1. Perubahan memenuhi definisi MAJOR §69 dua kali: **orchestrator** baru (Lapis
   2) dan **execution runtime** berubah (Claude Code native → loop Messages API
   yang dijalankan binary m2s).
2. `v0.1.1` adalah kelas "hook validation fixes" (contoh §69) — PATCH untuk
   perubahan orchestration menipu semver.
3. **Bukan 0.2.0**: trigger §65 (Rulix, plugin marketplace, database, durable
   queue, multi-project bottleneck, container isolation, automated
   observability) **tidak terpicu**. Kenaikan ke 0.2.0 adalah kelas
   Rulix/plugin distribution.
4. Tidak menunggu "production-proven": §69 `1.0.0 Production-proven` adalah
   deskripsi state, bukan gate semver. Production-proofing = fase hardening (F6).
5. **Container isolation dan automated observability (§65) tidak diadopsi di
   v1.0.0** — keduanya tetap kandidat v1.1.0, tercatat sebagai not-trigger,
   bukan diabaikan.

## Rationale

Semver §69 mengklasifikasi perubahan berdasarkan *kelas perubahan*, bukan
besarnya diff. Menambah lapisan orkestrator + mengganti runtime eksekusi adalah
MAJOR tanpa ambiguitas. Satu-satunya pembelaan label v0.1.1 adalah "perubahan
ini masih kecil karena pilot" — itu argumen produk, bukan semver; semver sudah
memutuskan.

## Consequences

- `README.md` badge judul: `v0.1.0` → `1.0.0`.
- Dokumen arsitektur baru: `M2S-VSH-Lite-v1.0.0-Architecture.md`.
- Entri "Keputusan Versi 1.0.0" di `docs/decisions/open-questions.md`.
- Dokumen v0.1.0 tetap ada sebagai riwayat; §72 doc v1.0.0 mencatat override.

## Backward compatibility / rollback

- Versi hanya label + dokumen; tidak ada perubahan runtime dari keputusan ini
  sendiri. Rollback = ubah label dan arsipkan doc v1.0.0.

## Menimpa

- Pembacaan v0.1.1 sebagai label yang valid untuk perubahan ini; status §69
  untuk perubahan orchestrator/runtime.

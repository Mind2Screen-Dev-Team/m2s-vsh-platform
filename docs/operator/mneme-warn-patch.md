# Patch human-only — Mneme: ADR-009 entry + keputusan pertahankan warn

`.mneme/project_memory.json` ditolak `deny` di `.claude/settings.json`
(component-inventory §7): engineering dilarang mengedit, TL/SA yang punya
(§6.4, §27). Karena itu perubahan berikut **dijalankan manusia** — disalin ke
`.mneme/project_memory.json`.

## Keputusan: pertahankan `mode: warn` (5 Agustus 2026)

Mneme **tidak di-flip ke strict**. Alasan (§6.6 roadmap Mneme):

1. **Data warn belum cukup.** Fase warn berjalan singkat; belum ada catatan
   false positive yang diteliti. §6.6: "Perbaiki decision scope dan false
   positive" sebelum strict.
2. **Strict global = blokir edit** yang melanggar forbidden. Tanpa bukti plugin
   meminimalkan false positive, risiko kerja agent terhenti lebih besar dari
   nilainya. Dokumentasi arsitektur menempatkan strict sebagai tujuan akhir,
   bukan langkah pertama.
3. **Keputusan baru masih masuk.** ADR-009 baru diputuskan 2026-08-05. Memory
   harus lengkap dulu sebelum enforcement diperketat.

**Next:** evaluasi flip strict setelah satu siklus warn penuh tanpa false
positive, atau bila plugin memperlihatkan dukungan per-decision strict.

## Patch: tambah ADR-009 ke decisions[]

Tempel blok berikut ke array `decisions` di `.mneme/project_memory.json`
(sesuaikan koma/format JSON):

```json
,
    {
      "id": "ADR-009",
      "title": "Enforcement untuk repo klien private + status item tertunda",
      "status": "accepted",
      "date": "2026-08-05",
      "scope": ["client-private", "gitlab-evaluation", "merge-queue", "phase7-archive"],
      "summary": "D-02 diterima sementara tanpa enforcement utk klien private (upgrade GitHub Team bila klien masuk; GitLab ditolak — approval wajib + merge queue Premium). D-03 tertutup via migrasi org ADR-008. Contract Phase 7 invalid dipertahankan sbg arsip. Merge queue ditunda v0.1.0.",
      "forbidden": [
        "migrasi ke GitLab tanpa re-evaluasi klien private pertama",
        "perbaiki contract Phase 7 arsip (memalsukan riwayat)",
        "aktifkan merge queue tanpa trigger nyata"
      ]
    }
```

Verifikasi setelah menempel: `python3 -c "import json; json.load(open('.mneme/project_memory.json'))"`
(jangan error), lalu `make verify` hijau.

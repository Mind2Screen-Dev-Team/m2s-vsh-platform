# M2S-VSH Lite

Workflow pengembangan perangkat lunak berbasis **agen AI** yang terstruktur,
dapat diaudit, dan dikerjakan paralel.

## Apa itu M2S-VSH Lite?

M2S-VSH Lite adalah sebuah **cara kerja** (workflow) di mana beberapa agen AI
dengan peran berbeda membantu mengembangkan produk. Bukan satu bot yang
mengerjakan semuanya — melainkan **tim AI** dengan tugas dan batas yang jelas,
diatur lewat repositori pusat.

Tujuannya sederhana: membuktikan bahwa beberapa peran AI dapat bekerja bersama
secara terstruktur, paralel, dan tidak saling menimpa pekerjaan — sementara
keputusan penting tetap di tangan manusia.

## Mengapa ini dibuat?

Proyek ini lahir dari pertanyaan: *bisakah AI mengembangkan perangkat lunak
seperti tim sungguhan, bukan seperti "satu asisten yang mengerjakan semua"?*

Jawabannya diwujudkan sebagai workflow dengan tiga prinsip:

1. **Terstruktur** — setiap pekerjaan agen punya kontrak dan batas yang jelas.
2. **Paralel** — backend dan frontend dikerjakan bersamaan, tidak berurutan.
3. **Diaudit** — semua yang dikerjakan agen tercatat dan bisa diperiksa ulang.

## Tujuan

- Menjalankan pipeline pengembangan dari perencanaan sampai rilis dengan agen AI.
- Memisahkan peran dengan jelas: manajer proyek, analis teknis, desainer,
  engineer backend/frontend, penguji mutu (QA), penulis dokumentasi.
- Menguji mekanisme kerja paralel antar repo aplikasi.
- Menjaga **manusia sebagai penentu akhir** — agen tidak boleh mengambil
  keputusan yang tak bisa dibatalkan.

## Struktur project

Project ini terdiri dari satu repositori pengatur + satu atau lebih
repositori aplikasi.

| Repositori | Isi | Peran |
|---|---|---|
| `m2s-vsh-platform` | aturan, kontrak tugas, runner, template, dokumentasi | sumber pengatur |
| `m2s-vsh-project-backend` | server Go — endpoint status | repo aplikasi (backend) |
| `m2s-vsh-project-frontend` | aplikasi Next.js — dashboard status | repo aplikasi (frontend) |

> Repositori aplikasi tidak harus selalu backend+frontend. Arsitektur mendukung
> backend, frontend, mobile, fullstack, dan kombinasi lain.

## Bagaimana cara kerjanya?

**Alur singkat:**

1. **Kontrak dulu.** Analis teknis menulis kontrak tugas — apa yang dikerjakan,
   di repo mana, file apa yang boleh diubah.
2. **Kerjakan paralel.** Engineer backend dan frontend bekerja bersamaan,
   masing-masing di area terpisah.
3. **Diperiksa.** Setiap pekerjaan lewat pemeriksaan otomatis (path, kontrak)
   dan ulasan manusia.
4. **Manusia yang menggabungkan.** Perubahan tidak langsung masuk ke cabang
   utama (`main`) — harus lewat alur `agent` → `develop` → `staging` → `main`,
   dan penggabungan akhir dilakukan manusia.

**Prinsip kunci:**

- **Kontrak tugas** — setiap pekerjaan agen punya spesifikasi tertulis:
  repo, cabang, file yang boleh disentuh, kriteria selesai.
- **Eksekusi terisolasi** — tiap agen bekerja di ruang kerja terpisah;
  tidak ada dua agen menulis file yang sama.
- **Gate manusia** — persetujuan kontrak, persetujuan desain, dan penggabungan
  akhir selalu di tangan manusia.
- **Repositori aplikasi independen** — backend dan frontend terhubung lewat
  kontrak di repositori pengatur, bukan lewat kode yang saling merujuk.

## Status

Project ini masih dalam pengembangan. Dokumentasi dan panduan setup lengkap
ada di repositori pengatur (`docs/`).

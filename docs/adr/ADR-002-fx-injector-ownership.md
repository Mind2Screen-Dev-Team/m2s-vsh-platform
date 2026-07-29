# ADR-002 — Kepemilikan Dependency Injection Registry (Uber Fx)

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 29 Juli 2026 |
| **Pemutus** | Pemilik arsitektur (Mindtoscreen) |
| **Versi arsitektur** | 0.1.0 |
| **Berlaku pada** | `m2s-vsh-project-backend` (Go) |
| **Terkait** | §29.6, §67, R-03, R-04, R-19 |

---

## Konteks

Repository referensi struktur backend (`fajarcandraaa/thousand-sunny`) memakai
**Uber Fx** untuk dependency injection, dengan pendaftaran modul terpusat di
`app/injector/**` dan `internal/fx.modules`.

Konsekuensinya: **setiap modul domain baru wajib didaftarkan di file terpusat.**

Ini berbenturan langsung dengan acceptance test §67 yang menuntut dua task backend
berjalan paralel tanpa overlap:

```
BE-101 → internal/payroll/**      ┐
BE-102 → internal/attendance/**   ┘  keduanya harus mengedit app/injector/**
```

Dokumen arsitektur §29.6 sudah memperingatkan kategori masalah ini — ia menyebut
"route registry" sebagai contoh shared file yang wajib bersingle owner. Pada
codebase Go dengan Fx, **DI injector adalah wujud nyata dari peringatan tersebut**.

Tanpa penanganan eksplisit, dua task backend paralel akan:
- gagal reservasi karena keduanya membutuhkan path yang sama, atau
- lolos reservasi lalu bertabrakan saat merge

Keduanya menggagalkan acceptance §67.

**Keadaan yang menguntungkan:** repository backend masih **kosong** (`size: 0`).
Konvensi dapat ditetapkan sejak awal, bukan ditambal belakangan.

---

## Temuan pendukung dari repository referensi

| Temuan | Dampak |
|---|---|
| `internal/` diorganisasi per domain dengan konvensi penamaan file (handler, repo, service, fx module), bukan folder berlapis | Reservasi `internal/<domain>/**` bersih dan tidak beririsan antar domain |
| **Tidak ada direktori `tests/`** — unit test colocated `_test.go` | Reservasi domain otomatis mencakup test-nya. Path `tests/unit/**` di §20.6 **tidak berlaku** untuk repo ini |
| `gen/` berisi kode hasil SQLC dan GORM Gen | `gen/**` wajib masuk `forbidden_paths` — §16.1 melarang agent mengedit generated file |
| `pkg/` memakai konvensi prefiks `x` (`xlog`, `xresp`, …) | Paket bersama; bukan milik task domain |
| Build lewat `Makefile` | Quality gate memanggil target `make`, bukan perintah `go` mentah |

---

## Keputusan

**1. Setiap domain memiliki file modul Fx sendiri.**

```
internal/<domain>/fx.go        →  var Module = fx.Module("<domain>", …)
```

File ini **dimiliki task domain** dan masuk `allowed_paths` task tersebut.
Dengan demikian mayoritas pekerjaan wiring tidak lagi menyentuh file terpusat.

**2. File agregasi terpusat menjadi shared file dengan owner tunggal.**

```
app/injector/modules.go        →  hanya mengimpor dan merangkai <domain>.Module
```

File ini masuk `forbidden_paths` **semua** task implementasi domain.

**3. Pendaftaran modul baru dilakukan lewat task `WIRE-*` tersendiri**, yang
dijalankan **setelah** task domain terkait selesai.

```
BE-101 (payroll)     ┐
                     ├─→  WIRE-101  →  daftarkan kedua modul  →  PR
BE-102 (attendance)  ┘
```

`WIRE-*` adalah task kecil, berumur pendek, dan menyentuh satu file. Ia memegang
reservasi eksklusif atas `app/injector/**`.

**4. `gen/**` masuk `forbidden_paths` default.** Regenerasi hanya lewat
`make sqlc-gen` / `make gorm-gen`, dan hanya dalam task yang secara eksplisit
memilikinya.

**5. Path reservasi backend mengikuti pola:**

```yaml
allowed_paths:
  - internal/<domain>/**          # termasuk _test.go colocated
forbidden_paths:
  - app/injector/**
  - internal/fx.modules
  - gen/**
  - database/migrations/**
  - go.mod
  - go.sum
  - constant/**
  - config/**
  - pkg/**
```

---

## Rasional

### Mengapa bukan "injector di forbidden, TL/SA yang mendaftarkan"

Menjadikan TL/SA sebagai pendaftar tunggal memang sederhana, tetapi menciptakan
bottleneck pada setiap penambahan modul dan mencampur peran: §18.5 melarang TL/SA
melakukan routine implementation. Mendaftarkan modul adalah pekerjaan implementasi.

### Mengapa bukan refactor penuh agar tanpa file terpusat

Pola registrasi otomatis lewat `init()` menghilangkan file terpusat, tetapi
mengorbankan keterbacaan dan urutan inisialisasi yang eksplisit — dua hal yang
justru menjadi alasan memakai Fx. Ini melanggar ladder Ponytail yang diadopsi §5.2:
pilih implementasi paling sederhana, jangan menambah abstraksi yang belum dibutuhkan.

### Mengapa `WIRE-*` sebagai task terpisah

Ia mengikuti §29.6 secara harfiah: shared file memiliki owner tunggal pada satu
waktu. Serialisasi hanya terjadi pada satu file kecil, sementara pekerjaan domain —
bagian yang mahal — tetap sepenuhnya paralel.

Biayanya adalah satu task tambahan per gelombang, bukan per task domain.

---

## Konsekuensi

### Positif

- Acceptance §67 menjadi dapat dipenuhi: BE-101 dan BE-102 benar-benar paralel
- Konflik merge pada injector hilang secara struktural, bukan diserahkan pada keberuntungan
- Reservasi path menjadi bersih dan mudah divalidasi

### Negatif

- Menambah satu task (`WIRE-*`) per gelombang implementasi
- Modul domain belum aktif sampai `WIRE-*` selesai — integration test harus
  dijadwalkan setelahnya
- Konvensi `internal/<domain>/fx.go` harus ditegakkan sejak file pertama;
  bila terlewat, hazard-nya kembali

### Yang harus divalidasi di Phase 8

Acceptance §67 diperluas dengan pemeriksaan tambahan:

| Kode | Uji |
|---|---|
| AC-8.8 | BE-101 dan BE-102 **tidak** menyentuh `app/injector/**` |
| AC-8.9 | Reservasi menolak task ketiga yang meminta `app/injector/**` selagi `WIRE-101` aktif |
| AC-8.10 | Tidak ada task domain yang menyentuh `gen/**` |

---

## Catatan penerapan bila stack berbeda

Keputusan ini spesifik untuk backend Go dengan Uber Fx. Q4 menetapkan stack dapat
berbeda per project. Pola yang **berlaku umum** adalah:

> Setiap codebase memiliki satu atau lebih *registry file* yang wajib disentuh
> setiap penambahan unit baru. File semacam itu harus diidentifikasi **sebelum**
> task pertama dijalankan, dimasukkan ke `forbidden_paths`, dan ditangani task
> registrasi tersendiri.

Padanan pada stack lain:

| Stack | Registry file yang setara |
|---|---|
| Next.js | route manifest, `middleware.ts`, barrel `index.ts`, provider root |
| NestJS | root `AppModule` |
| Laravel | service provider registry, route file |
| Django | `INSTALLED_APPS`, root `urls.py` |

Identifikasi registry file menjadi bagian wajib dari Definition of Ready TL/SA (§18.8).

---

## Rollback

Bila pola `WIRE-*` terbukti terlalu memberatkan:

1. Gabungkan pendaftaran ke dalam task domain, dan serialisasikan task backend
   (kehilangan paralelisme, acceptance §67 tidak terpenuhi), **atau**
2. Pindah ke registrasi otomatis berbasis `init()` sebagai dedicated refactor task

Keduanya memerlukan ADR pengganti.

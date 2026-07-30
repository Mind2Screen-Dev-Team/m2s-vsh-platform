# Matriks Uji Overlap Path

Spesifikasi normatif untuk deteksi konflik reservasi path. Menutup **R-03**;
acceptance **AC-2.3…AC-2.6**.

| | |
|---|---|
| **Sumber aturan** | §30 § Reservation rules, §29.6 |
| **Implementasi** | `cmd/m2s/**` (Go) — ADR-004 #2 |
| **Bentuk uji** | table-driven test, dijalankan CI |
| **Dipakai oleh** | `reserve-paths`, `validate-task`, PreToolUse hook Phase 3 (§59) |

---

## 1. Semantik glob yang ditetapkan

Dokumen arsitektur memakai `**` tanpa mendefinisikannya. Definisi berikut mengikat
implementasi.

| Pola | Arti | Contoh cocok | Contoh tidak cocok |
|---|---|---|---|
| `a/b.go` | exact file | `a/b.go` | `a/b.go.bak`, `a/c/b.go` |
| `a/**` | seluruh isi `a/` secara rekursif, termasuk `a/` sendiri | `a/x.go`, `a/b/c/d.go` | `ab/x.go`, `a.go` |
| `a/*` | satu segmen langsung di bawah `a/` | `a/x.go`, `a/b` | `a/b/c.go` |
| `*.go` | satu segmen berakhiran `.go` pada akar | `main.go` | `a/main.go` |

**Aturan tambahan:**

- Pemisah selalu `/`, tidak pernah `\`
- Path relatif terhadap akar repository; absolut dan segmen `..` ditolak schema
- `*` tidak melewati pemisah `/`; hanya `**` yang melewatinya
- Pencocokan **case-insensitive** — lihat §3

---

## 2. Definisi konflik

Dua himpunan path konflik bila **irisan berkas yang mungkin dicakup keduanya tidak
kosong**. Bukan kesamaan string.

Konsekuensi langsung dari §30:

> - Glob overlap dianggap conflict.
> - Parent path conflict dengan child path.
> - Exact shared file hanya boleh satu active writer.

Perhatikan bahwa `internal/payroll/**` dan `internal/payroll/period/**` **konflik**
meski kedua string berbeda — inilah kasus yang R-03 nyatakan akan diloloskan
implementasi naif berbasis kesamaan string.

---

## 3. Case-sensitivity

**Keputusan: pencocokan case-insensitive.**

Alasan: macOS (APFS) dan Windows default case-insensitive, sedangkan Linux
case-sensitive. Bila reservasi `Internal/Payroll/**` dan `internal/payroll/**`
dianggap tidak konflik, dua agent memperoleh izin menulis berkas yang **sama** pada
mesin operator macOS — kegagalan yang justru ingin dicegah.

Memperlakukan keduanya sebagai konflik bersifat konservatif: pada Linux ia menolak
pasangan yang sesungguhnya aman. Penolakan palsu berbiaya satu task tertunda;
peloloskan palsu berbiaya dua writer pada satu berkas.

---

## 4. Matriks uji

R-03 menuntut **minimal 12 kasus**. Berikut 24, dikelompokkan per kategori. Kolom
**Konflik** adalah hasil yang wajib dihasilkan implementasi.

### 4.1 Identik dan disjoint

| # | A | B | Konflik | Alasan |
|---|---|---|---|---|
| 1 | `internal/payroll/**` | `internal/payroll/**` | **ya** | identik |
| 2 | `internal/payroll/**` | `internal/attendance/**` | tidak | subtree terpisah |
| 3 | `go.mod` | `go.sum` | tidak | berkas berbeda |

### 4.2 Parent/child — kasus inti R-03

| # | A | B | Konflik | Alasan |
|---|---|---|---|---|
| 4 | `internal/payroll/**` | `internal/payroll/period/**` | **ya** | B subtree dari A |
| 5 | `internal/**` | `internal/payroll/period/close.go` | **ya** | A mencakup B |
| 6 | `internal/payroll/**` | `internal/payroll` | **ya** | A mencakup direktori itu sendiri |
| 7 | `**` | `go.mod` | **ya** | A mencakup seluruh repo |

### 4.3 Prefiks yang menyesatkan

| # | A | B | Konflik | Alasan |
|---|---|---|---|---|
| 8 | `internal/pay/**` | `internal/payroll/**` | tidak | `pay` bukan segmen induk `payroll` |
| 9 | `internal/payroll/**` | `internal/payroll-legacy/**` | tidak | segmen berbeda meski berprefiks sama |
| 10 | `a/b` | `a/bc` | tidak | exact berbeda |

Kategori ini menangkap kesalahan implementasi `strings.HasPrefix` tanpa memeriksa
batas segmen.

### 4.4 Glob vs exact file

| # | A | B | Konflik | Alasan |
|---|---|---|---|---|
| 11 | `internal/payroll/**` | `internal/payroll/enum.go` | **ya** | exact di dalam glob |
| 12 | `internal/payroll/*` | `internal/payroll/enum.go` | **ya** | satu segmen, cocok |
| 13 | `internal/payroll/*` | `internal/payroll/period/close.go` | tidak | `*` tidak melewati `/` |
| 14 | `*.go` | `main.go` | **ya** | keduanya pada akar |
| 15 | `*.go` | `internal/main.go` | tidak | `*` tidak melewati `/` |

### 4.5 Case-sensitivity — §3

| # | A | B | Konflik | Alasan |
|---|---|---|---|---|
| 16 | `internal/payroll/**` | `Internal/Payroll/**` | **ya** | case-insensitive |
| 17 | `go.mod` | `GO.MOD` | **ya** | berkas sama pada APFS/NTFS |
| 18 | `internal/payroll/**` | `INTERNAL/attendance/**` | tidak | segmen kedua tetap berbeda |

### 4.6 Shared file dengan owner — R-04

| # A | B | Konflik | Alasan |
|---|---|---|---|
| 19 | `internal/payroll/enum.go` milik BE-101 | `internal/payroll/enum.go` milik BE-102 | **ya** | §29.6 single owner |
| 20 | `internal/payroll/**` (BE-101, owner `enum.go`) | `internal/payroll/enum.go` (BE-102) | **ya** | owner ≠ pengklaim |

### 4.7 Normalisasi

| # | A | B | Konflik | Alasan |
|---|---|---|---|---|
| 21 | `internal/payroll/` | `internal/payroll` | **ya** | trailing slash tidak bermakna |
| 22 | `./internal/payroll/**` | `internal/payroll/**` | **ya** | `./` dinormalisasi |
| 23 | `internal//payroll/**` | `internal/payroll/**` | **ya** | pemisah ganda dinormalisasi |

### 4.8 Forbidden mengalahkan allowed

| # | Keadaan | Hasil | Alasan |
|---|---|---|---|
| 24 | `allowed: internal/**`, `forbidden: internal/auth/**` | tulis ke `internal/auth/x.go` **ditolak** | forbidden diperiksa lebih dulu |

---

## 5. Yang TIDAK dijamin matriks ini

Batas yang disengaja, agar tidak dibaca sebagai jaminan menyeluruh:

| Tidak tertangkap | Sebab | Penanganan |
|---|---|---|
| Semantic overlap | dua task pada berkas berbeda dapat mengklaim `BR` yang sama | §31 — requirement/business-rule ID, review TL/SA |
| Symlink keluar subtree | pencocokan bersifat leksikal, bukan resolusi filesystem | PreToolUse hook Phase 3 (§59) memeriksa path hasil resolusi |
| Unicode normalisasi (NFC/NFD) | `é` dapat direpresentasikan dua cara; APFS menormalisasi, ext4 tidak | belum ditangani — dicatat sebagai batas |
| Konflik urutan merge | dua task tanpa overlap path tetap dapat saling merusak secara semantik | §29.8 merge queue |

---

## 6. Status verifikasi

Prototipe algoritma pencocokan dijalankan terhadap matriks ini pada 30 Juli 2026,
sebelum `reservation.schema.json` ditulis, untuk memastikan tidak ada kasus yang
saling bertentangan atau mustahil dipenuhi.

| Kategori | Kasus | Hasil |
|---|---|---|
| §4.1 identik & disjoint | 1–3 | ✅ |
| §4.2 parent/child | 4–7 | ✅ |
| §4.3 prefiks menyesatkan | 8–10 | ✅ |
| §4.4 glob vs exact | 11–15 | ✅ |
| §4.5 case-sensitivity | 16–18 | ✅ |
| §4.7 normalisasi | 21–23 | ✅ |
| §4.6 shared file owner | 19–20 | ⏳ butuh konteks reservasi |
| §4.8 forbidden > allowed | 24 | ⏳ butuh konteks reservasi |

21 dari 24 kasus terbukti dapat dipenuhi satu algoritma, dan seluruhnya **simetris**
— `overlap(a,b)` selalu sama dengan `overlap(b,a)`.

Tiga kasus tersisa memerlukan dua dokumen reservasi atau satu task contract sebagai
input, bukan sepasang pola. Diverifikasi bersama implementasi `reserve-paths`.

**Prototipe bukan implementasi.** Ia membuktikan matriks konsisten, bukan bahwa
`cmd/m2s` sudah benar. Implementasi wajib memuat ulang seluruh 24 kasus sebagai
table-driven test.

---

## 7. Kaitan acceptance

| Kode | Terpenuhi oleh |
|---|---|
| AC-2.3 | §4.2 — parent/child |
| AC-2.4 | §4.4 — glob vs exact file |
| AC-2.5 | §4.5 — case-sensitivity |
| AC-2.6 | §4.6 — shared file single owner |

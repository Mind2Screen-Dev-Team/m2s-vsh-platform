# ADR-004 — Format Task Contract dan Bahasa Implementasi Runner

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 30 Juli 2026 |
| **Pemutus** | Pemilik arsitektur (Mindtoscreen) |
| **Versi arsitektur** | 0.1.0 |
| **Menimpa** | §30 (lokasi worktree, pelepasan reservasi), §34 (field kurang) |
| **Terkait** | §36, §58, Q8, Q11, Q12, Q13, Q15, R-03, R-04, ADR-003 |

---

## Konteks

Phase 1 (§58, setelah pertukaran ADR-003) menghasilkan tiga schema dan lima script.
Sebelum file pertama ditulis, empat hal harus ditetapkan karena menentukan bentuk
seluruh fase.

### 1. Format kontrak belum tunggal

Dokumen memberi tiga sinyal yang tidak sejalan:

| Sumber | Sinyal |
|---|---|
| §58 | "Buat task JSON/YAML schema" — keduanya disebut |
| §34, §35, §30 | seluruh contoh ditulis **YAML** |
| §36 | nama file `task.schema.json`, `handoff.schema.json`, `reservation.schema.json` |
| Q15 | runner menyuntikkan `<worktree>/.task/contract.json` |

### 2. Bahasa implementasi script belum ditetapkan

§36 menamai artefak `.sh`. Namun deteksi glob overlap yang dituntut R-03 —
parent/child, glob vs exact file, case-sensitivity, minimal 12 kasus uji — adalah
logika non-trivial. Kesalahan pada bagian ini meloloskan dua writer ke file yang
sama, yaitu kegagalan yang justru ingin dicegah arsitektur ini.

### 3. §30 memuat teks yang sudah ditimpa keputusan Phase 0

| Aspek | §30 / §34 | Keputusan Phase 0 |
|---|---|---|
| Lokasi worktree | `.claude/worktrees/BE-101` | `$HOME/.m2s/worktrees/<repository>/<task-id>` (Q8) |
| Pelepasan reservasi | "dilepas setelah PR dibuat" | dilepas saat **merge**; status antara `reserved-pending-merge` (Q12) |
| `shared_file_ownership` | tidak tersedia | wajib ada (R-04) |

Schema tidak dapat ditulis tanpa memilih salah satu. Menyalin §30 apa adanya akan
menghasilkan artefak yang bertentangan dengan keputusan yang sudah disetujui.

### 4. Q11 mengikat bentuk permukaan runner

Q11 membatasi tool `Bash` Project Manager ke pola persis `scripts/<runner>.sh`.
Binary Go tidak memenuhi pola tersebut.

---

## Keputusan

**1. JSON Schema sebagai validator, YAML sebagai format tulis, JSON sebagai format
transport.**

| Lapisan | Format | Lokasi |
|---|---|---|
| Definisi validasi | JSON Schema (draft 2020-12) | `schemas/*.schema.json` |
| Kontrak yang ditulis manusia/agent | YAML | `control/tasks/specifications/*.yaml` |
| Snapshot yang dibaca hook | JSON | `<worktree>/.task/contract.json` |

Runner melakukan materialisasi YAML → JSON saat menyuntikkan contract ke worktree,
sesuai Q15. YAML tidak pernah masuk worktree.

**2. Script runner diimplementasikan dalam Go, dibungkus wrapper `.sh`.**

```
scripts/validate-task.sh   → memanggil bin/m2s validate-task
scripts/reserve-paths.sh   → memanggil bin/m2s reserve-paths
scripts/launch-task.sh     → memanggil bin/m2s launch-task
scripts/collect-result.sh  → memanggil bin/m2s collect-result
scripts/release-reservation.sh → memanggil bin/m2s release-reservation
```

Wrapper `.sh` dipertahankan agar pola `scripts/<runner>.sh` pada Q11 tetap berlaku
tanpa perubahan, dan agar §36 tetap terpenuhi secara harfiah. Wrapper tidak memuat
logika — hanya meneruskan argumen dan exit code.

**3. Keputusan Phase 0 menang atas §30 dan §34.**

| Aspek | Yang berlaku |
|---|---|
| Lokasi worktree | `$HOME/.m2s/worktrees/<repository>/<task-id>`, override `M2S_WORKTREE_ROOT` (Q8) |
| Status reservasi | `active` → `reserved-pending-merge` → `released` (Q12) |
| Pelepasan | saat merge, bukan saat PR dibuat (Q12) |
| Field tambahan | `shared_file_ownership` pada `task.schema.json` (R-04) |

Contoh YAML pada §30 dan §34 diperlakukan sebagai **ilustrasi**, bukan definisi
normatif. Definisi normatif adalah file `schemas/*.schema.json`.

**4. Schema diberi `schema_version` dan diperlakukan sebagai breaking-change
surface.** Sesuai `component-inventory.md` § Schemas: perubahan schema membutuhkan
version bump. `task.schema.json` dimulai pada `1.0`, mengikuti `schema_version: 1.0`
pada §34.

---

## Rasional

### Mengapa YAML untuk penulisan, JSON untuk transport

Kontrak dibaca dan ditinjau manusia pada tahap review task — YAML mendukung komentar
dan lebih ringkas. Sebaliknya `.task/contract.json` dibaca hook pada setiap
PreToolUse; JSON menghilangkan kebutuhan parser YAML di jalur panas dan menghapus
ambiguitas tipe YAML (`yes`/`no`, versi tanpa kutip, indentasi).

Pemisahan ini juga menjaga sifat read-only snapshot: file yang dibaca hook adalah
hasil materialisasi runner, bukan file yang sama yang diedit manusia.

### Mengapa Go, bukan Bash

Alasan utama adalah R-03. Deteksi overlap harus menangani:

```
internal/payroll/**        vs  internal/payroll/period/**   → conflict (parent/child)
internal/payroll/**        vs  internal/attendance/**       → tidak conflict
internal/payroll/config.go vs  internal/payroll/**          → conflict (exact vs glob)
Internal/Payroll/**        vs  internal/payroll/**          → conflict (case-insensitive FS)
```

Logika ini menuntut unit test. Go menyediakannya secara native, dan matriks 12 kasus
R-03 menjadi tabel uji yang dijalankan CI — bukan pemeriksaan manual.

Alasan pendukung: Go sudah menjadi stack backend default (Q4) dan toolchain tersedia
pada mesin operator (go1.26.4). Tidak ada dependency runtime baru; binary
terkompilasi menghilangkan ketergantungan pada versi `bash`, `jq`, dan `grep` yang
berbeda antar mesin.

### Mengapa wrapper `.sh` tetap ada

Q11 mengunci `Bash` PM ke pola `scripts/<runner>.sh`. Mengganti pola itu berarti
membuka kembali keputusan yang sudah ditutup dan melonggarkan satu-satunya batasan
tool PM yang bersifat teknis. Wrapper tipis jauh lebih murah daripada mengubah Q11.

Wrapper juga menjaga §36 tetap benar secara harfiah: `scripts/` berisi file `.sh`
dengan nama persis seperti yang didaftarkan dokumen arsitektur.

### Mengapa §30 tidak diikuti apa adanya

Q8 dan Q12 dibuat setelah §30, melalui checkpoint yang disetujui pemilik arsitektur,
dan masing-masing menutup celah yang dinyatakan eksplisit: A-01 (worktree di dalam
path terlarang) dan A-05 (celah reservasi antara PR dan merge). Mengembalikan teks
§30 berarti membuka kembali kedua celah tersebut.

---

## Konsekuensi

### Positif

- Deteksi overlap dapat diuji otomatis; R-03 memperoleh jalur penutupan yang nyata
- Hook membaca JSON tanpa parser tambahan
- Q11 tetap utuh tanpa perubahan
- Schema menjadi satu-satunya definisi normatif, menghapus tiga sumber yang
  bertentangan

### Negatif dan biaya

- Menambah **build step**. Binary `bin/m2s` harus dikompilasi sebelum runner dapat
  dipakai, dan harus tersedia di mesin operator maupun CI
- Menambah **satu lapisan tak langsung**: `scripts/*.sh` → `bin/m2s`. Pembacaan
  sepintas pada `scripts/` tidak lagi memperlihatkan logika sebenarnya
- **Dua representasi kontrak** (YAML dan JSON) berarti dua kemungkinan sumber
  kesalahan bila materialisasi tidak setia. Dimitigasi: materialisasi divalidasi
  ulang terhadap JSON Schema yang sama setelah konversi
- `bin/**` perlu masuk `.gitignore` atau dirilis sebagai artefak — belum diputuskan,
  lihat bagian terbuka
- Contoh pada §30 dan §34 kini berbeda dari implementasi. Pembaca dokumen arsitektur
  saja akan memperoleh gambaran yang usang

### Yang belum diputuskan

| Pertanyaan | Kapan |
|---|---|
| `bin/m2s` di-commit, di-build lokal, atau dirilis sebagai artefak? | sebelum script pertama ditulis |
| Apakah CI memverifikasi wrapper `.sh` benar-benar tipis? | Phase 4 (§60) |

---

## Alternatif yang ditolak

### Bash penuh dengan `jq`

**Ditolak.** Memenuhi §36 secara harfiah dan tanpa build step, tetapi menempatkan
logika deteksi overlap — bagian yang kegagalannya paling mahal — pada bahasa tanpa
kerangka uji bawaan dan dengan perilaku glob yang bergantung pada shell dan
filesystem.

### Go tanpa wrapper `.sh`

**Ditolak.** Memaksa perubahan Q11 dan §36. Penghematan satu lapisan tak langsung
tidak sebanding dengan membuka kembali batasan tool PM.

### YAML sepanjang alur, termasuk `.task/contract.json`

**Ditolak.** Bertentangan dengan Q15 yang menyebut `contract.json` secara eksplisit,
dan menambah parser YAML pada jalur hook yang dieksekusi setiap PreToolUse.

### JSON sepanjang alur, termasuk file yang ditulis manusia

**Ditolak untuk v0.1.0.** Konsisten dan paling sederhana, tetapi kontrak task
ditinjau manusia dan JSON tidak mendukung komentar. Dapat ditinjau ulang bila
ternyata kontrak lebih sering dihasilkan agent daripada ditulis manusia.

---

## Rollback

1. **Format:** ganti materialisasi YAML→JSON menjadi penyalinan langsung; JSON
   Schema tetap berlaku karena tidak bergantung pada format sumber
2. **Bahasa:** ganti isi `scripts/*.sh` dengan implementasi Bash; permukaan yang
   dilihat PM tidak berubah, sehingga Q11 dan definisi agent tidak terpengaruh
3. Tandai ADR ini `Superseded`

Wrapper `.sh` membuat rollback bahasa bersifat lokal — tidak ada konsumen di luar
`scripts/` yang mengetahui implementasinya Go.

---

## Dampak pada dokumen arsitektur

Sesuai §70:

| Aspek | Isi |
|---|---|
| **Reason** | Tiga sumber format yang bertentangan; deteksi overlap R-03 menuntut logika teruji; §30/§34 memuat teks yang telah ditimpa Q8/Q12/R-04 |
| **Affected roles** | Project Manager (permukaan runner tetap `scripts/*.sh`); seluruh writer agent (bentuk `.task/contract.json`) |
| **Migration steps** | Tulis 3 JSON Schema → implementasi Go + wrapper → matriks uji 12 kasus → uji Done §58 |
| **Backward compatibility** | Tidak ada dampak — belum ada schema maupun script yang dibuat |
| **Rollback** | Lihat bagian di atas |
| **Versi** | Tetap 0.1.0 |

Klasifikasi §69: penetapan format artefak dan bahasa implementasi tidak mengubah
control model, orchestrator, responsibility boundary, maupun execution runtime.
Diperlakukan sebagai amandemen baseline pra-rilis.

# ADR-005 — Empat Role Engineering Tambahan

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 30 Juli 2026 |
| **Pemutus** | Pemilik arsitektur (Mindtoscreen) |
| **Versi arsitektur** | 0.1.0 |
| **Menimpa** | §57 (jumlah agent), §17–§25 (daftar role) |
| **Terkait** | §29.2, §37, §70, D-05, ADR-003, ADR-004 |

---

## Konteks

Dokumen arsitektur mendefinisikan sembilan role (§17–§25) dan §57 menyebut
"9 project agent". Kesembilannya mengasumsikan satu bentuk project: backend
terpisah dari frontend.

Pemeriksaan portabilitas menemukan asumsi itu tidak cukup untuk pekerjaan klien
yang bentuknya beragam:

| Bentuk project | Repo | Dapat dinyatakan? |
|---|---|---|
| A | `backend` + `frontend` | ✅ |
| B | satu repo `fullstack` | ⚠️ repo bisa, role tidak tepat |
| C | `backend` + `frontend` + `mobile` | ❌ role dan task type tidak ada |

Jumlah dan susunan repository **tidak pernah** menjadi kendala — `repository`
pada `common.schema.json` adalah string bebas, bukan enum, dan tidak ada tempat
yang mengunci berapa repo per project. Yang menjadi kendala adalah dua daftar
tertutup: `role` dan `task.type`.

Kata "mobile", "android", "ios", maupun "flutter" tidak muncul sama sekali pada
dokumen arsitektur. Task mobile karena itu **ditolak schema**, bukan sekadar
berjalan tanpa panduan.

### Mengapa repo fullstack tetap aman

§29.2 berbunyi *"satu task hanya boleh menulis pada satu repository"* — bukan
"satu repository hanya boleh ditulis satu task". Dua task pada repo fullstack
karena itu sah selama path-nya terpisah, dan isolasinya dijamin reservasi path,
bukan batas repository. Justru pada bentuk inilah deteksi overlap paling
berperan.

---

## Keputusan

**1. Empat role engineering ditambahkan.** Total menjadi 13.

| Role | Cakupan |
|---|---|
| `fullstack-engineer` | satu repo memuat backend dan frontend sekaligus |
| `mobile-engineer` | mobile lintas platform (React Native, Flutter) |
| `android-developer` | khusus Android native |
| `ios-developer` | khusus iOS native |

**2. Keempatnya adalah `writerRole`.** Mereka menulis kode aplikasi, sehingga
wajib memegang reservasi path dan wajib melaporkan `changed_files`.

**3. Dua `task.type` ditambahkan**, bukan empat: `fullstack-implementation` dan
`mobile-implementation`.

Platform dinyatakan oleh **role**, jenis pekerjaan oleh **type**. Menambah
`android-implementation` dan `ios-implementation` akan menduplikasi informasi
yang sudah dibawa `ownership.role`.

**4. Nama role wajib kebab-case huruf kecil.** `ios-developer`, bukan
`iOS-developer`.

Nama role menjadi nama berkas `.claude/agents/<role>.md` (§37). Huruf besar
berisiko berperilaku berbeda antara macOS/Windows yang case-insensitive dan
Linux yang case-sensitive — alasan yang sama dengan `path-overlap-matrix.md` §3.
Ditegakkan `TestRoleNamesAreFileSafe`.

**5. Phase 2 (§57) menghasilkan 13 definisi agent**, bukan 9.

---

## Rasional

### Mengapa memisahkan mobile menjadi tiga role

`mobile-engineer` saja tidak cukup, dan tiga role terpisah bukan pemborosan:

- **Quality gate berbeda.** Android memanggil Gradle, iOS memanggil `xcodebuild`
  dan menuntut mesin macOS. Satu role dengan dua stack membuat `quality_gates`
  ambigu.
- **Path boundary berbeda.** Project native memiliki `app/src/main/**` versus
  `ios/**`; cross-platform memiliki satu pohon sumber bersama.
- **Reservasi lebih tajam.** Pada repo mobile cross-platform, task Android dan
  iOS dapat berjalan paralel pada folder platform masing-masing tanpa dianggap
  konflik.

`mobile-engineer` tetap ada karena React Native dan Flutter memang satu basis
kode — memaksanya dipecah dua akan menghasilkan konflik reservasi yang tidak
perlu.

### Mengapa bukan memakai frontend-engineer untuk mobile

Mungkin untuk React Native, memaksakan untuk Flutter atau native. Perbedaannya
bukan hanya bahasa: siklus build, penandatanganan artifact, dan distribusi store
tidak memiliki padanan pada frontend web.

### Mengapa ini perubahan arsitektur, bukan penyesuaian schema

§57 menyebut angka "9 project agent" secara eksplisit, dan §17–§25 mendaftar
kesembilannya satu per satu. Menambah role mengubah responsibility boundary —
kategori yang §69 klasifikasikan MAJOR. Karena v0.1.0 belum dirilis, ia
diperlakukan sebagai amandemen baseline.

---

## Konsekuensi

### Positif

- Bentuk project A, B, dan C seluruhnya dapat dinyatakan
- Repo fullstack memperoleh role yang tepat, bukan backend-engineer yang
  dipaksa menulis frontend
- Task mobile tidak lagi ditolak schema

### Negatif dan biaya

- **Phase 2 (§57) membesar 44%** — 13 definisi agent, bukan 9. Setiap definisi
  memerlukan batas tool dan batas path sendiri
- **Empat role tanpa §17–§25 padanannya.** Kesembilan role lama memiliki bagian
  dokumen arsitektur yang merinci purpose, owns, responsibilities, allowed,
  prohibited, writable paths, dan definition of done. Keempat role baru **belum**
  memilikinya — harus ditulis saat Phase 2, dan itu pekerjaan yang tidak dapat
  disalin begitu saja
- **`ios-developer` menuntut runner bermesin macOS.** Belum ada mekanisme yang
  menyatakan prasyarat platform pada task contract. Dicatat sebagai terbuka
- Daftar role terduplikasi pada `handoff.schema.json` (blok
  `implementation-complete`) karena `code-reviewer` harus terkecuali secara
  eksplisit. Dijaga `TestWriterRolesMustReportChanges` agar tidak menyimpang

### Yang belum diputuskan

| Pertanyaan | Kapan |
|---|---|
| Bagaimana task menyatakan prasyarat platform runner, misal iOS wajib macOS? | Phase 2 (§57) |
| Apakah `mobile-engineer` dan `android-developer` boleh aktif bersamaan pada satu repo? | Phase 2 (§57) |
| Distribusi role per repository — Q10 hanya mencakup sembilan role lama | Phase 2 (§57) |

---

## Alternatif yang ditolak

### Mempertahankan sembilan role

**Ditolak.** Membatasi sistem pada project backend+frontend, sementara target
pekerjaan klien bentuknya beragam. Task mobile akan ditolak schema tanpa jalan
keluar selain memalsukan role.

### Satu role `mobile-engineer` untuk semua platform

**Ditolak.** Menyembunyikan perbedaan quality gate Gradle versus `xcodebuild`,
dan membuat task Android dan iOS pada satu repo saling konflik reservasi padahal
folder platformnya terpisah.

### Menambah `android-implementation` dan `ios-implementation` pada task.type

**Ditolak.** Platform sudah dinyatakan `ownership.role`; menambahkannya pada
`type` menciptakan dua sumber untuk satu fakta, yang dapat saling bertentangan.

### Role generik `engineer` dengan field stack

**Ditolak untuk v0.1.0.** Paling ringkas, tetapi memindahkan batas tool dan batas
path dari definisi agent ke data — sedangkan §37 menempatkan batas itu pada
frontmatter agent yang termasuk human-only write. Batas yang berada di data lebih
mudah diubah agent daripada batas yang berada di konfigurasi.

---

## Rollback

1. Hapus keempat role dari `role` dan `writerRole` pada `common.schema.json`
2. Hapus `fullstack-implementation` dan `mobile-implementation` dari `task.type`
3. Hapus keempatnya dari blok `implementation-complete` pada `handoff.schema.json`
4. Naikkan `schema_version` — task dan reservasi yang sudah memakainya menjadi
   tidak sah
5. Tandai ADR ini `Superseded`

Rollback aman selama belum ada task memakai role baru. Setelah ada, ia menuntut
migrasi data.

---

## Dampak pada dokumen arsitektur

Sesuai §70:

| Aspek | Isi |
|---|---|
| **Reason** | Sembilan role mengasumsikan project backend+frontend; bentuk fullstack dan mobile tidak dapat dinyatakan dan ditolak schema |
| **Affected roles** | Empat role baru; `task.type` bertambah dua; Phase 2 (§57) menghasilkan 13 agent |
| **Migration steps** | Perbarui enum pada tiga schema → tulis §17–§25 padanan untuk empat role baru pada Phase 2 → tetapkan distribusi per repository |
| **Backward compatibility** | Aditif. Task, reservasi, dan handoff yang sudah ada tetap sah — tidak ada nilai yang dihapus atau diubah artinya |
| **Rollback** | Lihat bagian di atas |
| **Versi** | Tetap 0.1.0 |

Klasifikasi §69: perubahan responsibility boundary umumnya MAJOR. Karena v0.1.0
belum dirilis, diperlakukan sebagai amandemen baseline, bukan kenaikan versi.
`schema_version` tetap `1.0` karena perubahan bersifat aditif.

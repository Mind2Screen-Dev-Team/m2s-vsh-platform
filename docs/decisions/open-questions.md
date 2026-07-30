# Open Questions & Ambiguity Register

**Tanggal:** 29 Juli 2026
**Sumber:** analisis `docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md`
**Status:** 13 tertutup, 4 tertutup sebagian, 2 terbuka, 1 menunggu konfirmasi, 5 menunggu uji empiris

Register ini melacak setiap bagian dokumen arsitektur yang ambigu, kontradiktif,
atau belum dapat langsung diimplementasikan. Kode `A-*` untuk ambiguitas arsitektur,
`D-*` untuk keputusan turunan yang masih terbuka, `V-*` untuk yang butuh verifikasi.

---

## Ringkasan status

| Kode | Isu | Status |
|---|---|---|
| A-01 | `.claude/**` forbidden vs lokasi worktree | ✅ tertutup — Q8 |
| A-02 | PM punya `Bash` sehingga hook Edit/Write terlewat | 🟡 sebagian — Q11 + R-07 |
| A-03 | Reviewer `plan` mode vs writable path | ✅ tertutup — Q9 |
| A-04 | TL/SA melanggar one-task-one-repository | ✅ tertutup — Q14 |
| A-05 | Reservasi dilepas saat PR, tapi retry kembali `running` | ✅ tertutup — Q12 |
| A-06 | Mneme fail-open tapi jadi required CI gate | ✅ tertutup — Q18 |
| A-07 | Penempatan konfigurasi Ponytail | ✅ tertutup — Q17 |
| A-08 | `git checkout` diblokir vs kebutuhan runner | ✅ tertutup — Q13 |
| A-09 | Blocklist berbasis string mudah dielakkan | 🟡 sebagian — R-08, V-04 |
| A-10 | Worktree install deps vs larangan install | ✅ tertutup |
| A-11 | `--add-dir` disebut read-only padahal memberi write | 🟡 sebagian — V-02 |
| A-12 | Identitas & nama control repository | ✅ tertutup — Q1 |
| A-13 | Repository pilot belum ditentukan | ✅ tertutup — Q2 |
| A-14 | Base branch belum ditentukan | ✅ tertutup — Q3 + D-01 |
| A-15 | Quality gate command belum diketahui | ✅ tertutup — Q4 |
| A-16 | Distribusi agent per repository | ✅ tertutup — Q10 |
| D-01 | Default branch pilot `master` vs control `main` | ✅ tertutup — dieksekusi 2026-07-29 |
| D-02 | Repo klien private tetap tanpa enforcement | 🔴 **terbuka** |
| D-03 | Pembatasan hak push/merge hanya untuk repo organization | 🔴 **terbuka** |
| D-04 | Skala severity ditetapkan schema, bukan arsitektur | 🟡 menunggu konfirmasi |
| D-05 | Field `project` belum dipakai runner (multi-project) | ✅ diputuskan — model A untuk v0.1.0, model B ke v0.2.0 |
| V-01…V-05 | Butuh uji empiris | ⏳ Phase 1 dan 3 (§58, §59) |

---

## Isu tertutup

### A-01 — `.claude/**` forbidden sekaligus lokasi worktree

**Kontradiksi:** §34 menempatkan `.claude/**` pada `paths.forbidden`, sementara §30
menaruh worktree di `.claude/worktrees/BE-101`. Akar kerja agent berada di dalam
path yang dilarang ditulisnya.

**Resolusi (Q8):** worktree dipindah ke luar repo —
`$HOME/.m2s/worktrees/<repository>/<task-id>`, override lewat `M2S_WORKTREE_ROOT`.
`.claude/**` tetap forbidden seutuhnya tanpa pengecualian.

---

### A-03 — Reviewer plan mode vs writable path

**Kontradiksi:** §23.6 memberi reviewer writable path `reviews/code/**`, sedangkan
Appendix A memberinya `permissionMode: plan` dan tool tanpa `Write`/`Edit`.
Dalam plan mode seluruh write diblokir.

**Resolusi (Q9):** reviewer mengembalikan structured output; runner yang menuliskan
ke `reviews/code/**`. Sifat read-only murni dipertahankan.

---

### A-04 — TL/SA melanggar one-task-one-repository

**Kontradiksi:** writable path TL/SA (§18.6) menyebar ke minimal tiga lokasi,
sedangkan §29.2 melarang satu task menulis lebih dari satu repository.

**Resolusi (Q14):** seluruh artifact TL/SA dipusatkan di control repository.
Hanya `.mneme/project_memory.json` yang berada di application repo, ditangani
task type `MNEME-*` terpisah.

---

### A-05 — Celah reservasi antara PR dan merge

**Kontradiksi:** §30 melepas reservasi saat PR dibuat; §33 mengizinkan
`changes-requested → running`. Antara PR dan merge, path tidak terlindungi.

**Resolusi (Q12):** reservasi ditahan sampai **merge**, dengan status antara
`reserved-pending-merge`.

---

### A-06 — Mneme fail-open sebagai required gate

**Kontradiksi:** §6.3 mengakui hook Mneme fail-open; §43 item 11 menjadikannya
gate wajib.

**Resolusi (Q18):** CI gate fail-closed. Hook lokal wajib `exit 2` (lihat T-01).
Setiap hook security memiliki self-test.

---

### A-07 — Penempatan konfigurasi Ponytail

**Ambiguitas:** §5.5 menyajikan env var dalam blok `bash` tanpa menyebut lokasinya.

**Resolusi (Q17):** `.claude/settings.json` blok `env`. Version-controlled dan
tidak dapat diubah engineer secara diam-diam (§5.6).

---

### A-08 — `git checkout` diblokir vs kebutuhan runner

**Kontradiksi:** §42.2 memblokir `git checkout`/`git switch`, padahal
`git worktree add` melakukan checkout internal.

**Resolusi (Q13):** runner beroperasi di luar sesi agent, sehingga hook agent tidak
berlaku padanya. Agent tidak pernah menjalankan perintah worktree/checkout.

---

### A-10 — Install dependency di worktree vs larangan install

**Kontradiksi semu:** §42.6 mengizinkan hook menginstal project dependencies;
§16.5 melarang agent menginstal package.

**Resolusi:** tidak bertentangan, karena install dilakukan **runner** pada tahap
bootstrap worktree, sebelum sesi agent dimulai. Ditegaskan eksplisit agar tidak
disalahtafsirkan sebagai izin bagi agent.

---

### A-12, A-13, A-15, A-16

Tertutup oleh Q1, Q2, Q4, dan Q10. Lihat `phase-0-decision-log.md`.

---

### A-14 — Base branch, dan D-01 — ketidakkonsistenan default branch

**Tertutup 2026-07-29** dengan eksekusi normalisasi pada kedua repo pilot.

**Koreksi terhadap catatan sebelumnya:** dokumen ini semula mencatat default branch
kedua repo pilot sebagai `master` yang perlu di-*rename*. Pemeriksaan API menunjukkan
kondisi sebenarnya berbeda: kedua repo **sepenuhnya kosong** — `isEmpty: true`,
`diskUsage: 0`, **nol branch**. `master` hanya nilai setting `default_branch` bawaan
akun, bukan branch yang ada. Karena itu tidak ada rename yang bisa dilakukan, dan
flip setting ditolak API:

```
422 Cannot update default branch for an empty repository.
    Please init the repository and push first.
```

**Konsekuensi:** `main`, `develop`, dan `staging` semuanya memerlukan commit untuk
bisa eksis. Normalisasi karena itu berbentuk **seed commit**, bukan rename.

**Yang dieksekusi** pada `m2s-vsh-project-backend` dan `m2s-vsh-project-frontend`:

1. Seed commit di `main` — `README.md` (peran, stack, tabel branch, rujukan ADR-001)
   dan `.gitignore` per stack mengikuti Q4 (Go untuk backend, Next.js untuk frontend).
2. `default_branch` di-set ke `main` via API setelah push.
3. `develop` dan `staging` dibuat dari commit `main` yang sama.

**Kondisi akhir terverifikasi:**

| Repository | Default branch | Branch |
|---|---|---|
| `m2s-vsh-platform` | `main` | `main` |
| `m2s-vsh-project-backend` | `main` | `develop`, `main`, `staging` |
| `m2s-vsh-project-frontend` | `main` | `develop`, `main`, `staging` |

**Temuan sampingan:** kedua repo pilot berada di akun personal `fajarcandraaa` dan
bersifat **public**, bukan di org `Mind2Screen-Dev-Team` seperti diasumsikan D-02.
Public repo pada akun Free tetap mendukung branch protection — ini menguntungkan
**AC-0.6** dan tidak terhalang plan Free. D-02 tetap terbuka untuk repo klien private.

**Belum dikerjakan:** control repo `m2s-vsh-platform` hanya punya `main`. Apakah
control repo juga memerlukan `develop`/`staging` belum diputuskan; §44 hanya
mengatur repo project.

---

## Isu tertutup sebagian

### A-02 — PM punya `Bash` sehingga hook Edit/Write terlewat

**Sisa risiko:** `Bash` dapat menulis file lewat `>`, `tee`, `cp`, `sed -i`,
`python -c` tanpa memicu hook Edit/Write.

**Mitigasi terpasang (Q11):** tool `Agent` dicabut; `Bash` dibatasi hook ke pola
persis `scripts/<runner>.sh`; write dibatasi `control/**`.

**Sisa yang belum tertutup:** validasi write-effect dari perintah Bash arbitrer.
Dilacak sebagai **R-07**, ditangani **Phase 3 (§59)** dengan CI changed-path
validation sebagai jaring kedua, lalu dijadikan required check pada
**Phase 4 (§60)**.

---

### A-09 — Blocklist string mudah dielakkan

**Sisa risiko:** daftar §42.2 adalah pencocokan teks. `git checkout` dapat menjadi
`g=checkout; git $g`; `rm -rf` dapat menjadi `rm -r -f` atau `find … -delete`.

**Diperkuat oleh temuan verifikasi:** dokumentasi Claude Code sendiri menyatakan
filter perintah Bash dapat fail-open pada input yang tidak ter-parse, dan
menyarankan permission system sebagai enforcement keras.

**Posisi yang ditetapkan:** hook adalah **defense-in-depth**, bukan boundary.
Boundary sebenarnya = `permissions.deny` + CI + branch protection.
Dilacak sebagai **R-08**; uji elakan dijadwalkan sebagai **V-04**.

---

### A-11 — `--add-dir` disebut read-only

**Sisa risiko:** §46 menyebut `--add-dir` "hanya untuk read context" sambil mengakui
ia memberi file access. Ini konvensi, bukan enforcement.

**Mitigasi rencana:** `permissions.deny` untuk Edit/Write pada direktori tambahan;
runner tidak pernah memberikan `--add-dir` ke repository yang bukan target task.

**Belum diverifikasi:** perilaku write sebenarnya. Dilacak sebagai **V-02**.

---

## Isu terbuka

### D-02 — Repository klien private tetap tanpa enforcement 🔴

**Kondisi:** organization `Mind2Screen-Dev-Team` berada pada plan **Free** dengan
29 repository private. Seluruhnya tidak dapat memasang branch protection maupun
rulesets.

**Dampak:** jalur repository public membuktikan model arsitektur bekerja, tetapi
**tidak dapat diterapkan pada project klien** yang wajib private. Untuk klien,
lapisan anti-overlap #7 dan #8 tetap tidak tertegakkan.

**Opsi:**

| Opsi | Biaya | Catatan |
|---|---|---|
| Upgrade org ke GitHub Team | ~$4/seat/bulan × 14 seat | satu-satunya jalur yang mendukung target akhir |
| Tetap Free, terima tanpa enforcement untuk klien | 0 | melanggar prinsip #7 dokumen arsitektur |

**Hambatan:** peran pemilik arsitektur di org adalah `member`, bukan `owner`.
Upgrade memerlukan persetujuan owner organization.

**Menunggu keputusan.** Tidak memblokir Phase 0–8 pilot.

---

### D-03 — Pembatasan hak push/merge hanya tersedia untuk repo organization 🔴

**Ditemukan 30 Juli 2026** saat mengeksekusi branch protection.

**Kondisi:** ketiga repo berada di akun personal `fajarcandraaa`. Status public membuka
branch protection, tetapi **tidak** membuka pembatasan *siapa* yang boleh push atau
merge. Fitur itu org-only. Terverifikasi:

```
PUT /repos/fajarcandraaa/m2s-vsh-project-backend/branches/main/protection
422 Only organization repositories can have users and team restrictions
```

**Mengoreksi asumsi sebelumnya.** README dan ADR-001 semula menyatakan status public
sudah cukup untuk menegakkan lapisan anti-overlap #7 dan #8. Itu tidak akurat: public
adalah syarat perlu, bukan syarat cukup.

**Dampak:** prasyarat ADR-001 #3 dan #4 tidak dapat dipenuhi. Model dua identitas
`m2s-worker`/`m2s-approver` tetap dapat dibuat, tetapi **hak-nya tidak dapat dibatasi
di sisi GitHub** — `m2s-worker` tidak dapat dilarang me-merge. Akibatnya keputusan #4
ADR-001 (tidak ada identitas agent ber-scope admin) juga kehilangan penegakannya.
Acceptance §66 #9 belum dapat diuji.

**Opsi:**

| Opsi | Biaya | Catatan |
|---|---|---|
| Pindahkan repo pilot ke organization (Free) | 0 | perlu diverifikasi apakah restriction tersedia pada org plan Free atau menuntut Team |
| Pindahkan + upgrade org ke Team | ~$4/seat/bulan | jalur yang pasti; menutup D-02 sekaligus |
| Tetap di akun personal | 0 | lapisan #7 dan #8 tetap soft rule sepanjang pilot |

**Belum diverifikasi:** apakah org plan **Free** sudah cukup untuk push restriction,
atau harus Team. Ini menentukan apakah D-03 dapat ditutup tanpa biaya. Perlu diuji
sebelum keputusan diambil.

**Bertautan dengan D-02** — keduanya bermuara pada status organization. Sebaiknya
diputuskan bersamaan.

**Menunggu keputusan.** Tidak memblokir Phase 1–3 (§58, §57, §59); memblokir
aktivasi ADR-001, yang bergantung pada Phase 4 (§60 — GitHub Workflow).

---

### D-04 — Skala severity ditetapkan schema, bukan dokumen arsitektur 🟡

**Ditemukan 30 Juli 2026** saat menulis `handoff.schema.json`.

**Kondisi:** §22.8 menuntut *"defects mempunyai severity"* dan §23.7 menuntut
*"setiap finding memiliki severity"*, tetapi dokumen arsitektur **tidak pernah
mendefinisikan nilai severity**. Pemindaian menemukan beberapa skala terpakai
bercampur: `high/medium/low`, `critical/major/minor/trivial`, dan `blocker`.

**Yang ditetapkan schema:**

```
blocker  → menghalangi merge
major    → wajib ditangani, tidak menghalangi merge
minor    → sebaiknya ditangani
nit      → preferensi, boleh diabaikan
```

**Alasan pemilihan:** `blocker` sudah dipakai dokumen dan memetakan langsung ke
`request-changes` (§23.7). `critical` dihindari karena tumpang tindih makna dengan
`blocker` tanpa menambah daya pisah. Empat tingkat cukup untuk memutuskan
merge/tidak-merge plus dua tingkat catatan.

**Status:** keputusan schema, **bukan** kutipan arsitektur. Bila pemilik arsitektur
menghendaki skala lain, `handoff.schema.json` perlu version bump dan seluruh handoff
yang sudah tersimpan perlu dipetakan ulang.

**Menunggu konfirmasi pemilik arsitektur.** Tidak memblokir Phase 1.

---

### D-05 — Field `project` belum dipakai runner ✅

**Ditemukan 30 Juli 2026** saat memeriksa portabilitas sistem ke project lain.

**Kondisi:** `task.schema.json` memiliki field `task.project` (§34 mencontohkan
`project: tumbuh`) dan §36 mendaftar direktori `control/projects/`. Namun runner
**tidak pernah membacanya**:

| Tempat | Perilaku sekarang | Akibat bila satu control repo melayani banyak project |
|---|---|---|
| Nama berkas reservasi | `<task-id>.yaml` | dua project dengan `BE-101` menulis ke berkas yang sama |
| Pencocokan konflik | `res.Repository() == repository` | dua project yang masing-masing punya repo bernama `backend` dianggap bertabrakan |
| `reservation.schema.json` | tidak punya field `project`, `additionalProperties: false` | tidak dapat ditambahkan tanpa mengubah schema |

**Dua model yang mungkin:**

| Model | Isi | Status |
|---|---|---|
| A | satu control repo per project — clone lalu ganti isinya | ✅ **dipilih untuk v0.1.0** |
| B | satu control repo, banyak project (disiratkan `control/projects/`) | ⏳ ditunda ke **v0.2.0** |

Dokumen arsitektur tidak memilih secara tegas: §56 menyebut "satu pilot project",
sementara §36 menyediakan `control/projects/` yang jamak.

### Keputusan 30 Juli 2026 — model A untuk v0.1.0

**Model A dipakai sepanjang pilot.** Ia bekerja tanpa perubahan apa pun: clone
control repo, ganti isi `control/`, arahkan `M2S_CONTROL_ROOT`. Setiap project
memiliki registry sendiri sehingga isolasinya mutlak.

**Model B ditunda ke v0.2.0.** Ini sejalan dengan §65, yang sudah mencantumkan
*"multi-project coordination menjadi bottleneck"* sebagai pemicu kenaikan versi.
Model B karena itu bukan pekerjaan yang terlupakan, melainkan lingkup versi
berikutnya.

**Biaya yang diterima pada model A:** `schemas/`, `cmd/m2s`, dan `scripts/`
terduplikasi di setiap clone. Perbaikan bug harus disalin ke seluruh project — dan
ini menyentuh komponen penegak batas, tempat penyimpangan versi paling berbahaya.
Bila jumlah project bertambah sampai penyalinan itu menjadi sumber kesalahan,
pemicu §65 terpenuhi dan model B dikerjakan.

**Yang membuat penundaan ini murah:** beralih A → B tetap ringan selama registry
masih kosong saat project baru dimulai. Yang mahal adalah menambah kunci `project`
setelah ada riwayat reservasi nyata, karena menuntut penamaan ulang berkas dan
pemindahan worktree. Empat titik pada tabel di atas adalah daftar kerja yang siap
dipakai saat v0.2.0 dibuka.

**Yang tidak menjadi kendala:** jumlah dan susunan repository per project.
`repository` adalah string bebas, bukan enum. Bentuk backend+frontend, fullstack
repo tunggal, maupun backend+frontend+mobile seluruhnya dapat dinyatakan —
role-nya ditambahkan ADR-005.

**Biaya menunda:** menambah kunci reservasi setelah ada riwayat reservasi nyata
jauh lebih mahal daripada sekarang, karena menuntut migrasi berkas dan penulisan
ulang pencocokan konflik.

**Tertutup sebagai keputusan sadar**, bukan dibiarkan menggantung: model A dipakai
v0.1.0, model B menjadi lingkup v0.2.0 sesuai pemicu §65. Tidak memblokir Phase 1.

---

## Bagian dokumen arsitektur yang ditimpa ADR

Daftar ini mencegah teks lama terbaca sebagai masih berlaku. Bila bekerja pada
bagian di bawah, baca ADR-nya lebih dulu.

| Bagian | Ditimpa oleh | Isi yang berubah |
|---|---|---|
| §17.6, §18.5 | ADR-001 | PM & TL/SA boleh approve/merge ke non-`main` — **belum berlaku efektif**, lihat D-03 |
| §57, §58 | ADR-003 | Urutan Phase 1 dan 2 ditukar |
| §57, §17–§25 | ADR-005 | Role bertambah 9 → 13; Phase 2 menghasilkan 13 definisi agent |
| §30 | ADR-004 (Q8) | Lokasi worktree: `$HOME/.m2s/worktrees/<repository>/<task-id>`, bukan `.claude/worktrees/` |
| §30 | ADR-004 (Q12) | Reservasi dilepas saat **merge**, bukan saat PR dibuat; status antara `reserved-pending-merge` |
| §34 | ADR-004 (R-04) | Ditambah field `shared_file_ownership` |
| §30, §34 | ADR-004 | Contoh YAML di kedua bagian bersifat **ilustratif**; definisi normatif ada pada `schemas/*.schema.json` |
| §20.6 | Q4 | Path `tests/unit/**` tidak berlaku untuk repo Go — unit test colocated `_test.go` |

---

## Menunggu uji empiris

| Kode | Pertanyaan | Dijadwalkan |
|---|---|---|
| V-01 | Apakah hook dapat membaca file di luar cwd? | *tidak lagi relevan — Q15 menghilangkan kebutuhannya* |
| V-02 | Perilaku write `--add-dir` yang sebenarnya | Phase 3 (§59) |
| V-03 | Presedensi `settings.json` vs `settings.local.json` untuk `deny` | Phase 3 (§59) |
| V-04 | Apakah `permissions.deny` Bash dapat dielakkan variabel shell | Phase 3 (§59) |
| V-05 | Perilaku `WorktreeCreate` terhadap worktree buatan runner | Phase 1 (§58) |

---

## Catatan pemeliharaan

Register ini diperbarui pada setiap checkpoint fase. Isu yang ditutup tidak dihapus,
melainkan ditandai beserta resolusinya, agar keputusan dapat ditinjau ulang bila
asumsinya berubah.

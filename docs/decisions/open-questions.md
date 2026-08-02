# Open Questions & Ambiguity Register

**Tanggal:** 29 Juli 2026
**Sumber:** analisis `docs/architecture/M2S-VSH-Lite-v0.1.0-Architecture.md`
**Status:** 14 tertutup, 4 tertutup sebagian, 2 terbuka, 5 menunggu uji empiris

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
| D-03 | Pembatasan hak push/merge hanya untuk repo organization | 🟠 **terbuka, biaya terjawab** — Team **tidak** dibutuhkan; ruleset `restrict updates` membukanya tanpa migrasi (ADR-007 #7). Menunggu GitHub App + V-05 |
| D-04 | Skala severity ditetapkan schema, bukan arsitektur | ✅ tertutup — disetujui 2026-07-31 |
| D-05 | Field `project` belum dipakai runner (multi-project) | ✅ diputuskan — model A untuk v0.1.0, model B ke v0.2.0 |
| V-01…V-08 | Butuh uji empiris | ⏳ Phase 1, 3, 4 (§58, §59, §60) — V-04, V-06, V-07, V-08 sudah terjawab |

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

**Opsi — diperbarui 31 Juli 2026 (Phase 4, ADR-007 #7), DIKOREKSI 1 Agustus 2026
oleh uji empiris V-06/V-07/V-08.** Opsi Team tetap **dicabut**. Tetapi kesimpulan
"ruleset membuka restriction tanpa migrasi" **hanya sebagian benar** — lihat V-08.

| Opsi | Biaya | Membuka | Catatan |
|---|---|---|---|
| **Ruleset `restrict updates` + bypass `User`/`RepositoryRole`** | 0 | push/merge restriction **untuk identitas manusia** | tanpa migrasi. Terverifikasi: ruleset menolak push pemilik repo (`GH013`) bila bypass kosong, dan melepasnya bila masuk list. `bypass_mode: pull_request` memisahkan hak merge dari hak push langsung. **Batasnya:** GitHub App tidak dapat menjadi bypass actor di repo personal (V-08) |
| Pindahkan repo pilot ke organization (Free) | 0 | push restriction **+ GitHub App sebagai bypass actor** + merge queue | **Free terkonfirmasi cukup untuk repo public.** Satu-satunya jalur yang memenuhi ADR-001 #5 secara utuh, karena App harus menjadi bagian *owner organization*. Org `Mind2Screen-Dev-Team` sudah ada |
| ~~Pindahkan + upgrade org ke Team~~ | ~~~$4/seat/bulan~~ | — | **dicabut.** Tidak dibutuhkan. Team hanya untuk D-02 (repo private) |
| Tetap di akun personal tanpa ruleset | 0 | — | lapisan #7 dan #8 tetap soft rule sepanjang pilot |

**Yang berubah karena V-08.** ADR-007 #7 menyimpulkan dari dokumentasi bahwa jalur
"ruleset tanpa migrasi" cukup untuk ADR-001 #5 (model dua identitas berbasis GitHub
App). Uji API membantahnya:

```
PUT /repos/fajarcandraaa/m2s-vsh-rules-probe/rulesets/20155906
422 Actor GitHub Actions integration must be part of the ruleset source
    or owner organization
```

Jadi di repo akun personal, ruleset dapat membatasi **manusia dan role**, tetapi
**tidak dapat memberi pengecualian kepada GitHub App**. Model `m2s-worker`/
`m2s-approver` sebagaimana ADR-001 #5 merancangnya (GitHub App, bukan machine user)
karena itu **menuntut organization** — bukan karena push restriction, melainkan
karena bypass actor.

**Kabar baiknya dari V-06:** ruleset **mengikat pemilik repo** (`GH013`,
`current_user_can_bypass: "never"`), berbeda dari classic protection di mana admin
selalu dapat push. Itu membuat ADR-001 **#4** ("tidak ada identitas agent ber-scope
admin") dapat ditegakkan — sesuatu yang classic protection tidak pernah bisa
berikan, bahkan di organization.

**Bertautan dengan D-02** — keduanya bermuara pada status organization, tetapi
**tidak lagi pada biaya yang sama**. D-02 (repo klien private) tetap menuntut Team;
D-03 tidak.

**Resolusi 2 Agustus 2026 — transfer org selesai (ADR-008).** Ketiga repo sudah di
`Mind2Screen-Dev-Team`; ruleset `agent-push-restriction` (bypass `Integration`
App `m2s-approver`) dan `agent-worker-restriction` (kunci develop/staging dari
push langsung worker) menunggu App dibuat — lihat `docs/operator/org-migration.md`
§Langkah ADR-001 #5. Verifikasi pasca-transfer (ADR-008 §Keputusan #4, wajib):
branch protection `validate-changed-paths` **ikut** di develop+staging backend+
frontend, GitGuardian **ikut** (check-run `GitGuardian Security Checks`, app
`gitguardian`, `conclusion: "success"` pada SHA `39ab5f2`), rulesets 0 (baseline,
belum dipasang). V-09 mencatat uji aktual `Integration` bypass di org.

**Menunggu keputusan.** Tidak memblokir Phase 1–4. Yang masih tertahan adalah
restriction *siapa* boleh merge lewat identitas App — dan V-07 menunjukkan itu
menuntut migrasi ke organization, bukan sekadar membuat App. Kriteria Done §60
karena itu tercapai sebagian (ADR-007 #8).

---

### D-04 — Skala severity ditetapkan schema, bukan dokumen arsitektur ✅

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

**Disetujui pemilik arsitektur 31 Juli 2026.** Skala `blocker/major/minor/nit`
ditetapkan berlaku. Tidak ada version bump karena belum ada handoff tersimpan yang
memakai skala lain.

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
| Appendix A | ADR-006 | Baseline frontmatter 13 role; PM **tanpa** tool `Agent` (Q11); `effort` ditetapkan per role |
| §30 | ADR-004 (Q8) | Lokasi worktree: `$HOME/.m2s/worktrees/<repository>/<task-id>`, bukan `.claude/worktrees/` |
| §30 | ADR-004 (Q12) | Reservasi dilepas saat **merge**, bukan saat PR dibuat; status antara `reserved-pending-merge` |
| §34 | ADR-004 (R-04) | Ditambah field `shared_file_ownership` |
| §30, §34 | ADR-004 | Contoh YAML di kedua bagian bersifat **ilustratif**; definisi normatif ada pada `schemas/*.schema.json` |
| §20.6 | Q4 | Path `tests/unit/**` tidak berlaku untuk repo Go — unit test colocated `_test.go` |

---

## Menunggu uji empiris

| Kode | Pertanyaan | Status |
|---|---|---|
| V-01 | Apakah hook dapat membaca file di luar cwd? | *tidak lagi relevan — Q15 menghilangkan kebutuhannya* |
| V-02 | Perilaku write `--add-dir` yang sebenarnya | ⏳ **tetap terbuka** — butuh sesi Claude Code live, tidak dapat diuji dari bash. `permissions.deny` Edit/Write pola path terpasang sebagai mitigasi; verifikasi perilaku ditunda ke pilot Phase 7 (§63) |
| V-03 | Presedensi `settings.json` vs `settings.local.json` untuk `deny` | ⏳ **tetap terbuka** — butuh sesi live. Konvensi Claude Code: `.local.json` menimpa project settings, tetapi `deny` bersifat aditif (union). Perlu konfirmasi empiris sebelum diandalkan |
| V-04 | Apakah `permissions.deny` Bash dapat dielakkan variabel shell | ✅ **TERKONFIRMASI dapat dielakkan** (Phase 3, 2026-07-31) — `g=checkout; git $g` dan `find -delete` lolos hook. Limitation R-08 diterima; boundary = CI + `permissions.deny` path. Lihat `capability-verification.md` §10 |
| V-06 | Apakah pemilik repo otomatis ter-bypass ruleset tanpa masuk bypass list | ✅ **TIDAK ter-bypass** (diuji 2026-08-01, repo `m2s-vsh-rules-probe`) — ruleset `restrict updates` dengan `bypass_actors: []` menolak push pemilik: `GH013: Cannot update this protected ref`, dan API melaporkan `current_user_can_bypass: "never"`. **Berbeda dari classic protection**, yang menyatakan admin *"always able to push to a protected branch"*. Konsekuensi: ruleset mengikat pemilik repo, sehingga ADR-001 **#4** (bukan hanya #3) dapat ditegakkan tanpa organization |
| V-07 | Apakah role "Maintain" tersedia sebagai bypass actor di repo personal | ✅ **`RepositoryRole` diterima** (diuji 2026-08-01) — `actor_id` 2, 4, dan 5 seluruhnya diterima API di repo akun personal. Ketersediaan di **picker UI** belum dilihat; ruleset probe sengaja dibiarkan hidup untuk itu. Tidak memblokir apa pun |
| V-08 | Apakah `actor_type: Integration` (GitHub App) dapat menjadi bypass actor di repo **akun personal** | 🔴 **TIDAK — ditolak** (diuji 2026-08-01): `422 Actor GitHub Actions integration must be part of the ruleset source or owner organization`. **Mengoreksi ADR-007 #7 dan D-03**, yang menyimpulkan dari dokumentasi bahwa hanya `OrganizationAdmin` yang tak berlaku di repo personal. Uji API membantahnya: App harus menjadi bagian dari **owner organization**. Lihat D-03 |
| V-09 | Apakah `actor_type: Integration` (GitHub App) dapat menjadi bypass actor di repo **organization** | ⏳ **terbukti mungkin — menunggu App.** 1 Agustus 2026: setelah transfer ke `Mind2Screen-Dev-Team`, alasan V-08 tidak lagi berlaku — App kini bagian dari *owner organization* (ADR-008). Uji aktual (ruleset `agent-push-restriction` dengan bypass `Integration` App `m2s-approver`) menunggu App dibuat, Fase C org-migration.md |
| V-05 | Perilaku `WorktreeCreate` terhadap worktree buatan runner | ⏳ **terbuka, dan lebih buruk dari yang tercatat** — klaim "guard secret terpasang" **tidak akurat**: `worktree-lifecycle.sh` ada di disk dan self-test-nya lulus, tetapi **tidak terdaftar di `.claude/settings.json`** (hanya event `PreToolUse`, `PostToolUse`, `SubagentStop` terdaftar; tidak ada `WorktreeCreate`/`WorktreeRemove`). Ditemukan Phase 4, 2026-08-01. Hook tidak pernah dipanggil runtime, sehingga mitigasi R-26 dan §42.6 **belum berjalan**. Lihat T-02 |

---

## Catatan pemeliharaan

Register ini diperbarui pada setiap checkpoint fase. Isu yang ditutup tidak dihapus,
melainkan ditandai beserta resolusinya, agar keputusan dapat ditinjau ulang bila
asumsinya berubah.

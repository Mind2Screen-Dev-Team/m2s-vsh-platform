# ADR-007 — Penegakan GitHub Workflow

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 31 Juli 2026 |
| **Pemutus** | Pemilik arsitektur (Mindtoscreen) |
| **Versi arsitektur** | 0.1.0 |
| **Menimpa** | §60 butir "Merge queue"; §29.8 dan §13.6 sejauh keduanya menyandarkan mitigasi pada merge queue |
| **Terkait** | ADR-001, ADR-004, Q16, D-02, D-03, R-12, R-19, R-20, R-24, §16.5, §29.7, §43, §44, §47, §60, §66, §68 |

---

## Konteks

Phase 4 (§60) menuntut lima artefak: PR template, CODEOWNERS, required checks,
review policy, merge queue. Kriteria Done-nya satu kalimat: *"worker menghasilkan
PR dan tidak dapat merge sendiri."*

Verifikasi keadaan nyata sebelum menulis satu berkas pun menemukan enam hal yang
mengubah bentuk fase ini. Empat di antaranya mengoreksi catatan fase sebelumnya.

**1. CI yang dicatat sudah live belum pernah berjalan sekali pun.** Phase 3
mengirim `.github/workflows/path-enforcement.yml` dan mencatatnya sebagai jaring
kedua R-07 yang aktif. Trigger-nya `pull_request: branches: [develop, staging]`,
sedangkan control repo hanya memiliki `main` — kondisi yang sudah tercatat sendiri
di open-questions sebagai "belum dikerjakan". Repo aplikasi memiliki
`develop`/`staging` tetapi tidak memiliki workflow apa pun. Bukti:
`GET /repos/…/actions/runs` mengembalikan `total_count: 0` pada **ketiga** repo.
Berkasnya ada, jalurnya mati. Klaim "CI live" tidak akurat.

**2. Bentuk job yang dikirim Phase 3 tidak dapat dijadikan required check.** Job
itu dijaga `if: startsWith(github.head_ref, 'agent/')` pada level job. GitHub
tidak melaporkan status apa pun untuk job yang di-skip. Dijadikan required check,
ia memblokir setiap PR yang tidak cocok — termasuk seluruh PR manusia — secara
permanen. Ini persis bahaya yang ADR-001 peringatkan: *"status check yang tidak
pernah dilaporkan akan memblokir seluruh merge tanpa jalan keluar."* Ia satu kelas
dengan R-24: penjaga yang tampak ada tetapi tidak menahan apa pun, dengan
tambahan bahwa yang ini justru menahan hal yang salah.

**3. D-03 memiliki jalan keluar tanpa biaya, dan asumsi biayanya salah.** D-03
mencatat pembatasan hak push/merge sebagai org-only berdasarkan `422 Only
organization repositories can have users and team restrictions`, dengan opsi
termahal berupa upgrade ke GitHub Team (~$4/seat/bulan). Galat itu benar, tetapi
kesimpulan biayanya tidak. Dua koreksi:

- **Repository rulesets** menyediakan rule *restrict updates* — *"only users with
  bypass permissions can push to branches or tags whose name matches the pattern
  you specify"* — dengan bypass list. REST API menyatakan satu-satunya
  `actor_type` yang tidak berlaku di repo personal adalah `OrganizationAdmin`;
  `Integration` (GitHub App) berlaku. Karena ADR-001 #5 memang memilih GitHub App
  dan bukan machine user, pola ini mereplikasi apa yang classic `restrictions`
  tolak — **tanpa migrasi dan tanpa biaya**.
- Bila migrasi tetap dipilih, **org plan Free sudah cukup untuk repo public**:
  *"You can enable branch restrictions in public repositories owned by a GitHub
  Free organization."* Ketiga repo pilot public.

Plan Team tidak dibutuhkan untuk apa pun di Phase 4. Ia hanya dibutuhkan D-02
(repo klien private).

**4. Merge queue mustahil di tempat repo berada, bukan sekadar mahal.** Q16
menunda merge queue dari v0.1.0 dengan alasan *"merge queue pada repository
private memerlukan plan berbayar. Setelah repository dijadikan public, merge queue
menjadi tersedia."* Bagian kedua tidak akurat: *"Pull request merge queues are
available in any public repository owned by an organization."* Syaratnya
kepemilikan **organization**, bukan visibilitas, dan tidak ada plan yang membukanya
untuk repo akun personal. Jadi penundaannya bukan keputusan biaya melainkan
konsekuensi tempat repo berada.

**5. CODEOWNERS sebagai gate menghasilkan deadlock.** Pemilik satu-satunya adalah
pemilik repo, yang juga author setiap PR. GitHub: *"Pull request authors cannot
approve their own pull requests."* Mengaktifkan `require_code_owner_review`
menghasilkan syarat yang mustahil dipenuhi. Sebaliknya: *"Repository owners and
administrators can merge a pull request even if it hasn't received an approving
review"* — jadi selama pemilik tunggal, CODEOWNERS menahan agent, bukan manusia.

**6. Satu kontradiksi dokumen belum tercatat.** Q16 memutuskan merge queue
ditunda, tetapi §60 masih mencantumkannya sebagai deliverable, §29.8 menyandarkan
mitigasi overlap semantik padanya (*"Merge queue menyerialisasi integration
order"*), §13.6 dan §32 menyebutnya, dan `path-overlap-matrix.md` memetakan
mitigasi "konflik urutan merge" ke §29.8. Tabel *"Bagian dokumen arsitektur yang
ditimpa ADR"* tidak mencatat satu pun dari ini.

---

## Keputusan

### 1. Artefak kanonik berada di control repo; distribusi lewat pull request

Tiga artefak ditulis satu kali di `templates/github/`:

| Berkas | Tujuan di repo aplikasi |
|---|---|
| `templates/github/workflows/path-enforcement.yml` | `.github/workflows/path-enforcement.yml` |
| `templates/github/CODEOWNERS` | `.github/CODEOWNERS` |
| `templates/github/PULL_REQUEST_TEMPLATE.md` | `.github/PULL_REQUEST_TEMPLATE.md` |

Perubahan selalu dimulai di control repo, lalu diturunkan. Salinan di repo
aplikasi tidak disunting langsung.

Distribusi memakai **pull request**, bukan push langsung, dan **manusia yang
merge**. §16.5 melarang seluruh agent *"mengubah CI protection, CODEOWNERS, atau
security configuration tanpa dedicated task"*; R-20 menempatkan penahannya pada
human approval. PR juga memberi hal yang tidak diberikan push: bukti bahwa
workflow benar-benar berjalan pada PR nyata — satu-satunya yang membedakan Phase 4
dari Phase 3.

### 2. Required check wajib berbentuk job yang selalu berjalan

Job `validate-changed-paths` **tidak boleh** memiliki `if:` pada level job.
Pemilahan branch terjadi di dalam step: step pertama menetapkan `mode`
(`agent`, `non-agent`, `merge-group`, `malformed`), step-step berikutnya dijaga
`if: steps.scope.outputs.mode == 'agent'` pada level **step**, dan step ringkasan
berjalan dengan `if: always()`.

Akibatnya job selalu melaporkan status. Branch non-agent **lolos secara
eksplisit** dengan pesan yang terbaca, bukan di-skip diam-diam. Branch yang mengaku
`agent/*` tetapi tidak mengikuti pola §44 ditolak dengan `mode=malformed` dan exit
bukan-nol — fail-closed dipertahankan.

Nama job `validate-changed-paths` adalah nama required check yang terpasang di
branch protection. Mengubahnya memutus proteksi tanpa gejala yang terlihat.

Aturan ini ditegakkan otomatis, bukan disandarkan pada kedisiplinan:
`make verify-github` menolak `if:` berindentasi empat spasi. Logikanya tinggal di
`tests/lib/check-github-artifacts.sh` dan dipakai bersama oleh target Makefile dan
`tests/negative/github-workflow.test.sh`, sehingga test negatif menguji penegak
yang sama alih-alih salinannya. Kasus regresi memakai bentuk Phase 3 yang asli.

### 3. Trigger `merge_group` dipasang sekarang; merge queue tidak diaktifkan

Workflow memuat `merge_group:` meski merge queue belum aktif. Docs GitHub:
*"you need to update the workflows to include the `merge_group` event as an
additional trigger. Otherwise, status checks will not be triggered when you add a
pull request to a merge queue. The merge will fail as the required status check
will not be reported."*

Merge queue sendiri **ditunda** — bukan karena biaya (Q16) melainkan karena ia
org-only pada setiap plan. Selama repo pilot berada di akun personal
`fajarcandraaa`, ia tidak tersedia. §60 butir "Merge queue" dan sandaran §29.8
padanya **ditimpa**: mitigasi urutan merge sementara bersandar pada reservasi path
(status `reserved-pending-merge`, Q12) dan penetapan urutan merge oleh TL/SA (§46).

Payload `merge_group` tidak memuat nama branch asal, sehingga task-id tak dapat
diturunkan di jalur itu. Validasi path sudah dijalankan pada event `pull_request`
untuk commit yang sama, jadi jalur `merge-group` melapor sukses. Ini batas yang
diketahui dan diterima, dicatat di `docs/operator/branch-protection.md`.

### 4. CODEOWNERS dipasang; `require_code_owner_review` tetap mati

Berkas CODEOWNERS dipasang karena ia lapis keempat mitigasi R-12 dan mitigasi
utama R-20 — satu-satunya lapis, bersama CI, yang menahan vektor PR. Cakupannya
mengikuti tuntutan R-20 (`.github/**`, bukan satu path) dan menyertakan
**kepemilikan atas dirinya sendiri** (`/.github/CODEOWNERS`); tanpa baris itu satu
PR dapat menghapus seluruh aturan di atasnya.

Setting `require_code_owner_review` **tidak diaktifkan** (lihat Konteks #5).
Aktivasi menunggu identitas `m2s-approver` (ADR-001 #5).

### 5. Required check diaktifkan hanya setelah ada bukti run hijau

Urutan wajib: pasang workflow → buka PR → amati satu run hijau → baru daftarkan
`validate-changed-paths` sebagai required check. ADR-001 sudah melarang urutan
sebaliknya. Menyalakan gate yang belum pernah melapor adalah cara paling langsung
mengunci repo dari luar.

### 6. Dua required approval tetap nol

`required_approving_review_count` tetap `0`. Hanya ada satu kolaborator dan GitHub
melarang self-approval, sehingga menaikkannya ke 2 memblokir seluruh merge. Ini
konsisten dengan #4 dan dengan status ADR-001 yang belum berlaku efektif.

### 7. D-03 diperbarui: Free cukup, dua jalur, Team dicoret

Opsi "upgrade org ke Team (~$4/seat/bulan)" dicabut dari D-03 sebagai jalur untuk
push restriction. Dua jalur yang tersisa, keduanya tanpa biaya:

| Jalur | Migrasi? | Membuka |
|---|---|---|
| Ruleset `restrict updates` + bypass list GitHub App | tidak | push/merge restriction |
| Pindah ke organization Free (`Mind2Screen-Dev-Team`) | ya | push restriction **dan** merge queue |

Dua hal yang dokumentasi tidak menjawab dicatat sebagai verifikasi empiris baru:
**V-06** apakah pemilik repo otomatis ter-bypass ruleset tanpa masuk bypass list,
dan **V-07** apakah role "Maintain" tersedia sebagai bypass actor di repo personal.
*(Dinomori ulang 1 Agustus 2026: V-05 sudah dipakai untuk perilaku `WorktreeCreate`.)*

> **Koreksi 1 Agustus 2026 — keputusan #7 di atas tidak sepenuhnya benar.**
> Uji empiris (repo `m2s-vsh-rules-probe`) menemukan **V-08**: `actor_type:
> Integration` **ditolak** di repo akun personal — `422 Actor GitHub Actions
> integration must be part of the ruleset source or owner organization`. Jadi
> ruleset di repo personal dapat membatasi manusia dan role, tetapi **tidak dapat
> memberi pengecualian kepada GitHub App**. Karena ADR-001 #5 memakai App, jalur
> "tanpa migrasi" **tidak cukup** untuk model dua identitas; itu menuntut
> organization.
>
> Yang tetap berlaku dari #7: opsi Team tetap dicabut (org **Free** cukup), dan
> V-06 menemukan sesuatu yang lebih kuat dari dugaan — ruleset **mengikat pemilik
> repo** (`GH013`, `current_user_can_bypass: "never"`), berbeda dari classic
> protection. Itu membuat ADR-001 **#4** dapat ditegakkan, hal yang classic
> protection tidak pernah bisa berikan bahkan di organization.
>
> Rincian: `docs/decisions/open-questions.md` D-03 dan
> `docs/operator/branch-protection.md`.

### 8. Phase 4 dinyatakan selesai-sebagian

Kriteria Done §60 (*"worker menghasilkan PR dan tidak dapat merge sendiri"*)
**tidak tercapai** pada fase ini. Paruh pertama tercapai; paruh kedua menuntut dua
GitHub App yang belum dibuat. Fase ditutup dengan pernyataan itu tertulis, bukan
dengan mengendurkan kriterianya.

Yang tercapai: PR template, CODEOWNERS, required check yang aman, review policy
terdokumentasi. Yang tidak: merge queue (#3), restriction siapa boleh merge (butuh
identitas kedua + ruleset), §66 #9 dan #11.

---

## Rasional

### Mengapa job selalu-jalan, bukan job kedua sebagai pelapor

Alternatifnya adalah membiarkan job ber-`if:` apa adanya lalu menambah job gate
yang `needs` job pertama dengan `if: always()`. Itu berfungsi, tetapi menambah
satu job yang keberadaannya hanya untuk mengakali bentuk job pertama, dan
`always()` yang digabung `needs` menuntut pemeriksaan eksplisit atas `result` —
lupa memeriksanya menghasilkan gate yang selalu hijau, yaitu R-24 lagi dengan
kemasan berbeda. Satu job yang selalu melapor lebih sedikit bagiannya dan
kegagalannya lebih sulit disalahpahami.

### Mengapa aturan bentuk ditegakkan Makefile, bukan dicatat di dokumen

Aturan "jangan pakai `if:` level job" adalah pengetahuan yang mudah hilang. Ia
tidak intuitif, tampak seperti optimasi yang wajar, dan konsekuensinya baru
terasa berbulan kemudian saat seseorang mengaktifkan required check. Repo ini
sudah memilih pola menegakkan hal semacam itu di Makefile (`verify-wrappers`
menahan wrapper yang menumbuhkan logika). `verify-github` mengikuti pola yang sama.

### Mengapa nama branch dilewatkan sebagai variabel env

Versi Phase 3 menulis `branch="${{ github.head_ref }}"` di dalam blok `run:`.
GitHub Actions mengganti ekspresi itu secara tekstual sebelum shell berjalan,
sehingga nama branch yang memuat metakarakter shell dieksekusi. Nama branch dapat
dipilih siapa pun yang dapat membuka PR. Melewatkannya lewat `env:` membuatnya
data, bukan kode. `verify-github` menahan bentuk lama.

### Mengapa CODEOWNERS dipasang meski tidak menjadi gate

Ia tetap memberi dua hal tanpa risiko deadlock: permintaan review otomatis yang
menandai siapa berwenang, dan jejak audit di UI. Dan ketika identitas kedua ada,
aktivasinya menjadi satu perubahan setting — bukan pekerjaan menulis berkas dari
nol pada saat yang paling sibuk.

---

## Konsekuensi

### Positif

- Required check untuk pertama kalinya benar-benar dapat dinyalakan tanpa
  mengunci repo. §66 #10 menjadi dapat diuji.
- Vektor PR pada R-12 dan R-20 memperoleh penahan pertamanya.
- D-03 kehilangan komponen biayanya. Tidak ada lagi alasan finansial yang
  menghalangi push restriction.
- Lubang script injection pada workflow Phase 3 tertutup, dan bentuknya ditahan
  otomatis agar tidak kembali.
- Klaim "CI live" yang tidak akurat terkoreksi dengan bukti terukur
  (`total_count`), bukan dengan pembacaan berkas.

### Negatif dan biaya

- Kriteria Done §60 tidak tercapai; fase ditutup selesai-sebagian.
- Merge queue tidak tersedia, sehingga mitigasi overlap semantik (R-01, §29.8)
  bersandar pada reservasi dan urutan merge manual — keduanya soft rule.
- Artefak ada di dua tempat (kanonik dan salinan repo aplikasi). Penyimpangan
  antar-salinan mungkin terjadi; `verify-github` memeriksa aturan bentuk pada
  keduanya tetapi tidak memeriksa kesamaan byte.
- Jalur `merge_group` melapor sukses tanpa memvalidasi ulang. Aman selama merge
  queue mati; harus ditinjau ulang saat diaktifkan.
- `.claude/settings.json` dan `Makefile` masuk daftar human-only write, sehingga
  dua bagian rencana ini (perluasan deny ke `.github/**` dan target
  `verify-github`) menuntut penerapan manusia. Ini bekerja sesuai desain, tetapi
  berarti fase tidak dapat ditutup tanpa satu langkah manual.

### Yang belum diputuskan

- Apakah repo pilot pindah ke `Mind2Screen-Dev-Team`. Ini satu-satunya jalan ke
  merge queue, dan menutup #3 sekaligus sebagian D-02.
- Apakah control repo memerlukan `develop`/`staging` (tercatat "belum
  dikerjakan" di open-questions). Untuk sekarang workflow control repo menargetkan
  `main`.
- V-06, V-07, V-08 **sudah diuji** 1 Agustus 2026 — lihat koreksi pada keputusan #7.
  Yang tersisa: ketersediaan role di **picker UI** (bagian visual V-07), dan apakah
  migrasi ke `Mind2Screen-Dev-Team` diambil supaya App dapat menjadi bypass actor.
- **T-04** — `worktree-lifecycle.sh` tidak terdaftar di `settings.json`, sehingga
  mitigasi R-26 dan §42.6 belum berjalan. Ditemukan saat Phase 4; menunggu keputusan
  pasang-atau-catat (`docs/operator/phase-4-human-only-patches.md` §3).
- Apakah `verify-github` juga memeriksa kesamaan byte antara artefak kanonik dan
  salinannya di repo aplikasi. Menuntut akses jaringan dari Makefile, yang belum
  pernah dilakukan target mana pun.

---

## Alternatif yang ditolak

### Mengaktifkan required check bersamaan dengan pemasangan workflow

Menghemat satu langkah. Ditolak: bila workflow ternyata gagal karena sebab yang
tidak terduga di lingkungan CI, repo terkunci dan pemulihannya menuntut mematikan
proteksi — tepat keadaan yang ADR-001 larang.

### Mengaktifkan `require_code_owner_review` sekarang

Terlihat memperkuat R-12. Ditolak: menghasilkan deadlock mutlak karena pemilik
tunggal adalah author. Gate yang mustahil dipenuhi bukan gate.

### Mengaktifkan merge queue dengan memindahkan repo ke organization pada fase ini

Menutup §60 secara utuh. Ditolak: migrasi kepemilikan repo mengubah URL remote,
identitas, dan seluruh proteksi yang sudah terpasang. Itu keputusan arsitektur
dengan permukaan yang jauh lebih luas dari §60, dan menggabungkannya ke fase ini
mencampur dua perubahan besar dalam satu langkah yang sulit dibatalkan sebagian.

### Menyalin aturan bentuk ke Makefile dan ke test negatif secara terpisah

Lebih sedikit berkas. Ditolak: test negatif kemudian menguji salinannya sendiri,
bukan penegak yang dipakai `make verify`. Keduanya dapat menyimpang tanpa satu
test pun berubah warna.

### Mencatat penundaan merge queue di open-questions saja, tanpa ADR

Lebih ringan. Ditolak: §60, §29.8, §13.6, §32 dan `path-overlap-matrix.md`
seluruhnya menyandarkan diri pada merge queue. Menimpa lima tempat dalam dokumen
arsitektur adalah simpangan yang konvensi repo ini tuntut dicatat sebagai ADR.

---

## Rollback

1. Hapus tiga artefak dari `templates/github/` dan salinannya di repo aplikasi.
2. Cabut `validate-changed-paths` dari `required_status_checks.contexts` pada
   `develop` dan `staging` kedua repo aplikasi. **Lakukan ini lebih dulu bila
   proteksi sudah aktif** — mencabut workflow tanpa mencabut required check
   meninggalkan gate yang tidak pernah melapor, yaitu kunci permanen.
3. Kembalikan `verify: check verify-wrappers verify-schemas verify-agents
   verify-hooks` dan hapus target `verify-github`.
4. Hapus `tests/lib/check-github-artifacts.sh` dan
   `tests/negative/github-workflow.test.sh`.
5. Kembalikan status D-03 ke bentuk sebelum ADR ini bila keputusan #7 dibatalkan.

Penegakan kembali ke keadaan Phase 3: hook lokal saja, dengan CI yang berkas-nya
ada tetapi jalurnya mati. Batas path kembali sepenuhnya dapat dielakkan lewat Bash
(R-07).

---

## Dampak pada dokumen arsitektur

| Bagian | Dampak |
|---|---|
| §60 butir "Merge queue" | **ditimpa** — ditunda, sebab org-only (bukan biaya seperti Q16 menyatakan) |
| §29.8 | **ditimpa sebagian** — serialisasi merge queue tidak tersedia; mitigasi bersandar reservasi + §46 |
| §13.6, §32 butir merge queue | sama, konsekuensi turunan |
| §60 butir "Required checks" | diperjelas — wajib berbentuk job selalu-jalan, dengan `merge_group` |
| §60 kriteria Done | dinyatakan **tercapai sebagian** |
| §66 #10 | menjadi dapat diuji |
| §66 #9, #11 | tetap tidak dapat diuji (identitas kedua belum ada) |
| Q16 | alasan penundaan dikoreksi: bukan plan berbayar, melainkan org-only |
| D-03 | opsi Team dicabut; dua jalur tanpa biaya dicatat |
| `path-overlap-matrix.md` | mitigasi "konflik urutan merge" perlu dibaca bersama ADR ini |

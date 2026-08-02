# ADR-008 — Migrasi Kepemilikan Repo ke Organization

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 1 Agustus 2026 |
| **Pemutus** | Pemilik arsitektur (Mindtoscreen) |
| **Versi arsitektur** | 0.1.0 |
| **Menimpa** | — |
| **Terkait** | ADR-001, ADR-007, D-02, D-03, V-06, V-07, V-08, R-12, R-20, §60, §66 |

---

## Konteks

Phase 4 (§60) selesai **sebagian** (ADR-007 #8). Kriteria Done-nya — *"worker
menghasilkan PR dan tidak dapat merge sendiri"* — paruh keduanya belum tercapai:
butuh dua identitas GitHub yang dipisah (ADR-001 #3), dan ADR-001 #5 menetapkan
bentuknya GitHub App, bukan machine user.

Verifikasi empiris V-08 menemukan bahwa jalur yang direncanakan **mustahil di
tempat repo berada sekarang**:

```
PUT /repos/fajarcandraaa/m2s-vsh-rules-probe/rulesets/20155906
422 Actor GitHub Actions integration must be part of the
    ruleset source or owner organization
```

Riset dokumentasi sebelumnya (ADR-007 #7) menyimpulkan hanya `OrganizationAdmin`
yang tak berlaku di repo akun personal. Uji API membantahnya: `actor_type:
Integration` (GitHub App) **juga ditolak**, karena App harus menjadi bagian dari
*owner organization*. Jadi model dua identitas berbasis App menuntut repo milik
organization, bukan sekadar preferensi — ini prasyarat teknis.

Empat blocker yang dibuka oleh kepemilikan organization:

| Blocker | Terbuka karena |
|---|---|
| GitHub App sebagai ruleset bypass actor (ADR-001 #5) | App bagian dari owner organization |
| Role granular + classic push restrictions | keduanya org-only |
| Merge queue (§60 butir kelima) | *"available in any public repository owned by an organization"* |
| Cabut hak merge dari worker | enforcement di sisi GitHub, bukan frontmatter |

Opsi jalur lain telah dievaluasi dan ditolak (lihat Alternatif): dua machine
account, fork ke org, manusia kedua.

## Keputusan

### 1. Tiga repo pindah ke organization `Mind2Screen-Dev-Team`

`m2s-vsh-platform` (control), `m2s-vsh-project-backend`, dan
`m2s-vsh-project-frontend` di-transfer dari akun personal `fajarcandraaa` ke org
`Mind2Screen-Dev-Team`. Ketiganya — bukan hanya dua — karena workflow di repo
aplikasi melakukan `actions/checkout` atas control repo
(`repository: fajarcandraaa/m2s-vsh-platform`); kalau control tinggal, ref menjadi
lintas-owner: rapuh, dan audit bercabang dua owner.

Org sudah ada (`Mind2Screen-Dev-Team`, plan **free**, 14 seat, dibuat 2024),
`members_can_create_public: true`, dan nama ketiga repo **tidak bentrok** di org
(terverifikasi API: 404 = belum ada). Biaya migrasi: **nol** — org Free cukup
untuk repo public (diverifikasi ADR-007 #7).

### 2. Mekanisme: transfer, bukan fork

Repo **di-transfer**, bukan di-fork. Fork menimbulkan dua jebakan yang tidak ada
di transfer:

- **PR default salah-arah**: PR dari branch dalam fork secara default menunjuk ke
  *parent/upstream*, bukan ke fork itu sendiri — bisa mengirim seluruh PR agent ke
  repo personal yang seharusnya ditinggalkan.
- **Fork network berbagi objek**: commit dalam satu jaringan fork terjangkau dari
  repo lain dalam jaringan yang sama; implikasi keamanan untuk repo yang menampung
  secret-handling perlu dinilai.

Transfer menjaga riwayat, PR, issue, dan star; URL lama di-redirect otomatis; dan
tidak menyisakan jaringan fork.

### 3. Referensi hardcoded diganti ke lokasi baru

Audit menemukan referensi `fajarcandraaa/` di beberapa lapis:

| Kelas | File | Efek kalau tidak diganti |
|---|---|---|
| **Kritis** | `templates/github/workflows/path-enforcement.yml` + 6 salinan ter-merge di repo aplikasi | CI putus — checkout control repo menunjuk repo lama |
| **Identitas kanonik** | `go.mod` (module path) + 5 import Go di `cmd/m2s/`, `internal/` | build lokal tetap jalan (ADR-004 #5) tapi identitas salah |
| **Kosmetik** | pr_url fixture, handoff example, ~55 dokumen | tak fungsional; auto-redirect |

Semua diganti ke `Mind2Screen-Dev-Team/...`. Pengecualian: fixture **CODEOWNERS di
test negatif** (`@fajarcandraaa`) — itu data uji, bukan target.

### 4. Urutan: persiapkan dulu, transfer belakangan

PR persiapan (ganti ref + ADR ini + runbook) di-merge **sebelum** transfer. Window
di mana CI menunjuk repo lama diminimalkan. Transfer adalah aksi manusia (butuh
hak owner org), bukan agent.

### 5. Penegakan ADR-001 pasca-transfer

Setelah org aktif: buat GitHub App `m2s-worker` + `m2s-approver`, pasang ruleset
`restrict updates` dengan bypass `actor_type: Integration` per App. Ini melengkapi
keputusan #3 dan #4 ADR-001 yang selama ini tidak dapat ditegakkan (D-03).

---

## Rasional

### Mengapa transfer, bukan fork (rincian)

Fork memberi kepemilikan org dengan biaya "tidak perlu memindahkan", tapi dua
jebakan di atas berbanding jauh dengan manfaatnya. Transfer adalah operasi resmi
yang menahan seluruh riwayat; fork adalah salinan yang berisiko mis-arah dan
berjaringan. Untuk repo yang baru mulai (backend/frontend nyaris kosong), biaya
transfer juga rendah.

### Mengapa go.mod ikut diganti, padahal build lokal tak peduli

Module path adalah identitas kanonik publik. Membiarkannya menunjuk akun lama
setelah repo pindah menciptakan dua nama untuk satu proyek — membingungkan
pembaca dan tooling. Ini koreksi identitas, bukan perubahan fungsional.

### Mengapa test negatif CODEOWNERS tak disentuh

`@fajarcandraaa` di `tests/negative/github-workflow.test.sh` adalah **fixture
kontrol negatif** — konten yang sengaja dimasukkan untuk diuji, bukan alamat
target. Menggantinya mengubah semantik test, bukan referensi migrasi. Pemilik
CODEOWNERS kanonik tetap manusia; alamat username pemilik bisa berganti saat
pemilik berganti, dan itu keputusan terpisah.

---

## Konsekuensi

### Positif

- ADR-001 #5 dapat ditegakkan; §66 #9 ("implementer tidak merge PR sendiri")
  menjadi dapat diuji.
- Merge queue tersedia (§60 butir kelima, §29.8).
- Lapisan anti-overlap #7 dan #8 keluar dari soft rule.
- Repo milik org memberi role granular dan audit yang lebih jelas.

### Negatif dan biaya

- URL berubah: `fajarcandraaa/*` → `Mind2Screen-Dev-Team/*`. Remote lokal
  pengguna perlu diperbarui (GitHub redirect otomatis membantu transisi).
- Window CI-putus bila urutan dilanggar (transfer sebelum ref diganti).
- Merujuk nama org di go.mod mengekspos identitas org ke publik — sesuai,
  karena repo memang milik org.

### Yang belum diputuskan

- Detail mekanik transfer — apakah branch protection, ruleset, dan required check
  **ikut** saat transfer, atau perlu didaftar ulang. **Belum diverifikasi ke
  dokumentasi** (riset web sempat terbatas); rencana memuat langkah verifikasi
  pasca-transfer sebagai wajib, bukan asumsi.
- Apakah GitGuardian (app_id 46505) ikut repos hasil transfer, atau perlu
  install ulang di org.
- Peran team vs manusia sebagai CODEOWNERS di org.

---

## Alternatif yang ditolak

### Dua machine account sebagai kolaborator

**Ditolak** — melanggar ToS dan tidak memisahkan hak. GitHub ToS §B.3:
*"You may maintain no more than one free machine account in addition to your free
Personal Account."* Dua akun otomasi gratis = pelanggaran eksplisit. Lebih
mendasar: repo akun personal hanya punya **dua permission level** — owner dan
collaborator-write — tanpa Read/Triage/Write/Maintain/Admin (itu skema
organization). Jadi worker dan approver mendapat hak **identik: write**, dan
kolaborator write otomatis boleh *"Create, merge, and close pull requests."*
Pemisahan worker/approver tidak dapat ditegakkan lewat permission, dan satu-satunya
penahan yang tersisa (`actor_type: User` di bypass list) tidak tercantum di daftar
eligible-actor dokumentasi — perilaku tak terdokumentasi, bisa berubah.

### Fork ke organization

**Ditolak** — lihat Keputusan #2 dan Rasional. Dua jebakan (PR salah-arah, fork
network) lebih mahal daripada operasi transfer yang resmi.

### Manusia kedua sebagai approver

**Ditolak oleh pemilik arsitektur** sejak ADR-001 — setiap merge ke `develop`
menunggu ketersediaan manusia, bertentangan dengan tujuan otomasi.

---

## Rollback

Transfer bersifat reversibel (transfer balik ke akun personal), tapi berisik:
remote pengguna sudah menunjuk URL baru, dan referensi hardcoded sudah diganti.
Rencana sebaliknya tidak dianjurkan.

Praktis: bila migrasi gagal di tengah, transfer-balik ketiga repo, pulihkan ref ke
`fajarcandraaa/` (PR terbalik), verifikasi CI. Harga: satu siklus penuh PR +
merge. Ini sebabnya verifikasi pasca-transfer adalah langkah wajib, bukan opsional.

---

## Dampak pada dokumen arsitektur

| Bagian | Dampak |
|---|---|
| §60 butir "Merge queue" | terbuka (ditunda di ADR-007 karena org-only) |
| ADR-001 prasyarat #3, #4, #5 | dapat dipenuhi setelah org aktif |
| D-03 | jalur org tersedia; keputusan migrasi diambil ADR ini |
| D-02 | tetap terbuka — repo klien private butuh plan Team, di luar cakupan |

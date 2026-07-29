# ADR-001 — Kewenangan Approval dan Merge bagi Agent

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 29 Juli 2026 |
| **Pemutus** | Pemilik arsitektur (Mindtoscreen) |
| **Versi arsitektur** | 0.1.0 |
| **Menimpa** | §17.6, §18.5 dokumen arsitektur |
| **Terkait** | R-06, R-20, ADR-002 |

---

## Konteks

Dokumen arsitektur v0.1.0 melarang PM Agent dan TL/SA Agent melakukan merge PR:

- §17.6 — PM Agent **prohibited**: "merge PR"
- §18.5 — TL/SA Agent **prohibited**: "merge PR"

Larangan ini berasal dari prinsip #6: *"Agent pembuat perubahan tidak boleh
menyetujui hasil kerjanya sendiri"*, dan lapisan anti-overlap #7 (§29.7):
*"Implementer tidak boleh menjadi reviewer task sendiri."*

Namun pemilik arsitektur menetapkan model operasional yang berbeda: TL/SA dan PM
adalah **agent**, dan keduanya berwenang melakukan approval serta merge ke branch
selain `main`. Manusia menjadi approver dan merger khusus untuk `main`.

Perlu diputuskan apakah model ini dapat ditegakkan, dan bagaimana caranya.

---

## Kendala teknis yang mengikat

Verifikasi menemukan satu batasan platform yang tidak dapat dinegosiasikan:

> **GitHub melarang author menyetujui pull request-nya sendiri, dan aturan itu
> melekat pada AKUN, bukan pada token.**

Konsekuensinya, bila seluruh agent memakai satu identitas GitHub:

```
backend-engineer  → membuka PR   sebagai akun X
tl-sa agent       → approve      sebagai akun X   ❌ ditolak GitHub
```

Alur yang diinginkan **tidak dapat berjalan sama sekali** dengan satu identitas.
Ini bukan persoalan kebijakan yang dapat dilonggarkan, melainkan batas platform.

Temuan kedua: pada plan GitHub Free, branch protection dan rulesets hanya tersedia
untuk repository **public**. Tanpa keduanya, pembatasan hak merge tidak dapat
ditegakkan sama sekali.

---

## Keputusan

**1. TL/SA Agent dan PM Agent berwenang melakukan approval dan merge ke branch
selain `main`.** §17.6 dan §18.5 ditimpa sejauh menyangkut branch `develop` dan
`staging`.

**2. Merge ke `main` tetap sepenuhnya milik manusia.** Tidak ada agent yang memiliki
hak apa pun atas `main`.

**3. Model dua identitas GitHub ditetapkan sebagai prasyarat.**

| Identitas | Role pemakai | Hak | Larangan |
|---|---|---|---|
| `m2s-worker` | Backend, Frontend, QA, DevOps, Technical Writer, Code Reviewer | push branch `agent/*`, create PR | approve, merge, akses `main` |
| `m2s-approver` | TL/SA, PM | approve PR, merge ke `develop`/`staging` | merge ke `main`, mengubah branch protection |
| Manusia | Pemilik arsitektur | approve + merge `main`, ubah proteksi | — |

**4. Tidak ada identitas agent yang memiliki scope admin repository.** Ini menutup
R-20 — branch protection hanya dapat diubah lewat UI/API oleh manusia.

**5. Implementasi identitas memakai GitHub App**, bukan machine user. Alasan:
tidak memakan seat, token berumur pendek, permission granular per instalasi.

---

## Rasional

### Mengapa prinsip #6 tetap terjaga

Prinsip #6 melarang agent menyetujui **hasil kerjanya sendiri**. Keputusan ini tidak
melanggarnya, karena:

- Agent yang menulis kode memakai identitas `m2s-worker`
- Agent yang menyetujui memakai identitas `m2s-approver`
- Keduanya adalah akun berbeda, sehingga GitHub menegakkan pemisahannya

Yang berubah bukan prinsipnya, melainkan **siapa yang menjalankan peran approver** —
dari manusia menjadi agent dengan identitas terpisah.

### Mengapa ini justru lebih kuat daripada honor system

Sebelum keputusan ini, §17.6 dan §18.5 adalah larangan tekstual di definisi agent —
yaitu **soft rule**. Dokumen arsitektur sendiri (prinsip #7) menyatakan prompt bukan
security boundary.

Dengan model dua identitas, larangan tersebut menjadi **hard rule**: implementer
secara teknis **tidak memiliki hak** untuk approve maupun merge. Tidak ada instruksi
yang dapat membuatnya melakukan hal itu.

### Mengapa `main` tetap milik manusia

§13.1 dan §50 menempatkan production release dan keputusan irreversible pada manusia.
Merge ke `main` adalah titik irreversible tersebut. Tidak ada alasan operasional yang
cukup kuat untuk memindahkannya ke agent pada v0.1.0.

---

## Konsekuensi

### Positif

- Lapisan anti-overlap #7 menjadi **ditegakkan platform**, bukan dijanjikan prompt
- Acceptance §66 #9 ("Implementer tidak merge PR sendiri") menjadi **dapat diuji**
- R-06 tertutup secara struktural
- Alur `develop` dapat berjalan tanpa menunggu ketersediaan manusia

### Negatif dan biaya

- Menambah kompleksitas setup: dua GitHub App, dua kredensial, dua konfigurasi runner
- Runner wajib memilih identitas yang benar berdasarkan role — kesalahan pemilihan
  akan menggagalkan alur atau, lebih buruk, memberi hak berlebih
- Kredensial `m2s-approver` menjadi aset bernilai tinggi. Kebocorannya berarti
  kemampuan merge otomatis ke `develop`
- Agent kini dapat me-merge kode yang salah ke `develop` tanpa mata manusia.
  Mitigasinya adalah required status checks — CI tetap menjadi gate deterministik

### Prasyarat yang wajib dipenuhi sebelum keputusan ini aktif

1. Repository berstatus public, atau organization di-upgrade ke GitHub Team
2. Branch protection aktif pada `main`, `staging`, dan `develop`
3. `main` — hanya manusia, agent tidak masuk daftar yang boleh push/merge
4. `develop`/`staging` — required status checks + 2 approval, `m2s-worker` tidak
   boleh merge
5. Kedua GitHub App terpasang dengan permission minimum

**Sampai kelima prasyarat terpenuhi, keputusan ini belum berlaku efektif** dan
seluruh merge dilakukan manusia.

---

## Alternatif yang ditolak

### Satu identitas untuk semua agent

**Ditolak** — secara teknis tidak dapat berjalan. GitHub menolak self-approval pada
level akun, sehingga TL/PM agent tidak akan dapat menyetujui PR yang dibuat agent
lain dengan akun yang sama.

### TL dan PM sebagai manusia

**Ditolak oleh pemilik arsitektur.** Secara keamanan ini opsi terkuat dan tidak
memerlukan identitas tambahan, tetapi bertentangan dengan tujuan otomasi:
setiap merge ke `develop` akan menunggu ketersediaan manusia.

### Sembilan identitas, satu per role

**Ditolak** — biaya operasional tidak sebanding. Pemisahan yang benar-benar
dibutuhkan hanyalah antara *yang menulis* dan *yang menyetujui*. Granularitas
lebih halus dapat ditambahkan pada v0.2.0 bila audit menunjukkan kebutuhannya.

---

## Rollback

Bila model ini bermasalah:

1. Cabut permission merge dari GitHub App `m2s-approver`
2. Kembalikan `develop` ke required human approval
3. Tandai ADR ini `Superseded`, buat ADR pengganti
4. Definisi agent TL/SA dan PM tidak perlu diubah — kewenangan ditegakkan di sisi
   GitHub, bukan di frontmatter

Rollback tidak memerlukan perubahan kode maupun konfigurasi Claude Code.

---

## Dampak pada dokumen arsitektur

Sesuai §70, perubahan arsitektur wajib mencatat:

| Aspek | Isi |
|---|---|
| **Reason** | Model operasional menghendaki otomasi merge ke `develop` tanpa menunggu manusia |
| **Affected roles** | PM Agent, TL/SA Agent, seluruh implementer, Code Reviewer |
| **Migration steps** | Buat 2 GitHub App → pasang branch protection → konfigurasi runner memilih identitas per role |
| **Backward compatibility** | Tidak ada — v0.1.0 belum pernah dirilis |
| **Rollback** | Lihat bagian di atas |
| **Versi** | Tetap 0.1.0 (pra-rilis, amandemen baseline) |

Klasifikasi §69: perubahan responsibility boundary umumnya berarti **MAJOR**.
Karena v0.1.0 belum dirilis, ini diperlakukan sebagai amandemen baseline, bukan
kenaikan versi.

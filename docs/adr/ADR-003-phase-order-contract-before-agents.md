# ADR-003 — Task Contract Dikerjakan Sebelum Core Agents

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 30 Juli 2026 |
| **Pemutus** | Pemilik arsitektur (Mindtoscreen) |
| **Versi arsitektur** | 0.1.0 |
| **Menimpa** | §57, §58 dokumen arsitektur |
| **Terkait** | §70, R-04, D-03, ADR-001 |

---

## Konteks

Roadmap §56–§64 menetapkan urutan:

```
Phase 1 (§57)  Core Agents            — 9 project agent
Phase 2 (§58)  Task Contract & Runner — schema, validasi, reservasi, launcher
```

Definisi agent pada `.claude/agents/**` (§37) memuat batas tool dan batas path yang
dipakai agent saat mengeksekusi task. Batas path itu merujuk bentuk yang ditetapkan
`task.schema.json` — artefak yang menurut §58 baru dibuat pada fase berikutnya.

Urutan ini menempatkan konsumen kontrak sebelum kontraknya ada.

### Ketergantungan yang teramati

| Artefak §57 | Bergantung pada | Sumber |
|---|---|---|
| `allowed_paths` / `forbidden_paths` per agent | bentuk field pada `task.schema.json` | §34, §37 |
| Field `shared_file_ownership` | `task.schema.json` | R-04 |
| Batas tool PM (`Bash` → `scripts/<runner>.sh`) | nama runner script yang dibuat §58 | Q11 |
| Kosakata status dan nama role | enum lintas schema | §34 |

Empat dari sembilan agent (PM, TL/SA, Backend, Frontend) memuat setidaknya satu
dari ketergantungan di atas.

### Konsekuensi bila urutan dipertahankan

Kesembilan definisi agent ditulis terhadap kontrak yang belum ada, lalu ditinjau
ulang setelah §58 selesai. Biaya penulisan ulang menyentuh 9 file pada 2 repository,
dan setiap perubahan memerlukan pull request karena branch protection telah aktif.

Lebih penting: kriteria **Done** §57 — *"setiap agent menunjukkan tool dan path
boundary yang berbeda"* — tidak dapat dibuktikan tanpa kosakata path yang stabil.
Fase akan ditutup atas dasar dokumen yang masih berubah.

---

## Keputusan

**1. Urutan Phase 1 dan Phase 2 ditukar.**

| Fase | Isi | Sumber |
|---|---|---|
| Phase 1 | Task Contract dan Runner | §58 |
| Phase 2 | Core Agents | §57 |

**2. Isi kedua fase tidak berubah.** Yang ditukar hanya urutan pengerjaan. Seluruh
butir §57 dan §58, termasuk kriteria **Done** masing-masing, tetap berlaku apa adanya.

**3. Penomoran fase lain tidak bergeser.** Phase 0 (§56) dan Phase 3–8 (§59–§64)
tetap pada nomornya. Hanya §57 dan §58 yang bertukar posisi.

**4. Rujukan wajib menyertakan nomor §.** Karena nomor fase kini tidak lagi sejajar
dengan urutan §, setiap rujukan fase pada dokumen ditulis sebagai `Phase 1 (§58)`
agar tidak ambigu.

---

## Rasional

### Mengapa kontrak lebih dulu

Task contract adalah antarmuka antara runner dan agent. Menulis konsumen sebelum
antarmukanya ada berarti menebak bentuk antarmuka tersebut. Menukar urutan
menghilangkan tebakan itu, bukan menghilangkan pekerjaannya.

### Mengapa ini konsisten dengan Q19

Q19 menetapkan *"yang bertahap adalah urutan pengerjaan, bukan cakupannya."*
Keputusan ini mengubah urutan dan tidak mengubah cakupan — kesembilan agent tetap
dibuat, seluruh artefak §58 tetap dibuat.

Perlu dicatat: kalimat Q19 tersebut pernah dipakai sebagai pembenaran menata ulang
penomoran fase secara diam-diam pada README, tanpa ADR. Itu keliru dan telah
dikoreksi pada commit `9955144`. ADR ini adalah cara yang benar untuk melakukan
perubahan yang sama — terbuka, menyebut bagian yang ditimpa, dan dapat diaudit.

### Mengapa bukan menggabungkan kedua fase

Menggabungkan menghilangkan checkpoint manusia di antara keduanya. §56–§64
menempatkan checkpoint pada setiap batas fase, dan Q19 memberi alasannya: bila
banyak komponen dinyalakan bersamaan, tidak ada cara mengisolasi komponen mana
yang rusak saat acceptance test gagal.

---

## Konsekuensi

### Positif

- Definisi agent ditulis terhadap kontrak yang sudah tetap
- Kriteria Done §57 dapat diuji dengan kosakata path yang stabil
- R-04 (`shared_file_ownership`) tertangani pada fase pertama, bukan menggantung
- Mengurangi pull request revisi pada `.claude/agents/**` di dua repository

### Negatif dan biaya

- Fase pertama tidak menghasilkan agent yang dapat dijalankan. Bukti kemajuan yang
  terlihat berkurang dibanding urutan semula
- Runner dan schema diuji tanpa agent nyata sebagai konsumennya. Sebagian asumsi
  bentuk kontrak baru terbukti benar atau salah pada fase berikutnya
- Nomor fase tidak lagi sejajar dengan nomor §, sehingga rujukan tanpa `§` menjadi
  ambigu. Dimitigasi oleh keputusan #4

### Yang tidak berubah

- Cakupan v0.1.0 (§2) — tetap utuh, sesuai Q19
- Kriteria Done kedua fase
- Checkpoint manusia pada setiap batas fase
- ADR-001 tetap belum berlaku efektif; D-03 tidak terpengaruh keputusan ini

---

## Alternatif yang ditolak

### Mempertahankan §57 apa adanya

**Ditolak.** Secara formal paling aman karena tidak menyimpang dari baseline, tetapi
memindahkan biaya ke penulisan ulang 9 file dan menutup fase atas dasar dokumen
yang belum stabil.

### Menggabungkan §57 dan §58 menjadi satu fase

**Ditolak** — menghilangkan checkpoint manusia di antara keduanya, bertentangan
dengan alasan bertahap pada Q19.

### Membuat subset agent lebih dulu, sisanya setelah kontrak

**Ditolak untuk v0.1.0.** Menghasilkan dua kelas agent dengan tingkat kematangan
berbeda dalam satu roadmap, dan tetap menanggung revisi pada subset pertama.
Dapat ditinjau ulang bila Phase 1 (§58) berjalan lebih lama dari perkiraan.

---

## Rollback

1. Kembalikan tabel roadmap README ke urutan §57 lalu §58
2. Tandai ADR ini `Superseded`
3. Artefak yang sudah dibuat tidak perlu dibuang — hanya urutannya yang kembali

Rollback tidak memerlukan perubahan kode. Tidak ada artefak yang menjadi tidak sah
akibat pembalikan urutan.

---

## Dampak pada dokumen arsitektur

Sesuai §70:

| Aspek | Isi |
|---|---|
| **Reason** | Definisi agent bergantung pada bentuk task contract yang menurut urutan semula baru dibuat pada fase berikutnya |
| **Affected roles** | Seluruh 9 project agent — definisinya bergeser satu fase |
| **Migration steps** | Perbarui tabel roadmap README; kerjakan §58 sebagai Phase 1; kerjakan §57 sebagai Phase 2 |
| **Backward compatibility** | Tidak ada dampak — belum ada artefak §57 maupun §58 yang dibuat |
| **Rollback** | Lihat bagian di atas |
| **Versi** | Tetap 0.1.0 |

Klasifikasi §69: perubahan ini tidak menyentuh control model, orchestrator,
responsibility boundary, maupun execution runtime — hanya urutan pengerjaan.
Diperlakukan sebagai amandemen baseline pra-rilis, bukan kenaikan versi.

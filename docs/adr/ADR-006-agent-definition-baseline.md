# ADR-006 — Baseline Definisi Agent

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 31 Juli 2026 |
| **Pemutus** | Pemilik arsitektur (Mindtoscreen) |
| **Versi arsitektur** | 0.1.0 |
| **Menimpa** | Appendix A (Recommended Role Frontmatter Baseline), §57 bersama ADR-005 |
| **Terkait** | ADR-005, Q9, Q10, Q11, Q14, A-03, §15, §37, §38, §69, §70 |

---

## Konteks

Phase 2 (§57) menghasilkan definisi agent. Kriteria Done-nya satu kalimat:
*"setiap agent menunjukkan tool dan path boundary yang berbeda."*

Satu-satunya panduan bentuk definisi agent adalah **Appendix A — Recommended Role
Frontmatter Baseline**. Appendix itu tidak lagi memadai, dan pada satu titik
justru menyesatkan:

**1. Ia memuat tiga contoh, bukan tiga belas.** `backend-engineer`,
`code-reviewer`, dan `project-manager` diberikan sebagai contoh; sepuluh role
sisanya tidak. ADR-005 menaikkan jumlah role dari 9 menjadi 13 tanpa menambah
contoh frontmatter.

**2. Frontmatter `project-manager` di sana memuat tool `Agent`** — bertentangan
langsung dengan **Q11**, yang mencabutnya dengan alasan role dijalankan runner
sebagai sesi top-level terpisah, bukan sebagai nested subagent. Teks Appendix A
yang dibaca apa adanya akan menghasilkan PM dengan kewenangan yang sudah
diputuskan dicabut.

**3. `effort` tidak disebut sama sekali.** `component-inventory.md` §6 mencatat
kesenjangan ini dan menjadwalkan penetapannya pada Phase 2, karena nilai tersebut
menjadi dasar pengukuran biaya token Phase 8 (§64).

Selain itu ADR-005 menyisakan tiga pertanyaan yang seluruhnya dijadwalkan tutup
pada Phase 2:

| # | Pertanyaan terbuka ADR-005 |
|---|---|
| 1 | Bagaimana task menyatakan prasyarat platform runner, misal iOS wajib macOS? |
| 2 | Apakah `mobile-engineer` dan `android-developer` boleh aktif bersamaan pada satu repo? |
| 3 | Distribusi role per repository — Q10 hanya mencakup sembilan role lama |

ADR ini menutup ketiganya sekaligus menetapkan baseline yang menggantikan
Appendix A.

---

## Keputusan

### 1. Baseline frontmatter untuk tiga belas role

Setiap role memiliki satu berkas `templates/agents/<role>.md`. Nama berkas sama
persis dengan nilai enum `role` pada `common.schema.json`, sesuai ADR-005 #4 dan
§37.

Tiga belas role dikelompokkan menjadi **tiga kelas boundary**. Kelas menentukan
`permissionMode` dan tool write; pembedaan lebih jauh terjadi pada writable path.

| Kelas | Role | `permissionMode` | Tool write |
|---|---|---|---|
| Read-only | `code-reviewer` | `plan` | tidak ada |
| Control-write | `project-manager`, `technical-lead-system-analyst` | `default` | `Edit`, `Write` — terbatas control repository |
| Worktree-write | sepuluh role sisanya | `default` | `Edit`, `Write` — pada worktree, dibatasi task contract |

**Tool `Agent` tidak diberikan kepada role mana pun.**

Q11 mencabutnya dari Project Manager. Alasan yang dipakai di sana tidak khusus PM:
role agent dijalankan runner sebagai sesi top-level terpisah, sehingga tidak ada
role yang perlu men-spawn siapa pun. Larangan itu karena itu diperluas menjadi
aturan tunggal yang berlaku bagi tiga belas role, dan dijaga
`TestNoAgentHasAgentTool`.

Field frontmatter yang boleh dipakai terbatas pada daftar terverifikasi
`component-inventory.md` §6: `name`, `description`, `tools`, `disallowedTools`,
`model`, `effort`, `permissionMode`, `background`, `isolation`, `maxTurns`,
`skills`, `hooks`, `mcpServers`, `color`. Field di luar itu ditolak
`TestAgentFrontmatterFieldsAreSupported`.

**`skills` dikosongkan pada v0.1.0.** Appendix A merujuk skill seperti
`project-backend-conventions` dan `review-rubric` yang belum pernah dibuat.
Frontmatter yang menyebut skill tidak ada akan gagal jauh dari sebabnya.
Pembuatan skill adalah pekerjaan Phase 5 (§61).

### 2. Nilai `effort` per role

Menutup kesenjangan yang dicatat `component-inventory.md` §6.

| `effort` | Role | Alasan |
|---|---|---|
| `high` | `project-manager`, `technical-lead-system-analyst`, `code-reviewer` | keputusan lintas komponen dan penilaian kualitas independen |
| `medium` | `ui-ux-designer`, `backend-engineer`, `frontend-engineer`, `qa-engineer`, `devops-release`, `fullstack-engineer`, `mobile-engineer`, `android-developer`, `ios-developer` | pekerjaan terarah di dalam batas yang sudah ditetapkan task contract |
| `low` | `technical-writer` | mengikuti implementasi yang sudah disetujui, bukan menentukannya |

Nilai ini adalah **baseline pengukuran token Phase 8 (§64)**. Perubahan sesudahnya
wajib dibandingkan terhadap angka ini, bukan ditetapkan ulang dari nol.

### 3. Prasyarat platform dinyatakan task contract, ditegakkan runner

Menutup pertanyaan terbuka ADR-005 #1.

Field opsional `execution.platform` ditambahkan pada `task.schema.json`:

```json
"platform": {
  "type": "string",
  "enum": ["any", "darwin", "linux"],
  "default": "any"
}
```

Runner memeriksanya pada `launch-task`, **sesudah** contract tervalidasi dan
**sebelum** worktree dibuat. Ketidakcocokan dengan `runtime.GOOS` menghasilkan
`exit 2` — kode yang sudah dipakai `main.go` untuk "kontrak ditolak", dibedakan
dari `exit 1` untuk "runner gagal berjalan".

Pemeriksaan dilakukan sebelum worktree dibuat agar penolakan tidak meninggalkan
worktree yatim yang harus dibersihkan manual.

Perubahan bersifat **aditif dan opsional**: `schema_version` tetap `1.0`, dan task
yang sudah ada tanpa field ini tetap sah dengan arti `any`.

**Mengapa runner, bukan frontmatter agent.** Frontmatter adalah prompt, dan
prinsip #3 menyatakan prompt bukan security boundary. Frontmatter tidak dapat
menolak eksekusi; runner dapat. Menaruh prasyarat pada task contract juga
menempatkannya di tempat yang sama dengan batas path — satu sumber untuk seluruh
prasyarat eksekusi.

**Mengapa enum, bukan string bebas.** Nilai di luar tiga ini tidak dapat
diverifikasi runner. `windows` sengaja tidak dimasukkan: tidak ada stack pada
scope v0.1.0 yang menuntutnya, dan enum lebih mudah ditambah daripada dikurangi.

### 4. `mobile-engineer` dan `android-developer` boleh aktif bersamaan pada satu repository

Menutup pertanyaan terbuka ADR-005 #2.

**Tidak ada aturan baru yang dibutuhkan.** Isolasi antar penulis dijamin
**reservasi path**, bukan batas role — persis yang K1 pada
`roles-extension-v0.1.0.md` tetapkan bagi repo fullstack. Alasan yang sama
berlaku bagi repo mobile.

Yang menentukan sah atau tidaknya adalah `internal/pathmatch`: dua task berjalan
paralel bila path-nya terpisah, dan ditolak begitu beririsan. Reservasi
`android/**` secara penuh oleh `mobile-engineer` **akan** berkonflik dengan task
`android-developer` pada repo yang sama — bukan karena rolenya berbeda, melainkan
karena path-nya beririsan.

Kewajiban yang timbul karena itu jatuh pada TL/SA saat menyusun task contract:
memecah path per platform, bukan mereservasi pohon `android/**` seluruhnya.
Catatan konflik §E2.6 sudah menyatakan hal ini; ADR ini menegaskan bahwa itulah
mekanisme yang berlaku, dan tidak ada larangan tambahan berbasis role.

### 5. Distribusi tiga belas role per repository

Menutup pertanyaan terbuka ADR-005 #3, memperluas Q10 dari sembilan role menjadi
tiga belas.

| Repository | Role yang di-deploy |
|---|---|
| control (`m2s-vsh-platform`) | `project-manager`, `technical-lead-system-analyst` |
| backend | `backend-engineer`, `qa-engineer`, `code-reviewer`, `devops-release`, `technical-writer` |
| frontend | `frontend-engineer`, `ui-ux-designer`, `qa-engineer`, `code-reviewer`, `devops-release`, `technical-writer` |
| fullstack (bentuk B) | `fullstack-engineer`, `qa-engineer`, `code-reviewer`, `devops-release`, `technical-writer` |
| mobile lintas platform (bentuk C) | `mobile-engineer`, `qa-engineer`, `code-reviewer`, `devops-release` |
| mobile native terpisah | `android-developer` atau `ios-developer`, `qa-engineer`, `code-reviewer`, `devops-release` |

Sesuai Q10, ini adalah mekanisme **least privilege**, bukan mekanisme paralelisme.
Paralelisme berasal dari proses dan worktree terpisah.

`code-reviewer` hadir pada seluruh repo aplikasi karena §47 menuntut review
terpisah dari implementasi. `devops-release` hadir pada seluruhnya karena
`.github/workflows/**` ada pada setiap repo.

**Pada Phase 2 hanya baris control yang dieksekusi.** Kelima baris lainnya menjadi
acuan Phase 7 (§63), ketika repo pilot benar-benar menerima definisi agent.
Template kanonik seluruh tiga belas role tetap ditulis sekarang, di control
repository, karena `component-inventory.md` §2 menempatkannya sebagai komponen
platform-global milik Human Workflow Maintainer.

---

## Rasional

### Mengapa template kanonik terpisah dari deployment

Definisi agent adalah komponen **platform-global** (`component-inventory.md` §2
butir 20), tetapi `.claude/agents/` yang aktif adalah komponen **project-specific**
yang hanya memuat subset relevan (Q10). Dua peran berbeda pada satu isi yang sama.

Menempatkan sumber kanonik pada `templates/agents/` dan salinan aktif pada
`.claude/agents/` membuat keduanya dapat dijaga konsisten oleh test, sekaligus
menjaga least privilege: repo backend tidak memuat definisi `ios-developer` yang
tidak pernah dipakainya.

Salinan dipilih, bukan symlink: Git tidak menjamin perilaku symlink lintas
platform, dan `TestDeployedAgentsMatchTemplates` menjaga keduanya tetap identik
tanpa bergantung pada mekanisme filesystem.

### Mengapa Appendix A ditimpa, bukan dilengkapi

Butir kedua pada Konteks bersifat menentukan: Appendix A memuat frontmatter yang
**bertentangan dengan keputusan yang sudah diambil**. Melengkapinya akan
menyisakan teks salah yang tetap terbaca berlaku — persis yang tabel "Bagian
dokumen arsitektur yang ditimpa ADR" pada `open-questions.md` ada untuk cegah.

### Mengapa boundary harus dapat diuji, bukan hanya terbaca

Kriteria Done §57 menuntut boundary yang berbeda. Menulis tiga belas berkas
serupa secara berurutan membuat penyalinan tanpa penyesuaian menjadi kegagalan
yang paling mungkin terjadi — dan kegagalan itu tidak terlihat saat dibaca, karena
setiap berkas tampak masuk akal sendiri-sendiri.

`TestAgentBoundariesAreDistinct` karena itu menuntut setiap pasangan role berbeda
pada sekurangnya satu dari: tool set, `permissionMode`, atau pola writable path.
Ia mengubah kriteria Done dari penilaian menjadi pemeriksaan.

---

## Konsekuensi

### Positif

- Ketiga pertanyaan terbuka ADR-005 tertutup pada fase yang dijadwalkan
- Prasyarat platform menjadi **ditegakkan runner**, bukan konvensi operator
- Pencabutan tool `Agent` berlaku seragam dan dapat diuji sebagai satu aturan
- Kriteria Done §57 dapat diperiksa, bukan dinilai
- Baseline `effort` tersedia sebelum pengukuran token Phase 8 dimulai

### Negatif dan biaya

- **Tiga belas berkas template menjadi permukaan pemeliharaan baru.** Perubahan
  §17–§25 atau §E1–§E4 kini menuntut pembaruan pada dua tempat. Dimitigasi dengan
  mewajibkan setiap template menyebut nomor § asalnya
- **Duplikasi antara `templates/agents/` dan `.claude/agents/`** dijaga test, bukan
  mekanisme filesystem. Test yang dihapus akan membuat keduanya menyimpang diam-diam
- **`execution.platform` menambah field yang harus diisi benar.** Task
  `ios-developer` yang lupa mencantumkan `darwin` akan berjalan pada runner yang
  salah dan gagal pada `quality_gates`, bukan pada `launch-task`. Penegakan
  hubungan role ⇄ platform belum ada
- **Enum `platform` tidak memuat `windows`.** Project yang menuntutnya akan tertolak
  schema dan memerlukan perubahan enum

### Yang belum diputuskan

| Pertanyaan | Kapan |
|---|---|
| Apakah `ios-developer` wajib menyiratkan `platform: darwin` secara otomatis? | Phase 3 (§59), bersama hook validation |
| Skill dan `.claude/rules/` yang dirujuk definisi agent | Phase 5 (§61) |
| Deployment sebelas template ke repo pilot | Phase 7 (§63) |

---

## Alternatif yang ditolak

### Menulis definisi agent langsung ke `.claude/agents/` control repo saja

**Ditolak.** Control repo hanya menjalankan dua role (Q10). Sebelas role sisanya
tidak akan memiliki tempat, sehingga kriteria Done §57 tidak dapat dibuktikan bagi
mayoritas role, dan `component-inventory.md` §2 butir 20 tetap tidak terpenuhi.

### Men-deploy seluruh tiga belas role ke ketiga repository sekarang

**Ditolak.** Melanggar least privilege Q10 dan menyentuh repo pilot sebelum hook
enforcement Phase 3 (§59) tersedia. Repo backend tidak memerlukan `ios-developer`.

### Menyatakan prasyarat platform pada frontmatter agent

**Ditolak.** Frontmatter adalah prompt. Prinsip #3 menyatakan prompt bukan security
boundary, dan frontmatter tidak memiliki jalan untuk menolak eksekusi. Prasyarat
yang tidak dapat menolak bukan prasyarat, melainkan catatan.

### Melarang `mobile-engineer` dan `android-developer` aktif bersamaan

**Ditolak.** Larangan berbasis role akan menduplikasi pekerjaan yang sudah
dilakukan deteksi overlap path, dan menolak kasus yang sebenarnya sah: satu task
pada `shared/mobile/**` dan satu lagi pada `android/app/**` tidak beririsan sama
sekali.

### Menetapkan `effort` seragam bagi seluruh role

**Ditolak.** Menghapus informasi yang justru dibutuhkan Phase 8. Biaya token
`technical-writer` dan `technical-lead-system-analyst` tidak sebanding, dan
menyamakannya membuat pengukuran kehilangan dasar pembanding.

---

## Rollback

1. Hapus `templates/agents/` dan `.claude/agents/*.md`
2. Hapus `execution.platform` dari `task.schema.json` dan pemeriksaannya dari
   `cmd/m2s/commands.go`
3. Hapus `internal/contract/agents_test.go` dan target `verify-agents`
4. Kembalikan status Phase 2 pada README menjadi belum selesai
5. Tandai ADR ini `Superseded`; ketiga pertanyaan ADR-005 kembali terbuka

Rollback aman selama belum ada task memakai `platform` selain `any`. Karena field
opsional dan default-nya `any`, task yang sudah ada tidak terdampak.

---

## Dampak pada dokumen arsitektur

Sesuai §70:

| Aspek | Isi |
|---|---|
| **Reason** | Appendix A memuat tiga contoh untuk tiga belas role, memberi PM tool `Agent` yang sudah dicabut Q11, dan tidak menyebut `effort` yang dijadwalkan ditetapkan pada fase ini |
| **Affected roles** | Seluruh tiga belas role — masing-masing memperoleh definisi, kelas boundary, dan nilai `effort` |
| **Migration steps** | Tulis ADR ini → tambah `execution.platform` dan penegakannya → tulis tiga belas template → deploy dua ke `.claude/agents/` control → tulis test boundary → perbarui dokumen turunan |
| **Backward compatibility** | Aditif. `schema_version` tetap `1.0`; task, reservasi, dan handoff yang sudah ada tetap sah |
| **Rollback** | Lihat bagian di atas |
| **Versi** | Tetap 0.1.0 |

Klasifikasi §69: penetapan batas tool per role menyentuh responsibility boundary,
yang umumnya MAJOR. Karena v0.1.0 belum dirilis, ia diperlakukan sebagai amandemen
baseline, bukan kenaikan versi.

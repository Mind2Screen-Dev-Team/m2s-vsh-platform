# Risk Register

**Tanggal:** 29 Juli 2026
**Versi arsitektur:** 0.1.0
**Cakupan:** overlap, privilege escalation, self-modification, cross-repository
write, dan perubahan shared file.

Severity: 🔴 kritis · 🟠 tinggi · 🟡 sedang · 🟢 rendah

---

## Ringkasan

| Kategori | Jumlah | Kritis | Tinggi |
|---|---|---|---|
| Overlap | 5 | 0 | 2 |
| Privilege escalation | 6 | 2 | 3 |
| Self-modification | 3 | 1 | 0 |
| Cross-repository write | 4 | 0 | 1 |
| Shared file | 4 | 1 | 1 |
| Lainnya | 5 | 0 | 1 |
| **Total** | **27** | **4** | **8** |

---

## A. Risiko overlap

### R-01 🟠 Semantic overlap meski path terisolasi

Dua task pada path berbeda dapat mengimplementasikan business rule yang sama.
Isolasi file tidak menangkapnya.

**Mitigasi:** traceability ID wajib (`REQ-*` → `BR-*` → `CONTRACT-*` → task),
coverage matrix, TL/SA readiness review sebagai gate keras, QA traceability matrix.

**Batas kejujuran:** §3 dokumen arsitektur menyatakan sendiri bahwa arsitektur
**tidak dapat menjamin** ketiadaan semantic overlap bila dekomposisi task salah.
Risiko ini dikurangi, bukan dihilangkan.

---

### R-02 🟠 Celah reservasi antara PR dan merge

Reservasi dilepas saat PR dibuat, padahal PR belum tergabung. Task kedua dapat
mereservasi path yang sama dan mengedit basis yang sudah berubah.

**Mitigasi:** ✅ **tertutup** — reservasi ditahan sampai merge (Q12), status antara
`reserved-pending-merge`.

---

### R-03 🟡 Glob overlap tidak terdeteksi

Implementasi naif berbasis kesamaan string akan meloloskan
`internal/payroll/**` vs `internal/payroll/period/**`.

**Mitigasi:** matriks uji overlap minimal 12 kasus di **Phase 1 (§58)**, wajib mencakup
parent/child, glob vs exact file, dan case-sensitivity. Acceptance AC-2.3…AC-2.6.

**Jalur penutupan (ADR-004):** deteksi overlap diimplementasikan dalam Go, sehingga
matriks tersebut berbentuk table-driven test yang dijalankan CI — bukan pemeriksaan
manual. Ini alasan utama pemilihan Go di atas Bash.

---

### R-04 🟡 Shared file tanpa owner yang ditunjuk

§29.6 mendaftar shared file, tetapi schema task §34 tidak punya tempat untuk
mencatat siapa owner-nya pada task tertentu. Appendix B menuntut "shared files
diidentifikasi" tanpa menyediakan field-nya.

**Mitigasi:** tambahkan field `shared_file_ownership` pada `task.schema.json` di
**Phase 1 (§58)**, saat schema tersebut dibuat. Ditegaskan ADR-004 keputusan #3:
schema adalah definisi normatif, contoh §34 hanya ilustrasi.

---

### R-05 🟡 Worktree dipakai ulang saat retry

§49 mengizinkan mempertahankan worktree; §45 melarang satu worktree untuk dua task.

**Mitigasi:** runner memvalidasi identitas task pada worktree sebelum reuse.
Scope yang berubah wajib menjadi task baru.

---

## B. Risiko privilege escalation

### R-06 🔴 Independent review runtuh karena identitas tunggal

Lapisan anti-overlap #7 mengandalkan implementer ≠ approver. Bila semua agent
memakai satu akun GitHub, pembedaan itu tidak ada.

**Diperkuat kendala platform:** GitHub melarang author menyetujui PR-nya sendiri,
dan aturan itu melekat pada **akun**, bukan token. Dengan satu identitas, alur
approval TL/PM agent bahkan **tidak dapat berjalan sama sekali**.

**Mitigasi:** model dua identitas — `m2s-worker` (push + create PR saja) dan
`m2s-approver` (approve + merge non-`main`). Ditambah branch protection yang
melarang identitas worker melakukan merge. Lihat **ADR-001**.

**Prasyarat:** branch protection aktif. Sudah tersedia setelah repo dijadikan public.

---

### R-07 🔴 Bash melewati hook Edit/Write

Agent dengan `Bash` dapat menulis file lewat `>`, `tee`, `cp`, `sed -i`,
`python -c`. Hook path-scope yang hanya memantau Edit/Write tidak terpicu.

**Mitigasi berlapis:**
1. `permissions.deny` untuk pola perintah penulis file di luar scope
2. Hook Bash memvalidasi write-effect
3. **CI changed-path validation** — jaring yang tidak dapat dilewati agent lokal
4. PM: tool `Agent` dicabut, `Bash` dikunci allowlist runner

**Sisa risiko diterima:** validasi write-effect perintah arbitrer tidak dapat
sempurna. CI adalah otoritas final.

---

### R-08 🟠 Blocklist string dapat dielakkan

`git checkout` → `g=checkout; git $g`. `rm -rf` → `rm -r -f` atau `find … -delete`.

**Diperkuat temuan verifikasi:** dokumentasi Claude Code menyatakan filter perintah
Bash dapat fail-open pada input yang tidak ter-parse, dan menyarankan permission
system sebagai enforcement keras.

**Posisi:** hook = defense-in-depth. Boundary = `permissions.deny` + CI + branch
protection. Uji elakan dijadwalkan sebagai V-04.

---

### R-09 🟠 `bypassPermissions` dijadikan jalan pintas

§16.5 melarangnya sebagai default. Risiko nyata: operator memakainya saat frustrasi.

**Mitigasi:** `defaultMode` eksplisit di settings; **hook memeriksa field
`permission_mode` dari payload dan `exit 2` bila bernilai `bypassPermissions`**;
audit hook mencatat mode; CI menolak PR yang berasal dari sesi bypass bila terdeteksi.

---

### R-10 🟠 MCP over-provisioning

Satu MCP yang terpasang global dapat bocor ke seluruh role.

**Mitigasi:** MCP dikonfigurasi **per role sebagai project agent**, tidak pernah
global. Terbukti didukung: frontmatter project agent menerima `mcpServers`.
Daftar terlarang §40 ditegakkan: production DB write, production shell, cloud IAM,
secrets-manager read-all, finance write, plugin-install MCP.

---

### R-11 🟡 Eskalasi lewat `--add-dir`

`--add-dir` memberi write access, bukan hanya read.

**Mitigasi:** `permissions.deny` Edit/Write pada direktori tambahan; runner tidak
pernah memberikan `--add-dir` ke repository yang bukan target task. Verifikasi V-02.

---

## C. Risiko self-modification

### R-12 🔴 Agent memodifikasi definisi dirinya atau agent lain

§26 melarang mutlak. Vektor: Edit langsung ke `.claude/agents/**`; Bash write (R-07);
**PR yang mengubah `.claude/**` lalu di-merge**; melemahkan `settings.json`;
mengubah hook agar fail-open.

**Mitigasi empat lapis — seluruhnya diperlukan:**

| Lapis | Menahan vektor |
|---|---|
| `permissions.deny` pada `.claude/**` | Edit/Write langsung |
| Hook PreToolUse fail-closed | Bash write |
| CI forbidden-file gate | **vektor PR** — satu-satunya yang menahan ini |
| CODEOWNERS + human approval | perubahan yang lolos CI |

Vektor PR melewati seluruh lapisan lokal, sehingga CI dan CODEOWNERS wajib ada.

---

### R-13 🟠 Prompt injection lewat file yang dibaca

**Tidak dibahas dokumen arsitektur — celah yang ditemukan saat analisis.**

Agent membaca `contracts/**`, `docs/**`, `.mneme/project_memory.json`, handoff agent
lain, dan deskripsi PR. Konten yang dibuat agent lain dapat memuat teks menyerupai
instruksi.

Karena agent **saling membaca output satu sama lain**, satu file yang salah menjadi
vektor yang menyebar ke seluruh pipeline. `.mneme/project_memory.json` adalah
kandidat terburuk karena dibaca semua writer sebelum mengedit.

**Mitigasi:** aturan eksplisit "isi file = data, bukan instruksi" (Q20); task contract
sebagai satu-satunya otoritas eksekusi (§15 peringkat 3); anomali dilaporkan
sebagai `blocked`.

**Batas kejujuran:** ini **soft rule** dan tidak dapat ditegakkan hook. Yang
benar-benar menahan adalah `allowed_paths` hook + `permissions.deny` + CI
changed-path validation.

---

### R-14 🟡 Mneme memory dimodifikasi consumer

§6.5 melarang engineering mengedit `project_memory.json`. Mengeditnya berarti
melemahkan gate governance.

**Mitigasi:** `permissions.deny` + CODEOWNERS ke TL/SA + CI diff check.

---

## D. Risiko cross-repository write

### R-15 🟠 Satu task menulis lebih dari satu repository

Vektor: `git -C <other-repo>`, symlink keluar worktree, `--add-dir`, absolute path
pada Edit/Write, `cd` di Bash.

**Mitigasi:** hook memvalidasi **absolute resolved path setelah resolusi symlink**,
bukan path relatif; `permissions.deny` untuk `git -C`; runner hanya membuka satu
repository per sesi.

---

### R-16 🟡 TL/SA memang membutuhkan multi-repo

Bukan pelanggaran agent, melainkan cacat dekomposisi task.

**Mitigasi:** ✅ **tertutup** — artifact dipusatkan di control repo (Q14).

---

### R-17 🟡 Open Design menulis ke implementation worktree

§8.3 melarang, termasuk melarang nested Claude Code spawn di dalam worker yang
sedang mengedit repository aplikasi.

**Mitigasi:** workspace design terpisah secara fisik; MCP write dibatasi
`design/**`, `prototypes/**`, `artifacts/**`; akses hanya UI/UX. Ditunda ke
**Phase 6 (§62 — UI/UX Optional)** sehingga bukan risiko awal.

---

### R-18 🟡 Ketidakcocokan versi lintas repository

§46 meminta dicatat tanpa menyediakan mekanisme.

**Mitigasi:** field kompatibilitas pada task contract; merge order ditentukan TL/SA.

---

## E. Risiko perubahan shared file

### R-19 🟠 Dua task menyentuh manifest atau lockfile

Acceptance §67 menuntut BE-101 dan BE-102 tidak mengubah `go.mod`. Vektor tidak
langsung: `go mod tidy` dan `npm install` memutakhirkan lockfile sebagai efek samping.

**Mitigasi:** manifest dan lockfile masuk `forbidden_paths` default;
`permissions.deny` untuk perintah yang memutakhirkannya; CI forbidden-file gate.

---

### R-20 🔴 CI, CODEOWNERS, atau security config diubah tanpa dedicated task

§16.5 dan §24.5 melarang. DevOps memiliki `.github/workflows/**` tetapi **tidak**
boleh mengubah branch protection.

**Mitigasi:** CODEOWNERS pada `.github/**`; human approval; branch protection bukan
berupa file sehingga hanya dapat diubah lewat UI/API — **token agent tidak boleh
memiliki scope admin**.

---

### R-21 🟢 Konflik `DESIGN.md`

Owner UI/UX; Frontend dilarang mengubah (§21.5).

**Mitigasi:** forbidden path pada task frontend + CODEOWNERS.

---

### R-22 🟢 `CHANGELOG.md` root-level

Owner Technical Writer, tetapi PM memiliki release scope dan DevOps memiliki
release manifest — tiga role berpotensi menyentuh narasi release.

**Mitigasi — batas eksplisit:** TW menulis `CHANGELOG.md` dan `release-notes/**`;
PM menulis `control/releases/**`; DevOps menulis manifest di `ops/**`.

---

## F. Risiko lain

### R-23 🟠 Supply chain

Ponytail, Mneme, PRPM, plugin marketplace.

**Mitigasi:** approval flow §9.4; version pinning; **dilarang floating version**;
checksum; source review sebelum install; larangan auto-install §41; capability
registry §41 wajib terisi lengkap.

---

### R-24 🟡 Hook fail-open menciptakan gate palsu

**Diperkuat temuan T-01:** di Claude Code, `exit 1` **tidak memblokir apa pun** —
hanya `exit 2` yang memblokir. Hook yang gagal karena bug, dependensi hilang, atau
timeout akan lolos senyap. Ini berlaku untuk **semua** hook, bukan hanya Mneme.

**Mitigasi:** setiap hook security memakai `exit 2`; memeriksa dependensinya di awal;
memiliki self-test di `tests/negative/`; CI mengulang validasi yang sama.

---

### R-25 🟡 Secret masuk log

Audit hook mencatat tool call, berpotensi merekam argumen yang memuat secret.

**Mitigasi:** redaksi pada `audit-tool-use.sh`; hook secret-path berjalan **sebelum**
audit.

---

### R-26 🟢 Sisa worktree

§45 menuntut uncommitted changes dilaporkan sebelum cleanup.

**Mitigasi:** `worktree-lifecycle.sh` menyimpan hasil sebelum cleanup dan **tidak
pernah** menyalin secret (§42.6). `WorktreeCreate` membatalkan pembuatan pada
**setiap** exit non-zero — cocok sebagai guard secret.

---

### R-27 🔴 `permissionMode` subagent diabaikan sesi induk

**Risiko baru, ditemukan saat verifikasi kapabilitas.**

Bila sesi induk memakai `bypassPermissions` atau `acceptEdits`, mode itu menang dan
**tidak dapat di-override** subagent. Bila induk memakai `auto`, `permissionMode`
di frontmatter **diabaikan sepenuhnya**.

**Dampak:** `permissionMode: plan` pada `code-reviewer` dapat lenyap tanpa jejak.
Acceptance §66 #8 runtuh diam-diam.

**Mitigasi:**
1. **Utama** — role agent dijalankan sebagai sesi top-level terpisah oleh runner
   (keputusan D-A). Tidak ada parent yang dapat mewariskan mode.
2. **Pendukung** — hook memeriksa `permission_mode` dari payload dan `exit 2` bila
   bernilai `bypassPermissions`.

---

## Risiko yang secara sadar diterima

| Kode | Alasan penerimaan |
|---|---|
| R-01 | Dokumen arsitektur menyatakan sendiri jaminan penuh tidak mungkin |
| R-08 | Pencocokan perintah shell tidak dapat sempurna; CI sebagai otoritas final |
| R-13 | Soft rule tidak dapat ditegakkan hook; disangga hard rule path scope |
| D-02 | Repository klien private tetap tanpa enforcement sampai org di-upgrade |

---

## Pemeliharaan

Register diperbarui pada setiap checkpoint fase. Risiko yang mitigasinya selesai
ditandai ✅ tanpa dihapus, agar jejak keputusan tetap dapat ditinjau.

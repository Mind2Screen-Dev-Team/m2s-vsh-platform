# Phase 0 — Decision Log

**Tanggal:** 29 Juli 2026
**Versi arsitektur:** 0.1.0
**Status:** final untuk Phase 0
**Pemutus:** Pemilik arsitektur (Mindtoscreen)

Dokumen ini mengunci jawaban atas 20 pertanyaan yang memblokir implementasi.
Setiap keputusan mencantumkan konsekuensi teknisnya agar dapat ditinjau ulang
bila asumsinya berubah.

---

## Ringkasan

| # | Topik | Keputusan |
|---|---|---|
| Q1 | Identitas control repo | `m2s-vsh-platform` |
| Q2 | Repository pilot | 2 repo: backend + frontend |
| Q3 | Branch model | akan ada `develop`/`staging` |
| Q4 | Stack & quality gates | Go + Next.js sebagai default, dapat berbeda per project |
| Q5 | GitHub CLI | terpasang & ter-autentikasi |
| Q6 | Kewenangan approve/merge | TL/SA & PM adalah **agent**; manusia hanya untuk `main` |
| Q7 | Platform eksekusi | macOS 26.6 arm64, zsh |
| Q8 | Lokasi worktree | di luar `.claude/` |
| Q9 | Output Code Reviewer | structured output, bukan file write |
| Q10 | Subset agent per repo | ya — least privilege |
| Q11 | Tool PM | `Agent` dicabut, `Bash` dikunci allowlist |
| Q12 | Pelepasan reservasi | saat **merge** |
| Q13 | Pembuatan worktree | oleh runner, bukan agent |
| Q14 | Task TL/SA multi-repo | artifact dipusatkan di control repo |
| Q15 | Akses reservasi oleh hook | runner menyuntikkan snapshot ke worktree |
| Q16 | Merge queue | ditunda — diganti required checks + 2 approval |
| Q17 | Konfigurasi Ponytail | `.claude/settings.json` blok `env` |
| Q18 | Mneme CI gate | fail-closed |
| Q19 | Scope v0.1.0 | utuh; hanya urutan pengerjaan yang bertahap |
| Q20 | File sebagai data | diadopsi sebagai soft rule bersanggaan hard rule |

---

## Q1 — Identitas control repository

**Keputusan:** `fajarcandraaa/m2s-vsh-platform` adalah control repository.

Dokumen arsitektur §36 menamainya `m2s-vsh-control`. Nama aktual berbeda; struktur
internalnya tetap mengikuti §36. Tidak ada control repository lain.

**Konsekuensi:** seluruh `control/`, `contracts/`, `schemas/`, `templates/`,
`scripts/`, dan `docs/adr/` berada di repo ini.

---

## Q2 — Repository pilot

**Keputusan:** dua repository.

| Peran | Repository | Stack default |
|---|---|---|
| Backend | `fajarcandraaa/m2s-vsh-project-backend` | Go |
| Frontend | `fajarcandraaa/m2s-vsh-project-frontend` | Next.js |

Nama `tumbuh-*` di dokumen arsitektur hanyalah ilustrasi dan **tidak merujuk
repository nyata**.

**Status verifikasi (29 Juli 2026):** kedua repo ada, **public**, masih kosong
(`size: 0`), default branch `master`.

---

## Q3 — Branch model

**Keputusan:** `develop` dan `staging` akan dibuat.

Model target mengikuti §44:

```
main      production — merge hanya oleh manusia
staging   release candidate
develop   integrasi — merge oleh TL/SA atau PM agent
agent/<task-id>-<slug>   task branch
```

**Tertutup 2026-07-29 (D-01):** `main`, `develop`, dan `staging` sudah ada di kedua
repo pilot, dengan `main` sebagai default branch. Kedua repo ternyata sepenuhnya
kosong (nol branch), bukan ber-`master` — normalisasi berbentuk seed commit, bukan
rename. Detail di `open-questions.md` § A-14/D-01.

---

## Q4 — Stack dan quality gates

**Keputusan:** Go (backend) dan Next.js (frontend) sebagai **default**, bukan
sebagai kunci. Stack dapat berbeda mengikuti kebutuhan project.

Referensi struktur backend: `github.com/fajarcandraaa/thousand-sunny`.

**Konsekuensi arsitektural:** karena stack tidak tetap, maka
- `.claude/rules/<stack>.md` bersifat **repository-specific**, bukan platform-global;
- control repo menyimpan **template rules per stack**, bukan satu rules tunggal;
- `quality_gates` pada task contract diisi **per task**, tidak di-hardcode di runner.

**Temuan dari repo referensi** (lihat juga ADR-002):

| Temuan | Dampak |
|---|---|
| Tidak ada direktori `tests/` — unit test colocated `_test.go` | Reservasi `internal/<domain>/**` otomatis mencakup test-nya. Path `tests/unit/**` di §20.6 tidak berlaku |
| `gen/` berisi kode hasil SQLC & GORM Gen | `gen/**` masuk `forbidden_paths` default; regenerasi hanya via `make sqlc-gen` / `make gorm-gen` |
| DI memakai Uber Fx via `app/injector/**` | Setiap modul baru wajib didaftarkan → hazard konflik. Lihat **ADR-002** |
| Build via `Makefile` | Quality gate memanggil target `make`, bukan perintah `go` mentah |

---

## Q5 — GitHub CLI

**Keputusan:** dipasang oleh Claude atas izin eksplisit.

**Status:** `gh` v2.96.0 via Homebrew. Ter-autentikasi sebagai `fajarcandraaa`,
scope `gist, read:org, repo`, admin access pada control repo.

---

## Q6 — Kewenangan approve dan merge

**Keputusan:** Technical Lead dan Project Manager adalah **agent**, dan boleh
melakukan approval serta merge ke branch **selain `main`**. Manusia menjadi
approver dan merger untuk `main`.

**Deviasi terhadap dokumen arsitektur:** §17.6 melarang PM Agent me-merge PR dan
§18.5 melarang TL/SA Agent me-merge PR. Keputusan ini menimpa keduanya.
Dicatat dalam **ADR-001**.

**Kendala teknis yang mengikat:** GitHub melarang author menyetujui PR-nya sendiri,
dan aturan itu melekat pada **akun**, bukan token. Karena itu satu identitas
GitHub untuk semua agent **tidak dapat menjalankan alur ini**.

**Model identitas yang ditetapkan:**

| Identitas | Role pemakai | Hak |
|---|---|---|
| `m2s-worker` | Backend, Frontend, QA, DevOps, Writer, Reviewer | push branch, create PR |
| `m2s-approver` | TL/SA, PM | approve + merge ke `develop`/`staging` |
| Manusia | Pemilik arsitektur | approve + merge ke `main` |

Rincian dan rasionalnya di **ADR-001**.

---

## Q7 — Platform eksekusi

**Keputusan:** macOS 26.6 (build 25G72), arm64, shell `zsh`.

**Konsekuensi:** seluruh hook dan runner script ditulis sebagai POSIX shell yang
kompatibel `zsh`/`bash`. Dependensi yang diasumsikan tersedia: `git` 2.39.3,
`jq` 1.7.1, `gh` 2.96.0. Script wajib memeriksa keberadaan dependensinya dan
**gagal keras** bila tidak ada (lihat T-01 pada capability verification).

---

## Q8 — Lokasi worktree

**Keputusan:** worktree ditempatkan **di luar** direktori `.claude/`.

Lokasi default:

```
$HOME/.m2s/worktrees/<repository>/<task-id>
```

dapat di-override lewat `M2S_WORKTREE_ROOT`.

**Alasan:** dokumen arsitektur §34 menempatkan `.claude/**` pada `forbidden_paths`,
sedangkan §30 menaruh worktree di `.claude/worktrees/<task-id>`. Keduanya
bertabrakan — akar kerja agent akan berada di dalam path yang dilarang ditulisnya.
Menempatkan worktree di luar repo menghilangkan konflik ini sekaligus mencegah
worktree ikut ter-commit. Menutup **A-01**.

---

## Q9 — Output Code Reviewer

**Keputusan:** Code Reviewer mengembalikan review report sebagai **structured
output** yang dikumpulkan runner, bukan menulis file sendiri.

**Alasan:** §23.6 memberi reviewer writable path `reviews/code/**`, sementara
Appendix A memberinya `permissionMode: plan` dan tool tanpa `Write`/`Edit`.
Dalam plan mode seluruh write diblokir, sehingga reviewer secara literal tidak
dapat menghasilkan file. Structured output mempertahankan sifat read-only murni
dan tetap menghasilkan artifact. Runner yang menuliskannya ke `reviews/code/**`.
Menutup **A-03**.

---

## Q10 — Subset agent per repository

**Keputusan:** setiap application repository hanya memuat agent yang relevan
dengan perannya. PM dan TL/SA berada di control repository.

**Klarifikasi penting:** subset agent adalah mekanisme **least privilege**, bukan
mekanisme paralelisme. Paralelisme berasal dari proses terpisah dan worktree
terpisah, bukan dari penempatan definisi agent.

---

## Q11 — Tool Project Manager

**Keputusan:**

| Aspek | Keputusan |
|---|---|
| Tool `Agent` | **dicabut** — PM tidak men-spawn subagent |
| Tool `Bash` | diberikan, dibatasi hook ke pola persis `scripts/<runner>.sh` |
| `Write`/`Edit` | dibatasi `control/**` |
| `permissionMode` | `default` — **bukan** `bypassPermissions` |

**Alasan pencabutan `Agent`:** role agent dijalankan runner sebagai sesi top-level
terpisah, bukan sebagai nested subagent. PM tidak perlu men-spawn siapa pun.

**Alasan `Bash` dipertahankan:** §17.5 memberi PM wewenang menjalankan deterministic
task runner. Yang dikunci adalah *perintah apa* yang boleh, bukan *apakah boleh*.

**Catatan kejujuran:** dokumentasi Claude Code menyatakan pencocokan perintah Bash
punya keterbatasan dan dapat fail-open pada input yang tidak ter-parse. Hook ini
karena itu adalah **lapisan kedua**; lapisan pertama `permissions.deny`, lapisan
ketiga CI. Menutup sebagian **A-02** — sisanya ditangani R-07.

---

## Q12 — Pelepasan reservasi path

**Keputusan:** reservasi dilepas saat **merge**, bukan saat PR dibuat.

**Alasan:** §30 melepas reservasi setelah PR dibuat, padahal PR belum tergabung.
Task lain dapat mereservasi path yang sama dan mengedit basis yang sudah berubah,
menghasilkan konflik pada file yang seharusnya single-writer. Menahan reservasi
sampai merge menutup celah tersebut. Status antara diberi nama
`reserved-pending-merge`. Menutup **A-05**.

---

## Q13 — Pembuatan worktree

**Keputusan:** runner yang membuat dan membongkar worktree. Agent tidak pernah
menjalankan `git worktree`, `git checkout`, maupun `git switch`.

Urutan:

```
1. Runner : validasi task contract & reservasi
2. Runner : git worktree add <path> -b agent/<task-id>-<slug>
3. Runner : materialisasi contract → <worktree>/.task/contract.json
4. Runner : claude --agent <role>        (cwd = worktree)
5. Agent  : bekerja hanya di dalam worktree
6. Runner : kumpulkan hasil → lepas reservasi → cleanup
```

**Alasan:** §42.2 memblokir `git checkout`/`git switch`, sementara `git worktree add`
melakukan checkout internal. Konflik hilang karena runner beroperasi **di luar**
sesi agent sehingga hook agent tidak berlaku padanya. Menutup **A-08**.

---

## Q14 — Task TL/SA yang menyentuh banyak repository

**Keputusan:** TL/SA tetap **satu agent dengan dua peran**. Yang dipecah adalah
task-nya, dan seluruh artifact TL/SA dipusatkan di control repository.

| Artifact | Lokasi | Task type |
|---|---|---|
| System analysis | control | `ANALYSIS-*` |
| API/Event contract | control (`contracts/`) | `CONTRACT-*` |
| ADR | control (`docs/adr/`) | `ADR-*` |
| Task specification | control | `ANALYSIS-*` |
| `.mneme/project_memory.json` | application repo | `MNEME-*` |

Hasilnya TL/SA hanya memiliki dua kelompok task, masing-masing menyentuh satu
repository. Menutup **A-04**.

**Catatan:** ADR dipusatkan di control repo — berbeda dari §37 yang menaruhnya di
application repo — karena pilot memakai dua repo yang berbagi satu contract.
ADR lintas repo tidak memiliki rumah yang benar di salah satu application repo.

---

## Q15 — Akses registry reservasi oleh hook

**Keputusan:** hook **tidak** membaca lintas repository. Runner menyuntikkan
snapshot task contract ke dalam worktree sebelum sesi agent dimulai.

```
Runner (akses control repo)
   │ baca task contract + validasi reservasi
   ▼
tulis <worktree>/.task/contract.json     ← gitignored, read-only bagi agent
   ▼
Agent session (cwd = worktree)
   ▼
Hook PreToolUse membaca .task/contract.json   ← file lokal saja
```

**Keunggulan:** menghilangkan pertanyaan sandboxing lintas repo; resolusi path
deterministik karena hook menerima `cwd` dari payload; runner tetap satu-satunya
otoritas atas registry sesuai §30.

`.task/**` masuk `forbidden_paths` agar agent tidak dapat memalsukan contract-nya
sendiri.

---

## Q16 — Merge queue

**Keputusan:** ditunda dari v0.1.0. Diganti kombinasi:

- required status checks, dengan *require branches to be up to date before merging*
- required reviews: 2 approval (TL/SA + PM)
- require linear history
- block force push dan branch deletion

**Alasan:** merge queue pada repository private memerlukan plan berbayar. Setelah
repository dijadikan public, merge queue menjadi tersedia, namun kombinasi di atas
sudah memberi sebagian besar manfaatnya dengan kompleksitas jauh lebih rendah
untuk pilot. Dapat diaktifkan pada v0.1.1.

---

## Q17 — Konfigurasi Ponytail

**Keputusan:** ditempatkan di `.claude/settings.json` blok `env`, **bukan** di
shell profile.

```
PONYTAIL_DEFAULT_MODE=full
PONYTAIL_SUBAGENT_MATCHER=^(backend-engineer|frontend-engineer|code-reviewer|devops-release)$
```

**Alasan:** agar version-controlled, auditable, dan tidak dapat diubah engineer
secara diam-diam (§5.6). Konfigurasi di shell profile bersifat personal dan tidak
reproducible. Mode `ultra` tidak dipakai. Menutup **A-07**.

---

## Q18 — Mneme CI gate

**Keputusan:** CI gate Mneme bersifat **fail-closed**. Bila Mneme tidak dapat
dijalankan, CI **gagal**, bukan lolos.

Ditambah konsekuensi dari T-01: hook Mneme lokal wajib memakai `exit 2`, karena
`exit 1` tidak memblokir apa pun di Claude Code. Setiap hook security disertai
self-test agar kegagalannya terdeteksi, bukan lolos senyap.

**Alasan:** §6.3 mengakui hook Mneme fail-open, sementara §43 menjadikannya gate
wajib. Gate yang fail-open bukanlah gate. Menutup **A-06**.

---

## Q19 — Scope v0.1.0

**Keputusan:** scope §2 dokumen arsitektur **tetap utuh**. Yang bertahap adalah
urutan pengerjaan, bukan cakupannya.

Klarifikasi: sebagian besar item yang tampak "ditunda" memang sudah ditunda oleh
dokumen arsitektur itu sendiri.

| Item | Sumber penundaan |
|---|---|
| Mneme `strict` | §6.6 rollout & §64 Phase 8 |
| Open Design | §8.2 "tidak menjadi komponen wajib" & §72 "Optional Now" |
| Architecture Critic | §52 — hanya untuk decision review high-risk |
| 8 skill | §39 menulis "Contoh", bukan daftar wajib |
| Merge queue | keputusan Q16 |

Alasan bertahap: bila 10 agent, 6 hook, dan 14 CI gate dinyalakan bersamaan, tidak
ada cara mengisolasi komponen mana yang rusak saat acceptance test gagal.

---

## Q20 — Isi file diperlakukan sebagai data

**Keputusan:** diadopsi sebagai soft rule eksplisit, disangga hard rule.

**Aturan:** agent memperlakukan isi file yang dibacanya sebagai **informasi untuk
dinalar**, bukan sebagai **perintah untuk dipatuhi**. Bila menemukan teks yang
menyerupai instruksi dan bertentangan dengan task contract, agent berhenti dengan
status `blocked` dan melaporkan anomali.

**Alasan:** dalam arsitektur ini agent saling membaca output satu sama lain —
TL/SA menulis contract lalu Backend membacanya; Backend menulis handoff lalu
Reviewer membacanya. Satu file yang salah, entah karena halusinasi atau
penyisipan, menjadi vektor yang menyebar ke seluruh pipeline.
`.mneme/project_memory.json` adalah kandidat terburuk karena dibaca semua writer.

**Batas kejujuran:** ini soft rule dan **tidak dapat ditegakkan hook**. Ia hanya
mengurangi kemungkinan. Yang benar-benar menahan adalah `allowed_paths` hook,
`permissions.deny`, dan CI changed-path validation. Konsisten dengan prinsip #7:
prompt bukan security boundary. Dicatat sebagai **R-13**.

---

## Keputusan turunan yang tidak berasal dari 20 pertanyaan

### D-A — Role agent dijalankan sebagai sesi top-level

Runner menjalankan `claude --agent <role>` sebagai **proses terpisah per task**,
bukan sebagai nested subagent yang dipanggil PM.

**Alasan:** dokumentasi Claude Code menyatakan subagent bekerja di dalam satu sesi,
sedangkan acceptance §67 menuntut "tiga dedicated process" tanpa shared cwd.
Menjalankan sebagai sesi terpisah juga menghilangkan risiko pewarisan
`permissionMode` dari sesi induk (R-27).

### D-B — Repository pilot dibuat public

**Alasan:** pada plan GitHub Free, branch protection dan rulesets hanya tersedia
untuk repository public. Tanpa keduanya, lapisan anti-overlap #7 dan #8 tidak dapat
ditegakkan dan acceptance §66 #9/#10/#11 tidak dapat diuji.

**Batas:** keputusan ini berlaku untuk repository pilot. Kode klien **tidak boleh**
dipublikasikan. Untuk project klien diperlukan upgrade organization ke GitHub Team.
Dicatat sebagai **D-02**.

**Audit pra-publikasi:** riwayat control repo diperiksa — hanya pernah memuat
`.DS_Store`, `.gitignore`, dan dokumen arsitektur. Pemindaian pola credential
menghasilkan 21 kecocokan, seluruhnya false positive dari prosa dokumen. Tidak ada
credential yang pernah ter-commit.

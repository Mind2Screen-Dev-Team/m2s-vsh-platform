# ADR-010: Integrasi GitHub Projects dengan tracking m2s-vsh

**Tanggal:** 2026-08-07
**Decider:** owner arsitektur (Mindtoscreen)
**Status:** proposed (draft untuk review)

## Context

Tracking kerja m2s-vsh hari ini hidup di `control/tasks/` sebagai file YAML/markdown
di control repo: backlog di `control/backlog/`, spec task di
`control/tasks/specifications/`, status task di `control/tasks/status/` (dibakukan
oleh `docs/operator/task-status.md`), release scope di `control/releases/` (masih
`.gitkeep`, kosong). Workflow state machine 23 status didefinisikan di §33.

Dua pain muncul:

1. **Status tidak ter-persist.** `control/tasks/status/` — writable path PM §17.7 —
   tak pernah terisi jejak status untuk task Phase 7. `docs/operator/task-status.md`
   membakukan formatnya tapi tidak ada yang menulisnya. Riwayat task hilang;
   tidak ada yang membedakan "task berstatus apa" setelah selesai.
2. **Backlog tak terlihat.** Satu file `control/backlog/PILOT-1-phase-7.md`.
   Tidak ada kanban, tidak ada burndown, tidak ada milestone progress. PM dan
   manusia sama-sama tidak punya view per-sprint.

GitHub menyediakan fitur yang tepat menutup dua gap ini tanpa menambah infrastruktur:
**Projects** (kanban Board + Table + Roadmap timeline view), **Milestones** (grup
issue per rilis, progress bar), **Issues** (task item dengan label/assignee/timestamps),
**Insights** (burndown, metrics). Agent stack sudah memakai `gh` CLI
(`scripts/gh-app-token.sh`, `scripts/launch-task.sh`) dan workflow PR-based
(ADR-007) — integrasi tidak butuh dependency baru.

## Decision

1. **Source of truth tetap YAML di control repo.** GitHub Projects menjadi
   lapisan *presentation + history*, bukan kanonik. Agents tetap menulis spec dan
   status sebagai file git (`control/tasks/specifications/*.yaml`,
   `control/tasks/status/*.yaml`). Sync GitHub → hasil turunan, satu arah.
2. **Backlog markdown dikonversi menjadi Issues.** Satu issue per backlog item,
   `task.id` (`BE-201`, `FE-201`, …) sebagai prefix judul. Issue body menunjuk ke
   spec file canonical (`control/tasks/specifications/<id>.yaml`). Issue bukan
   source of truth — pointer.
3. **Status didorong ke Projects lewat GitHub Actions, bukan agent memanggil API.**
   Agent hanya meng-update YAML status (perilaku sekarang). Workflow
   `sync-project` (trigger: push ke `control/tasks/status/**`) membaca semua file
   status dan memetakan `taskStatus` ke kolom Project. Tidak ada agent yang
   menulis Projects langsung — konsisten dengan §17.7 writable paths dan prinsip
   "agent bekerja via git" ADR-004.
4. **Status 23-state dipetakan ke 6 kolom board.** Kolom Projects, bukan status
   penuh:

   | Kolom Project | taskStatus yang masuk |
   |---|---|
   | Backlog | `draft`, `needs-business-clarification`, `analysis-ready` |
   | Analysis | `technical-analysis`, `needs-technical-clarification`, `technical-ready` |
   | In Progress | `reserved`, `running` |
   | Review / QA | `implementation-complete`, `reviewing`, `changes-requested`, `qa-testing`, `defect-found`, `ci-passed` |
   | Merge Ready | `merge-ready` |
   | Done | `merged`, `documented`, `staging-verified`, `released` |
   | (tidak tampil di board) | terminal: `cancelled`, `failed`, `superseded`, `blocked` |

   Status detail tetap utuh di YAML (§33). Kolom adalah agregasi tampilan.
5. **Milestone untuk release scope.** `control/releases/` (kosong) diisi dengan
   daftar milestone per rilis. Milestone GitHub dipakai sebagai progress bar rilis;
   release notes tetap di GitHub Releases pada tag. Isi kanonik tetap file release
   di control repo.
6. **Roadmap view dipakai untuk timeline rilis.** Project mode Roadmap
   menampilkan item per-rilis di timeline. Tidak ada keputusan baru di sini —
   Roadmap hanyalah view dari data yang sama.
7. **Scope fase pertama terbatas pada control repo.** Backlog + status + milestone
   untuk `m2s-vsh-platform`. Issue lintas repo (backend/frontend task di
   m2s-vsh-project-backend/frontend) masuk fase berikutnya setelah pola terbukti.
   Ini membatasi blast radius dan memberi bukti sebelum skala.

## Rationale

### Mengapa source of truth tetap YAML, bukan Issues

Agents bekerja via git: contract, schema, runner (ADR-004), dan enforcement
(ADR-007) semua berdiri di atas file yang divalidasi dan di-review lewat PR.
Memindahkan status ke Issues berarti (a) agent wajib memanggil `gh api` untuk
menulis — menambah jalur tulis di luar git yang §17.7 tidak izinkan, (b) dua
sumber truth yang bisa konflik, (c) kehilangan riwayat git, validasi schema, dan
review PR pada status. YAML tetap kanonik menghindari semuanya. GitHub Projects
mengisi tepat bagian yang YAML tidak punya: visualisasi, timeline, agregasi.

### Mengapa sync lewat workflow, bukan agent langsung

Prinsip §26.4 "setiap writable path hanya satu active writer". Projects bukan
writable path yang didefinisikan; kalau agent menulis langsung, siapa writer-nya?
Workflow yang deterministic (satu arah, file → Project) mempertahankan satu
writer dan bisa di-audit lewat Actions log. Agent tidak berubah perilakunya —
perubahan hanya di lapisan turunan.

### Mengapa 6 kolom, bukan 23

Board 23 kolom tidak terbaca. Kolom adalah agregasi, bukan state baru — state
tetap 23 di YAML, konsisten §33. Sebaliknya, kalau Projects jadi source of truth,
kolom akan menggantikan state machine dan §33 harus ditulis ulang: mahal, tanpa
nilai. Detail transisi (`changes-requested → running`, `defect-found → running`)
tetap diwakili YAML; board menampilkan posisi kasar saja.

### Mengapa backlog jadi Issues

Backlog item = issue adalah pola GitHub standar; Issue memberi timestamps,
labels, assignee, comment thread, dan linking ke PR — semuanya yang file markdown
tidak punya. Issue sebagai pointer (bukan konten) menjaga spec kanonik tetap di
git yang bisa divalidasi. Kalau konten spec pindah ke issue body, schema dan
runner kehilangan input-nya.

### Opsi yang dipertimbangkan dan ditolak

| Opsi | Alasan ditolak |
|---|---|
| Issues sebagai source of truth penuh | Agent harus tulis via `gh api`; dua sumber truth; kehilangan validasi schema + review PR |
| Projects hanya manual (manusia isi, tanpa workflow) | Kembali ke pain #1: tidak ada yang ter-persist, tidak otomatis |
| GitLab / self-host (ADR-009) | Sudah ditolak; integrasi ini di GitHub karena enforcement stack sudah di GitHub |

## Consequences

- **Positif**: kanban + burndown + milestone progress + roadmap view gratis;
  riwayat task ter-persist tanpa manual write (workflow yang menulis); kolom
  board otomatis sinkron dengan YAML — tidak bisa terlewat.
- **Biaya**: satu workflow `sync-project` baru (wajib selalu-jalan, bentuk
  mengikuti ADR-007 #2); maintenance token Project; agent tidak terpengaruh
  langsung tetapi contract PM yang memakai Projects harus sadar Projects itu
  turunan.
- **Risiko auth**: Projects API hanya terima token dengan scope `project`
  (mutasi) atau `read:project` (query). Dua jalur valid untuk sync lewat
  Actions, dan **keduanya butuh konfigurasi baru**:
  1. **PAT classic** scope `project` + `repo`, disimpan sebagai secret org;
  2. **GitHub App** diberi permission read/write **organization projects**
     (permission repository projects TIDAK cukup — dokumen eksplisit), lalu
     App dipasang untuk semua repo yang butuh akses.
  `GITHUB_TOKEN` bawaan Actions **tidak punya scope project** — sync tidak bisa
  jalan tanpa PAT/App. Verifikasi scope App m2s-worker saat ini belum punya
  permission projects.
- **Yang tidak diputuskan**: issue lintas repo (fase 2), apakah roadmap view
  dipakai per-sprint atau per-rilis, siapa owner kolom kustom bila detail status
  perlu ditambah.

## Verifikasi plan Free

Temuan dari audit 2026-08-07 (org `Mind2Screen-Dev-Team` aktif `plan: free`,
token login `fajarcandraaa`):

| Fitur | Tersedia di Free? | Bukti |
|---|---|---|
| Projects v2 (Board/Table/Roadmap view) | ✅ | docs "About Projects": tiga layout tanpa syarat plan |
| Custom fields (single-select, date, number, iteration) | ✅ | docs "Adding metadata to your items": hingga 50 field, tanpa syarat plan |
| Built-in workflows / auto-add / auto-archive | ✅ | docs "Automating your projects" |
| Insights (current + historical charts, burndown-style) | ✅ | docs "About insights": chart pakai item project sebagai sumber, tanpa syarat plan |
| GraphQL API (`projectV2`) | ✅ | docs "Using the API to manage Projects": autentikasi PAT/App dengan scope `project` |
| Milestones + Releases | ✅ | bagian inti GitHub, tanpa syarat plan |

**Tidak ada fitur yang dipakai integrasi ini yang dikunci di balik plan berbayar.**
Yang berbeda antar plan: storage/rate limits dan user seats (irrelevant di sini —
limits jauh di atas volume m2s-vsh: 50.000 item per project, 50 fields). Ini
membalik arah ADR-009 D-02: alasan "repo klien private butuh upgrade Team"
tetap berdiri untuk branch protection, tetapi **Projects sendiri tidak menuntut
Team**. Integrasi Projects bisa jalan di Free.

## Kendala akses yang terverifikasi langsung

1. Token pribadi (`gh auth`, scope `gist, read:org, repo, workflow`) **menolak**
   query `projectsV2` — error `INSUFFICIENT_SCOPES`, butuh `read:project`.
2. Organisasi `Mind2Screen-Dev-Team` aktif di plan **Free** (`{"plan":"free"}`).
3. `scripts/gh-app-token.sh` tidak menyebut permission `projects`; App
   m2s-worker/m2s-approver belum diverifikasi punya akses org projects.

**Implikasi implementasi:** sebelum sync-project di-deploy, salah satu jalur
harus dikonfigurasi: re-auth PAT pribadi dengan scope `project`, atau App
m2s-worker diberi permission org projects (read/write) + re-install. Tanpa ini
langkah #1 (buat project + cek via API) sudah gagal. Ini pre-requisite, bukan
pilihan.

## Backward compatibility / rollback

- Tidak mengubah perilaku agent. YAML tetap ditulis seperti sekarang; workflow
  hanya menambah output turunan.
- Rollback = nonaktifkan workflow `sync-project`; file YAML utuh, Projects
  membeku pada snapshot terakhir. Tidak ada migrasi balik.
- Backlog existing (`PILOT-1-phase-7.md`) dikonversi sekali jalan ke Issues;
  file tetap dipertahankan sebagai arsip (pola sama dengan contract Phase 7
  invalid ADR-009 #3).

**Menimpa:** tidak ada keputusan ADR sebelumnya. Mengisi gap writable path
`control/releases/` yang belum berisi. Memperluas `docs/operator/task-status.md`
dengan langkah sync.

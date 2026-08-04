# Phase 8 §64 — Pengukuran Stabilization

**Tanggal:** 2026-08-04
**Sumber data:** riwayat PR 3 repo (`gh pr list`), riwayat commit, hasil `make verify`
**Cakupan:** Phase 0 sampai Phase 8 (2026-07-31 s.d. 2026-08-04)

§64 menuntut lima pengukuran sebagai dasar sebelum aturan mana pun dikurangi.
Dokumen ini mengisinya dengan angka nyata, bukan perkiraan. Yang tidak terekam
dinyatakan tidak terekam.

## 1. Review cycles

Diukur sebagai **PR per task** — berapa PR dibutuhkan untuk menyelesaikan satu
unit kerja. PR yang ditutup tanpa merge dihitung sebagai siklus yang terbuang.

| Repo | PR merged | PR closed tanpa merge | Rasio terbuang |
|---|---|---|---|
| `m2s-vsh-platform` (control) | 26 | 2 | 7% |
| `m2s-vsh-project-backend` | 9 | 5 | 36% |
| `m2s-vsh-project-frontend` | 6 | 1 | 14% |

**Siklus terbuang yang teridentifikasi:**

| PR | Sebab | Ditutup guard mana sekarang |
|---|---|---|
| control #14 `agent/phase-7-plan-specs` | branch planning memakai pola `agent/*`, bukan `worktree-*` | **H-02** |
| control #15 `agent/qa-102-report` | sama — planning artifact di branch agent | **H-02** |
| backend #5 `agent/merge-probe` | probe §66 #9 (bukan defect) | — |
| backend #6 `agent/approver-probe` | probe (bukan defect) | — |
| backend #7 `test/approver-ruleset-probe` | probe (bukan defect) | — |
| frontend #4 `test/secc66-probe` | probe (bukan defect) | — |

Setelah probe dikeluarkan (4 dari 8 adalah probe sengaja, bukan kegagalan),
**siklus terbuang nyata = 2**, keduanya pelanggaran naming yang kini ditahan H-02.

**Task dengan lebih dari satu PR (rework):**

| Task | PR | Sebab rework |
|---|---|---|
| BE-102 | #9 (endpoint), #10 (CORS), #11 (CORS v2) | contract melarang `go.mod` → CI gagal → remediasi manual (2 PR tambahan) |
| BE-102 contract | control #18, #19 | contract diperbaiki dua kali agar realistis |
| FE-102 contract | control #20 | contract melarang berkas scaffolding `create-next-app` |

**Kesimpulan review cycle:** rework terbesar bukan karena kode salah, melainkan
karena **contract tidak realistis** — 4 PR tambahan (BE #10, #11, control #18,
#19, #20) semuanya akibat contract melarang berkas yang scaffolding wajib buat.
Itulah yang **H-03** (`execution.scaffold`) tutup: contract semacam itu kini
ditolak sebelum agent mulai.

## 2. Escaped defects

Defect yang lolos ke branch integrasi lalu ditemukan setelahnya.

| Defect | Ditemukan di | Dampak | Ditahan sekarang oleh |
|---|---|---|---|
| CORS header hilang pada `/api/v1/status` | QA-102 (setelah BE-102 merged) | 2 PR remediasi (#10, #11) | tidak ada guard — ini defect fungsional sejati, acceptance criteria BE-102 tidak menyebut CORS |
| Workflow app-repo menunjuk org `fajarcandraaa` (pra-migrasi) | Phase 8 A7 | CI 2 repo me-resolve control repo salah — tak terdeteksi sejak migrasi ADR-008 | perluasan `github-workflow.test.sh` (memeriksa org pada 3 salinan) |
| PR `agent/* → main` lolos tanpa check apa pun | Phase 8 A7 (uji dummy) | celah enforcement: `mergeStateStatus: CLEAN` | **branch protection** (`validate-changed-paths` required) |
| Ruleset Phase 4 kosong (`rules: null`) sejak dipasang | Phase 8 A7 | proteksi yang diyakini aktif ternyata tidak menegakkan apa pun | terungkap: ruleset butuh plan GitHub Team |

**Escaped defect = 4.** Satu fungsional (CORS), tiga **enforcement** — yaitu
proteksi yang diyakini aktif tetapi tidak menegakkan apa pun. Yang terakhir
paling mahal: selama ~3 hari sistem dianggap terlindungi ruleset yang kosong.

## 3. Path violations

| Sumber | Angka |
|---|---|
| Skenario §68 ditolak `enforcement.test.sh` | 15 lulus |
| Skenario §68 ditolak `github-workflow.test.sh` | 18 lulus |
| Pelanggaran path nyata oleh agent selama Phase 7 | **tidak terekam** |

`audit-tool-use.sh` mencatat tool use ke `<worktree>/.task/audit.log`, tetapi
worktree Phase 7 sudah dibersihkan sehingga log tidak tersisa. Jumlah kejadian
nyata `check-path` menolak tulis tidak dapat dihitung retrospektif.

**Yang diketahui:** CI menolak PR BE-102 karena `go.mod` di luar allowed —
itu path violation yang tertangkap CI (bukan hook), tercatat sebagai kegagalan
CI, bukan sebagai metrik.

## 4. Token / cost / time

| Metrik | Status |
|---|---|
| Baseline statis per role (kata, model, effort, maxTurns) | ✅ `docs/operator/phase-8-token-baseline.md` |
| Token nyata per sesi | ❌ tidak terekam — payload `PostToolUse` tidak membawa `usage` |
| Waktu per task | 🟡 dapat diturunkan dari timestamp PR (lihat bawah) |

**Durasi kasar dari timestamp PR:**

| Task | Mulai (PR pertama) | Selesai (PR terakhir) | Rentang |
|---|---|---|---|
| BE-102 (+ CORS) | 2026-08-03 10:45 | 2026-08-03 14:10 | ~3,5 jam |
| FE-102 | 2026-08-03 10:49 | 2026-08-03 10:49 | satu PR |
| Phase 7 keseluruhan | 2026-08-03 06:44 | 2026-08-03 15:56 | ~9 jam |
| Phase 8 keseluruhan | 2026-08-03 15:03 | 2026-08-04 07:35 | ~16,5 jam |

Angka ini wall-clock, bukan waktu agent — termasuk jeda review manusia.

## 5. Rules yang tidak efektif

§64 menuntut pengurangan aturan yang tidak efektif. Kandidat dari bukti di atas:

| Aturan | Bukti | Rekomendasi |
|---|---|---|
| Ruleset `agent-push-restriction`, `agent-worker-restriction` | `rules: null` sejak dipasang — tak menegakkan apa pun (plan Free) | **Sudah ditangani**: `apply-rulesets.sh` diberi catatan, `main-agent-block` diarsip. Jangan mengandalkan ruleset sampai plan naik |
| `enforce_admins: false` pada branch protection | admin dapat bypass — termasuk saat merge PR Phase 8 sendiri | Pertahankan untuk sekarang (satu-satunya jalan keluar bila CI macet), tetapi catat sebagai celah yang diketahui |
| `required_approving_review_count: 0` | tidak ada review wajib di control repo | Pertahankan — pengembang tunggal; naikkan ke 1 saat tim bertambah |

**Tidak ada aturan yang direkomendasikan dihapus.** Yang ditemukan justru
sebaliknya: dua "aturan" yang dianggap aktif ternyata kosong.

## 6. Mneme strict untuk decisions stabil

§64 menyebut pertimbangan menaikkan Mneme dari `warn` ke `strict`. **Belum
dievaluasi** — memerlukan data seberapa sering Mneme memberi peringatan selama
Phase 7/8, dan itu tidak terekam. Ditunda sampai ada pencatatan.

## Ringkasan

| §64 | Status |
|---|---|
| Ukur token/cost/time | 🟡 baseline statis + durasi PR; token nyata tidak terekam |
| Ukur review cycles | ✅ 2 siklus terbuang nyata, 5 PR rework akibat contract tak realistis |
| Ukur escaped defects | ✅ 4 (1 fungsional, 3 enforcement) |
| Ukur path violations | 🟡 33 skenario uji lulus; kejadian nyata tidak terekam |
| Kurangi rules tak efektif | ✅ dievaluasi — tidak ada yang dihapus; 2 ruleset ternyata kosong |
| Mneme strict | ⬜ ditunda, data tidak ada |

**Temuan utama Phase 8:** sumber rework terbesar bukan kode, melainkan contract
yang tidak realistis (5 PR) dan proteksi yang diyakini aktif tetapi kosong
(3 escaped enforcement defect). Keduanya kini ditahan H-03 dan branch protection.

## Prasyarat sebelum pengukuran ini dapat diulang

1. Token per sesi — `docs/operator/phase-8-token-tracking.md`.
2. `audit.log` diselamatkan sebelum worktree dibersihkan (saat ini hilang bersama
   worktree).
3. Status task ter-persist — `docs/operator/task-status.md`; tanpa itu durasi task
   hanya dapat diturunkan dari timestamp PR.

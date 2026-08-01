# Branch Protection & Required Checks — Panduan Operator

**Fase:** Phase 4 (§60)
**Cakupan:** required status check `validate-changed-paths`, CODEOWNERS, PR
template, repository rulesets sebagai pengganti classic push restriction, dan
batas yang masih terbuka (D-02, D-03).

Dokumen ini menjelaskan cara memasang lapis penegakan di sisi GitHub, urutan yang
wajib diikuti, dan tiga cara mengunci repo secara permanen bila urutannya
dilanggar. Ia untuk **operator manusia**, bukan agent — seluruh aksi di sini
menuntut hak yang tidak dimiliki agent mana pun (§16.5, ADR-001 #4).

---

## Prinsip

> Prompt bukan security boundary (prinsip #3). Batas ditegakkan permissions,
> hooks, runner, Git, dan CI — bukan instruksi di prompt.

Lapis yang ditambahkan fase ini berada di ujung rantai, sesudah seluruh lapis
lokal dapat dilewati:

| Lapis | Menahan | Dapat dilewati agent? |
|---|---|---|
| `permissions.deny` (settings.json) | Edit/Write pada pola terlarang | tidak untuk pola persis |
| PreToolUse hook | write di luar scope, command destruktif | Bash arbitrer (R-07), elakan string (R-08) |
| **Required check `validate-changed-paths`** | path di luar scope pada PR | **tidak** — otoritas final |
| **CODEOWNERS** | vektor PR ke `.claude/**`, `.github/**` | tidak, tapi bukan gate (lihat Perangkap 2) |
| **Ruleset `restrict updates`** | siapa boleh push/merge | tidak — tetapi belum terpasang |
| Merge queue | urutan merge, overlap semantik | **tidak tersedia** (org-only, ADR-007 #3) |

Hook adalah **defense-in-depth**. CI adalah **otoritas final**. Branch protection
adalah **penahan terakhir** — dan satu-satunya yang tidak berupa berkas, sehingga
hanya dapat diubah lewat UI atau API, bukan lewat PR (mitigasi R-20).

---

## Tiga perangkap deadlock

Ketiganya menghasilkan keadaan yang sama: **merge terblokir permanen, dan
pemulihannya menuntut mematikan proteksi**. ADR-001 menyatakannya langsung:
*"mewajibkan approval yang tidak mungkin diberikan, atau status check yang tidak
pernah dilaporkan, akan memblokir seluruh merge tanpa jalan keluar."*

### Perangkap 1 — required check yang di-skip

Job dengan `if:` pada **level job** tidak melaporkan status ketika kondisinya tidak
terpenuhi. GitHub menunggu status itu selamanya.

```yaml
jobs:
  validate-changed-paths:
    if: startsWith(github.head_ref, 'agent/')   # ← JANGAN. Gate palsu.
```

Bentuk yang benar memindahkan pemilahan ke dalam step, sehingga job selalu
melapor — branch non-agent lolos **eksplisit**, bukan di-skip. Ditegakkan
otomatis:

```bash
make verify-github
```

Memeriksa manual bahwa suatu PR memang melapor, bukan di-skip:

```bash
gh api "repos/fajarcandraaa/<repo>/commits/<sha>/check-runs" \
  --jq '.check_runs[] | {name, status, conclusion}'
```

`conclusion: "success"` benar. `conclusion: "skipped"` adalah perangkap — jangan
jadikan required check.

### Perangkap 2 — approval yang mustahil diberikan

GitHub: *"Pull request authors cannot approve their own pull requests."* Aturan itu
melekat pada **akun**, bukan token. Selama hanya ada satu kolaborator:

- `required_approving_review_count: 2` → tidak akan pernah terpenuhi
- `require_code_owner_review: true` → sama, karena owner tunggal = author

Karena itu keduanya **sengaja dibiarkan nol/mati** (ADR-007 #4, #6). Aktivasi
menunggu identitas `m2s-approver` (ADR-001 #5).

Catat juga sisi lain: *"Repository owners and administrators can merge a pull
request even if it hasn't received an approving review."* Jadi required review
bukan penahan bagi pemilik repo. Ia menahan agent.

### Perangkap 3 — `merge_group` absen saat merge queue aktif

Docs GitHub: *"Otherwise, status checks will not be triggered when you add a pull
request to a merge queue. The merge will fail as the required status check will not
be reported."*

Workflow sudah memuat `merge_group:` sejak sekarang meski merge queue belum aktif,
supaya aktivasi nanti tidak menuntut perubahan workflow. `make verify-github`
menahan penghapusannya.

---

## Urutan pemasangan — jangan dibalik

```
1. workflow terpasang di repo aplikasi   (lewat PR, manusia merge)
2. PR nyata dibuka                        → workflow berjalan
3. amati SATU run hijau                   → bukti ia melapor
4. baru daftarkan required check
```

Langkah 3 bukan formalitas. Ia satu-satunya yang membedakan required check yang
berfungsi dari Perangkap 1.

### Langkah 1 — pasang artefak

Artefak kanonik ada di control repo `templates/github/`. Salinlah ke repo
aplikasi lewat PR, jangan push langsung (§16.5, R-20):

| Sumber | Tujuan |
|---|---|
| `templates/github/workflows/path-enforcement.yml` | `.github/workflows/path-enforcement.yml` |
| `templates/github/CODEOWNERS` | `.github/CODEOWNERS` |
| `templates/github/PULL_REQUEST_TEMPLATE.md` | `.github/PULL_REQUEST_TEMPLATE.md` |

#### Ke branch mana — tidak sama untuk ketiganya

Diverifikasi terhadap dokumentasi GitHub, karena ketiganya berperilaku berbeda:

| Berkas | Branch yang dibaca | Konsekuensi |
|---|---|---|
| `CODEOWNERS` | **base branch PR** — *"the CODEOWNERS file must be on the base branch of the pull request"* | wajib ada di **setiap** branch yang menjadi base: `develop`, `staging`, `main`. Satu salinan di `main` **tidak** melindungi PR ke `develop` |
| `PULL_REQUEST_TEMPLATE.md` | **default branch saja** — *"You must create templates on the repository's default branch"* | cukup di `main`. Di `develop` saja tidak akan muncul sama sekali |
| `workflows/*.yml` | `on.pull_request.branches` menyaring **base branch** | daftarkan setiap base yang dipakai. Berkas perlu ada di branch yang menjadi base PR |

Karena itu distribusi Phase 4 memakai **satu PR per base branch per repo** (enam
PR untuk dua repo), bukan satu PR ke `develop` saja.

#### Catatan jalur: `git checkout` diblokir di sesi agent

§42.2 memblokir `git checkout`/`switch`/`worktree`, dan A-08 menetapkan manajemen
branch milik runner. Distribusi karena itu dilakukan lewat **GitHub API** (`git/blobs`
→ `git/trees` → `git/commits` → `git/refs`), tanpa checkout lokal sama sekali. Ini
bukan siasat mengelakkan hook — ia jalur yang sejalan dengan A-08, dan jejaknya
terekam di API.

⚠️ **Yang perlu diketahui:** `permissions.deny` pada `Edit`/`Write` **tidak** menahan
tulisan lewat Bash (R-07) maupun lewat API. Jadi deny `.github/**` melindungi
**control repo dari suntingan tool Edit/Write**, bukan menutup seluruh jalur menuju
`.github/` di repo mana pun. Penahan sesungguhnya untuk repo aplikasi tetap
**CODEOWNERS + human merge**, dan itulah sebabnya distribusi wajib lewat PR.

Push workflow menuntut token ber-scope `workflow`. Bila ditolak:

```bash
gh auth refresh -s workflow
```

### Langkah 2–3 — buktikan ia berjalan

```bash
# Sebelum Phase 4 angka ini 0 di ketiga repo — berkas ada, jalurnya mati.
gh api "repos/fajarcandraaa/<repo>/actions/runs" --jq '.total_count'

# Lima run terakhir.
gh api "repos/fajarcandraaa/<repo>/actions/runs" \
  --jq '.workflow_runs[0:5][] | {name, event, head_branch, conclusion}'
```

Yang dicari: `total_count > 0`, dan `conclusion: "success"` pada PR dari branch
**non-agent** (bukan `skipped`).

### Langkah 4 — daftarkan required check

Nama context = nama job = `validate-changed-paths`.

```bash
for repo in m2s-vsh-project-backend m2s-vsh-project-frontend; do
  for branch in develop staging; do
    gh api -X PUT "repos/fajarcandraaa/$repo/branches/$branch/protection" \
      --input - <<'JSON'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["validate-changed-paths"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 0
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_conversation_resolution": true
}
JSON
  done
done
```

Catatan pada payload:

- `strict: true` — *require branches to be up to date before merging*, pengganti
  sebagian manfaat merge queue (Q16).
- `required_approving_review_count: 0` — **sengaja**. Lihat Perangkap 2.
- `require_code_owner_reviews: false` — **sengaja**. Lihat Perangkap 2.
- `restrictions: null` — classic restriction org-only, ditolak `422` di repo
  personal (D-03). Gunakan ruleset di bawah.
- `main` **tidak** disentuh perintah ini. Merge ke `main` sepenuhnya milik manusia
  (ADR-001 #2).

Verifikasi:

```bash
gh api "repos/fajarcandraaa/<repo>/branches/develop/protection" \
  --jq '{checks: .required_status_checks.contexts,
         approvals: .required_pull_request_reviews.required_approving_review_count,
         admins: .enforce_admins.enabled}'
```

---

## Membatasi siapa boleh push — ruleset, bukan classic protection

Classic protection menolak pembatasan hak di repo akun personal:

```
PUT /repos/fajarcandraaa/<repo>/branches/main/protection
422 Only organization repositories can have users and team restrictions
```

**Repository rulesets tidak punya batasan itu.** Rule *restrict updates*: *"only
users with bypass permissions can push to branches or tags whose name matches the
pattern you specify."*

### Hasil uji empiris 1 Agustus 2026 (repo `m2s-vsh-rules-probe`)

Diuji langsung terhadap API, bukan disimpulkan dari dokumentasi:

| Uji | Hasil | Arti |
|---|---|---|
| Ruleset `restrict updates`, bypass **kosong**, push sebagai **pemilik repo** | ❌ `GH013: Cannot update this protected ref`; `current_user_can_bypass: "never"` | **Ruleset mengikat pemilik repo** (V-06). Berbeda dari classic protection, di mana admin *"always able to push"*. Ini yang membuat ADR-001 **#4** dapat ditegakkan |
| Pemilik dimasukkan bypass list (`User`, `always`) | ✅ push lolos (`51d15cd..266f050`) | bypass list berfungsi |
| `bypass_mode: "pull_request"`, push **langsung** | ❌ `GH013` | hak merge lewat PR terpisah dari hak push langsung — persis yang model dua identitas butuhkan |
| `actor_type: RepositoryRole`, id 2 / 4 / 5 | ✅ ketiganya diterima | role tersedia sebagai bypass actor di repo personal (V-07) |
| `actor_type: OrganizationAdmin` | ❌ `422 ruleset source must be in an organization` | sesuai dokumentasi |
| **`actor_type: Integration` (GitHub App)** | ❌ **`422 Actor GitHub Actions integration must be part of the ruleset source or owner organization`** | **V-08 — mengoreksi rencana semula** |

### ⚠️ V-08 mengubah jalur ADR-001 #5

Dokumentasi menyiratkan hanya `OrganizationAdmin` yang tak berlaku di repo personal.
Uji API membantahnya: **GitHub App juga tidak dapat menjadi bypass actor** kecuali ia
bagian dari owner organization.

Akibatnya, di repo akun personal ruleset dapat membatasi **manusia dan role**, tetapi
**tidak dapat memberi pengecualian kepada App**. Model `m2s-worker`/`m2s-approver`
sebagaimana ADR-001 #5 merancangnya (GitHub App, bukan machine user) karena itu
**menuntut migrasi ke organization** — bukan karena push restriction, melainkan karena
bypass actor.

Yang **bisa** dilakukan sekarang tanpa migrasi: mengunci `develop`/`staging` dengan
`restrict updates` dan memberi bypass kepada identitas **manusia** dengan
`bypass_mode: pull_request`. Itu menahan seluruh push langsung termasuk milik pemilik
repo — lebih kuat dari classic protection, tetapi belum memisahkan worker dari
approver.

```bash
# Lihat ruleset yang ada (0 di ketiga repo pilot per 1 Agustus 2026).
gh api "repos/fajarcandraaa/<repo>/rulesets" --jq 'length'

# Kunci develop/staging; hanya bypass yang tercantum boleh mengubahnya.
# Ganti actor_id dengan hasil `gh api user --jq .id`.
gh api -X POST "repos/fajarcandraaa/<repo>/rulesets" --input - <<'JSON'
{
  "name": "agent-push-restriction",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": { "include": ["refs/heads/develop", "refs/heads/staging"], "exclude": [] }
  },
  "rules": [{ "type": "update" }],
  "bypass_actors": [
    { "actor_type": "User", "actor_id": 0, "bypass_mode": "pull_request" }
  ]
}
JSON
```

**Uji di repo sekali-pakai lebih dulu.** Repo probe
`fajarcandraaa/m2s-vsh-rules-probe` sengaja dibiarkan hidup dengan ruleset aktif;
bypass picker UI dapat diperiksa di
`https://github.com/fajarcandraaa/m2s-vsh-rules-probe/rules/20155906`. Hapus repo itu
bila sudah tidak diperlukan.

---

## Troubleshooting

**PR terblokir "Expected — Waiting for status to be reported".**
Perangkap 1. Required check terdaftar tetapi job-nya di-skip. Periksa:

```bash
gh api "repos/fajarcandraaa/<repo>/commits/<sha>/check-runs" \
  --jq '.check_runs[] | {name, conclusion}'
```

Pemulihan: cabut context dari `required_status_checks.contexts`, perbaiki bentuk
job, pasang ulang. Jangan mematikan `enforce_admins` sebagai jalan pintas (R-09).

**PR terblokir, menuntut approval yang tak dapat diberikan.**
Perangkap 2. Turunkan `required_approving_review_count` ke `0` dan
`require_code_owner_reviews` ke `false` sampai identitas kedua ada.

**`422 Only organization repositories can have users and team restrictions`.**
Kirim `"restrictions": null` pada payload classic; pakai ruleset untuk membatasi
hak. Ini D-03, bukan galat konfigurasi.

**Push workflow ditolak `refusing to allow an OAuth App to create or update workflow`.**
Token tanpa scope `workflow`. `gh auth refresh -s workflow`.

**Workflow tidak berjalan sama sekali.**
Periksa trigger cocok dengan branch yang benar-benar ada. Ini kegagalan Phase 3:
`branches: [develop, staging]` di control repo yang hanya punya `main`.

```bash
gh api "repos/fajarcandraaa/<repo>/branches" --jq '.[].name'
```

---

## Batas yang diketahui (diterima)

| Batas | Sebab | Konsekuensi |
|---|---|---|
| Merge queue tidak tersedia | org-only pada setiap plan | mitigasi overlap semantik (§29.8, R-01) bersandar reservasi + urutan merge TL/SA (§46) |
| `required_approving_review_count: 0` | satu kolaborator, larangan self-approval | §66 #9 tidak dapat diuji; ADR-001 belum berlaku efektif |
| `require_code_owner_review` mati | owner tunggal = author PR | CODEOWNERS memberi jejak audit, bukan gate |
| Push restriction belum terpasang | menunggu GitHub App | agent yang membuka PR masih dapat me-merge PR-nya sendiri |
| Jalur `merge_group` melapor sukses tanpa validasi ulang | payload tidak memuat branch asal | aman selama merge queue mati; tinjau saat diaktifkan |
| Repo klien private tanpa enforcement | D-02 | satu-satunya hal yang benar-benar menuntut plan Team |
| Artefak ada di dua tempat | kanonik + salinan repo aplikasi | `verify-github` memeriksa bentuk, bukan kesamaan byte |

Rujukan: ADR-007 (keputusan fase ini), ADR-001 (kewenangan merge), D-02 dan D-03
(`docs/decisions/open-questions.md`), R-12/R-19/R-20/R-24
(`docs/decisions/risk-register.md`), `docs/operator/hook-enforcement.md` (lapis
lokal).

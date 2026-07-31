# Hook Enforcement — Panduan Operator

**Fase:** Phase 3 (§59)
**Cakupan:** enam hook `.claude/hooks/`, konfigurasi `.claude/settings.json`,
subcommand runner `check-path` dan `validate-changed-paths`, CI path validation.

Dokumen ini menjelaskan cara kerja penegakan batas path pada level runtime, cara
mengujinya, dan cara membaca jejaknya. Ia untuk **operator manusia**, bukan agent.

---

## Prinsip

> Prompt bukan security boundary (prinsip #3). Batas ditegakkan permissions,
> hooks, runner, Git, dan CI — bukan instruksi di prompt.

Penegakan berlapis. Tidak ada satu lapis pun yang cukup sendiri:

| Lapis | Menahan | Dapat dilewati? |
|---|---|---|
| `permissions.deny` (settings.json) | Edit/Write/Bash pada pola terlarang | tidak untuk pola persis |
| PreToolUse hook | write di luar scope, command destruktif, secret | Bash arbitrer (R-07), elakan string (R-08) |
| CI changed-path validation | path di luar scope pada PR | **tidak oleh agent lokal** — otoritas final |
| Branch protection | merge tak sah | Phase 4, org-only (D-03) |

Hook adalah **defense-in-depth**. CI adalah **otoritas final**.

---

## Enam hook

Seluruhnya fail-closed dengan **`exit 2`**. `exit 1` **tidak memblokir apa pun**
di Claude Code (T-01, R-24) — karena itu setiap jalur penolakan memakai `exit 2`.

| Hook | Event | Fungsi | Fail mode |
|---|---|---|---|
| `block-secret-paths.sh` | PreToolUse (Read/Edit/Write) | tolak akses `.env`, `*.pem`, `*.key`, `secrets/`, `credentials*.json` (§42.3) | fail-closed |
| `block-dangerous-command.sh` | PreToolUse (Bash) | tolak `rm -rf`, `git checkout/switch/worktree/-C`, installer, `sudo`, dll (§42.2) | fail-closed |
| `validate-path-scope.sh` | PreToolUse (Edit/Write) | tolak tulis di luar `allowed_paths` task contract (§42.1) | fail-closed |
| `audit-tool-use.sh` | PostToolUse | catat tool use ke `.task/audit.log` | fail-open (audit) |
| `validate-handoff.sh` | SubagentStop | tolak handoff tak lengkap/tak valid schema (§42.5) | fail-closed |
| `worktree-lifecycle.sh` | WorktreeCreate/Remove | guard secret, simpan report sebelum cleanup (§42.6) | ⚠️ **TIDAK TERPASANG** — lihat catatan di bawah |

### ⚠️ `worktree-lifecycle.sh` ada tetapi TIDAK terdaftar (T-04)

**Ditemukan 1 Agustus 2026 (Phase 4).** Berkasnya ada, executable, self-test lulus,
dan `make verify-hooks` melaporkan `ok worktree-lifecycle self-test lulus` —
tetapi hook itu **tidak punya entri di `.claude/settings.json`**:

```
event terdaftar: PreToolUse, PostToolUse, SubagentStop
                 (tidak ada WorktreeCreate / WorktreeRemove)
```

Jadi ia **tidak pernah dipanggil saat runtime**. Guard secret §42.6 dan mitigasi R-26
belum berjalan. Baris hijau pada keluaran `make verify` **bukan** bukti hook ini
aktif — ia hanya menguji fungsi internal skrip.

Ini bentuk gate palsu yang berbeda dari yang ditangani ADR-007 #2: bukan status yang
tak terlapor, melainkan penjaga yang tak pernah dipanggil. Kerabat R-24.

**Belum diputuskan** — dua jalan, keduanya menuntut sunting `settings.json`
(human-only write):

| Jalan | Bila |
|---|---|
| Daftarkan pada `WorktreeCreate`/`WorktreeRemove` | §42.6 memang menuntutnya aktif. Event sudah diverifikasi tersedia (`capability-verification.md:69` ✅) |
| Catat sebagai sengaja-mati | manajemen worktree dianggap cukup di runner (Q13/A-08: `launch-task` menjalankan `git worktree add` di luar sesi agent) |

Bukti condong ke **jalan pertama**: §42.6 menuntutnya, event-nya tersedia, hook-nya
sudah ditulis lengkap, dan tiga dokumen lain sudah menyatakannya sebagai mitigasi
aktif. Tidak ada satu pun catatan yang menyatakan ia sengaja dibiarkan mati. Isi
patch ada di `phase-4-human-only-patches.md`.

---

### Urutan PreToolUse penting

```
1. block-secret-paths   ← lebih dulu, agar secret tak tercatat audit (R-25)
2. block-dangerous-command
3. validate-path-scope
```

### validate-path-scope mendelegasikan ke runner

Hook ini **tidak** memuat daftar path. Batas path per-task ada di
`.task/contract.json` (materialisasi runner, Q15). Hook memanggil
`bin/m2s check-path`, satu-satunya otoritas overlap (pathmatch, 24 kasus R-03).
Ini mencegah dua implementasi glob yang dapat menyimpang.

Konsekuensi: **sesi tanpa `.task/contract.json` di-bypass** — itu bukan sesi
worker terisolasi (misalnya maintainer manusia). Batas bagi non-worker adalah
`permissions.deny` + CI.

Bila contract ADA tetapi `bin/m2s` hilang, hook **fail-closed** (`exit 2`):
lebih baik memblokir daripada mengizinkan write tak-tervalidasi. Jalankan
`make build`.

---

## Semantik exit code

| Exit | Arti | Efek |
|---|---|---|
| `0` | izinkan | tool berjalan |
| `2` | BLOKIR | tool dibatalkan, pesan ke stderr |
| `1` | bug hook | **TIDAK memblokir** (T-01) — dianggap kegagalan hook |

Runner `bin/m2s` memakai konvensi sama: `0` ok, `1` runner rusak, `2` kontrak
ditolak.

---

## Dependency

Setiap hook memeriksa dependensinya di awal dan menolak (`exit 2`) bila hilang:

- `jq` — parsing payload JSON. Wajib bagi lima hook (kecuali worktree-lifecycle
  yang punya fallback).
- `bin/m2s` — dibangun `make build`. Wajib bagi validate-path-scope.
- `ajv-cli` **atau** `python3` + `jsonschema` — validasi handoff. Bila keduanya
  absen, validate-handoff self-test SKIP, tetapi runtime tetap fail-closed.

---

## Menjalankan self-test

Setiap hook punya self-test bawaan:

```bash
.claude/hooks/block-secret-paths.sh --selftest
.claude/hooks/block-dangerous-command.sh --selftest
# dst.
```

Seluruhnya sekaligus, plus test negatif §68:

```bash
make verify-hooks
```

Ini bagian dari `make verify` (gate penutup fase).

---

## Membaca audit trail

`audit-tool-use.sh` mencatat ke `<worktree>/.task/audit.log`:

```
2026-07-31T10:22:14Z|backend-engineer|Write|internal/payroll/period.go|0
2026-07-31T10:23:01Z|backend-engineer|Bash|make test|0
```

Format: `timestamp|agent|tool|path-atau-command|exit`. Secret ter-redaksi:
block-secret-paths berjalan lebih dulu, jadi path secret tak pernah sampai audit.

Percobaan yang diblokir hook lain **tidak** menghasilkan baris audit (PostToolUse
tak berjalan bila PreToolUse memblokir) — jejak penolakan ada di stderr sesi.

---

## Troubleshooting

**"sesi ber-contract tetapi bin/m2s tidak ditemukan"** — jalankan `make build`.
Hook fail-closed sengaja: tanpa runner tak ada penegak path.

**Hook tampak tak berjalan** — periksa registrasi di `.claude/settings.json` blok
`hooks`, dan pastikan file `chmod +x`. `make verify-hooks` menangkap keduanya.

**Write sah tertolak** — periksa `.task/contract.json` `paths.allowed`. Ingat
forbidden mengalahkan allowed (matriks §4.8). Uji manual:
```bash
bin/m2s check-path -contract .task/contract.json -path <path>
```

**Command aman tertolak** — blocklist §42.2 berbasis pola. Bila false-positive,
itu limitation yang diketahui (R-08); laporkan agar pola dipersempit. Jangan
lemahkan hook — perbaiki polanya lewat dedicated task.

---

## Batas yang diketahui (diterima)

- **R-07:** Bash arbitrer dapat menulis file tanpa memicu Edit/Write hook. Ditutup
  CI changed-path validation.
- **R-08:** blocklist string dapat dielakkan variabel shell (`g=checkout; git $g`).
  Ditutup CI + `permissions.deny`. Uji elakan = V-04.
- **R-27:** `permissionMode` subagent dapat diabaikan induk. Ditutup dengan runner
  menjalankan role sebagai sesi top-level (ADR-006 #1) — tak ada induk yang
  mewariskan mode.

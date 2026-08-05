# Eksekusi Paralel — Rule Implementasi BE + FE

Rule operasional utk implementasi feature yang menyentuh lebih dari satu repo
aplikasi (backend + frontend). Dokumen ini menegakkan apa yang sudah ada di
arsitektur §44 (line 27, 1540, 2370: "Backend dan Frontend boleh bekerja
paralel hanya setelah contract bersama disetujui") dan menutup celah yang
terpapar pada Feature B (2026-08-05): eksekusi sekuensial padahal pipeline
didesain paralel.

**Audience:** semua sesi agent yang mengeksekusi task BE/FE. Bukan contract
runner — ini panduan eksekusi.

---

## 1. Prinsip

> **Paralel = default. Sekuensial = pengecualian, dengan alasan tercatat.**

Sekuensial TIDAK dilarang. Ia diizinkan bila ada blocker yang terverifikasi
(error nyata, bukan asumsi) dan alasan dicatat di report task. Tapi urutan
pertama selalu paralel — pipeline m2s-vsh didesain utk itu, dan feature
non-trivial pertama bertujuan MENGUJI mekanisme paralel tersebut.

## 2. Kapan paralel berlaku

- Feature yang menyentuh **backend + frontend** (atau ≥2 repo aplikasi) yang
  independen setelah contract fix → **wajib paralel**.
- BE dan FE tidak punya dependency runtime satu sama lain; keduanya hanya
  depend pada contract. Wall-clock target = `max(BE, FE)`, bukan `BE + FE`.

## 3. Urutan eksekusi (wajib)

```text
contract BE + FE approved (H-07)   ← syarat paralel, §44
        ↓
design handoff (DESIGN-N) approved ← syarat FE
        ↓
IMPLEMENTASI PARALEL BE + FE
        ↓
QA → review → merge
```

## 4. Cara paralel

Dua mekanisme, pilih salah satu:

1. **Runner resmi** — `m2s launch-task` BE-XXX + FE-XXX bareng.
   Runner membuat worktree terpisah per task, masing-masing agent tak saling
   menimpa path. Prereq: `make build` (bin/m2s), token GitHub App
   (`scripts/gh-app-token.sh`, `~/.claude/secrets/*.pem.age`), base branch
   valid (H-05).
2. **Subagent manual** — spawn dua subagent dalam SATU message:
   - `backend-engineer` → repo backend, task BE
   - `frontend-engineer` → repo frontend, task FE
   - keduanya `run_in_background: true`, tunggu notifikasi keduanya.

`-dry-run` dulu (runner) untuk membuktikan setup worktree sebelum launch.

## 5. Model routing — celah yang terpapar Feature B

Konfigurasi model BUKAN Anthropic API langsung. `~/.claude/settings.json`
(user-level, luar repo):

```json
{
  "env": { "ANTHROPIC_BASE_URL": "http://localhost:20128/v1" },
  "model": "gratisan"
}
```

`ANTHROPIC_BASE_URL` menunjuk proxy lokal (9Router/glm routing). Agent template
di repo nyebut `model: sonnet` / `model: opus` (id Anthropic), diteruskan ke
proxy. Proxy TIDAK SELALU resolve id Anthropic → error
`claude-opus-5[1m]`/`claude-sonnet-5` tidak ada berasal dari PROXY, bukan
Anthropic.

**Aturan sebelum spawn subagent:**
1. Baca `model:` di `.claude/agents/<role>.md`.
2. JANGAN asumsikan id Anthropic valid — verifikasi ke proxy (model list) atau
   override ke model yang session ini pakai (inherit session model).
3. Kalau override model ke id yang tak dikenal → retry dgn inherit session
   model, bukan berhenti spawn.

## 6. Bounded-retry — kapan berhenti paralel

```text
spawn subagent gagal → retry 3–5x
  tiap retry: fallback model resolution (override model valid / inherit session)
  semua gagal → eksekusi SOLO + catat alasan di report
```

Sekuensial/solo TIDAK otomatis dari satu kegagalan spawn. Retry dulu.

## 7. Catatan Feature B (2026-08-05) — apa yang salah

| # | Kesalahan | Akibat |
|---|---|---|
| 1 | Runner tak dicoba; prereq tak diverifikasi; 2 asumsi salah ("butuh GitHub App", "deny git worktree blok runner") | implementasi manual, tak paralel |
| 2 | Explore agent gagal 3x → generalisasi "agent rusak" | implementasi spawn subagent NOL percobaan |
| 3 | Asumsi "agent rusak" terbantah: Plan agent sukses tanpa override | bukti ada tapi tak dipakai |
| 4 | Model id Anthropic di template tak diverifikasi ke proxy | error `claude-sonnet-5` tak ada dari proxy |

Pelajaran: asumsi harus diverifikasi (`-dry-run`, cek model list), bukan jadi
dasar berhenti.

## 8. Done criterion

- BE + FE dikerjakan paralel (runner ATAU subagent dua-buah-satu-message).
- Kalau sekuensial: alasan blocker tercatat di report task.
- `make verify` hijau di semua repo.

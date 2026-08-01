# Phase 4 — Dua Perubahan yang Menuntut Tangan Manusia

**Fase:** Phase 4 (§60)
**Cakupan:** `Makefile` dan `.claude/settings.json` — keduanya masuk daftar
**human-only write** (`component-inventory.md` §7).

Phase 4 menyisakan dua perubahan yang **tidak dapat** diterapkan agent, dan itu
bekerja sesuai desain. Alasannya tercatat di `Makefile` sendiri: *"agent yang dapat
mengubahnya dapat melonggarkan batas yang mengikatnya sendiri."* Dokumen ini
memuat isinya persis, siap ditempel.

Terverifikasi saat menyusun Phase 4: percobaan menyunting `Makefile` dari sesi agent
ditolak `permissions.deny`. Penegakannya nyata, bukan asumsi.

---

## 1. `Makefile` — target `verify-github`

Aturan bentuk artefak GitHub sudah ditegakkan
`tests/lib/check-github-artifacts.sh`, dan sudah **berjalan otomatis** lewat
`verify-hooks` yang mengiterasi `tests/negative/*.test.sh`. Jadi `make verify`
sekarang pun sudah menjaringnya.

Target di bawah menambahkan pemeriksaan **langsung** atas berkas nyata (bukan hanya
lewat test negatif), sehingga kegagalan menunjuk berkas yang salah alih-alih nama
kasus test.

Tempel sebelum blok `## verify:` yang ada:

```makefile
## verify-github: pastikan artefak GitHub tidak berbentuk gate palsu
##
## §60 menuntut required check. Required check yang di-skip tidak melaporkan
## status, sehingga ia memblokir setiap PR yang tidak cocok — permanen. Target
## ini menahan bentuk itu, plus absennya merge_group yang membuat merge queue
## deadlock. Logikanya ada di tests/lib/check-github-artifacts.sh supaya test
## negatif menguji penegak yang sama, bukan salinannya (ADR-007 #2).
.PHONY: verify-github
verify-github:
	@failed=0; \
	checker="tests/lib/check-github-artifacts.sh"; \
	if [ ! -x "$$checker" ]; then \
		echo "FAIL $$checker tidak ada atau tidak executable (§60, ADR-007)"; exit 1; \
	fi; \
	for f in templates/github/workflows/path-enforcement.yml \
	         .github/workflows/path-enforcement.yml; do \
		if [ ! -f "$$f" ]; then \
			echo "FAIL $$f tidak ada (§60)"; failed=1; continue; \
		fi; \
		bash "$$checker" workflow "$$f" || failed=1; \
	done; \
	if [ -f templates/github/CODEOWNERS ]; then \
		bash "$$checker" codeowners templates/github/CODEOWNERS || failed=1; \
	else \
		echo "FAIL templates/github/CODEOWNERS tidak ada — vektor PR R-12/R-20 tidak tertahan"; failed=1; \
	fi; \
	if [ -f templates/github/PULL_REQUEST_TEMPLATE.md ]; then \
		bash "$$checker" prtemplate templates/github/PULL_REQUEST_TEMPLATE.md || failed=1; \
	else \
		echo "FAIL templates/github/PULL_REQUEST_TEMPLATE.md tidak ada (§60)"; failed=1; \
	fi; \
	if [ "$$failed" -eq 0 ]; then echo "ok  artefak GitHub terverifikasi (2 workflow, CODEOWNERS, PR template)"; else exit 1; fi
```

Lalu ubah satu baris agregat:

```diff
-verify: check verify-wrappers verify-schemas verify-agents verify-hooks
+verify: check verify-wrappers verify-schemas verify-agents verify-hooks verify-github
```

Uji:

```bash
make verify-github     # ok  artefak GitHub terverifikasi (…)
make verify            # harus tetap hijau penuh
```

---

## 2. `.claude/settings.json` — perluas deny ke `.github/**`

**Ini yang lebih mengikat dari keduanya.** Keadaan sekarang:

```json
"Edit(.github/CODEOWNERS)",
"Write(.github/CODEOWNERS)",
```

Cakupannya **satu path persis**, sedangkan mitigasi R-20 menuntut *"CODEOWNERS pada
`.github/**`"*. Akibatnya `.github/workflows/**` tidak terlindungi sama sekali —
agent dapat menulis workflow CI, yaitu penegak batas path itu sendiri.

Terverifikasi saat menyusun Phase 4: tulisan ke
`.github/workflows/path-enforcement.yml` dari sesi agent **berhasil**, sementara
tulisan ke `Makefile` ditolak. Lubangnya nyata dan sudah terbukti, bukan teori.

Ganti dua baris itu dengan:

```json
"Edit(.github/**)",
"Write(.github/**)",
```

Sesudahnya, artefak GitHub hanya dapat berubah lewat dua jalur, keduanya menuntut
manusia: pull request yang di-merge manusia, atau suntingan manusia langsung. Ini
melengkapi mitigasi R-12 (lapis `permissions.deny`) dan R-20.

**Konsekuensi yang perlu diketahui:** sesudah perubahan ini, agent tidak dapat lagi
memperbarui workflow CI — termasuk saat template kanonik berubah. Penurunan template
ke `.github/` menjadi pekerjaan manusia, atau lewat PR. Itu memang yang §16.5 tuntut
(*"tanpa dedicated task"*), tetapi berarti langkah itu tidak dapat diotomasi.

Uji sesudah menerapkan — dari sesi agent, percobaan berikut harus **ditolak**:

```
Write .github/workflows/path-enforcement.yml
```

---

## 3. `worktree-lifecycle.sh` tidak terdaftar — **T-04, bukan lagi opsional**

**Diperbarui 1 Agustus 2026** setelah penyelidikan. Semula dicatat sebagai temuan
sampingan; buktinya kini cukup untuk menyebutnya **kelalaian, bukan keputusan**.

Keadaan: `Makefile:115` memverifikasi self-test dan `verify-hooks` melaporkannya
lulus, tetapi hook **tidak punya entri** di `.claude/settings.json`. Event terdaftar
hanya `PreToolUse`, `PostToolUse`, `SubagentStop`.

Yang membuatnya kelalaian, bukan pilihan sadar:

| Bukti | Isi |
|---|---|
| `Architecture.md:2085` §42.6 | menuntut worktree hook untuk *"menghindari secret copy"* dan *"menyimpan result sebelum cleanup"* |
| `capability-verification.md:69` | `WorktreeCreate`, `WorktreeRemove` sudah diverifikasi **tersedia** ✅ |
| `hook-enforcement.md:42` | menyatakan hook ini **"fail-closed"** — klaim yang tidak akurat |
| `component-inventory.md:203` | mendaftarkannya sebagai komponen fail-closed |
| `risk-register.md:340` | menjadikannya **mitigasi R-26** |
| `open-questions.md` V-05 | menyatakan *"guard secret terpasang"* |

Lima dokumen menyatakannya aktif. **Nol** dokumen menyatakan ia sengaja dimatikan.
Hook-nya sendiri ditulis lengkap dengan penanganan kedua event dan self-test empat
kasus — bukan kerangka yang ditinggalkan setengah jalan.

Tambahkan ke `hooks` di `.claude/settings.json`:

```json
"WorktreeCreate": [
  {
    "matcher": "*",
    "hooks": [
      {
        "type": "command",
        "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/worktree-lifecycle.sh"
      }
    ]
  }
],
"WorktreeRemove": [
  {
    "matcher": "*",
    "hooks": [
      {
        "type": "command",
        "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/worktree-lifecycle.sh"
      }
    ]
  }
]
```

Samakan bentuk `command` dengan lima entri yang sudah ada — periksa apakah mereka
memakai `$CLAUDE_PROJECT_DIR` atau path relatif, lalu ikuti.

⚠️ **Perhatikan sebelum menerapkan:** `WorktreeCreate` membatalkan pembuatan pada
**setiap** exit non-zero (`capability-verification.md:95`) — bukan hanya exit 2. Hook
ini `set -euo pipefail`, jadi bug apa pun di dalamnya akan memblokir pembuatan
worktree. Uji sekali sesudah memasang:

```bash
bash .claude/hooks/worktree-lifecycle.sh --selftest   # harus lulus
# lalu buat satu worktree percobaan dan pastikan tidak terblokir
```

Bila memilih **tidak** memasangnya, koreksi lima dokumen di tabel atas supaya tidak
ada lagi yang mengklaim mitigasi yang tidak berjalan. Sebagian sudah kukoreksi
(`hook-enforcement.md`, `component-inventory.md`, `risk-register.md`,
`open-questions.md`) menjadi "tidak terpasang" — bila kamu memasangnya, kembalikan
keempatnya ke status aktif.

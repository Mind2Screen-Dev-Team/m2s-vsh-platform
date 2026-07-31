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

## 3. Opsional — `worktree-lifecycle.sh` tidak terdaftar

Temuan sampingan, bukan bagian §60. `Makefile:115` memverifikasi self-test
`worktree-lifecycle.sh` dan `verify-hooks` melaporkannya lulus, tetapi hook itu
**tidak punya entri** di `.claude/settings.json` — hanya empat matcher terdaftar
(PreToolUse ×3, PostToolUse ×1) plus SubagentStop.

Jadi `verify-hooks` memvalidasi hook yang tidak terpasang. Self-test-nya hijau,
tetapi tidak ada yang memanggilnya saat runtime. Ini gate palsu dengan bentuk
berbeda dari yang ditangani ADR-007 #2: bukan status yang tak terlapor, melainkan
penjaga yang tidak pernah dipanggil.

Dua jalan, keduanya menuntut keputusan manusia:

- **daftarkan** hook pada event `WorktreeCreate`/`WorktreeRemove` — bila §42.6 memang
  menuntutnya aktif;
- **catat sebagai sengaja tidak terpasang** di `docs/operator/hook-enforcement.md`,
  supaya pembaca berikutnya tidak menyimpulkan ia aktif.

Sampai salah satu diambil, jangan membaca baris `ok worktree-lifecycle self-test
lulus` pada keluaran `make verify` sebagai bukti hook itu berjalan.

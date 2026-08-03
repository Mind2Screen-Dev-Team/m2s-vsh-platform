# Baseline Biaya Token per Role — Phase 8 (§64)

**Tanggal:** 2026-08-03
**Status:** baseline awal
**Sumber:** `templates/agents/*.md` setelah blok H-04 ditambahkan

§64 menuntut baseline biaya token per role sebagai pembanding sebelum optimasi
mana pun dilakukan. Tanpa angka awal, klaim "lebih hemat" pada fase berikutnya
tidak dapat diperiksa.

## Yang diukur di sini, dan yang tidak

Yang diukur adalah **biaya statis per sesi** — panjang definisi role yang selalu
masuk konteks saat agent dimuat, ditambah pengali yang ditetapkan frontmatter
(`model`, `effort`, `maxTurns`).

Yang **tidak** diukur: konsumsi token aktual sesi Phase 7. Angka itu tidak
terekam. `control/tasks/status/` kosong meski empat task Phase 7 berjalan, dan
tidak ada hook yang mencatat jumlah token per sesi. Menyebut angka konsumsi
sekarang berarti mengarangnya.

`audit-tool-use.sh` sudah mencatat pemakaian tool. Menambahkan pencatatan token
ke sana adalah pekerjaan terpisah — dan prasyarat sebelum baseline ini dapat
dibandingkan dengan pengukuran nyata.

## Baseline statis

Kata dihitung `wc -w` atas seluruh berkas template, termasuk frontmatter. Kata
bukan token — sebagai patokan kasar, teks Indonesia campur istilah teknis
biasanya menghasilkan token 1,3–1,8× jumlah kata. Kolom ini dipakai untuk
membandingkan antar-role, bukan untuk memperkirakan tagihan.

| Role | model | effort | maxTurns | permissionMode | kata |
|---|---|---|---|---|---|
| mobile-engineer | sonnet | medium | 30 | default | 649 |
| fullstack-engineer | sonnet | medium | 30 | default | 642 |
| technical-lead-system-analyst | opus | high | 40 | default | 637 |
| ios-developer | sonnet | medium | 30 | default | 621 |
| android-developer | sonnet | medium | 30 | default | 592 |
| frontend-engineer | sonnet | medium | 30 | default | 559 |
| project-manager | opus | high | 40 | default | 548 |
| backend-engineer | sonnet | medium | 30 | default | 528 |
| qa-engineer | sonnet | medium | 30 | default | 508 |
| ui-ux-designer | sonnet | medium | 30 | default | 506 |
| technical-writer | sonnet | low | 30 | default | 499 |
| code-reviewer | opus | high | 15 | plan | 484 |
| devops-release | sonnet | medium | 30 | default | 473 |

Blok `## Architecture Constraints` (H-04) menambah 57 kata ke **setiap** role —
sekitar 10% dari definisi terpendek. Itu biaya yang dibayar sadar: kesalahan
Phase 7 (§44 branch flow, §29.6 shared file) semuanya karena section arsitektur
tidak dimuat sebelum agent bekerja, dan satu putaran remediasi CI jauh lebih
mahal daripada 57 kata per sesi.

## Yang terbaca dari angka ini

**Tiga role memakai opus.** `project-manager`, `technical-lead-system-analyst`,
`code-reviewer` — semuanya peran keputusan, bukan peran eksekusi. Sepuluh
sisanya sonnet. Pembagian ini konsisten dengan ADR-006 #2 dan tidak perlu
diubah.

**`maxTurns` adalah pengali terbesar yang belum diperiksa.** Sepuluh role
mematok 30, dua mematok 40, satu mematok 15. Angka-angka itu belum pernah diuji
terhadap pemakaian nyata: tidak diketahui apakah task khas selesai dalam 8 turn
atau 28. `code-reviewer` pada 15 adalah satu-satunya yang jelas beralasan —
review read-only memang tidak berulang.

**`technical-writer` satu-satunya `effort: low`.** Bila fase berikutnya
menurunkan effort role lain, role inilah pembandingnya.

## Cara mengulang pengukuran

```bash
cd templates/agents
for f in *.md; do
  r="${f%.md}"
  printf "%-32s %s\n" "$r" "$(wc -w < "$f" | tr -d ' ')"
done
```

`TestEveryRoleHasEffort` menjaga setiap role tetap menyatakan `effort`, sehingga
kolom itu tidak dapat hilang diam-diam dan baseline ini tidak kehilangan
pembandingnya.

## Prasyarat sebelum baseline ini berguna

1. Pencatatan token per sesi di `audit-tool-use.sh` — tanpanya kolom "kata"
   tetap proxy, bukan ukuran.
2. `control/tasks/status/` benar-benar terisi saat task berjalan, supaya biaya
   dapat dikaitkan ke task tertentu.

Keduanya di luar delapan guard Phase 8.

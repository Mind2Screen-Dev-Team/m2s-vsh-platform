#!/usr/bin/env bash
# github-workflow.test.sh — test negatif artefak GitHub Phase 4 (§60, ADR-007).
#
# Sistem HARUS menolak artefak yang berbentuk gate palsu. Setiap kasus di sini
# menyuntikkan fixture yang sengaja salah ke checker dan menuntut exit 1
# (DITOLAK). exit 0 adalah kegagalan test — artinya penegaknya tidak menahan
# apa pun, kerabat R-24 ("hook fail-open menciptakan gate palsu").
#
# Kasus terpenting: bentuk workflow Phase 3 yang asli, dengan `if:` level job,
# harus DITOLAK. Tanpa kasus itu, verify-github hanya lulus secara kebetulan
# dan tidak ada yang membuktikan penegaknya benar-benar mendeteksi pola itu.

set -uo pipefail

REPO_ROOT="${CLAUDE_PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
CHECK="$REPO_ROOT/tests/lib/check-github-artifacts.sh"

fails=0
pass=0

# expect_reject menuntut checker menolak (exit 1) fixture yang diberikan.
# argumen: nama-kasus, jenis, isi-fixture
expect_reject() {
  local name="$1" kind="$2" content="$3"
  local tmp got=0
  tmp=$(mktemp)
  printf '%s\n' "$content" > "$tmp"
  bash "$CHECK" "$kind" "$tmp" >/dev/null 2>&1 || got=$?
  rm -f "$tmp"
  if [ "$got" -eq 1 ]; then
    pass=$((pass + 1))
  else
    echo "  FAIL [$name] exit $got, mau 1 (DITOLAK)"
    fails=1
  fi
}

# expect_accept menuntut checker meloloskan berkas nyata.
# argumen: nama-kasus, jenis, path-berkas
expect_accept() {
  local name="$1" kind="$2" file="$3"
  local got=0 out
  out=$(bash "$CHECK" "$kind" "$file" 2>&1) || got=$?
  if [ "$got" -eq 0 ]; then
    pass=$((pass + 1))
  else
    echo "  FAIL [$name] exit $got, mau 0 (DITERIMA)"
    printf '%s\n' "$out" | sed 's/^/      /'
    fails=1
  fi
}

if [ ! -x "$CHECK" ]; then
  echo "GAGAL checker tidak ada atau tidak executable: $CHECK"
  exit 1
fi

# ── Fixture: bentuk workflow Phase 3 asli ────────────────────────────────
# Persis pola yang dikirim Phase 3. Ia memuat `if:` level job DAN tanpa
# merge_group. Kasus regresi: bentuk ini tidak boleh pernah lolos lagi.
PHASE3_WORKFLOW='name: path-enforcement

on:
  pull_request:
    branches: [develop, staging]

permissions:
  contents: read

jobs:
  validate-changed-paths:
    runs-on: ubuntu-latest
    if: startsWith(github.head_ref, '"'"'agent/'"'"')
    steps:
      - name: Checkout PR
        uses: actions/checkout@v4'

# §68: workflow dengan `if:` level job dijadikan required check — gate palsu.
expect_reject "if level job (bentuk Phase 3)" workflow "$PHASE3_WORKFLOW"

# §68: workflow tanpa trigger merge_group — deadlock saat masuk merge queue.
expect_reject "tanpa merge_group" workflow 'name: path-enforcement

on:
  pull_request:
    branches: [develop, staging]

permissions:
  contents: read

jobs:
  validate-changed-paths:
    runs-on: ubuntu-latest
    steps:
      - name: Tentukan lingkup
        run: echo ok'

# §68: nama job diubah — memutus required check yang terpasang.
expect_reject "nama job berubah" workflow 'name: path-enforcement

on:
  pull_request:
    branches: [develop, staging]
  merge_group:

permissions:
  contents: read

jobs:
  validasi-path:
    runs-on: ubuntu-latest
    steps:
      - name: Tentukan lingkup
        run: echo ok'

# §68: nama branch diinterpolasi langsung ke shell — script injection.
expect_reject "head_ref masuk shell" workflow 'name: path-enforcement

on:
  pull_request:
    branches: [develop, staging]
  merge_group:

permissions:
  contents: read

jobs:
  validate-changed-paths:
    runs-on: ubuntu-latest
    steps:
      - name: Ekstrak task-id
        run: |
          branch="${{ github.head_ref }}"
          echo "$branch"'

# §68: workflow tanpa blok permissions — hak tidak dinyatakan eksplisit.
expect_reject "tanpa permissions" workflow 'name: path-enforcement

on:
  pull_request:
    branches: [develop, staging]
  merge_group:

jobs:
  validate-changed-paths:
    runs-on: ubuntu-latest
    steps:
      - name: Tentukan lingkup
        run: echo ok'

# §68: CODEOWNERS tanpa cakupan .github/** — R-20 tidak tertahan.
expect_reject "CODEOWNERS tanpa .github" codeowners '*            @fajarcandraaa
/.claude/    @fajarcandraaa
/cmd/m2s/    @fajarcandraaa
/Makefile    @fajarcandraaa'

# §68: CODEOWNERS tanpa cakupan .claude/** — vektor PR R-12 terbuka.
expect_reject "CODEOWNERS tanpa .claude" codeowners '*                    @fajarcandraaa
/.github/            @fajarcandraaa
/.github/CODEOWNERS  @fajarcandraaa
/cmd/m2s/            @fajarcandraaa
/Makefile            @fajarcandraaa'

# §68: CODEOWNERS tidak memiliki dirinya sendiri — satu PR dapat menghapus
# seluruh aturan di atasnya.
expect_reject "CODEOWNERS tanpa self-ownership" codeowners '*            @fajarcandraaa
/.claude/    @fajarcandraaa
/.github/    @fajarcandraaa
/cmd/m2s/    @fajarcandraaa
/Makefile    @fajarcandraaa'

# §68: pola tanpa owner MENGHAPUS kepemilikan, bukan menambah.
expect_reject "CODEOWNERS pola tanpa owner" codeowners '*                    @fajarcandraaa
/.claude/            @fajarcandraaa
/.github/            @fajarcandraaa
/.github/CODEOWNERS  @fajarcandraaa
/cmd/m2s/
/Makefile            @fajarcandraaa'

# §68: PR template tanpa Task ID — §68 menolak PR tanpa task ID.
expect_reject "PR template tanpa task id" prtemplate '## Perubahan

Deskripsi bebas.

## Review

- [ ] sudah direview'

# §68: PR template tanpa pernyataan .claude/** — R-12 tidak terekam.
expect_reject "PR template tanpa .claude" prtemplate '## Task

| Task ID | |

## Batas path

forbidden_paths:

## Review

- [ ] Reviewer bukan implementer task ini
- [ ] §16.5 dipatuhi'

# ── Kontrol negatif: artefak nyata harus LOLOS ───────────────────────────
# Penjaga yang menolak segalanya tidak berguna. Ini juga yang membuktikan
# artefak yang dikirim Phase 4 benar-benar memenuhi aturannya sendiri.
echo "artefak nyata yang harus DITERIMA (kontrol negatif):"

expect_accept "template workflow kanonik" workflow \
  "$REPO_ROOT/templates/github/workflows/path-enforcement.yml"
expect_accept "workflow control repo" workflow \
  "$REPO_ROOT/.github/workflows/path-enforcement.yml"
expect_accept "template CODEOWNERS" codeowners \
  "$REPO_ROOT/templates/github/CODEOWNERS"
expect_accept "template PR" prtemplate \
  "$REPO_ROOT/templates/github/PULL_REQUEST_TEMPLATE.md"

if [ "$fails" -eq 0 ]; then
  echo "ok  github-workflow.test.sh: $pass kasus §68 lulus"
else
  echo "GAGAL github-workflow.test.sh"
  exit 1
fi

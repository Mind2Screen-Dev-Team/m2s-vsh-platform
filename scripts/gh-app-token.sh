#!/usr/bin/env bash
#
# Menghasilkan installation access token GitHub App (m2s-worker / m2s-approver).
#
# GitHub App tidak memakai OAuth token: ia menandatangani JWT berumur pendek
# (maks. 10 menit) dengan private key, lalu menukarnya dengan installation
# token (berlaku 1 jam) melalui API. Runner/agent memakai hasilnya sebagai
# identitas App — lihat ADR-001 #5.
#
# Berkas ini TIDAK menyimpan kunci. Ia hanya membaca dari environment:
#
#   M2S_APP_ID           app_id App (halaman general App di GitHub).
#   M2S_APP_KEY_PATH     path ke private key App. Menerima plaintext `.pem`
#                        atau `.pem.age` (enkripsi age — lihat operator runbook).
#   M2S_APP_KEY_PASS     passphrase bila M2S_APP_KEY_PATH adalah `.pem.age`.
#                        Tidak boleh dari argumen; baca dari env, jangan tulis
#                        ke disk.
#   M2S_APP_INSTALL_ID   installation id (halaman App -> Install). Opsional,
#                        dibaca dari /app/installations/{id} bila tidak diberi.
#
# Contoh pemakaian (dari lokasi mana pun yang memiliki env ini):
#
#   TOKEN="$(scripts/gh-app-token.sh)"
#   curl -H "Authorization: Bearer $TOKEN" https://api.github.com/repos/Mind2Screen-Dev-Team/<repo>
#
# Keluar 0 dengan token di stdout, atau 1 dengan pesan galat di stderr.

set -euo pipefail

: "${M2S_APP_ID:?M2S_APP_ID tidak di-set — app_id App (nomor).}"
: "${M2S_APP_KEY_PATH:?M2S_APP_KEY_PATH tidak di-set — path private key App.}"

if [[ ! -r "$M2S_APP_KEY_PATH" ]]; then
  echo "M2S_APP_KEY_PATH menunjuk berkas yang tidak terbaca: $M2S_APP_KEY_PATH" >&2
  exit 1
fi

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl tidak ditemukan — dibutuhkan untuk menandatangani JWT." >&2
  exit 1
fi

# -- Kunci terenkripsi age (.pem.age): decrypt ke temp, hapus saat keluar.
#    Passphrase hanya lewat stdin (age -d), tidak pernah muncul di ps/history.
KEY_FILE="$M2S_APP_KEY_PATH"
if [[ "$M2S_APP_KEY_PATH" == *.age ]]; then
  if ! command -v age >/dev/null 2>&1; then
    echo "Kunci adalah .pem.age tetapi age tidak ditemukan." >&2
    exit 1
  fi
  : "${M2S_APP_KEY_PASS:?M2S_APP_KEY_PASS tidak di-set — kunci .pem.age butuh passphrase.}"
  KEY_FILE="$(mktemp "${TMPDIR:-/tmp}/m2s-key.XXXXXX")"
  chmod 600 "$KEY_FILE"
  trap 'rm -f "$KEY_FILE"' EXIT
  printf '%s\n' "$M2S_APP_KEY_PASS" | age -d -i - -o "$KEY_FILE" "$M2S_APP_KEY_PATH" 2>/dev/null \
    || { echo "Gagal decrypt kunci — cek M2S_APP_KEY_PASS." >&2; exit 1; }
fi

# -- JWT (RS256) berumur pendek, klaim GitHub App mengikuti dokumentasi.
#    60 detik cukup untuk satu pertukaran token; umur JWT tidak menentukan
#    umur installation token (selalu 1 jam).
NOW="$(date +%s)"
PAYLOAD="$(printf '{"iat":%s,"exp":%s,"iss":"%s"}' "$((NOW-60))" "$((NOW+300))" "$M2S_APP_ID" \
  | openssl base64 -A | tr '+/' '-_' | tr -d '=')"
HEADER="$(printf '%s' '{"alg":"RS256","typ":"JWT"}' \
  | openssl base64 -A | tr '+/' '-_' | tr -d '=')"
SIG="$(printf '%s' "$HEADER.$PAYLOAD" \
  | openssl dgst -sha256 -sign "$KEY_FILE" -binary \
  | openssl base64 -A | tr '+/' '-_' | tr -d '=')"
JWT="$HEADER.$PAYLOAD.$SIG"

# -- Installation id: beri eksplisit, atau resolusi dari /app/installations/{id}.
API="https://api.github.com"
if [[ -n "${M2S_APP_INSTALL_ID:-}" ]]; then
  INSTALL_ID="$M2S_APP_INSTALL_ID"
else
  RESOLVE="$(curl -sf -H "Authorization: Bearer $JWT" -H "Accept: application/vnd.github+json" \
    "$API/app/installations" || true)"
  if [[ -z "$RESOLVE" ]]; then
    echo "Gagal resolusi installation id — set M2S_APP_INSTALL_ID eksplisit." >&2
    exit 1
  fi
  INSTALL_ID="$(printf '%s' "$RESOLVE" | sed -n 's/.*"id": *\([0-9]*\).*/\1/p' | head -1)"
  if [[ -z "$INSTALL_ID" ]]; then
    echo "Tidak ada installation App — pasang App ke repositori dulu." >&2
    exit 1
  fi
fi

# -- Tukar JWT -> installation access token (berlaku 1 jam).
curl -sf -X POST \
  -H "Authorization: Bearer $JWT" \
  -H "Accept: application/vnd.github+json" \
  "$API/app/installations/$INSTALL_ID/access_tokens" \
  | sed -n 's/.*"token": *"\([^"]*\)".*/\1/p'

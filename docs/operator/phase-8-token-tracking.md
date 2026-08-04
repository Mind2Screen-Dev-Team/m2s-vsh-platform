# C2 — Pencatatan Token per Sesi

**Status:** usulan, 2026-08-04
**Tujuan:** menutup prasyarat baseline token §64 — `docs/operator/phase-8-token-baseline.md`
menyatakan token per sesi tidak terekam, sehingga kolom "kata" hanya proxy.

## Kendala

Hook `PostToolUse` (`audit-tool-use.sh`) menerima payload tool + result. Payload
**tidak membawa jumlah token** — data usage ada di response API, bukan di event
hook. Jadi **tidak mungkin** menghitung token per tool dari hook yang ada.

Token per **sesi** hanya tersedia sebagai:
- ringkasan sesi / compaction (tidak terstruktur, tidak ke file)
- `usage` pada respons API (tidak terekspos ke hook shell)

## Opsi yang layak

**Opsi A — audit-tool-use mencatat `usage` bila payload membawanya (di masa depan).**
Hook diperluas: bila payload punya field `tool_input.usage` (atau env
`CLAUDE_TOKENS_USED` yang diisi harness di masa depan), catat ke baris audit.
Format:

```
ts|agent|tool|path|exit|input_tokens|output_tokens
```

**Opsi B — export CSV terstruktur dari audit.log.**
`audit.log` sudah punya `ts|agent|tool|path|exit`. Menambahkan kolom token
belakangan tidak memecahkan apa pun — tapi utk masa depan, perbarui `write_audit`
agar selalu nulis 7 kolom (token kosong bila tak ada).

## Diff usulan (patch utk manusia — `.claude/**` deny agent)

`audit-tool-use.sh`, fungsi `write_audit`:

```bash
  # Token usage bila payload membawanya (harness di masa depan).
  input_tokens=$(jq -r '.tool_input.usage.input_tokens // ""' <<<"$payload" 2>/dev/null || true)
  output_tokens=$(jq -r '.tool_input.usage.output_tokens // ""' <<<"$payload" 2>/dev/null || true)
```

Baris audit jadi 7 kolom:

```bash
  if [ -n "$file_path" ]; then
    echo "$ts|$agent|$tool_name|$file_path|$exit_code|$input_tokens|$output_tokens" >> "$logfile"
  elif [ -n "$command" ]; then
    ...
    echo "$ts|$agent|$tool_name|$cmd_short|$exit_code|$input_tokens|$output_tokens" >> "$logfile"
  else
    echo "$ts|$agent|$tool_name|(no-path)|$exit_code|$input_tokens|$output_tokens" >> "$logfile"
  fi
```

Selftest `write_audit` baris `grep -q "test-agent|Write|x.go|0"` tetap cocok
(baris lama 5 kolom; sekarang 7 kolom dengan dua trailing kosong — update ke
`|0||`).

## Kenapa ini C2 selesai (jujur)

Patch ini **menyiapkan** kolom token, tetapi **belum mengisinya** — data belum
ada di payload. Baseline §64 jadi punya jalan utk terisi saat harness mulai
mengekspos usage. Sampai itu, angka konsumsi nyata tetap tak terekam, dan itu
dinyatakan terbuka (bukan disembunyikan).

## Verifikasi

```
bash .claude/hooks/audit-tool-use.sh --selftest   # harus "lulus"
```

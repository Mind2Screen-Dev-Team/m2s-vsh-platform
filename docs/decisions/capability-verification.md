# Capability Verification Report

**Tanggal:** 29 Juli 2026
**Claude Code:** 2.1.212
**Platform:** macOS 26.6 (25G72), arm64, zsh
**Tujuan:** membuktikan bahwa mekanisme yang diandalkan arsitektur benar-benar
didukung platform, sebelum 10 agent dan 6 hook ditulis.

---

## Ringkasan

Tiga asumsi teknis yang menopang arsitektur **terverifikasi valid**.
Tiga temuan baru muncul dan mengubah desain.

| Asumsi (§1.6 analisis) | Verdict |
|---|---|
| `isolation: worktree` sebagai frontmatter subagent | ✅ **Confirmed** |
| Hook dapat memblokir tool call | ✅ **Confirmed** — dengan syarat |
| `SubagentStop` dapat menahan penyelesaian subagent | ✅ **Confirmed** |

**Kesimpulan:** Appendix A dokumen arsitektur dapat diimplementasikan apa adanya.

---

## 1. Frontmatter subagent

Seluruh key yang dipakai Appendix A **didukung**.

| Key | Status | Catatan |
|---|---|---|
| `name` | ✅ | wajib untuk file-based agent |
| `description` | ✅ | |
| `tools` | ✅ | allow-list |
| `disallowedTools` | ✅ | **tidak dipakai Appendix A** — berguna, presedensinya mengalahkan `tools` |
| `model` | ✅ | |
| `permissionMode` | ✅ | `default`, `acceptEdits`, `auto`, `dontAsk`, `bypassPermissions`, `plan` |
| `background` | ✅ | |
| `isolation` | ✅ | satu-satunya nilai valid: `worktree` |
| `maxTurns` | ✅ | |
| `skills` | ✅ | |
| `hooks` | ✅ | per-agent — **hanya untuk project agent** |
| `mcpServers` | ✅ | per-agent — **hanya untuk project agent** |

**Temuan pendukung penting:** plugin-shipped agent **tidak mendukung** `hooks`,
`mcpServers`, dan `permissionMode`, secara sengaja karena alasan keamanan.

Ini memvalidasi §13.5 dokumen arsitektur yang menyatakan role agent harus disimpan
sebagai **project agent**, bukan plugin. Keputusan itu terbukti benar secara teknis,
bukan sekadar preferensi.

**Resolusi tool:** tanpa kedua field → mewarisi seluruh tool parent; `tools` saja →
hanya yang terdaftar; `disallowedTools` saja → semua kecuali yang terdaftar;
keduanya → `disallowedTools` menang.

---

## 2. Hook events

Seluruh event yang dibutuhkan §42 **tersedia**.

| Kebutuhan §42 | Event | Status |
|---|---|---|
| §42.1 Path scope | `PreToolUse` | ✅ |
| §42.2 Dangerous command | `PreToolUse` | ✅ |
| §42.3 Secret paths | `PreToolUse` | ✅ |
| §42.4 Format & audit | `PostToolUse` | ✅ |
| §42.5 Handoff validation | `SubagentStop` | ✅ |
| §42.6 Worktree lifecycle | `WorktreeCreate`, `WorktreeRemove` | ✅ |

Event lain yang tersedia dan berpotensi berguna: `SessionStart`, `SessionEnd`,
`Setup`, `UserPromptSubmit`, `Stop`, `StopFailure`, `PermissionRequest`,
`PermissionDenied`, `PostToolUseFailure`, `PostToolBatch`, `SubagentStart`,
`TaskCreated`, `TaskCompleted`, `PreCompact`, `PostCompact`, `InstructionsLoaded`,
`ConfigChange`, `CwdChanged`, `FileChanged`, `Notification`.

---

## 3. T-01 — `exit 1` TIDAK memblokir apa pun

**Severity: tinggi. Mengubah cara seluruh hook ditulis.**

| Exit code | Efek |
|---|---|
| `0` | sukses — stdout diparse sebagai JSON bila ada |
| `1` atau lainnya | **non-blocking error.** Eksekusi **tetap lanjut** |
| `2` | **blocking error.** stderr menjadi alasan penolakan |

Ini berlawanan dengan konvensi Unix. Hook yang gagal karena bug, `jq` tidak
terpasang, timeout, atau OS error akan **fail-open secara diam-diam**.

Ini persis kelemahan yang diakui §6.3 untuk Mneme — dan ternyata berlaku untuk
**semua** hook, bukan hanya Mneme.

**Pengecualian:** `WorktreeCreate` — di sini **setiap** exit non-zero membatalkan
pembuatan worktree. Cocok untuk guard "jangan pernah menyalin secret ke worktree" (§42.6).

**Konsekuensi wajib:**
1. Setiap hook security memakai `exit 2`, tidak pernah `exit 1`.
2. Setiap hook memeriksa keberadaan dependensinya di awal dan `exit 2` bila hilang.
3. Setiap hook memiliki self-test di `tests/negative/`.
4. CI mengulang validasi yang sama — hook lokal tidak pernah menjadi gate tunggal.

### Cara memblokir tool call

Dua mekanisme, **pilih salah satu**:

```bash
# Opsi A — exit 2, alasan ke stderr
echo "Blocked: path di luar allowed_paths" >&2
exit 2
```

```json
// Opsi B — exit 0 + JSON ke stdout
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "Path di luar allowed_paths"
  }
}
```

`permissionDecision` menerima `allow`, `deny`, `ask`, `defer`.
**Diam tidak berarti menyetujui** — exit 0 tanpa output berarti "tanpa keputusan",
dan alur permission normal tetap berjalan.

---

## 4. T-02 / R-27 — `permissionMode` subagent dapat diabaikan parent

**Severity: kritis.**

Aturan pewarisan yang terdokumentasi:

- Bila sesi induk memakai `bypassPermissions` atau `acceptEdits`, mode itu **menang
  dan tidak dapat di-override** subagent.
- Bila sesi induk memakai `auto`, subagent mewarisi `auto` dan `permissionMode`
  di frontmatter-nya **diabaikan sepenuhnya**.

**Dampak:** `permissionMode: plan` pada `code-reviewer` dapat lenyap tanpa jejak
bila operator menjalankan sesi induk dengan bypass. Acceptance §66 #8
("Reviewer tidak mengedit code") runtuh diam-diam.

**Mitigasi yang ditetapkan:**

1. **Utama** — role agent dijalankan sebagai **sesi top-level terpisah** oleh runner
   (keputusan D-A), bukan sebagai nested subagent. Tidak ada parent yang dapat
   mewariskan mode.
2. **Pendukung** — payload hook memuat field `permission_mode`. Hook memeriksanya
   dan `exit 2` bila bernilai `bypassPermissions`.

---

## 5. T-03 — Subagent berjalan dalam satu sesi

Dokumentasi menyatakan subagent bekerja **di dalam satu sesi**, dan mengarahkan
penggunaan background agent untuk menjalankan banyak sesi independen secara paralel.

Sementara acceptance §67 menuntut **"tiga dedicated process"** dan
"tidak ada shared cwd".

**Resolusi:** runner menjalankan `claude --agent <role>` sebagai proses terpisah
per task. Memenuhi §67 secara harfiah, memberi paralelisme nyata, dan sekaligus
menutup R-27.

---

## 6. Payload hook — field yang berguna untuk enforcement

Field yang diterima hook lewat stdin:

| Field | Kegunaan bagi M2S |
|---|---|
| `cwd` | resolusi path deterministik — dasar validasi `allowed_paths` |
| `tool_name`, `tool_input` | inti validasi path scope dan dangerous command |
| `permission_mode` | **mendeteksi dan menolak sesi `bypassPermissions`** — mitigasi R-09 & R-27 |
| `agent_type` | satu hook dapat menegakkan aturan berbeda per role tanpa duplikasi script |
| `agent_id` | membedakan pemanggilan dari dalam subagent vs main thread |
| `session_id`, `transcript_path` | audit trail (§42.4) |
| `tool_use_id` | korelasi log |

`agent_type` berisi nilai frontmatter `name`, bukan nama file.

**Catatan:** `transcript_path` ditulis asinkron dan dapat tertinggal dari percakapan
langsung; untuk mengambil output akhir subagent gunakan `last_assistant_message`.

---

## 7. `SubagentStop` — memblokir penyelesaian

Dapat memblokir, dua cara:

```bash
echo "Handoff tanpa test evidence" >&2
exit 2
```

```json
{ "decision": "block", "reason": "Handoff tanpa test evidence" }
```

`"block"` adalah satu-satunya nilai yang diterima. Matcher `SubagentStop`
memfilter berdasarkan **agent type**.

Untuk umpan balik non-error yang tidak menghentikan, tersedia
`hookSpecificOutput.additionalContext`.

**Catatan:** pada frontmatter subagent, hook `Stop` otomatis dikonversi menjadi
`SubagentStop`.

---

## 8. Enforcement GitHub

| Cek | Sebelum repo public | Sesudah |
|---|---|---|
| Classic branch protection | `403 Upgrade to GitHub Pro` | `404 Branch not protected` ✅ |
| Rulesets | `403 Upgrade to GitHub Pro` | `[]` ✅ |

Perubahan dari `403` ke `404`/`[]` membuktikan fitur kini dapat diakses dan hanya
belum dikonfigurasi.

**Konteks plan:**

| Entitas | Plan | Implikasi |
|---|---|---|
| Akun `fajarcandraaa` | Free | repo private tanpa proteksi |
| Org `Mind2Screen-Dev-Team` | **Free**, 14 seat terisi, 29 repo private | seluruh repo private di org juga tanpa proteksi |
| Peran di org | `member` | upgrade plan perlu persetujuan owner |

**Batas yang harus diingat:** jalur repo public membuktikan modelnya bekerja,
tetapi **tidak** menyelesaikan kebutuhan project klien yang wajib private.
Dicatat sebagai **D-02**.

---

## 9. Lingkungan

| Komponen | Versi | Status |
|---|---|---|
| Claude Code | 2.1.212 | ✅ |
| git | 2.39.3 (Apple Git-145) | ✅ worktree didukung |
| jq | 1.7.1-apple | ✅ |
| gh | 2.96.0 | ✅ auth `fajarcandraaa`, scope `gist, read:org, repo` |
| Homebrew | 6.0.13 | ✅ |
| Ponytail | env-var di `.claude/settings.json` | ✅ terpasang — warn mode, matcher 8 role (Phase 5 §61) |
| Mneme | `.mneme/project_memory.json` | ✅ terpasang — 8 ADR, warn mode, owner TL/SA (Phase 5 §61) |

`~/.claude/` tidak memiliki `agents/` maupun `CLAUDE.md`. Tidak ada konfigurasi
global yang berpotensi bertabrakan dengan konfigurasi project.

---

## 10. Verifikasi empiris Phase 3

| # | Pertanyaan | Status |
|---|---|---|
| V-01 | Apakah hook dapat membaca file di luar cwd? | *(dibuat tidak relevan oleh Q15 — runner menyuntikkan snapshot)* |
| V-02 | Perilaku write `--add-dir` yang sebenarnya | belum |
| V-03 | Presedensi `settings.json` project vs `settings.local.json` untuk aturan `deny` | belum |
| V-04 | Apakah `permissions.deny` Bash dapat dielakkan lewat variabel shell | **✅ TERKONFIRMASI dapat dielakkan** |
| V-05 | Perilaku `WorktreeCreate` hook terhadap worktree yang dibuat runner di luar sesi | belum |

### V-04: Bash blocklist dapat dielakkan (Phase 3, §59)

**Hasil uji empiris:**

```bash
# Shell var evasion — LOLOS
echo '{"tool_input":{"command":"g=checkout; git $g main"}}' \
  | bash .claude/hooks/block-dangerous-command.sh
# exit 0 — tidak terblokir

# String quoting — LOLOS
echo '{"tool_input":{"command":"find . -delete"}}' \
  | bash .claude/hooks/block-dangerous-command.sh
# exit 0 — tidak terblokir

# Pola utuh terdeteksi — TERBLOKIR
echo '{"tool_input":{"command":"rm -rf build"}}' \
  | bash .claude/hooks/block-dangerous-command.sh
# exit 2 — terblokir
```

**Kesimpulan:** `permissions.deny` Bash(pattern) dan `block-dangerous-command.sh`
**tidak dapat menangkap elakan shell-var/quoting** (R-08). Limitation ini diterima
karena:

1. Hook adalah **defense-in-depth**, bukan boundary (R-08 §42.2 eksplisit).
2. Boundary sebenarnya: CI changed-path validation (tidak dapat dielakkan agent
   lokal, R-07) + `permissions.deny` pola path (bukan command string).
3. Pola blocklist akan tetap menangkap perintah naif mayoritas — elakan butuh
   pengetahuan khusus dan ditelusuri V-04 dedicated.

**Mitigasi berlapis:**

- CI `validate-changed-paths` (R-07) — otoritas final atas changed file
- `permissions.deny` Edit/Write pola path — terhadap `.claude/**`, `cmd/m2s/**`, dll
- Branch protection Phase 4 — mencegah merge tak sah ke develop/main

---

## Sumber

- [Create custom subagents — Claude Code Docs](https://code.claude.com/docs/en/sub-agents)
- [Hooks reference — Claude Code Docs](https://code.claude.com/docs/en/hooks)
- [Plugins reference — Claude Code Docs](https://code.claude.com/docs/en/plugins-reference)
- [Claude Code settings — Claude Code Docs](https://code.claude.com/docs/en/settings)
- [Repo referensi struktur Go — thousand-sunny](https://github.com/fajarcandraaa/thousand-sunny)

# Status ADR-001 #5 — Dua GitHub App + Ruleset Bypass

**Fase:** ADR-008 §Langkah ADR-001 #5
**Rujukan:** ADR-008 (kepemilikan org), ADR-001 (kewenangan merge), V-06..V-10
(`docs/decisions/open-questions.md`), `docs/operator/org-migration.md` §3b,
`docs/operator/branch-protection.md`.
**Kondisi akhir:** 2 Agustus 2026 (setelah PR #8 merged ke `main` `93fd578`).

Dokumen ini merangkum posisi final model dua identitas (`m2s-worker` /
`m2s-approver`) setelah migrasi org, plus sisa checklist operasional.

---

## Posisi final — terbukti lewat uji nyata

### Identitas GitHub App (terpasang di org)

| App | app_id | installation id | repo |
|---|---|---|---|
| `m2s-worker` | `4461216` | `150656551` | 3 repo (selected) |
| `m2s-approver` | `4461262` | `150657089` | 3 repo (selected) |

- Permission: `Contents: Read & write`, `Pull requests: Read & write`.
- Private key tiap App dienkripsi `age` (`~/.claude/secrets/*.pem.age`),
  passphrase berbeda per App (password manager). Plaintext `.pem` dihapus.
  `scripts/gh-app-token.sh` terima `.pem.age` + `M2S_APP_KEY_PASS`.
- Membuat App = UI-only (tak ada REST endpoint); dilakukan manusia owner org
  (`mind2screen`; `fajarcandraaa` dinaikkan jadi owner/admin org utk itu).

### Ruleset (develop/staging)

Payload kanonik: `templates/github/rulesets/*.json`,
pemasang: `tools/apply-rulesets.sh`. Terapkan di repo yg punya branch
develop/staging (`backend`, `frontend`). `platform` = control, hanya `main`,
ruleset ada tapi menarget ref yg tak ada -> siaga tak-aktif | Bypass actor utk
App datang dari `M2S_APPROVER_ID`.

| Ruleset | bypass_actors | Efek |
|---|---|---|
| `agent-push-restriction` | App `m2s-approver` `Integration:always` | approver bisa merge ke develop/staging lewat PR |
| `agent-worker-restriction` | App `m2s-approver` `always` + `OrganizationAdmin` `pull_request` | blokir push langsung workder; approver tetap boleh |

Branch protection `required_approving_review_count = 1` di `develop`+`staging`
`backend` & `frontend` (semula 0 utk hindari Perangkap 2).

### Alur akhir (terbukti end-to-end, backend PR #8)

1. `m2s-worker` buka PR (author) — check `validate-changed-paths` hijau.
2. Tanpa review -> merge **ditolak** `405 … At least 1 approving review is required`.
3. `m2s-approver` submit review `APPROVE` (`m2s-approver[bot]`).
4. Merge oleh `m2s-approver[bot]` -> `merged: true`.

Author ≠ reviewer (worker vs approver) -> bukan self-approval, prinsip #6
terjaga. Audit trail utuh (PR + review + merged_by).

### Acceptance teruji

- **§66 #9** (implementer tidak merge sendiri): worker merge ditolak —
  backend PR #5, frontend PR #4. ✅
- **V-10** (approver merge butuh `bypass_mode: always`, bukan `pull_request`);
  approver harus bypass **dua** ruleset (`agent-push-restriction` +
  `agent-worker-restriction`; semula `agent-worker-restriction` cuma
  `OrganizationAdmin` sehingga menahan approver). ✅ resolved.

### Artefak berguna

- `scripts/gh-app-token.sh` — token installation dari App (JWT pendek + POST
  access_tokens). Env `M2S_APP_ID`/`M2S_APP_KEY_PATH`/`M2S_APP_KEY_PASS`/
  `M2S_APP_INSTALL_ID`. Dukung `.pem.age`.
- `templates/github/rulesets/agent-push-restriction-approver.json` +
  `agent-worker-restriction.json` — payload kanonik.
- `tools/apply-rulesets.sh` — pasang ruleset 3 repo, isi actor approver dari
  `M2S_APPROVER_ID`.

---

## Sisa checklist

- [ ] **Repo probe `fajarcandraaa/m2s-vsh-rules-probe`** — sengaja DIBIARKAN
      (keputusan user, 2 Agustus). Masih jadi acuan uji V-06/V-07/V-08 di
      `branch-protection.md`/`open-questions.md`. Bila dihapus nanti: DELETE
      permanen (impl).
- [ ] **Clone lokal backend/frontend** masih `behind 2` — pull keputusan user
      (clone aktif). Remote sudah `Mind2Screen-Dev-Team/...`.
- [x] **Phase 5 (§61 Tool Pilot)** — **SELESAI** (PR #11, 3 Agustus 2026).
      6 skill terinstal (stop-slop, ui-ux-pro-max, emilkowalski, Ponytail, Mneme,
      DESIGN.md), 2 agent baru (frontend-engineer, technical-writer),
      `.mneme/project_memory.json` aktif (warn mode), `design/DESIGN.md` tersedia.
- [ ] Bila `platform` (control) punya branch `develop`/`staging` di waktu depan,
      ruleset `agent-*` jadi aktif otomatis.

## Catatan pemeliharaan

Posisi final 3 Agustus 2026 (Phase 5 §61 selesai via PR #11). Perbarui bila
konfigurasi ruleset/proteksi/repo berubah.
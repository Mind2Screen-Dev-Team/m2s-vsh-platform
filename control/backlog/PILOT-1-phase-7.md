# PILOT-1 — Phase 7 Multi-Repo Pilot

**ID:** PILOT-1
**Phase:** 7 (§63)
**Status:** ✅ done — 3 Agustus 2026 (9 PR merged)
**Created:** 2026-08-03

## Feature: System Status

Pilot feature pertama end-to-end pipeline multi-repo. Sebuah endpoint health sederhana
di backend + card UI di frontend.

## Task Breakdown

| Task | Type | Repository | Depends On |
|---|---|---|---|
| CONTRACT-102 | contract-change | m2s-vsh-platform | — |
| BE-102 | backend-implementation | m2s-vsh-project-backend | CONTRACT-102 |
| FE-102 | frontend-implementation | m2s-vsh-project-frontend | CONTRACT-102 |
| QA-102 | integration | m2s-vsh-platform | BE-102, FE-102 |

## Dependencies

```
CONTRACT-102 → BE-102 + FE-102 (parallel) → QA-102 → Code Review → merge
```

## Scope

- **Backend:** GET `/api/v1/status` → JSON `{status, version, uptime_seconds}`
- **Frontend:** StatusCard component consuming the endpoint
- **QA:** Acceptance + integration test verifying contract compliance

## Acceptance Criteria

- [x] 1. Backend dan Frontend berjalan paralel dengan contract yang sama
- [x] 2. Tidak ada overlap path antara BE-102 dan FE-102
- [x] 3. QA-102 acceptance test lulus (PASS 5/5)
- [x] 4. `make verify` hijau di control repo

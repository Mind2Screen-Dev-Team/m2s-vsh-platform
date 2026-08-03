# ADR D-P7-4: Lokasi `contracts/` — Control Repo

**Tanggal:** 2026-08-03
**Decider:** TL/SA
**Status:** approved

## Context

Contract bisa berada di:
- Control repo (`m2s-vsh-platform/contracts/`)
- Dedicated repo (`m2s-vsh-contracts/`)

## Decision

**Control repo cukup untuk pilot Phase 7.**

`contracts/CONTRACT-102.yaml` di `m2s-vsh-platform`.

## Rationale

1. **Phase 7 pilot** — hanya 1 contract, tidak ada inter-dependency.
2. **component-inventory.md §3** — TL/SA sudah designated owner `contracts/**`.
3. **Simplicity** — tidak ada CI tambahan untuk dedicated repo.
4. **Easy migration** — bila contract count naik >10, mudah migrasi ke dedicated
   repo later (per ADR process).

## Consequences

- Contract read oleh BE dan FE agent (via `inputs:` dalam task spec).
- `contracts/` di control repo version-controlled, visible ke semua stakeholder.
- Tidak perlu MCP cross-repo untuk contract retrieval.
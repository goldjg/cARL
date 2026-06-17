<!-- version: 1.1.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR. Update
it when scope is explicitly amended. If a requested action falls outside
approved scope, stop and escalate before proceeding.

Use this contract to distinguish active PR constraints, completed PR
constraints, durable invariants, and intentional amendments. Completed
PR constraints are historical evidence unless explicitly promoted to
durable invariants.

---

> **Starting a new PR?** Copy `current-pr-contract.template.md` into
> this file, fill in each section, and set contract status to `active`.

---

## Goal

Bootstrap cARL (`goldjg/cARL`) by importing the complete `.github/`
governance structure from `goldjg/coding-agent-baselines` and performing
a full productisation rebrand from AADLC to cARL (Cognitive Agent
Runtime Layer). Produce supporting documentation (README, VISION,
ARCHITECTURE, ROADMAP, GLOSSARY). Runtime semantics are unchanged — this
is a rename and productisation exercise, not a redesign.

## Contract status

closed

## Non-goals

- Implementing CI invariant enforcement
- Structured memory schema or YAML-based artefacts beyond what is imported
- Multi-repo governance tooling
- Non-Copilot agent support
- Pack marketplace
- Rust, Go, or C# language packs
- Any application or infrastructure code changes

## Carry-forward rules

The following constraints are promoted to durable invariants and must be
preserved in all future PRs:

- Agents must operate in Plan-only mode by default; Assisted or Automatic
  mode requires explicit user approval.
- No secrets may be committed to the repository.
- Security baseline (no hard-coded secrets, input validation, least
  privilege) applies to all future changes.
- `.github/carl/current-pr-contract.md` is the canonical active-contract
  slot; agents must read it before implementation.
- The three-layer model (operating model → instruction packs → governance
  artefacts) must be preserved in any structural changes.

## Approved scope

- `.github/copilot-instructions.md` — root operating model
- `.github/carl/` — governance artefacts directory (memory cache, PR
  contract, invariants, trust boundaries, tool policy, repo map, plans)
- `.github/instructions/core/` — all instruction pack files
- `README.md`, `VISION.md`, `ARCHITECTURE.md`, `ROADMAP.md`, `GLOSSARY.md`
- Rebrand of all `AADLCv2` references to `cARLv2`
- Path updates: `.github/aadlc/` → `.github/carl/`

## Intentional amendments

This is the founding PR; there are no prior constraints to amend. All
durable invariants listed under Carry-forward rules are established here
for the first time.

## Forbidden scope

- Changes to application or infrastructure code
- Implementing any roadmap item deferred in `ROADMAP.md`
- Adding CI workflows or automated enforcement
- Changing the runtime semantics of any instruction pack

## Architectural constraints

- The three-layer model must be accurately reflected in all documentation.
- Instruction packs must remain modular and independently loadable.
- Governance artefacts must stay under `.github/carl/`.
- No new directories or files outside the approved scope listed above.

## Security constraints

- No secrets, tokens, or credentials may appear in any committed file.
- Instruction packs must not weaken the security baseline described in
  `.github/copilot-instructions.md`.

## Files expected to change

- `.github/copilot-instructions.md`
- `.github/carl/*.md`, `.github/carl/*.yml`, `.github/carl/*.json`
- `.github/carl/plans/*.md`
- `.github/instructions/core/*.instructions.md`
- `README.md`, `VISION.md`, `ARCHITECTURE.md`, `ROADMAP.md`, `GLOSSARY.md`

## Tests / validation

- Manual review: all `AADLCv2` / `aadlc` references replaced with `cARLv2` / `carl`.
- Manual review: all `.github/carl/` paths resolve to committed files.
- Manual review: documentation accurately describes the three-layer model.
- No automated test suite exists for this documentation-only bootstrap PR.

## Stop conditions

- Any change that weakens the security baseline.
- Any change that implements a deferred roadmap item without explicit
  scope amendment.

## Escalation triggers

- Requests to modify instruction pack runtime semantics.
- Ambiguity about whether a new file falls within approved scope.

## Context reset notes

This contract is **closed**. PR #1 has been merged. Promote the
carry-forward rules above to `invariants.yml` if not already present.
For the next PR, copy `current-pr-contract.template.md` into this file
and populate all sections before implementation begins.
